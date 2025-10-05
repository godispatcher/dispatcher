package server

import (
	"errors"
	"sync"
	"time"

	"github.com/godispatcher/dispatcher/model"
)

// StreamClientPool provides a simple connection pool for the Stream API client.
// It manages up to `size` persistent StreamClient connections that can be used
// concurrently by multiple goroutines. Each underlying StreamClient remains
// single-inflight; concurrency is achieved by using different pooled clients.
//
// Typical usage:
//   pool, _ := NewStreamClientPoolFromHTTPPort("127.0.0.1", "9000", 4, 5*time.Second)
//   defer pool.Close()
//   resp, err := pool.Send(model.Document{Department:"product", Transaction:"list"})
//
// Acquire/Release are also exposed for advanced scenarios.
//
// Timeouts: DialTimeout is used for creating connections. You can adjust per-call
// read/write timeouts by setting Pool.ReadWriteTimeout which is applied to
// borrowed clients for the duration of Send; you can also adjust it manually on
// acquired clients before Release.
//
// The pool is lazy: connections are created on demand up to the pool size.
// If a connection errors during use, it is closed and replaced on the next Acquire.
// Close() closes all currently pooled connections and prevents further use.
//
// Note: Pool focuses on minimalism and avoids background goroutines.
// If you need smarter health checks or ping, extend as needed.

type StreamClientPool struct {
	host             string
	port             string
	size             int
	dialTimeout      time.Duration
	ReadWriteTimeout time.Duration

	mu    sync.Mutex
	conns chan *StreamClient
	// created tracks how many clients have been created so far; never exceeds size
	created int
	closed  bool
}

// NewStreamClientPool creates a pool for the given host:port with the specified size.
// Size must be > 0.
func NewStreamClientPool(host, port string, size int, dialTimeout time.Duration) (*StreamClientPool, error) {
	if size <= 0 {
		return nil, errors.New("pool size must be > 0")
	}
	p := &StreamClientPool{
		host:        host,
		port:        port,
		size:        size,
		dialTimeout: dialTimeout,
		// reasonable default per-call timeout
		ReadWriteTimeout: 15 * time.Second,
		conns:            make(chan *StreamClient, size),
	}
	return p, nil
}

// NewStreamClientPoolFromHTTPPort derives the stream port from an HTTP port and creates a pool.
func NewStreamClientPoolFromHTTPPort(host, httpPort string, size int, dialTimeout time.Duration) (*StreamClientPool, error) {
	return NewStreamClientPool(host, deriveStreamPortLocal(httpPort), size, dialTimeout)
}

// Acquire returns a StreamClient from the pool, creating one if necessary.
// The caller must Release the client when done.
func (p *StreamClientPool) Acquire() (*StreamClient, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pool is closed")
	}
	// Try to take an existing connection without holding the lock too long.
	select {
	case c := <-p.conns:
		p.mu.Unlock()
		return c, nil
	default:
		// none available; create if we can
		if p.created >= p.size {
			// wait for an existing one to become available
			p.mu.Unlock()
			c := <-p.conns
			return c, nil
		}
		p.created++
		host := p.host
		port := p.port
		dialTimeout := p.dialTimeout
		p.mu.Unlock()

		cli, err := NewStreamClient(host, port, dialTimeout)
		if err != nil {
			p.mu.Lock()
			p.created-- // rollback creation count on failure
			p.mu.Unlock()
			return nil, err
		}
		// inherit the pool's per-call timeout as default
		cli.ReadWriteTimeout = p.ReadWriteTimeout
		return cli, nil
	}
}

// Release returns a StreamClient back to the pool.
// If the provided error is non-nil, the connection is closed and discarded.
// If the pool is closed or already full, the connection is closed.
func (p *StreamClientPool) Release(c *StreamClient, err error) {
	if c == nil {
		return
	}
	// If an error occurred during use, drop the connection.
	if err != nil {
		_ = c.Close()
		p.mu.Lock()
		if p.created > 0 {
			p.created--
		}
		p.mu.Unlock()
		return
	}
	// Reset per-call timeout to pool default in case caller changed it.
	c.ReadWriteTimeout = p.ReadWriteTimeout

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		_ = c.Close()
		if p.created > 0 {
			p.created--
		}
		return
	}
	select {
	case p.conns <- c:
		// returned to pool
	default:
		// pool already full; close extra
		_ = c.Close()
		if p.created > 0 {
			p.created--
		}
	}
}

// Send borrows a client, performs the request and returns the response.
// It returns the client to the pool, discarding it on error.
func (p *StreamClientPool) Send(doc model.Document) (model.Document, error) {
	c, err := p.Acquire()
	if err != nil {
		return model.Document{}, err
	}
	resp, sendErr := c.Send(doc)
	p.Release(c, sendErr)
	return resp, sendErr
}

// Close closes the pool and all currently idle connections.
// Ongoing borrowed connections continue to work but will be discarded on Release.
func (p *StreamClientPool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.conns)
	// Drain and close all currently pooled connections.
	for c := range p.conns {
		_ = c.Close()
		if p.created > 0 {
			p.created--
		}
	}
	p.mu.Unlock()
	return nil
}

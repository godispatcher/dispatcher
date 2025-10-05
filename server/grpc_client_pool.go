package server

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/godispatcher/dispatcher/model"
)

// GrpcClientPool provides a simple pool of GrpcClient connections.
// Note: a single gRPC ClientConn supports concurrent RPCs; pooling is optional.
// This pool is provided to mirror existing StreamClientPool ergonomics and to
// allow isolation across endpoints if desired.
type GrpcClientPool struct {
	host           string
	port           string
	size           int
	dialTimeout    time.Duration
	RequestTimeout time.Duration

	mu      sync.Mutex
	conns   chan *GrpcClient
	created int
	closed  bool
}

// NewGrpcClientPool creates a pool for host:port with the given size (>0).
func NewGrpcClientPool(host, port string, size int, dialTimeout time.Duration) (*GrpcClientPool, error) {
	if size <= 0 {
		return nil, errors.New("pool size must be > 0")
	}
	p := &GrpcClientPool{
		host:           host,
		port:           port,
		size:           size,
		dialTimeout:    dialTimeout,
		RequestTimeout: 15 * time.Second,
		conns:          make(chan *GrpcClient, size),
	}
	return p, nil
}

// NewGrpcClientPoolFromStreamPort derives the gRPC port from a stream port (typically +1) and creates a pool.
func NewGrpcClientPoolFromStreamPort(host, streamPort string, size int, dialTimeout time.Duration) (*GrpcClientPool, error) {
	return NewGrpcClientPool(host, incrementPortLocal(streamPort), size, dialTimeout)
}

// Acquire returns a GrpcClient from the pool, creating a new one if capacity allows.
func (p *GrpcClientPool) Acquire() (*GrpcClient, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, errors.New("pool is closed")
	}
	select {
	case c := <-p.conns:
		p.mu.Unlock()
		return c, nil
	default:
		if p.created >= p.size {
			p.mu.Unlock()
			c := <-p.conns
			return c, nil
		}
		p.created++
		host := p.host
		port := p.port
		dialTimeout := p.dialTimeout
		p.mu.Unlock()

		cli, err := NewGrpcClient(host, port, dialTimeout)
		if err != nil {
			p.mu.Lock()
			p.created--
			p.mu.Unlock()
			return nil, err
		}
		return cli, nil
	}
}

// Release returns a client to the pool; if err is non-nil, the client is closed and discarded.
func (p *GrpcClientPool) Release(c *GrpcClient, err error) {
	if c == nil {
		return
	}
	if err != nil {
		_ = c.Close()
		p.mu.Lock()
		if p.created > 0 {
			p.created--
		}
		p.mu.Unlock()
		return
	}
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
		_ = c.Close()
		if p.created > 0 {
			p.created--
		}
	}
}

// Execute borrows a client, runs a unary Execute RPC using RequestTimeout, and returns it.
func (p *GrpcClientPool) Execute(doc model.Document) (model.Document, error) {
	c, err := p.Acquire()
	if err != nil {
		return model.Document{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), p.RequestTimeout)
	defer cancel()
	resp, callErr := c.Execute(ctx, doc)
	p.Release(c, callErr)
	return resp, callErr
}

// Close closes the pool and all idle connections.
func (p *GrpcClientPool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	close(p.conns)
	for c := range p.conns {
		_ = c.Close()
		if p.created > 0 {
			p.created--
		}
	}
	p.mu.Unlock()
	return nil
}

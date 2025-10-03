package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/godispatcher/dispatcher/model"
)

// StreamClient is a lightweight NDJSON (line-delimited JSON) client
// for the ServStreamApi TCP server.
//
// Usage:
//
//	c, _ := NewStreamClientFromHTTPPort("127.0.0.1", "9000")
//	defer c.Close()
//	resp, err := c.Send(model.Document{Department: "product", Transaction: "list"})
//
// The client maintains a single persistent connection and supports sequential calls.
// It is safe for concurrent use only for one-inflight request at a time; if you need
// true concurrency, create multiple clients.
//
// Timeouts: A write+read deadline is applied per Send call if ReadWriteTimeout > 0.
// DialTimeout is used when establishing the connection.
type StreamClient struct {
	conn             net.Conn
	reader           *bufio.Reader
	mu               sync.Mutex
	ReadWriteTimeout time.Duration
}

// NewStreamClient dials the given host:port and returns a connected client.
// host examples: "127.0.0.1" or "localhost". port example: "9001".
func NewStreamClient(host, port string, dialTimeout time.Duration) (*StreamClient, error) {
	if strings.TrimSpace(host) == "" {
		return nil, errors.New("host is required")
	}
	if strings.TrimSpace(port) == "" {
		return nil, errors.New("port is required")
	}
	addr := net.JoinHostPort(host, port)
	if dialTimeout <= 0 {
		dialTimeout = 5 * time.Second
	}
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("dial %s failed: %w", addr, err)
	}
	return &StreamClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
		// default per-call timeout
		ReadWriteTimeout: 15 * time.Second,
	}, nil
}

// NewStreamClientFromHTTPPort derives the stream port from an HTTP port and dials it.
// If httpPort is numeric, the stream port is httpPort+1 (e.g., 9000 -> 9001).
// Otherwise a "-stream" suffix is appended.
func NewStreamClientFromHTTPPort(host, httpPort string, dialTimeout time.Duration) (*StreamClient, error) {
	return NewStreamClient(host, deriveStreamPortLocal(httpPort), dialTimeout)
}

// deriveStreamPortLocal mirrors server.deriveStreamPort without exporting it.
func deriveStreamPortLocal(httpPort string) string {
	if httpPort == "" {
		return "9001"
	}
	if n, err := strconv.Atoi(httpPort); err == nil {
		return strconv.Itoa(n + 1)
	}
	return strings.TrimSpace(httpPort) + "-stream"
}

// Send writes a single line JSON document and reads a single line JSON response.
// If the response's Type is "Error" and Error is set, an error is returned alongside the document.
func (c *StreamClient) Send(doc model.Document) (model.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return model.Document{}, errors.New("client is closed")
	}

	// Apply a per-call deadline if configured
	if c.ReadWriteTimeout > 0 {
		_ = c.conn.SetDeadline(time.Now().Add(c.ReadWriteTimeout))
	} else {
		// clear deadline
		_ = c.conn.SetDeadline(time.Time{})
	}

	// Marshal and write followed by a newline
	b, err := json.Marshal(doc)
	if err != nil {
		return model.Document{}, err
	}
	if _, err := c.conn.Write(append(b, '\n')); err != nil {
		return model.Document{}, err
	}

	// Read one line response
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return model.Document{}, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return model.Document{}, errors.New("empty response")
	}
	var out model.Document
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		return model.Document{}, err
	}
	if strings.EqualFold(out.Type, "Error") && out.Error != nil {
		return out, fmt.Errorf("remote error: %v", out.Error)
	}
	return out, nil
}

// Close closes the underlying TCP connection.
func (c *StreamClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

// Call is a convenience for a one-shot request using a derived stream port from an HTTP port.
// It dials, sends the document, reads the response, and closes the connection.
func Call(host, httpPort string, doc model.Document) (model.Document, error) {
	cli, err := NewStreamClientFromHTTPPort(host, httpPort, 5*time.Second)
	if err != nil {
		return model.Document{}, err
	}
	defer cli.Close()
	return cli.Send(doc)
}

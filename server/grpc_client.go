package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/godispatcher/dispatcher/model"
)

// Uses the JSON codec defined in server/grpc.go to avoid .proto generation.

// GrpcClient is a lightweight client for the ServGrpcApi defined in server/grpc.go.
// It handles unary Execute and bidirectional Stream calls with the JSON codec enforced.
type GrpcClient struct {
	conn *grpc.ClientConn
}

// NewGrpcClient dials a gRPC server on host:port with the JSON codec enforced.
// dialTimeout controls how long to wait for the initial connection.
func NewGrpcClient(host, port string, dialTimeout time.Duration) (*GrpcClient, error) {
	if host == "" {
		return nil, errors.New("host is required")
	}
	if port == "" {
		port = "9002" // default grpc port used by server if not specified
	}
	addr := net.JoinHostPort(host, port)

	// Ensure our JSON codec is registered and force it for all calls on this connection.
	encoding.RegisterCodec(jsonCodec{})

	if dialTimeout <= 0 {
		dialTimeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s failed: %w", addr, err)
	}
	return &GrpcClient{conn: conn}, nil
}

// NewGrpcClientFromStreamPort derives the gRPC port from the given stream port (typically +1) and dials it.
func NewGrpcClientFromStreamPort(host, streamPort string, dialTimeout time.Duration) (*GrpcClient, error) {
	return NewGrpcClient(host, incrementPortLocal(streamPort), dialTimeout)
}

// Execute performs a unary call to /dispatcher.Dispatcher/Execute with the provided document.
func (c *GrpcClient) Execute(ctx context.Context, doc model.Document) (model.Document, error) {
	if c == nil || c.conn == nil {
		return model.Document{}, errors.New("client is closed")
	}
	var out model.Document
	err := c.conn.Invoke(ctx, "/dispatcher.Dispatcher/Execute", &doc, &out)
	return out, err
}

// GrpcBidiStream wraps a gRPC ClientStream to send/receive model.Document messages.
type GrpcBidiStream struct {
	cs grpc.ClientStream
}

// Send sends a single Document message on the stream.
func (s *GrpcBidiStream) Send(doc model.Document) error { return s.cs.SendMsg(&doc) }

// Recv receives a single Document message from the stream.
func (s *GrpcBidiStream) Recv() (model.Document, error) {
	var out model.Document
	err := s.cs.RecvMsg(&out)
	return out, err
}

// CloseSend closes the client-sending direction of the stream.
func (s *GrpcBidiStream) CloseSend() error { return s.cs.CloseSend() }

// Stream opens a bidirectional stream to /dispatcher.Dispatcher/Stream.
// The caller should CloseSend on the returned stream when done sending.
func (c *GrpcClient) Stream(ctx context.Context) (*GrpcBidiStream, error) {
	if c == nil || c.conn == nil {
		return nil, errors.New("client is closed")
	}
	desc := &grpc.StreamDesc{ServerStreams: true, ClientStreams: true}
	cs, err := c.conn.NewStream(ctx, desc, "/dispatcher.Dispatcher/Stream")
	if err != nil {
		return nil, err
	}
	return &GrpcBidiStream{cs: cs}, nil
}

// Close closes the underlying gRPC connection.
func (c *GrpcClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// CallUnary is a convenience helper that creates a client, performs a unary Execute, and closes it.
func CallUnary(host, port string, doc model.Document, timeout time.Duration) (model.Document, error) {
	cli, err := NewGrpcClient(host, port, timeout)
	if err != nil {
		return model.Document{}, err
	}
	defer cli.Close()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return cli.Execute(ctx, doc)
}

// incrementPortLocal mirrors the server's naive increment logic without importing unexported funcs.
func incrementPortLocal(port string) string {
	var n int
	if _, err := fmt.Sscanf(port, "%d", &n); err == nil {
		return fmt.Sprintf("%d", n+1)
	}
	return port
}

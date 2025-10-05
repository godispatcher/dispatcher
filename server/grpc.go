package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/godispatcher/dispatcher/department"
	"github.com/godispatcher/dispatcher/model"
)

// jsonCodec allows us to use JSON payloads over gRPC so we can reuse model.Document
// without introducing .proto code-gen for minimal changes.
type jsonCodec struct{}

func (jsonCodec) Name() string                               { return "json" }
func (jsonCodec) Marshal(v interface{}) ([]byte, error)      { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v interface{}) error { return json.Unmarshal(data, v) }

// Register and start a gRPC server that mirrors the Stream API behavior.
func ServGrpcApi(register *department.RegisterDispatcher) {
	// Ensure our JSON codec is available and forced so clients don't need to specify it explicitly
	encoding.RegisterCodec(jsonCodec{})

	port := deriveGrpcPort(register.GRPCPort, register.StreamPort)
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("grpc api listen error on port %s: %v", port, err)
		return
	}

	// Force using JSON codec on server side for compatibility with non-proto messages
	grpcServer := grpc.NewServer(grpc.ForceServerCodec(jsonCodec{}))

	// Register service using a manual ServiceDesc
	service := &dispatcherGrpcService{}
	grpcServer.RegisterService(&grpc.ServiceDesc{
		ServiceName: "dispatcher.Dispatcher",
		HandlerType: (*dispatcherGrpcService)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "Execute",
				Handler:    service.executeUnaryHandler,
			},
		},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "Stream",
				Handler:       service.streamBidiHandler,
				ServerStreams: true,
				ClientStreams: true,
			},
		},
	}, service)

	log.Printf("grpc api listening on :%s (json codec)\n", port)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("grpc serve error: %v", err)
		}
	}()
}

func deriveGrpcPort(grpcPort, streamPort string) string {
	if grpcPort != "" {
		return grpcPort
	}
	// If stream port provided, pick next port; else default to 9002
	if streamPort != "" {
		return incrementPort(streamPort)
	}
	return "9002"
}

func incrementPort(port string) string {
	// naive: try to parse last 1-2 digits; if fail, just return given
	var n int
	if _, err := fmt.Sscanf(port, "%d", &n); err == nil {
		return fmt.Sprintf("%d", n+1)
	}
	return port
}

type dispatcherGrpcService struct{}

// Unary Execute handler
func (s *dispatcherGrpcService) executeUnaryHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	req := new(model.Document)
	if err := dec(req); err != nil {
		return nil, err
	}
	handler := func(ctx context.Context, reqAny interface{}) (interface{}, error) {
		resp := executeDocument(*req)
		return &resp, nil
	}
	if interceptor == nil {
		return handler(ctx, req)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dispatcher.Dispatcher/Execute",
	}
	return interceptor(ctx, req, info, handler)
}

// Bidirectional stream handler
func (s *dispatcherGrpcService) streamBidiHandler(srv interface{}, stream grpc.ServerStream) error {
	for {
		req := new(model.Document)
		if err := stream.RecvMsg(req); err != nil {
			return err
		}
		resp := executeDocument(*req)
		if err := stream.SendMsg(&resp); err != nil {
			return err
		}
	}
}

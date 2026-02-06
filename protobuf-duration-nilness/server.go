package main

import (
	"context"
	"fmt"
	"net"

	"buf.build/go/protovalidate"
	taskv1 "github.com/fredrikaverpil/go-playground/protobuf-duration-nilness/gen/task/v1"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"google.golang.org/grpc"
)

// echoServer echoes back whatever Task it receives.
type echoServer struct {
	taskv1.UnimplementedTaskServiceServer
}

func (s *echoServer) CreateTask(
	_ context.Context,
	req *taskv1.CreateTaskRequest,
) (*taskv1.CreateTaskResponse, error) {
	return &taskv1.CreateTaskResponse{Task: req.GetTask()}, nil
}

// listenAndServe starts the gRPC server on the given address and blocks until it's stopped.
func listenAndServe(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	validator, err := protovalidate.New()
	if err != nil {
		return fmt.Errorf("create validator: %w", err)
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(protovalidate_middleware.UnaryServerInterceptor(validator)),
	)
	taskv1.RegisterTaskServiceServer(srv, &echoServer{})
	return srv.Serve(lis)
}

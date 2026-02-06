package main

import (
	"context"
	"fmt"
	"net"

	taskv1 "github.com/fredrikaverpil/go-playground/protobuf-duration-nilness/gen/task/v1"
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
	srv := grpc.NewServer()
	taskv1.RegisterTaskServiceServer(srv, &echoServer{})
	return srv.Serve(lis)
}

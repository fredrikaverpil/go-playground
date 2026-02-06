package main

import (
	"context"
	"net"
	"testing"
	"time"

	taskv1 "github.com/fredrikaverpil/go-playground/protobuf-duration-nilness/gen/task/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

func newTestClient(t *testing.T) taskv1.TaskServiceClient {
	t.Helper()

	// Pick a random free port for the test server.
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := lis.Addr().String()
	_ = lis.Close()

	go func() {
		if err := listenAndServe(addr); err != nil {
			t.Logf("serve: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return taskv1.NewTaskServiceClient(conn)
}

func TestDurationNilVsZero_GRPC(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	t.Run("unset duration stays nil over gRPC", func(t *testing.T) {
		resp, err := client.CreateTask(ctx, &taskv1.CreateTaskRequest{
			Task: &taskv1.Task{Name: "unset"},
		})
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		if resp.GetTask().GetMaxDuration() != nil {
			t.Fatal("expected nil duration after gRPC round-trip")
		}
	})

	t.Run("zero duration stays non-nil over gRPC", func(t *testing.T) {
		resp, err := client.CreateTask(ctx, &taskv1.CreateTaskRequest{
			Task: &taskv1.Task{
				Name:        "zero",
				MaxDuration: durationpb.New(0),
			},
		})
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		d := resp.GetTask().GetMaxDuration()
		if d == nil {
			t.Fatal("expected non-nil duration after gRPC round-trip")
		}
		if d.GetSeconds() != 0 || d.GetNanos() != 0 {
			t.Fatalf("expected zero duration, got seconds=%d nanos=%d", d.GetSeconds(), d.GetNanos())
		}
	})

	t.Run("5 minute duration preserved over gRPC", func(t *testing.T) {
		resp, err := client.CreateTask(ctx, &taskv1.CreateTaskRequest{
			Task: &taskv1.Task{
				Name:        "five-min",
				MaxDuration: durationpb.New(5 * time.Minute),
			},
		})
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		d := resp.GetTask().GetMaxDuration()
		if d == nil {
			t.Fatal("expected non-nil duration after gRPC round-trip")
		}
		if d.GetSeconds() != 300 {
			t.Fatalf("expected 300 seconds, got %d", d.GetSeconds())
		}
	})
}

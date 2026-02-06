package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/grpc/codes"
)

const databaseURI = "projects/my-project/instances/my-instance/databases/my-db"

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	logFile, err := os.OpenFile("spanner.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log file: %v\n", err)
		return 1
	}
	defer func() { _ = logFile.Close() }()

	if err := os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010"); err != nil {
		fmt.Fprintf(os.Stderr, "set env: %v\n", err)
		return 1
	}

	go startContainer(logFile)
	defer stopContainer()

	if err := waitForPorts([]string{"9010", "9020"}, 30*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "wait for emulator: %v\n", err)
		return 1
	}

	ctx := context.Background()
	if err := createInstanceAndDatabase(ctx, databaseURI); err != nil {
		fmt.Fprintf(os.Stderr, "create instance and database: %v\n", err)
		return 1
	}

	return m.Run()
}

func startContainer(logFile *os.File) {
	// Remove any leftover container from a previous run.
	_ = exec.Command("docker", "rm", "-f", "spanner-emulator").Run()
	cmd := exec.Command("docker",
		"run", "--rm",
		"-p", "9020:9020",
		"-p", "9010:9010",
		"--name", "spanner-emulator",
		"gcr.io/cloud-spanner-emulator/emulator",
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	fmt.Println("Starting Spanner emulator...")
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 2 {
			// Container was already stopped or removed; safe to ignore.
			return
		}
		fmt.Fprintf(os.Stderr, "docker run: %v\n", err)
	}
}

func stopContainer() {
	fmt.Println("Stopping Spanner emulator...")
	if err := exec.Command("docker", "stop", "spanner-emulator").Run(); err != nil {
		fmt.Fprintf(os.Stderr, "docker stop: %v\n", err)
	}
}

// waitForPorts polls until all ports are reachable or the timeout expires.
func waitForPorts(ports []string, timeout time.Duration) error {
	deadline := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return fmt.Errorf("ports %v not reachable after %v", ports, timeout)
		case <-ticker.C:
			if allOpen(ports) {
				return nil
			}
		}
	}
}

func allOpen(ports []string) bool {
	for _, port := range ports {
		conn, err := net.Dial("tcp", "localhost:"+port)
		if err != nil {
			return false
		}
		_ = conn.Close()
	}
	return true
}

func createInstanceAndDatabase(ctx context.Context, uri string) error {
	if err := createInstance(ctx, uri); err != nil {
		return fmt.Errorf("create instance: %w", err)
	}
	if err := recreateDatabase(ctx, uri); err != nil {
		return fmt.Errorf("recreate database: %w", err)
	}
	return nil
}

func createInstance(ctx context.Context, uri string) error {
	matches := regexp.MustCompile("projects/(.*)/instances/(.*)/databases/.*").FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("invalid database URI %q", uri)
	}
	project, instanceID := matches[1], matches[2]
	instanceName := "projects/" + project + "/instances/" + instanceID

	adminClient, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("create instance admin client: %w", err)
	}
	defer func() { _ = adminClient.Close() }()

	_, err = adminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{Name: instanceName})
	if err == nil {
		return nil // Already exists.
	}
	if spanner.ErrCode(err) != codes.NotFound {
		return fmt.Errorf("get instance: %w", err)
	}

	_, err = adminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     "projects/" + project,
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("create instance: %w", err)
	}
	return nil
}

// recreateDatabase drops the existing database (if any) and creates a fresh one.
func recreateDatabase(ctx context.Context, uri string) error {
	matches := regexp.MustCompile("^(.*)/databases/(.*)$").FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("invalid database URI %q", uri)
	}
	parent, dbName := matches[1], matches[2]

	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("create database admin client: %w", err)
	}
	defer func() { _ = adminClient.Close() }()

	// Drop if it already exists.
	_, err = adminClient.GetDatabase(ctx, &databasepb.GetDatabaseRequest{Name: uri})
	if err != nil && spanner.ErrCode(err) != codes.NotFound {
		return fmt.Errorf("get database: %w", err)
	}
	if err == nil {
		if err := adminClient.DropDatabase(ctx, &databasepb.DropDatabaseRequest{Database: uri}); err != nil {
			return fmt.Errorf("drop database: %w", err)
		}
	}

	op, err := adminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          parent,
		CreateStatement: "CREATE DATABASE `" + dbName + "`",
	})
	if err != nil {
		return fmt.Errorf("create database: %w", err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for database creation: %w", err)
	}
	return nil
}

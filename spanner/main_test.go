package main

import (
	"context"
	"fmt"
	"log"
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

const (
	db = "projects/my-project/instances/my-instance/databases/my-db"
)

func TestMain(m *testing.M) {
	// Open the log file.
	logFile, err := os.OpenFile("spanner.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	// Set the SPANNER_EMULATOR_HOST environment variable.
	os.Setenv("SPANNER_EMULATOR_HOST", "localhost:9010")

	// Start the emulator.
	defer stopContainer()
	go startContainer(logFile)

	// Wait until emulator is ready.
	checkPorts([]string{"9020", "9010"})

	// Create the instance and database.
	ctx := context.Background()
	if os.Getenv("SPANNER_EMULATOR_HOST") != "" {
		err := createInstanceAndDatabase(ctx, db, true)
		if err != nil {
			log.Panicf("createDatabaseAndInstance: %v", err)
		}
	} else {
		log.Panicf("SPANNER_EMULATOR_HOST is not set")
	}

	// Run the test.
	exitCode := m.Run()

	// Manually run defer commands in case of test failure.
	// TODO: why is not defer being run automatically?
	// ...is it because of the os.Exit?
	if exitCode != 0 {
		stopContainer()
		logFile.Close()

		// Exit with non-exit status code.
		os.Exit(exitCode)
	}
}

func startContainer(logFile *os.File) {
	dockerArgs := []string{
		"run",
		"--rm",
		"-p",
		"9020:9020",
		"-p",
		"9010:9010",
		"--name",
		"spanner-emulator",
		// "-d",
		"gcr.io/cloud-spanner-emulator/emulator",
	}
	cmd := exec.Command("docker", dockerArgs...)
	fmt.Println("Starting Spanner emulator...")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err := cmd.Run()
	if err != nil {
		if err.Error() == "exit status 2" {
			return
		} else {
			log.Printf("'docker run' exited with %s\n", err)
		}
	}
}

func stopContainer() {
	fmt.Println("Stopping Spanner emulator...")
	stopCmd := exec.Command("docker", "stop", "spanner-emulator")
	err := stopCmd.Run()
	if err != nil {
		log.Printf("'docker stop' failed with %s\n", err)
	}
}

func checkPorts(ports []string) {
	timeout := time.After(30 * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-timeout:
			fmt.Println("Checking ports timed out")
			os.Exit(1)
		case <-tick:
			allOpen := true
			for _, port := range ports {
				conn, err := net.Dial("tcp", "localhost:"+port)
				if err != nil {
					allOpen = false
					break
				}
				conn.Close()
			}
			if allOpen {
				return
			}
		}
	}
}

func createInstanceAndDatabase(ctx context.Context, uri string, drop bool) error {
	if err := createInstance(ctx, uri); err != nil {
		return err
	}
	return createDatabase(ctx, uri, drop)
}

func createInstance(ctx context.Context, uri string) error {
	matches := regexp.MustCompile("projects/(.*)/instances/(.*)/databases/.*").FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("invalid instance id %s", uri)
	}
	instanceName := "projects/" + matches[1] + "/instances/" + matches[2]

	instanceAdminClient, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return err
	}
	defer instanceAdminClient.Close()

	_, err = instanceAdminClient.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: instanceName,
	})
	if err != nil && spanner.ErrCode(err) != codes.NotFound {
		return err
	}
	if err == nil {
		// instance already exists
		return nil
	}
	_, err = instanceAdminClient.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     "projects/" + matches[1],
		InstanceId: matches[2],
	})
	if err != nil {
		return err
	}
	return nil
}

func createDatabase(ctx context.Context, uri string, drop bool) error {
	matches := regexp.MustCompile("^(.*)/databases/(.*)$").FindStringSubmatch(uri)
	if matches == nil || len(matches) != 3 {
		return fmt.Errorf("invalid database id %s", uri)
	}

	databaseAdminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	_, err = databaseAdminClient.GetDatabase(ctx, &databasepb.GetDatabaseRequest{Name: uri})
	if err != nil && spanner.ErrCode(err) != codes.NotFound {
		return err
	}
	if err == nil {
		// Database already exists
		if drop {
			if err = databaseAdminClient.DropDatabase(ctx, &databasepb.DropDatabaseRequest{Database: uri}); err != nil {
				return err
			}
		} else {
			return nil
		}
	}

	op, err := databaseAdminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          matches[1],
		CreateStatement: "CREATE DATABASE `" + matches[2] + "`",
		ExtraStatements: []string{},
	})
	if err != nil {
		return err
	}
	if _, err = op.Wait(ctx); err != nil {
		return err
	}
	return nil
}

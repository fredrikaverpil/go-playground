package main

import (
	"context"
	"embed"
	"strings"
	"testing"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	"google.golang.org/api/option"
)

//go:embed schema/*.sql
var schemaFS embed.FS

//go:embed seed/*.sql
var seedFS embed.FS

// applySchema reads DDL files from the embedded schema/ FS and applies them via UpdateDatabaseDdl.
func applySchema(t *testing.T, ctx context.Context, files ...string) {
	t.Helper()
	adminClient, err := database.NewDatabaseAdminClient(ctx, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("create admin client: %v", err)
	}
	t.Cleanup(func() { adminClient.Close() })

	var statements []string
	for _, file := range files {
		data, err := schemaFS.ReadFile("schema/" + file)
		if err != nil {
			t.Fatalf("read schema file %s: %v", file, err)
		}
		statements = append(statements, splitStatements(string(data))...)
	}

	op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   databaseURI,
		Statements: statements,
	})
	if err != nil {
		t.Fatalf("update DDL: %v", err)
	}
	if err := op.Wait(ctx); err != nil {
		t.Fatalf("wait for DDL: %v", err)
	}
}

// applySeed reads DML files from the embedded seed/ FS and applies them via ReadWriteTransaction.
func applySeed(t *testing.T, ctx context.Context, client *spanner.Client, files ...string) {
	t.Helper()
	var statements []spanner.Statement
	for _, file := range files {
		data, err := seedFS.ReadFile("seed/" + file)
		if err != nil {
			t.Fatalf("read seed file %s: %v", file, err)
		}
		for _, s := range splitStatements(string(data)) {
			statements = append(statements, spanner.NewStatement(s))
		}
	}

	_, err := client.ReadWriteTransaction(ctx, func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
		_, err := txn.BatchUpdate(ctx, statements)
		return err
	})
	if err != nil {
		t.Fatalf("apply seed data: %v", err)
	}
}

// newClient creates a spanner.Client and registers cleanup via t.Cleanup.
func newClient(t *testing.T, ctx context.Context) *spanner.Client {
	t.Helper()
	client, err := spanner.NewClient(ctx, databaseURI, option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	return client
}

// splitStatements splits SQL text on semicolons and discards empty entries.
func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	var result []string
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

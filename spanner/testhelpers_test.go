package main

import (
	"context"
	"database/sql"
	"embed"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	_ "github.com/googleapis/go-sql-spanner"
	"google.golang.org/api/option"
)

//go:embed schema/*.sql
var schemaFS embed.FS

//go:embed seed/*.sql
var seedFS embed.FS

// applySchema reads DDL files from the embedded schema/ FS and applies them via UpdateDatabaseDdl.
func applySchema(tb testing.TB, ctx context.Context, files ...string) {
	tb.Helper()
	adminClient, err := database.NewDatabaseAdminClient(ctx, option.WithoutAuthentication())
	if err != nil {
		tb.Fatalf("create admin client: %v", err)
	}
	tb.Cleanup(func() { _ = adminClient.Close() })

	var statements []string
	for _, file := range files {
		data, err := schemaFS.ReadFile("schema/" + file)
		if err != nil {
			tb.Fatalf("read schema file %s: %v", file, err)
		}
		statements = append(statements, splitStatements(string(data))...)
	}

	op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
		Database:   databaseURI,
		Statements: statements,
	})
	if err != nil {
		tb.Fatalf("update DDL: %v", err)
	}
	if err := op.Wait(ctx); err != nil {
		tb.Fatalf("wait for DDL: %v", err)
	}
}

// applySeed reads DML files from the embedded seed/ FS and applies them via ReadWriteTransaction.
func applySeed(tb testing.TB, ctx context.Context, client *spanner.Client, files ...string) {
	tb.Helper()
	var statements []spanner.Statement
	for _, file := range files {
		data, err := seedFS.ReadFile("seed/" + file)
		if err != nil {
			tb.Fatalf("read seed file %s: %v", file, err)
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
		tb.Fatalf("apply seed data: %v", err)
	}
}

// newClient creates a spanner.Client and registers cleanup via tb.Cleanup.
func newClient(tb testing.TB, ctx context.Context) *spanner.Client {
	tb.Helper()
	client, err := spanner.NewClient(ctx, databaseURI, option.WithoutAuthentication())
	if err != nil {
		tb.Fatalf("create client: %v", err)
	}
	tb.Cleanup(func() { client.Close() })
	return client
}

// newDB opens a database/sql connection to the Spanner emulator and registers cleanup.
func newDB(tb testing.TB, _ context.Context) *sql.DB {
	tb.Helper()
	host := os.Getenv("SPANNER_EMULATOR_HOST")
	dsn := host + "/" + databaseURI + ";usePlainText=true"
	db, err := sql.Open("spanner", dsn)
	if err != nil {
		tb.Fatalf("open database/sql: %v", err)
	}
	tb.Cleanup(func() { _ = db.Close() })
	return db
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

package main

// Variant 3: Error-returning options.
//
// The option function returns an error, allowing validation at construction
// time rather than at use time. The constructor short-circuits on the first
// error.
//
//	type DatabaseOption func(*database) error
//
// This is useful when invalid configuration should be caught early, for example
// checking port ranges, mutually exclusive flags, or required fields.

import (
	"errors"
	"fmt"
	"testing"
)

// database is a sample struct configured via error-returning options.
type database struct {
	host     string
	port     int
	database string
	poolSize int
}

// databaseOption configures a [database] and may return an error.
type databaseOption func(*database) error

// withHost sets the database host.
func withHost(host string) databaseOption {
	return func(db *database) error {
		if host == "" {
			return errors.New("host must not be empty")
		}
		db.host = host
		return nil
	}
}

// withPort sets the database port.
func withPort(port int) databaseOption {
	return func(db *database) error {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port %d out of range [1, 65535]", port)
		}
		db.port = port
		return nil
	}
}

// withDatabase sets the database name.
func withDatabase(name string) databaseOption {
	return func(db *database) error {
		if name == "" {
			return errors.New("database name must not be empty")
		}
		db.database = name
		return nil
	}
}

// withPoolSize sets the connection pool size.
func withPoolSize(size int) databaseOption {
	return func(db *database) error {
		if size < 1 {
			return fmt.Errorf("pool size must be positive, got %d", size)
		}
		db.poolSize = size
		return nil
	}
}

// newDatabase creates a database connection with defaults, then applies
// options. It returns an error if any option fails validation.
func newDatabase(opts ...databaseOption) (*database, error) {
	db := &database{
		host:     "localhost",
		port:     5432,
		database: "postgres",
		poolSize: 10,
	}
	for _, opt := range opts {
		if err := opt(db); err != nil {
			return nil, fmt.Errorf("apply database option: %w", err)
		}
	}
	return db, nil
}

func TestErrorOption_defaults(t *testing.T) {
	db, err := newDatabase()
	if err != nil {
		t.Fatalf("newDatabase() error: %v", err)
	}
	if db.host != "localhost" {
		t.Errorf("host = %q, want %q", db.host, "localhost")
	}
	if db.port != 5432 {
		t.Errorf("port = %d, want %d", db.port, 5432)
	}
}

func TestErrorOption_overrides(t *testing.T) {
	db, err := newDatabase(
		withHost("db.example.com"),
		withPort(3306),
		withDatabase("myapp"),
		withPoolSize(25),
	)
	if err != nil {
		t.Fatalf("newDatabase() error: %v", err)
	}
	if db.host != "db.example.com" {
		t.Errorf("host = %q, want %q", db.host, "db.example.com")
	}
	if db.port != 3306 {
		t.Errorf("port = %d, want %d", db.port, 3306)
	}
	if db.database != "myapp" {
		t.Errorf("database = %q, want %q", db.database, "myapp")
	}
	if db.poolSize != 25 {
		t.Errorf("poolSize = %d, want %d", db.poolSize, 25)
	}
}

func TestErrorOption_invalidPort(t *testing.T) {
	_, err := newDatabase(withPort(0))
	if err == nil {
		t.Fatal("expected error for port 0, got nil")
	}
}

func TestErrorOption_invalidHost(t *testing.T) {
	_, err := newDatabase(withHost(""))
	if err == nil {
		t.Fatal("expected error for empty host, got nil")
	}
}

func TestErrorOption_negativePoolSize(t *testing.T) {
	_, err := newDatabase(withPoolSize(-1))
	if err == nil {
		t.Fatal("expected error for negative pool size, got nil")
	}
}

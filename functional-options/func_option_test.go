package main

// Variant 1: Simple function type.
//
// This is the most common variant, popularized by Dave Cheney. The option type
// is a plain function that mutates the target struct. It is simple, concise,
// and sufficient for the vast majority of use cases.
//
//	type Option func(*Server)
//
// See server.go for the full implementation.

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestFuncOption_defaults(t *testing.T) {
	s := NewServer()
	if s.addr != ":8080" {
		t.Errorf("addr = %q, want %q", s.addr, ":8080")
	}
	if s.readTimeout != 5*time.Second {
		t.Errorf("readTimeout = %v, want %v", s.readTimeout, 5*time.Second)
	}
	if s.writeTimeout != 10*time.Second {
		t.Errorf("writeTimeout = %v, want %v", s.writeTimeout, 10*time.Second)
	}
	if s.maxConns != 100 {
		t.Errorf("maxConns = %d, want %d", s.maxConns, 100)
	}
}

func TestFuncOption_overrides(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	s := NewServer(
		WithAddr(":9090"),
		WithReadTimeout(30*time.Second),
		WithWriteTimeout(60*time.Second),
		WithMaxConns(500),
		WithLogger(logger),
	)
	if s.addr != ":9090" {
		t.Errorf("addr = %q, want %q", s.addr, ":9090")
	}
	if s.readTimeout != 30*time.Second {
		t.Errorf("readTimeout = %v, want %v", s.readTimeout, 30*time.Second)
	}
	if s.writeTimeout != 60*time.Second {
		t.Errorf("writeTimeout = %v, want %v", s.writeTimeout, 60*time.Second)
	}
	if s.maxConns != 500 {
		t.Errorf("maxConns = %d, want %d", s.maxConns, 500)
	}
	if s.logger != logger {
		t.Error("logger not set correctly")
	}
}

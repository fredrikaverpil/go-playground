// Package main demonstrates the functional options pattern in Go.
//
// The pattern was popularized by Dave Cheney and Rob Pike. It provides a clean,
// extensible API for configuring structs without requiring breaking changes when
// new options are added.
//
// Key benefits:
//   - Sensible defaults via the zero value or explicit defaults.
//   - Callers only specify what they want to override.
//   - Adding new options is backwards-compatible.
//   - Self-documenting: each option is a named function.
//
// See: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
package main

import (
	"log/slog"
	"time"
)

// Server is an example HTTP server with configurable options.
type Server struct {
	addr         string
	readTimeout  time.Duration
	writeTimeout time.Duration
	maxConns     int
	logger       *slog.Logger
}

// Option configures a [Server].
type Option func(*Server)

// WithAddr sets the listen address. Defaults to ":8080".
func WithAddr(addr string) Option {
	return func(s *Server) {
		s.addr = addr
	}
}

// WithReadTimeout sets the read timeout. Defaults to 5s.
func WithReadTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = d
	}
}

// WithWriteTimeout sets the write timeout. Defaults to 10s.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = d
	}
}

// WithMaxConns sets the maximum number of concurrent connections. Defaults to 100.
func WithMaxConns(n int) Option {
	return func(s *Server) {
		s.maxConns = n
	}
}

// WithLogger sets a custom logger. Defaults to [slog.Default].
func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.logger = l
	}
}

// NewServer creates a [Server] with the given options applied on top of defaults.
func NewServer(opts ...Option) *Server {
	s := &Server{
		addr:         ":8080",
		readTimeout:  5 * time.Second,
		writeTimeout: 10 * time.Second,
		maxConns:     100,
		logger:       slog.Default(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

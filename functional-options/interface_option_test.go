package main

// Variant 2: Interface-based options.
//
// Used by gRPC (grpc.DialOption) and many Google Cloud client libraries. The
// option is an interface with an unexported apply method. This prevents callers
// from implementing the interface outside the package, giving the library full
// control over valid options.
//
//	type ClientOption interface {
//	    apply(*client)
//	}
//
// A private struct implements the interface for each option, or a function
// adapter type can be used for convenience (shown below).

import (
	"testing"
	"time"
)

// client is a sample struct configured via interface-based options.
type client struct {
	endpoint string
	timeout  time.Duration
	retries  int
}

// clientOption is the interface that all options must implement.
type clientOption interface {
	apply(*client)
}

// clientOptionFunc is a convenience adapter so plain functions can be used as
// options without defining a new struct for each one.
type clientOptionFunc func(*client)

func (f clientOptionFunc) apply(c *client) { f(c) }

// withEndpoint sets the client endpoint.
func withEndpoint(endpoint string) clientOption {
	return clientOptionFunc(func(c *client) {
		c.endpoint = endpoint
	})
}

// withTimeout sets the client timeout.
func withTimeout(d time.Duration) clientOption {
	return clientOptionFunc(func(c *client) {
		c.timeout = d
	})
}

// retryOption is an example of a struct-based option, useful when the option
// carries multiple related values or needs its own validation logic.
type retryOption struct {
	retries int
}

func (o retryOption) apply(c *client) {
	c.retries = o.retries
}

// withRetries sets the number of retries.
func withRetries(n int) clientOption {
	return retryOption{retries: n}
}

// newClient creates a client with defaults, then applies options.
func newClient(opts ...clientOption) *client {
	c := &client{
		endpoint: "https://api.example.com",
		timeout:  10 * time.Second,
		retries:  3,
	}
	for _, opt := range opts {
		opt.apply(c)
	}
	return c
}

func TestInterfaceOption_defaults(t *testing.T) {
	c := newClient()
	if c.endpoint != "https://api.example.com" {
		t.Errorf("endpoint = %q, want %q", c.endpoint, "https://api.example.com")
	}
	if c.timeout != 10*time.Second {
		t.Errorf("timeout = %v, want %v", c.timeout, 10*time.Second)
	}
	if c.retries != 3 {
		t.Errorf("retries = %d, want %d", c.retries, 3)
	}
}

func TestInterfaceOption_overrides(t *testing.T) {
	c := newClient(
		withEndpoint("https://staging.example.com"),
		withTimeout(30*time.Second),
		withRetries(5),
	)
	if c.endpoint != "https://staging.example.com" {
		t.Errorf("endpoint = %q, want %q", c.endpoint, "https://staging.example.com")
	}
	if c.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", c.timeout, 30*time.Second)
	}
	if c.retries != 5 {
		t.Errorf("retries = %d, want %d", c.retries, 5)
	}
}

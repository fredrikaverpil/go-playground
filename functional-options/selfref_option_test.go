package main

// Variant 4: Self-referential options (Rob Pike's original).
//
// Each option function returns the previous value of the option it set. This
// allows callers to save and restore previous configuration, which is useful
// for temporary overrides in tests or scoped configuration changes.
//
//	type CacheOption func(*cache) CacheOption
//
// See: https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html

import (
	"testing"
	"time"
)

// cache is a sample struct configured via self-referential options.
type cache struct {
	maxItems int
	ttl      time.Duration
	evict    bool
}

// cacheOption configures a [cache] and returns the previous option value.
type cacheOption func(*cache) cacheOption

// withMaxItems sets the maximum number of cached items.
func withMaxItems(n int) cacheOption {
	return func(c *cache) cacheOption {
		prev := c.maxItems
		c.maxItems = n
		return withMaxItems(prev)
	}
}

// withTTL sets the time-to-live for cache entries.
func withTTL(d time.Duration) cacheOption {
	return func(c *cache) cacheOption {
		prev := c.ttl
		c.ttl = d
		return withTTL(prev)
	}
}

// withEviction enables or disables automatic eviction.
func withEviction(enabled bool) cacheOption {
	return func(c *cache) cacheOption {
		prev := c.evict
		c.evict = enabled
		return withEviction(prev)
	}
}

// newCache creates a cache with defaults, then applies options.
func newCache(opts ...cacheOption) *cache {
	c := &cache{
		maxItems: 1000,
		ttl:      5 * time.Minute,
		evict:    true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func TestSelfRefOption_defaults(t *testing.T) {
	c := newCache()
	if c.maxItems != 1000 {
		t.Errorf("maxItems = %d, want %d", c.maxItems, 1000)
	}
	if c.ttl != 5*time.Minute {
		t.Errorf("ttl = %v, want %v", c.ttl, 5*time.Minute)
	}
	if !c.evict {
		t.Error("evict = false, want true")
	}
}

func TestSelfRefOption_overrides(t *testing.T) {
	c := newCache(
		withMaxItems(500),
		withTTL(1*time.Minute),
		withEviction(false),
	)
	if c.maxItems != 500 {
		t.Errorf("maxItems = %d, want %d", c.maxItems, 500)
	}
	if c.ttl != 1*time.Minute {
		t.Errorf("ttl = %v, want %v", c.ttl, 1*time.Minute)
	}
	if c.evict {
		t.Error("evict = true, want false")
	}
}

func TestSelfRefOption_saveAndRestore(t *testing.T) {
	c := newCache()

	// Apply an option and capture the undo function.
	undo := withMaxItems(42)(c)
	if c.maxItems != 42 {
		t.Fatalf("maxItems after set = %d, want %d", c.maxItems, 42)
	}

	// Restore the previous value.
	undo(c)
	if c.maxItems != 1000 {
		t.Errorf("maxItems after restore = %d, want %d", c.maxItems, 1000)
	}
}

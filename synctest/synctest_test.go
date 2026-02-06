// Package synctest_test demonstrates the testing/synctest package, which lets
// you test concurrent code inside a "time bubble" with a fake clock and
// deterministic goroutine synchronization.
//
// See https://go.dev/blog/synctest and https://go.dev/blog/testing-time
package synctest_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"testing/synctest"
	"time"
)

// TestFakeClock shows that time inside a bubble starts at 2000-01-01 00:00:00
// UTC and advances instantly when all goroutines are blocked on time.Sleep.
func TestFakeClock(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()

		// This does not actually wait 10 seconds of real time.
		time.Sleep(10 * time.Second)

		elapsed := time.Since(start)
		if elapsed != 10*time.Second {
			t.Fatalf("elapsed = %v, want 10s", elapsed)
		}

		// Verify that the bubble clock started at 2000-01-01.
		year := start.Year()
		if year != 2000 {
			t.Fatalf("start year = %d, want 2000", year)
		}
	})
}

// TestWaitSynchronizesGoroutines shows that synctest.Wait blocks until all
// goroutines in the bubble are durably blocked, giving you a synchronization
// point for assertions.
func TestWaitSynchronizesGoroutines(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		done := false
		go func() {
			done = true
		}()

		// Wait for the goroutine above to finish.
		synctest.Wait()

		if !done {
			t.Fatal("expected goroutine to have completed")
		}
	})
}

// TestContextDeadline demonstrates testing context.WithDeadline deterministically.
// Without synctest, this kind of test is either slow (sleeping real time) or
// flaky (racing against the scheduler).
func TestContextDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const timeout = 5 * time.Second
		ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(timeout))
		defer cancel()

		// Sleep to just before the deadline.
		time.Sleep(timeout - time.Nanosecond)
		synctest.Wait()

		if err := ctx.Err(); err != nil {
			t.Fatalf("before deadline: ctx.Err() = %v, want nil", err)
		}

		// Sleep past the deadline.
		time.Sleep(time.Nanosecond)
		synctest.Wait()

		if err := ctx.Err(); err != context.DeadlineExceeded {
			t.Fatalf("after deadline: ctx.Err() = %v, want DeadlineExceeded", err)
		}
	})
}

// TestContextAfterFunc demonstrates testing context.AfterFunc, verifying that
// the function is called only after cancellation.
func TestContextAfterFunc(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())

		called := false
		context.AfterFunc(ctx, func() {
			called = true
		})

		synctest.Wait()
		if called {
			t.Fatal("AfterFunc called before cancel")
		}

		cancel()
		synctest.Wait()

		if !called {
			t.Fatal("AfterFunc not called after cancel")
		}
	})
}

// TestPipeCopy demonstrates testing concurrent I/O with io.Pipe. Data written
// to the pipe writer is available on the reader side after synctest.Wait.
func TestPipeCopy(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		reader, writer := io.Pipe()
		defer func() { _ = writer.Close() }()

		var buf bytes.Buffer
		go func() { _, _ = io.Copy(&buf, reader) }()

		want := "hello, synctest"
		if _, err := writer.Write([]byte(want)); err != nil {
			t.Fatal(err)
		}
		synctest.Wait()

		if got := buf.String(); got != want {
			t.Fatalf("copied %q, want %q", got, want)
		}
	})
}

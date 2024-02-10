package main

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

// --- used for test 1

type SpySleeper struct {
	Calls int
}

func (s *SpySleeper) Sleep() {
	s.Calls++
}

// --- used for test 2

const (
	write = "write"
	sleep = "sleep"
)

type SpyCountdownOperations struct {
	Calls []string
}

func (s *SpyCountdownOperations) Sleep() {
	s.Calls = append(s.Calls, sleep)
}

// Implements io.Writer interface - so it can be used in test
func (s *SpyCountdownOperations) Write(p []byte) (n int, err error) {
	s.Calls = append(s.Calls, write)
	return
}

// -- used for test 3

type SpyTime struct {
	durationSlept time.Duration
}

func (s *SpyTime) Sleep(duration time.Duration) {
	s.durationSlept = duration
}

// -- the tests

func TestCountdown(t *testing.T) {
	t.Run("prints 3 to Go!", func(t *testing.T) {
		// use spySleeper
		buffer := &bytes.Buffer{}
		Countdown(buffer, &SpySleeper{})

		got := buffer.String()
		want := `3
2
1
Go!`

		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("sleep before every print", func(t *testing.T) {
		spySleepPrinter := &SpyCountdownOperations{}
		Countdown(spySleepPrinter, spySleepPrinter)

		want := []string{
			write,
			sleep,
			write,
			sleep,
			write,
			sleep,
			write,
		}

		if !reflect.DeepEqual(want, spySleepPrinter.Calls) {
			t.Errorf("wanted calls %v got %v", want, spySleepPrinter.Calls)
		}
	})

	t.Run("sleep with configurable sleeper", func(t *testing.T) {
		sleepTime := 5 * time.Second

		spyTime := &SpyTime{}
		sleeper := ConfigurableSleeper{sleepTime, spyTime.Sleep}
		sleeper.Sleep()

		if spyTime.durationSlept != sleepTime {
			t.Errorf("should have slept for %v but slept for %v", sleepTime, spyTime.durationSlept)
		}
	})
}

package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Sleeper interface {
	Sleep()
}

type DefaultSleeper struct{}

func (d *DefaultSleeper) Sleep() {
	time.Sleep(1 * time.Second)
}

type ConfigurableSleeper struct {
	duration time.Duration
	sleep    func(time.Duration)
}

func (c *ConfigurableSleeper) Sleep() {
	c.sleep(c.duration)
}

func Countdown(out io.Writer, sleeper Sleeper) error {
	for i := 3; i > 0; i-- {
		if _, err := fmt.Fprintln(out, i); err != nil {
			return err
		}
		sleeper.Sleep()
	}
	_, err := fmt.Fprint(out, "Go!")
	return err
}

func main() {
	fmt.Println("Counting down with sleeper...")
	sleeper := &DefaultSleeper{}
	if err := Countdown(os.Stdout, sleeper); err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println("Counting down with configurable sleeper...")
	configurableSleeper := &ConfigurableSleeper{1 * time.Second, time.Sleep}
	if err := Countdown(os.Stdout, configurableSleeper); err != nil {
		fmt.Println("error:", err)
	}
}

package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

func main() {
	// Use all defaults.
	s1 := NewServer()
	fmt.Println("=== Default server ===")
	printServer(s1)

	// Override specific options.
	s2 := NewServer(
		WithAddr(":9090"),
		WithReadTimeout(30*time.Second),
		WithMaxConns(500),
		WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
	)
	fmt.Println("\n=== Custom server ===")
	printServer(s2)
}

func printServer(s *Server) {
	fmt.Printf("  addr:          %s\n", s.addr)
	fmt.Printf("  readTimeout:   %s\n", s.readTimeout)
	fmt.Printf("  writeTimeout:  %s\n", s.writeTimeout)
	fmt.Printf("  maxConns:      %d\n", s.maxConns)
	fmt.Printf("  logger:        %v\n", s.logger)
}

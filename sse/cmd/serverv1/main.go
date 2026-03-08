package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/events", sseHandler)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}

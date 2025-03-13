package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/events", sseHandler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("error starting server: %s\n", err)
	}
}

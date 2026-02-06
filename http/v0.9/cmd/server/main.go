package main

import (
	"log"
	"net/http"
)

func main() {
	addr := "127.0.0.1:9000"
	s := Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := w.Write([]byte("Hello World!")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
		}),
	}
	log.Printf("Listening on %s", addr)
	if err := s.ServeAndListen(); err != nil {
		log.Fatal(err)
	}
}

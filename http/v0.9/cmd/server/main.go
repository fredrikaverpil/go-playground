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
			w.Write([]byte("Hello World!"))
		}),
	}
	log.Printf("Listening on %s", addr)
	if err := s.ServeAndListen(); err != nil {
		log.Fatal(err)
	}
}

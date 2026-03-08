package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func Greet(writer io.Writer, name string) error {
	_, err := fmt.Fprintf(writer, "Hello, %s", name)
	return err
}

func MyGreeterHandler(w http.ResponseWriter, _ *http.Request) {
	if err := Greet(w, "world"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	if err := Greet(os.Stdout, "Elodie"); err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:              ":5001",
		Handler:           http.HandlerFunc(MyGreeterHandler),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

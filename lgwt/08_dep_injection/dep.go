package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func Greet(writer io.Writer, name string) error {
	_, err := fmt.Fprintf(writer, "Hello, %s", name)
	return err
}

func MyGreeterHandler(w http.ResponseWriter, r *http.Request) {
	if err := Greet(w, "world"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	if err := Greet(os.Stdout, "Elodie"); err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":5001", http.HandlerFunc(MyGreeterHandler)))
}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

// Different kind of readers passed into the same
// readerToStdout function
func main() {
	err := filePrinter()
	if err != nil {
		log.Fatal(err)
	}

	err = stringPrinter()
	if err != nil {
		log.Fatal(err)
	}

	err = connectionPrinter()
	if err != nil {
		log.Fatal(err)
	}
}

func filePrinter() error {
	file, err := os.Open("fileops/text.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = file.Close() }()

	readerToStdout(file, 64)

	return nil
}

func stringPrinter() error {
	s := strings.NewReader("very short but interesting string")

	readerToStdout(s, 2)

	return nil
}

func connectionPrinter() error {
	conn, err := net.Dial("tcp", "google.com:80")
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// fire off a request
	fmt.Fprint(conn, "GET / HTTP/1.0\r\n\r\n")

	readerToStdout(conn, 64)

	return nil
}

func readerToStdout(r io.Reader, bufSize int) error {
	buf := make([]byte, bufSize)

	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		} else if n > 0 {
			fmt.Println(string(buf[:n]))
		}
	}
}

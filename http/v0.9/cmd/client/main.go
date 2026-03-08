package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	var d net.Dialer
	conn, err := d.DialContext(context.Background(), "tcp", "127.0.0.1:9000")
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	if _, err := conn.Write([]byte("GET /this/is/a/test\r\n")); err != nil {
		log.Fatalf("err: %s", err)
	}

	body, err := io.ReadAll(conn)
	if err != nil {
		log.Fatalf("err: %s", err)
	}

	fmt.Println(string(body))
}

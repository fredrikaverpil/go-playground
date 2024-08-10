package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine() // GET /path/to/resource
	if err != nil {
		return
	}

	fields := strings.Fields(string(line))
	if len(fields) < 2 {
		return
	}

	r := &http.Request{
		Method: fields[0],
		URL: &url.URL{
			Scheme: "http",
			Path:   fields[1],
		},
		Proto:      "HTTP/0.9",
		ProtoMajor: 0,
		ProtoMinor: 9,
		RemoteAddr: conn.RemoteAddr().String(),
	}

	log.Printf("Method: %s, URL: %s, Proto: %s", r.Method, r.URL, r.Proto)

	s.Handler.ServeHTTP(newWriter(conn), r)
}

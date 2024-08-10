package main

import (
	"net"
	"net/http"
)

type responseBodyWriter struct {
	conn net.Conn
}

func newWriter(c net.Conn) http.ResponseWriter {
	return &responseBodyWriter{conn: c}
}

func (r *responseBodyWriter) Header() http.Header {
	// unsupported with HTTP/0.9
	return nil
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
	return r.conn.Write(b)
}

func (r *responseBodyWriter) WriteHeader(statusCode int) {
	// unsupported with HTTP/0.9
}

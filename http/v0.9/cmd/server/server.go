package main

import (
	"context"
	"log"
	"net"
	"net/http"
)

type Server struct {
	Addr    string
	Handler http.Handler
}

func (s *Server) ServeAndListen() error {
	if s.Handler == nil {
		panic("http server started without a handler")
	}
	var lc net.ListenConfig
	l, err := lc.Listen(context.Background(), "tcp", s.Addr)
	if err != nil {
		return err
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Printf("Failed to close listener: %v", err)
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			return err
		}

		go s.handleConnection(conn)
	}
}

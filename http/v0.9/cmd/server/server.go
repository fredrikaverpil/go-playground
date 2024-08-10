package main

import (
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
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.handleConnection(conn)
	}
}

package server

import (
	"log"
	"net"
	"net/http"
	"time"
)

type HTTPServer struct {
	Server *http.Server
}

func NewHTTP(handler http.Handler, addr string) *HTTPServer {
	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return &HTTPServer{Server: s}
}

func (s *HTTPServer) Listen() error {
	log.Printf("http listening on %s", s.Server.Addr)
	ln, err := net.Listen("tcp", s.Server.Addr)
	if err != nil { return err }
	return s.Server.Serve(ln)
}

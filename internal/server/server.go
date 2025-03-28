package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	Addr     string
	Handler  func(w response.ResponseWriter, r *request.Request)
	Listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error listening on port %d: %v", port, err)
	}

	srv := &Server{Addr: addr, Listener: lst}
	go srv.listen()

	return srv, nil
}
func (s *Server) Close() error {
	if !s.closed.CompareAndSwap(false, true) {
		return fmt.Errorf("server already closed")
	}

	err := s.Listener.Close()
	if err != nil {
		return fmt.Errorf("error closing server listener: %v", err)
	}

	log.Println("Server has been closed gracefully")
	return nil
}

func (s *Server) listen() {
	for {
		if s.closed.Load() {
			log.Println("Server is shutting down, stopping listener")
			return
		}

		conn, err := s.Listener.Accept()
		if err != nil {
			if s.closed.Load() {
				log.Println("Listener closed, exiting")
			}
			log.Printf("Error accepting connection: %v", err)
		}
		go s.handle(conn)
	}

}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
			return
		}
	}()

	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\r\n"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

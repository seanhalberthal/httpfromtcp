package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	Addr     string
	Handler  Handler
	Listener net.Listener
	closed   atomic.Bool
}

// Serve creates HTTP Listener on a given port
func (s *Server) Serve(port int) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	lst, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error listening on port %d: %v", port, err)
	}

	s.Addr = addr
	s.Listener = lst
	go s.listen()
	return s, nil
}

// Close closes the server
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

// listen is the Listener that is called by Serve
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
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}

}

// handle is responsible for formatting a successful connection
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		WriteError(conn, HandlerError{StatusCode: 400, Message: err.Error()})
		return
	}

	buffer := &bytes.Buffer{}

	handlerErr := s.Handler(buffer, req)
	if handlerErr != nil {
		WriteError(conn, *handlerErr)
		return
	}

	headers := response.GetDefaultHeaders(buffer.Len())

	err = response.WriteStatusLine(conn, response.StatusOK)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("Error writing headers: %v", err)
		return
	}

	_, err = conn.Write(buffer.Bytes())
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

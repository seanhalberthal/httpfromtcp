package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"log"
	"net"
)

func main() {
	// Create a TCP listener on port 42069
	TCPListener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Printf("error creating TCP listener: %v\n", err)
		return
	}

	// Ensure the TCP Listener is closed when the main function exits
	defer func() {
		err := TCPListener.Close()
		if err != nil {
			log.Printf("error closing TCP listener: %v\n", err)
			return
		}
	}()

	// Accept incoming connections in a loop
	for {
		conn, err := TCPListener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("new connection from %s\n", conn.RemoteAddr())

		// Create a request that reads data from the open connection
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("error parsing request: %v\n", err)
			continue
		}

		// Format the request line
		reqLine := &request.RequestLine{
			HttpVersion:   req.RequestLine.HttpVersion,
			RequestTarget: req.RequestLine.RequestTarget,
			Method:        req.RequestLine.Method,
		}

		// Print the request line properties to the terminal
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", reqLine.Method)
		fmt.Printf("- Target: %s\n", reqLine.RequestTarget)
		fmt.Printf("- Version: %s\n", reqLine.HttpVersion)
	}
}

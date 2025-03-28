package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func main() {
	srv := &server.Server{
		Handler: testHandler,
	}

	_, err := srv.Serve(port)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func(srv *server.Server) {
		err := srv.Close()
		if err != nil {
			log.Fatalf("Error closing server: %v", err)
		}
	}(srv)
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func testHandler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		return &server.HandlerError{
			StatusCode: 400,
			Message:    "Your problem is not my problem\n",
		}
	case "/myproblem":
		return &server.HandlerError{
			StatusCode: 500,
			Message:    "Woopsie, my bad\n",
		}
	default:
		_, err := io.WriteString(w, "All good, frfr\n")
		if err != nil {
			return &server.HandlerError{
				StatusCode: 500,
				Message:    "Error writing response\n",
			}
		}
		return nil
	}
}

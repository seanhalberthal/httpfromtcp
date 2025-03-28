package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"io"
	"log"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteError(w io.Writer, h HandlerError) {
	handlerError := fmt.Sprintf("Status: %d, %s", h.StatusCode, h.Message)
	_, err := w.Write([]byte(handlerError))
	if err != nil {
		log.Println(err)
	}
}

package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
)

type HandlerError struct {
	StatusCode int
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteError(w io.Writer, h HandlerError) {
	var statusCode response.StatusCode

	switch h.StatusCode {
	case 200:
		statusCode = response.StatusOK
	case 400:
		statusCode = response.StatusBadRequest
	case 500:
		statusCode = response.StatusInternalError
	default:
		statusCode = response.StatusInternalError
	}

	err := response.WriteStatusLine(w, statusCode)
	if err != nil {
		log.Printf("Error writing status line: %v", err)
		return
	}

	headers := response.GetDefaultHeaders(len(h.Message))
	err = response.WriteHeaders(w, headers)
	if err != nil {
		log.Printf("Error getting default headers: %v", err)
		return
	}

	_, err = w.Write([]byte(h.Message))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
}

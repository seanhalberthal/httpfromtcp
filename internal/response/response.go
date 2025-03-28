package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"net"
	"strconv"
)

type ResponseWriter struct {
	conn       net.Conn // Holds connection to write to
	headers    headers.Headers
	statusCode int
	body       []byte
}

type StatusCode int

const (
	StatusOK            StatusCode = 200
	StatusBadRequest    StatusCode = 400
	StatusInternalError StatusCode = 500
)

// SetHeader sets a key-value header pair
func (rw *ResponseWriter) SetHeader(key, value string) {
	rw.headers[key] = value
}

// WriteHeader sets the HTTP status code
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

// Write writes the response body
func (rw *ResponseWriter) Write(body []byte) {
	rw.body = body
}

// SendResponse formats the response and sends it over the connection
func (rw *ResponseWriter) SendResponse() error {
	response := fmt.Sprintf("HTTP/1.1 %d OK\r\n", rw.statusCode)
	for k, v := range rw.headers {
		response += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	response += "\r\n" + string(rw.body)

	_, err := rw.conn.Write([]byte(response))
	return err
}

// WriteStatusLine handles writing the HTTP status of an incoming request
func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var statusLine string

	switch statusCode {
	case StatusOK:
		statusLine = "HTTP/1.1 200 OK\r\n"
	case StatusBadRequest:
		statusLine = "HTTP/1.1 400 Bad Request\r\n"
	case StatusInternalError:
		statusLine = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		statusLine = fmt.Sprintf("HTTP/1.1 %d \r\n", int(statusCode))
	}

	_, err := w.Write([]byte(statusLine))
	return err
}

// GetDefaultHeaders sets default headers based on a given content-length
func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	// Set the required headers
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		// Format: "Key: Value\r\n"
		headerLine := fmt.Sprintf("%s: %s\r\n", k, v)

		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return err
		}

	}

	// After all headers, write the final CRLF that separates headers from body
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}

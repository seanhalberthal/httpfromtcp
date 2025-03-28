package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"net"
)

type ResponseWriter struct {
	conn       net.Conn // Holds connection to write to
	headers    headers.Headers
	statusCode int
	body       []byte
}

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

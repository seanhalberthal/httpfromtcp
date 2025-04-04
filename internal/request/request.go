package request

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Define possible states
const (
	requestStateParsingLine = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

// Request defines data structure for an incoming request
type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       int
	Body        []byte
}

// RequestLine defines data structure for the start-line (RFC 9110)
type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

// RequestFromReader creates a new Request from a reader input
func RequestFromReader(r io.Reader) (*Request, error) {
	// Initialize request
	req := &Request{
		Headers: headers.NewHeaders(),
		state:   requestStateParsingLine,
	}

	// Create buffer with 1024 bytes
	buf := make([]byte, 1024)

	// Create buffer for leftover data
	var leftover []byte

	// Start infinite loop
	for {
		// Read data in chunks from the reader
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}

		// Check that chunk is not empty
		if n == 0 {
			break
		}

		// Combine leftover data with newly read data
		data := append(leftover, buf[:n]...)

		// Pass each chunk to parse method
		bytesProcessed, err := req.parse(data)
		if err != nil {
			return nil, err
		}

		// Store any remaining data which hasn't yet been processed
		leftover = data[bytesProcessed:]

		// If parsing is complete, break out of the loop
		if req.state == requestStateDone {
			break
		}
	}

	// If not in the done state, request must not have been complete
	if req.state != requestStateDone {
		return nil, fmt.Errorf("incomplete request")
	}

	if req.state == requestStateParsingBody {
		contentLengthStr := string(req.Headers.Get("Content-Length"))
		if contentLengthStr != "" {
			contentLength, err := strconv.Atoi(contentLengthStr)
			if err != nil {
				return nil, err
			}
			if len(req.Body) < contentLength {
				return nil, fmt.Errorf("body too short")
			}
		}
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0

	// Check if state is not done, then call parseSingle on data
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}

		if n == 0 {
			break // No more parsing can be done with the current data
		}

		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateParsingLine:
		// Convert data to string and check if it contains "\r\n"
		stringData := string(data)
		endOfLine := strings.Index(stringData, "\r\n")

		// If there's no line break then more data is needed
		if endOfLine == -1 {
			return 0, nil
		}

		// Grab the request line (without the "\r\n") and split on whitespace
		line := stringData[:endOfLine]
		parts := strings.Split(line, " ")

		// Check if request line is formatted correctly
		if len(parts) != 3 {
			return 0, fmt.Errorf("invalid request line: %s", stringData)
		}

		// Grab relevant parts of request line
		method := parts[0]
		requestTarget := parts[1]
		httpVersion := parts[2]

		// Check that method is formatted correctly
		for _, char := range method {
			if !unicode.IsUpper(char) || !unicode.IsLetter(char) {
				return 0, fmt.Errorf("invalid method: %s - must only contain uppercase letters", method)
			}
		}

		// Check HTTP version is 1.1
		if !strings.HasPrefix(httpVersion, "HTTP/") {
			return 0, fmt.Errorf("invalid http version: %s - must be HTTP/1.1", httpVersion)
		}

		versionNumber := strings.TrimPrefix(httpVersion, "HTTP/")
		if versionNumber != "1.1" {
			return 0, fmt.Errorf("invalid http version: %s - must be HTTP/1.1", httpVersion)
		}

		// Set the request line parts to the RequestLine object
		r.RequestLine.Method = method
		r.RequestLine.RequestTarget = requestTarget
		r.RequestLine.HttpVersion = versionNumber

		// Set state to requestStateParsingHeaders
		r.state = requestStateParsingHeaders
		return endOfLine + 2, nil // + 2 for "\r\n"

	// Parse the headers using Headers.Parse
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}

		// Set state to requestStateParsingBody
		if done {
			r.state = requestStateParsingBody
		}

		return n, nil

	case requestStateParsingBody:
		// Get Content-Length header
		contentLengthStr := r.Headers.Get("Content-Length")

		// Check if Content-Length is empty
		if contentLengthStr == "" {
			r.state = requestStateDone
			return 0, nil
		}

		// Parse Content-Length to int
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("invalid content length: %s", contentLengthStr)
		}

		// If Content-Length == 0 -> no body to parse
		if contentLength == 0 {
			r.state = requestStateDone
			return 0, nil
		}

		// Determine how many bytes we can process from current data
		bytesToRead := len(data)
		if bytesToRead > contentLength-len(r.Body) {
			bytesToRead = contentLength - len(r.Body)
		}

		// Append to the body
		r.Body = append(r.Body, data[:bytesToRead]...)

		// Check if we've read the entire body
		if len(r.Body) == contentLength {
			r.state = requestStateDone
		}

		return bytesToRead, nil

	default:
		break
	}
	return 0, nil
}

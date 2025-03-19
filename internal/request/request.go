package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(r io.Reader) (*Request, error) {
	// Read data from the reader
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Convert data into a string and split by line using HTTP newline character "\r\n"
	stringData := string(data)
	lines := strings.Split(stringData, "\r\n")
	firstLine := lines[0]

	// Split the first line into method, request target and HTTP version
	parts := strings.Split(firstLine, " ")

	// Check that all 3 parts exist
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", firstLine)
	}

	// Extract method, request target, and HTTP version from parts
	method := parts[0]
	requestTarget := parts[1]
	httpVersion := parts[2]

	// Check method is all uppercase letters
	for _, char := range method {
		if !unicode.IsUpper(char) || !unicode.IsLetter(char) {
			return nil, fmt.Errorf("invalid method: %s - must only contain uppercase letters", method)
		}
	}

	// Check HTTP version is 1.1
	if !strings.HasPrefix(httpVersion, "HTTP/") {
		return nil, fmt.Errorf("invalid http version: %s - must be HTTP/1.1", httpVersion)
	}

	versionNumber := strings.TrimPrefix(httpVersion, "HTTP/")

	if versionNumber != "1.1" {
		return nil, fmt.Errorf("invalid http version: %s - must be HTTP/1.1", httpVersion)
	}

	// Instantiate and return request
	return &Request{
		RequestLine: RequestLine{
			Method:        method,
			RequestTarget: requestTarget,
			HttpVersion:   versionNumber,
		},
	}, nil
}

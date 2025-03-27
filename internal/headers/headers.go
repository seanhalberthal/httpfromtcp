package headers

import (
	"fmt"
	"regexp"
	"strings"
)

// Headers defines the headers map type with a key-value pair of strings
type Headers map[string]string

// Valid character specification from RFC 9110
var tcharRegex = regexp.MustCompile(`^[!#$%&'*+\-.^_` + "`" + `|~0-9a-zA-Z]+$`)

// Parse parses incoming data stream and maps it to a header map
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	// Convert data to string and check if it only contains \r\n (end of headers)
	stringData := string(data)
	if strings.HasPrefix(stringData, "\r\n") {
		return 2, true, nil // Consumed 2 bytes and is done
	}

	bytesProcessed := 0

	// Initialize infinite loop
	for {
		// Check if we have enough data to process, if not then exit
		if bytesProcessed >= len(stringData) {
			return bytesProcessed, false, nil
		}

		// Check if we've reached end of headers (empty line)
		if len(stringData[bytesProcessed:]) >= 2 && stringData[bytesProcessed:bytesProcessed+2] == "\r\n" {
			return bytesProcessed, true, nil // End of headers
		}

		// Find the end of the current header line in the remaining data
		remainingData := stringData[bytesProcessed:]
		endOfLine := strings.Index(remainingData, "\r\n")
		// No complete header line found, needs more data
		if endOfLine == -1 {
			return bytesProcessed, false, nil
		}

		// Grab the request line
		requestLine := remainingData[:endOfLine]

		// Find the colon (which separates key from value)
		colon := strings.IndexByte(requestLine, ':')
		if colon == -1 {
			// No colon in the line == invalid header format
			return 0, false, fmt.Errorf("invalid header format: %s - must include colon", requestLine)
		}

		// Extract the key and value
		key := requestLine[:colon]
		value := requestLine[colon+1:]

		// Check for whitespace between colon and key
		if strings.TrimSpace(key) != key {
			return 0, false, fmt.Errorf("invalid header format: %s - spaces between colon and key", requestLine)
		}

		// Check if key has invalid characters using validateKey
		if err = validateKey(key); err != nil {
			return 0, false, err
		}

		// Trim any leading whitespace from the key and value
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Convert key to lowercase before adding it to the map
		key = strings.ToLower(key)

		// Check if header already exists in map
		if existingValue, exists := h[key]; exists {
			// If key exists, append new value with comma separator
			h[key] = existingValue + ", " + value
		} else {
			// If key doesn't exist, simply set the value
			h[key] = value
		}

		// Update bytesProcessed to move past this header line (including \r\n)
		bytesProcessed += endOfLine + 2
		return endOfLine + 2, false, nil
	}
}

// NewHeaders Creates a new Headers map
func NewHeaders() Headers {
	return make(map[string]string)
}

// validateKey checks if the key uses only valid tchar characters
func validateKey(key string) error {
	if !tcharRegex.MatchString(key) {
		return fmt.Errorf("invalid characters in header key %s", key)
	}
	return nil
}

// Get returns the value a header by its key, case-insensitive
func (h Headers) Get(key string) string {
	lowerKey := strings.ToLower(key)
	for k, v := range h {
		if strings.ToLower(k) == lowerKey {
			return v
		}
	}
	return ""
}

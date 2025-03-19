package headers

import (
	"fmt"
	"regexp"
	"strings"
)

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

	// Find the end of the current header line
	endOfLine := strings.Index(stringData, "\r\n")
	if endOfLine == -1 {
		return 0, false, nil
	}

	// Grab the header
	headerLine := stringData[:endOfLine]

	// Find the colon (which separates key from value)
	colon := strings.IndexByte(headerLine, ':')
	if colon == -1 {
		// No colon in the line == invalid header format
		return 0, false, fmt.Errorf("invalid header format: %s - must include colon", headerLine)
	}

	// Extract the key and value
	key := headerLine[:colon]
	value := headerLine[colon+1:]

	// Check for whitespace between colon and key
	if strings.TrimSpace(key) != key {
		return 0, false, fmt.Errorf("invalid header format: %s - spaces between colon and key", headerLine)
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

	// Return the number of bytes consumed (the line itself + "\r\n")
	return endOfLine + 2, false, nil
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

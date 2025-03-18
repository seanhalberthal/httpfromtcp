package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		// Create a channel that provides lines of text from the connection
		linesChannel := getLinesChannel(conn)
		for line := range linesChannel {
			fmt.Printf("%s\n", line)
		}
	}
}

// Reads data from the connection and sends complete lines to a channel
func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		// Ensures the channel gets closed when goroutine exits
		defer close(lines)
		// Ensures the connection gets closed when goroutine exits
		defer func(f net.Conn) {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(conn)

		// Create 8 byte buffer to read data from the connection and variable to accumulate parts
		// for the current line
		bytes := make([]byte, 8)
		currentLine := ""

		for {
			// Reads data from the connection
			n, err := conn.Read(bytes)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("error reading from connection: %v\n", err)
				break
			}

			// Convert the bytes read into a string and split them by newline
			stringBytes := string(bytes[:n])
			parts := strings.Split(stringBytes, "\n")

			// Store each complete part as a line in the channel
			for i := 0; i < len(parts)-1; i++ {
				lines <- currentLine + parts[i]
				currentLine = ""
			}

			// Store the last part (which may not be a full line)
			currentLine += parts[len(parts)-1]

		}

		// Send any remaining data to the channel as the last line
		if currentLine != "" {
			lines <- currentLine
		}
	}()

	return lines
}

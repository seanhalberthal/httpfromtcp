package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	TCPListener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Printf("error creating TCP listener: %v\n", err)
		return
	}

	defer func() {
		err := TCPListener.Close()
		if err != nil {
			log.Printf("error closing TCP listener: %v\n", err)
			return
		}
	}()

	for {
		conn, err := TCPListener.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v\n", err)
			continue
		}
		fmt.Printf("new connection from %s\n", conn.RemoteAddr())

		linesChannel := getLinesChannel(conn)
		for line := range linesChannel {
			fmt.Printf("%s\n", line)
		}
	}
}

func getLinesChannel(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer func(f net.Conn) {
			err := f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(conn)

		bytes := make([]byte, 8)
		currentLine := ""

		for {
			n, err := conn.Read(bytes)
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("error reading from connection: %v\n", err)
				break
			}

			stringBytes := string(bytes[:n])
			parts := strings.Split(stringBytes, "\n")

			for i := 0; i < len(parts)-1; i++ {
				lines <- currentLine + parts[i]
				currentLine = ""
			}

			currentLine += parts[len(parts)-1]

		}
		if currentLine != "" {
			lines <- currentLine
		}
	}()

	return lines
}

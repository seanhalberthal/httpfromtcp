package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// Resolve UDP Address, hardcoded to localhost:42069
	resolve, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Printf("error resolving UDP address: %v\n", err)
	}

	// Prepare UDP connection
	conn, err := net.DialUDP("udp", nil, resolve)
	if err != nil {
		log.Printf("error connecting to UDP: %v\n", err)
	}

	// Defer the closing of the connection
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("error closing UDP connection: %v\n", err)
		}
	}(conn)

	// Create new Reader to read input
	reader := bufio.NewReader(os.Stdin)

	// Start an infinite loop
	for {
		// Print "> " whilst awaiting user input
		fmt.Print("> ")

		// Read user input
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading input: %v\n", err)
		}

		// Write user input to the UDP connection
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error writing to UDP: %v\n", err)
		}
	}
}

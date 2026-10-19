package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:6380")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Read error: %v", err)
			}
			return
		}

		ackMessage := strings.ToUpper(strings.TrimSpace(message))
		response := fmt.Sprintf("Response: %s\n\n", ackMessage)

		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("Server write error: %v", err)
		}
	}
}

package main

import (
	"bufio"
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
		args, err := readCommand(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read command: %v", err)
			}
			return
		}

		if len(args) == 0 {
			log.Printf("No args: %v", err)
			return
		}

		command := strings.ToUpper(args[0])

		var response string

		switch command {
		case "PING":
			if len(args) != 1 {
				response = "-ERR PING arguments not supported yet\r\n"
			} else {
				response = "+PONG\r\n"
			}

		default:
			response = "-ERR unknown command\r\n"
		}

		if _, err := io.WriteString(conn, response); err != nil {
			log.Printf("Server write error: %v", err)
			return
		}
	}
}

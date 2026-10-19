package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"

	"github.com/magomed066/redis-clone/internal/store"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:6380")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	db := store.NewStore()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		go handleConnection(conn, db)
	}
}

func handleConnection(conn net.Conn, db *store.Store) {
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

		case "ECHO":
			if len(args) != 2 {
				response = "-ERR wrong number of arguments for ECHO\r\n"
			} else {
				response = encodeBulkString(args[1])
			}

		case "SET":
			if len(args) != 3 {
				response = "-ERR wrong number of arguments for SET\r\n"
			} else {
				db.Set(args[1], args[2])
				response = "+OK\r\n"
			}

		case "GET":
			if len(args) != 2 {
				response = "-ERR wrong number of arguments for GET\r\n"
			} else {
				val, exists := db.Get(args[1])
				if !exists {
					response = "$-1\r\n"
				} else {
					response = encodeBulkString(val)
				}
			}

		case "INCR":
			if len(args) != 2 {
				response = "-ERR wrong number of arguments for INCR\r\n"
			} else {
				val, err := db.Incr(args[1])
				if err != nil {
					response = "-ERR " + err.Error() + "\r\n"
				} else {
					response = encodeInteger(val)
				}
			}

		case "DEL":
			if len(args) < 2 {
				response = "-ERR wrong number of arguments for DEL\r\n"
			} else {
				deleted := db.DeleteMany(args[1:])
				if deleted < 1 {
					response = encodeInteger(0)
				} else {
					response = encodeInteger(deleted)
				}
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

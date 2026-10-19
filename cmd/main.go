package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/magomed066/redis-clone/internal/store"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	listener, err := net.Listen("tcp", "127.0.0.1:6380")
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()

	db := store.NewStore()

	cleanup := make(chan struct{})

	go func() {
		defer close(cleanup)
		db.RunCleanup(ctx)
	}()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			log.Printf("Accept error: %v", err)
			continue
		}

		go handleConnection(conn, db)
	}

	stop()
	<-cleanup
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

		response := executeCommand(db, args)

		if _, err := io.WriteString(conn, response); err != nil {
			log.Printf("Server write error: %v", err)
			return
		}
	}
}

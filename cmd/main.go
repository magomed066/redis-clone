package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/magomed066/redis-clone/internal/store"
)

const snapshotPath = "snapshot.json"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
	)
	defer stop()

	db := store.NewStore()

	// Restore saved data before accepting clients.
	if err := db.LoadSnapshot(snapshotPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Println("No snapshot found; starting with an empty store")
		} else {
			return fmt.Errorf("load snapshot: %w", err)
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:6380")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	log.Println("Listening on 127.0.0.1:6380")

	// Run one expiration cleanup worker for the entire store.
	cleanupDone := make(chan struct{})

	go func() {
		defer close(cleanupDone)
		db.RunCleanup(ctx)
	}()

	// Closing the listener unblocks Accept during shutdown.
	listenerDone := make(chan struct{})

	go func() {
		defer close(listenerDone)

		<-ctx.Done()
		listener.Close()
	}()

	var clients sync.WaitGroup
	var acceptErr error

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() == nil {
				acceptErr = fmt.Errorf("accept connection: %w", err)
			}
			break
		}

		// Shutdown may have started just as Accept returned.
		if ctx.Err() != nil {
			conn.Close()
			break
		}

		clients.Add(1)

		go func(conn net.Conn) {
			defer clients.Done()
			handleConnection(ctx, conn, db)
		}(conn)
	}

	// Cancel workers even if the loop ended because of an error.
	stop()

	// Wait until all store activity has finished before saving.
	clients.Wait()
	<-cleanupDone
	<-listenerDone

	log.Println("Saving final snapshot")

	if err := db.SaveSnapshot(snapshotPath); err != nil {
		return errors.Join(
			acceptErr,
			fmt.Errorf("save final snapshot: %w", err),
		)
	}

	log.Println("Server stopped; snapshot saved")

	return acceptErr
}

func handleConnection(
	ctx context.Context,
	conn net.Conn,
	db *store.Store,
) {
	defer conn.Close()

	finished := make(chan struct{})
	watcherDone := make(chan struct{})

	// Close this client's connection when shutdown begins.
	go func() {
		defer close(watcherDone)

		select {
		case <-ctx.Done():
			conn.Close()

		case <-finished:
			// The handler finished before server shutdown.
		}
	}()

	// Ensure the watcher exits when this handler returns.
	defer func() {
		close(finished)
		<-watcherDone
	}()

	reader := bufio.NewReader(conn)

	for {
		if ctx.Err() != nil {
			return
		}

		args, err := readCommand(reader)
		if err != nil {
			if ctx.Err() == nil && !errors.Is(err, io.EOF) {
				log.Printf("Read command: %v", err)
			}
			return
		}

		if ctx.Err() != nil {
			return
		}

		response := executeCommand(db, args)

		if _, err := io.WriteString(conn, response); err != nil {
			if ctx.Err() == nil {
				log.Printf("Write response: %v", err)
			}
			return
		}
	}
}

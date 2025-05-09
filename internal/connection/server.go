package connection

import (
	"bufio"
	"fmt"
	"mini-sgbd/internal/model"
	"mini-sgbd/internal/pipeline"
	"net"
	"sync"
)

var (
	activeConnections int
	connMutex         sync.Mutex
	maxConnections    int
	connSemaphore     chan struct{}
)

func StartServer(address string, maxConns int) {
	maxConnections = maxConns
	connSemaphore = make(chan struct{}, maxConns)

	ln, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server listening on %s (max connections: %d)\n", address, maxConns)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("[Server] Connection error: %v\n", err)
			continue
		}

		select {
		case connSemaphore <- struct{}{}:
			connMutex.Lock()
			activeConnections++
			connMutex.Unlock()
			fmt.Printf("[Server] New client connected (active connections: %d/%d)\n", activeConnections, maxConnections)
			go handleClient(conn)
		default:
			fmt.Printf("[Server] Connection rejected - maximum connections reached (%d)\n", maxConnections)
			conn.Close()
		}
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		conn.Close()
		<-connSemaphore
		connMutex.Lock()
		activeConnections--
		connMutex.Unlock()
		fmt.Printf("[Server] Client disconnected (active connections: %d/%d)\n", activeConnections, maxConnections)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		submitCommand(conn, line)
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("[Client] Error: %v\n", err)
	}
}

func submitCommand(conn net.Conn, raw string) {
	cmd := &model.ParsedCommand{
		Conn: conn,
		Raw:  raw,
	}
	pipeline.EnqueueParse(cmd)
}

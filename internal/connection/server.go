package connection

import (
	"bufio"
	"fmt"
	"mini-sgbd/internal/model"
	"mini-sgbd/internal/pipeline"
	"net"
)

func StartServer(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Server listening on %s\n", address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("[Server] Connection error: %v\n", err)
			continue
		}
		fmt.Printf("[Server] New client connected\n")
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
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

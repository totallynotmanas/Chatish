package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

const (
	TCP_SERVER_ADDR = "127.0.0.1:5000"
	WEB_SERVER_ADDR = ":8080"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . [server|client|web]")
		fmt.Println("  server - Run the secure chat server (CLI)")
		fmt.Println("  client - Run the chat client (CLI)")
		fmt.Println("  web    - Run the secure chat server with web UI")
		os.Exit(1)
	}

	mode := os.Args[1]

	if mode == "server" {
		runTCPServer()
	} else if mode == "client" {
		runClient()
	} else if mode == "web" {
		runWebServer()
	} else {
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func runTCPServer() {
	// Initialize the server
	server := NewSecureServer()

	// Start the TCP listener
	listener, err := net.Listen("tcp", TCP_SERVER_ADDR)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", TCP_SERVER_ADDR, err)
	}
	defer listener.Close()

	fmt.Printf("[SERVER] Secure Multi-Room Chat Server started on %s\n", TCP_SERVER_ADDR)
	fmt.Println("[SERVER] Waiting for connections...")
	fmt.Println("")
	fmt.Println("=== TEST CREDENTIALS ===")
	fmt.Println("  alice / password123 (Admin)")
	fmt.Println("  bob / secure456 (Member)")
	fmt.Println("  charlie / guest789 (Guest)")
	fmt.Println("")

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		// Handle each client connection in a separate goroutine
		go server.HandleClient(conn)
	}
}

func runClient() {
	RunClientCLI(TCP_SERVER_ADDR)
}

func runWebServer() {
	// Initialize the secure server
	secureServer := NewSecureServer()

	// Start TCP server in background goroutine
	go func() {
		listener, err := net.Listen("tcp", TCP_SERVER_ADDR)
		if err != nil {
			log.Printf("Failed to start TCP server: %v", err)
			return
		}
		defer listener.Close()

		fmt.Printf("[TCP SERVER] Listening on %s\n", TCP_SERVER_ADDR)

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept TCP connection: %v", err)
				continue
			}
			go secureServer.HandleClient(conn)
		}
	}()

	// Start HTTP+WebSocket server
	runHTTPServer(secureServer, WEB_SERVER_ADDR)
}

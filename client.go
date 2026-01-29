package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// ClientConnection represents a client connection to the server
type ClientConnection struct {
	conn        net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	sessionKey  []byte
	username    string
	role        string
	currentRoom string
}

// ConnectToServer establishes a connection to the secure chat server
func ConnectToServer(serverAddr string) (*ClientConnection, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, err
	}

	client := &ClientConnection{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	return client, nil
}

// Close closes the connection
func (c *ClientConnection) Close() error {
	return c.conn.Close()
}

// PerformKeyExchange performs the simplified key exchange with the server
func (c *ClientConnection) PerformKeyExchange() error {
	// Receive server's public key
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}

	line = strings.TrimSpace(line)

	var msg Message
	serverMsgB64, err := base64.StdEncoding.DecodeString(line)
	if err != nil {
		return err
	}

	err = json.Unmarshal(serverMsgB64, &msg)
	if err != nil {
		return err
	}

	if msg.Type != "key_exchange" {
		return fmt.Errorf("expected key_exchange, got %s", msg.Type)
	}

	serverPublic, err := base64.StdEncoding.DecodeString(msg.Content)
	if err != nil {
		return err
	}

	// Generate client's DH parameters
	clientPrivate := make([]byte, 32)
	_, err = rand.Read(clientPrivate)
	if err != nil {
		return err
	}

	clientPublic := derivePublic(clientPrivate)

	// Send client public key
	response := &Message{
		Type:    "key_exchange",
		Content: base64.StdEncoding.EncodeToString(clientPublic),
	}

	jsonData, _ := json.Marshal(response)
	encoded := base64.StdEncoding.EncodeToString(jsonData)
	c.writer.WriteString(encoded + "\n")
	c.writer.Flush()

	// Derive shared session key
	c.sessionKey = deriveSharedSecret(clientPrivate, serverPublic)

	fmt.Println("[CLIENT] Session key established")
	return nil
}

// Authenticate performs authentication and MFA
func (c *ClientConnection) Authenticate(username, password string) error {
	fmt.Print("Enter username: ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	username = input.Text()

	fmt.Print("Enter password: ")
	input.Scan()
	password = input.Text()

	// Send credentials
	authMsg := &Message{
		Type:     "auth",
		Username: username,
		Content:  password,
	}

	err := c.sendMessage(authMsg)
	if err != nil {
		return err
	}

	// Receive response (should be OTP request)
	data, err := c.readMessage()
	if err != nil {
		return err
	}

	var response Message
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	if response.Type == "error" {
		return fmt.Errorf("authentication error: %s", response.Content)
	}

	if response.Type != "otp_required" {
		return fmt.Errorf("unexpected response: %s", response.Type)
	}

	// Read OTP from server console
	fmt.Println("[SERVER] Check server console for your OTP")
	fmt.Print("Enter your OTP: ")
	input.Scan()
	otp := input.Text()

	// Send OTP
	otpMsg := &Message{
		Type:    "otp",
		Content: otp,
	}

	err = c.sendMessage(otpMsg)
	if err != nil {
		return err
	}

	// Receive auth success
	data, err = c.readMessage()
	if err != nil {
		return err
	}

	var finalResponse Message
	err = json.Unmarshal(data, &finalResponse)
	if err != nil {
		return err
	}

	if finalResponse.Type == "error" {
		return fmt.Errorf("authentication failed: %s", finalResponse.Content)
	}

	if finalResponse.Type != "auth_success" {
		return fmt.Errorf("unexpected response: %s", finalResponse.Type)
	}

	c.username = username
	fmt.Printf("[CLIENT] %s\n", finalResponse.Content)
	return nil
}

// readMessage reads, decodes, and decrypts a message from the server
func (c *ClientConnection) readMessage() ([]byte, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(line)
	if err != nil {
		return nil, err
	}

	// Extract HMAC signature (last 32 bytes)
	if len(ciphertext) < 32 {
		return nil, fmt.Errorf("invalid message format")
	}

	messageData := ciphertext[:len(ciphertext)-32]
	signature := ciphertext[len(ciphertext)-32:]

	// Verify HMAC
	if !verifyHMAC(messageData, signature, c.sessionKey) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	// Decrypt
	plaintext, err := decryptAES(c.sessionKey, messageData)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// sendMessage encrypts and sends a message to the server
func (c *ClientConnection) sendMessage(msg *Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Encrypt
	ciphertext, err := encryptAES(c.sessionKey, jsonData)
	if err != nil {
		return err
	}

	// Compute HMAC
	signature := computeHMAC(ciphertext, c.sessionKey)

	// Combine ciphertext and signature
	packet := append(ciphertext, signature...)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(packet)

	// Send
	_, err = c.writer.WriteString(encoded + "\n")
	if err != nil {
		return err
	}

	return c.writer.Flush()
}

// JoinRoom sends a join request for a room
func (c *ClientConnection) JoinRoom(room string) error {
	msg := &Message{
		Type: "join",
		Room: room,
	}

	err := c.sendMessage(msg)
	if err != nil {
		return err
	}

	// Read response
	data, err := c.readMessage()
	if err != nil {
		return err
	}

	var response Message
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	if response.Type == "error" {
		return fmt.Errorf("join failed: %s", response.Content)
	}

	c.currentRoom = room
	fmt.Printf("[CLIENT] %s\n", response.Content)
	return nil
}

// SendChatMessage sends a chat message to the current room
func (c *ClientConnection) SendChatMessage(content string) error {
	if c.currentRoom == "" {
		return fmt.Errorf("not in any room")
	}

	msg := &Message{
		Type:    "chat",
		Content: content,
	}

	return c.sendMessage(msg)
}

// LeaveRoom leaves the current room
func (c *ClientConnection) LeaveRoom() error {
	if c.currentRoom == "" {
		return fmt.Errorf("not in any room")
	}

	msg := &Message{
		Type: "leave",
	}

	err := c.sendMessage(msg)
	if err != nil {
		return err
	}

	// Read response
	data, err := c.readMessage()
	if err != nil {
		return err
	}

	var response Message
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	c.currentRoom = ""
	fmt.Printf("[CLIENT] %s\n", response.Content)
	return nil
}

// ListRooms lists available rooms
func (c *ClientConnection) ListRooms() error {
	msg := &Message{
		Type: "list",
	}

	err := c.sendMessage(msg)
	if err != nil {
		return err
	}

	// Read response
	data, err := c.readMessage()
	if err != nil {
		return err
	}

	var response Message
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}

	fmt.Println("[ROOMS]", response.Content)
	return nil
}

// ListenForMessages listens for incoming messages from the server
func (c *ClientConnection) ListenForMessages(done chan bool) {
	for {
		data, err := c.readMessage()
		if err != nil {
			if err == io.EOF {
				fmt.Println("[SERVER] Connection closed")
				done <- true
				return
			}
			fmt.Printf("[ERROR] %v\n", err)
			continue
		}

		var msg Message
		err = json.Unmarshal(data, &msg)
		if err != nil {
			continue
		}

		// Display message
		if msg.Type == "system" {
			fmt.Printf("[%s] %s\n", msg.Room, msg.Content)
		} else if msg.Type == "chat" {
			fmt.Printf("[%s] %s: %s\n", msg.Room, msg.Username, msg.Content)
		} else if msg.Type == "error" {
			fmt.Printf("[ERROR] %s\n", msg.Content)
		}
	}
}

// ============================================
// CLIENT CLI INTERFACE
// ============================================

func RunClientCLI(serverAddr string) {
	// Connect to server
	client, err := ConnectToServer(serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer client.Close()

	fmt.Println("=== Secure Multi-Room Chat Client ===")
	fmt.Println("Connected to server at", serverAddr)

	// Perform key exchange
	err = client.PerformKeyExchange()
	if err != nil {
		fmt.Printf("Key exchange failed: %v\n", err)
		return
	}

	// Authenticate
	err = client.Authenticate("", "")
	if err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}

	// Start listening for messages
	done := make(chan bool)
	go client.ListenForMessages(done)

	// CLI loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nAvailable commands:")
	fmt.Println("  /list - List available rooms")
	fmt.Println("  /join <room> - Join a room")
	fmt.Println("  /leave - Leave current room")
	fmt.Println("  /msg <message> - Send a message")
	fmt.Println("  /quit - Exit")
	fmt.Print("\n> ")

	for scanner.Scan() {
		select {
		case <-done:
			return
		default:
		}

		input := scanner.Text()

		if input == "" {
			fmt.Print("> ")
			continue
		}

		parts := strings.Fields(input)
		command := parts[0]

		switch command {
		case "/list":
			client.ListRooms()

		case "/join":
			if len(parts) < 2 {
				fmt.Println("Usage: /join <room>")
			} else {
				room := strings.Join(parts[1:], " ")
				err := client.JoinRoom(room)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}

		case "/leave":
			err := client.LeaveRoom()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "/msg":
			if len(parts) < 2 {
				fmt.Println("Usage: /msg <message>")
			} else {
				message := strings.Join(parts[1:], " ")
				err := client.SendChatMessage(message)
				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}

		case "/quit":
			fmt.Println("Disconnecting...")
			return

		default:
			fmt.Println("Unknown command. Use /list, /join, /leave, /msg, or /quit")
		}

		fmt.Print("> ")
	}
}

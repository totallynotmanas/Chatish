package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// ============================================
// 1. AUTHENTICATION & AUTHORIZATION TYPES
// ============================================

type Role string

const (
	ROLE_GUEST  Role = "Guest"
	ROLE_MEMBER Role = "Member"
	ROLE_ADMIN  Role = "Admin"
)

type Room string

const (
	ROOM_GENERAL    Room = "General"
	ROOM_VIP        Room = "VIP"
	ROOM_ADMIN_LOGS Room = "Admin"
)

// User represents a registered user with credentials
type User struct {
	Username     string
	PasswordHash string
	PasswordSalt string
	Role         Role
}

// ClientSession represents an authenticated client connection
type ClientSession struct {
	conn          net.Conn
	username      string
	role          Role
	currentRoom   Room
	sessionKey    []byte // Shared encryption key
	sessionID     string
	authenticated bool
	reader        *bufio.Reader
	writer        *bufio.Writer
}

// SecureServer manages chat rooms, users, and authentication
type SecureServer struct {
	rooms         map[Room]map[string]*ClientSession // room -> username -> client
	users         map[string]*User                   // username -> user
	clientsMutex  sync.RWMutex
	usersMutex    sync.RWMutex
	sessionsMutex sync.RWMutex
	otpStore      map[string]string // username -> current OTP
	otpMutex      sync.Mutex
}

// ============================================
// 2. INITIALIZATION & SERVER SETUP
// ============================================

func NewSecureServer() *SecureServer {
	server := &SecureServer{
		rooms:    make(map[Room]map[string]*ClientSession),
		users:    make(map[string]*User),
		otpStore: make(map[string]string),
	}

	// Initialize empty rooms
	server.rooms[ROOM_GENERAL] = make(map[string]*ClientSession)
	server.rooms[ROOM_VIP] = make(map[string]*ClientSession)
	server.rooms[ROOM_ADMIN_LOGS] = make(map[string]*ClientSession)

	// Initialize mock user database
	server.initializeMockUsers()

	return server
}

// initializeMockUsers creates demo users with hashed passwords
func (s *SecureServer) initializeMockUsers() {
	mockUsers := []struct {
		username string
		password string
		role     Role
	}{
		{"alice", "password123", ROLE_ADMIN},
		{"bob", "secure456", ROLE_MEMBER},
		{"charlie", "guest789", ROLE_GUEST},
	}

	s.usersMutex.Lock()
	defer s.usersMutex.Unlock()

	for _, u := range mockUsers {
		salt := generateRandomSalt()
		hash := hashPassword(u.password, salt)
		s.users[u.username] = &User{
			Username:     u.username,
			PasswordHash: hash,
			PasswordSalt: salt,
			Role:         u.role,
		}
	}

	fmt.Println("[SERVER] Mock users initialized:")
	fmt.Println("  - alice / password123 (Admin)")
	fmt.Println("  - bob / secure456 (Member)")
	fmt.Println("  - charlie / guest789 (Guest)")
}

// ============================================
// 3. AUTHENTICATION (MFA)
// ============================================

// handleAuthentication performs login and MFA
func (s *SecureServer) handleAuthentication(session *ClientSession) bool {
	// Read credentials
	credentials := &Message{}
	data, err := session.readEncryptedMessage()
	if err != nil {
		fmt.Printf("[AUTH DEBUG] Failed to read credentials: %v\n", err)
		session.sendError("Failed to read credentials")
		return false
	}

	err = json.Unmarshal(data, credentials)
	if err != nil || credentials.Type != "auth" {
		fmt.Printf("[AUTH DEBUG] Invalid auth format. Error: %v, Type: %s\n", err, credentials.Type)
		session.sendError("Invalid authentication format")
		return false
	}

	username := credentials.Username
	password := credentials.Content

	// Step 1: Verify credentials
	s.usersMutex.RLock()
	user, exists := s.users[username]
	s.usersMutex.RUnlock()

	if !exists {
		session.sendError("User not found")
		return false
	}

	providedHash := hashPassword(password, user.PasswordSalt)
	if providedHash != user.PasswordHash {
		fmt.Printf("[AUTH DEBUG] Invalid password for user '%s'. Expected: %s, Got: %s\n", username, user.PasswordHash, providedHash)
		session.sendError("Invalid password")
		return false
	}

	fmt.Printf("[AUTH] User '%s' password verified (Role: %s)\n", username, user.Role)

	// Step 2: Generate OTP and send to client
	otp := generateOTP()
	s.otpMutex.Lock()
	s.otpStore[username] = otp
	s.otpMutex.Unlock()

	// Print OTP to server console (simulation)
	fmt.Printf("[MFA] OTP for '%s': %s (Valid for 2 minutes)\n", username, otp)

	// Notify client to enter OTP
	otpRequest := &Message{
		Type:    "otp_required",
		Content: "Enter your 6-digit OTP",
	}
	session.sendMessage(otpRequest)

	// Step 3: Read OTP from client
	otpMsg := &Message{}
	data, err = session.readEncryptedMessage()
	if err != nil {
		session.sendError("Failed to read OTP")
		return false
	}

	err = json.Unmarshal(data, otpMsg)
	if err != nil || otpMsg.Type != "otp" {
		session.sendError("Invalid OTP format")
		return false
	}

	s.otpMutex.Lock()
	storedOTP, otpExists := s.otpStore[username]
	delete(s.otpStore, username) // One-time use
	s.otpMutex.Unlock()

	if !otpExists || otpMsg.Content != storedOTP {
		session.sendError("Invalid OTP")
		return false
	}

	fmt.Printf("[MFA] OTP verified for '%s'\n", username)

	// Authentication successful
	session.username = username
	session.role = user.Role
	session.authenticated = true

	successMsg := &Message{
		Type:    "auth_success",
		Content: fmt.Sprintf("Authenticated as %s (%s)", username, user.Role),
	}
	session.sendMessage(successMsg)

	return true
}

// ============================================
// 4. AUTHORIZATION (ACL)
// ============================================

// canJoinRoom checks if a user can join a room based on ACL
func (s *SecureServer) canJoinRoom(role Role, room Room) bool {
	switch role {
	case ROLE_GUEST:
		return room == ROOM_GENERAL
	case ROLE_MEMBER:
		return room == ROOM_GENERAL || room == ROOM_VIP
	case ROLE_ADMIN:
		return true // Admins can join all rooms
	default:
		return false
	}
}

// ============================================
// 5. MESSAGE OPERATIONS
// ============================================

// readEncryptedMessage reads, decodes, and decrypts a message
func (session *ClientSession) readEncryptedMessage() ([]byte, error) {
	line, err := session.reader.ReadString('\n')
	if err != nil {
		fmt.Printf("[READ DEBUG] Read error: %v\n", err)
		return nil, err
	}

	// Remove newline
	line = strings.TrimSpace(line)
	fmt.Printf("[READ DEBUG] Received line (first 60 chars): %s\n", line[:min(60, len(line))])

	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(line)
	if err != nil {
		fmt.Printf("[READ DEBUG] Base64 decode error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[READ DEBUG] Decoded ciphertext length: %d bytes\n", len(ciphertext))

	// Extract HMAC signature (last 32 bytes)
	if len(ciphertext) < 32 {
		fmt.Printf("[READ DEBUG] Ciphertext too short: %d bytes\n", len(ciphertext))
		return nil, fmt.Errorf("invalid message format")
	}

	messageData := ciphertext[:len(ciphertext)-32]
	signature := ciphertext[len(ciphertext)-32:]

	fmt.Printf("[READ DEBUG] Message data length: %d, Signature: %v\n", len(messageData), signature[:8])

	// Verify HMAC
	if !verifyHMAC(messageData, signature, session.sessionKey) {
		fmt.Printf("[READ DEBUG] HMAC verification failed for message\n")
		return nil, fmt.Errorf("HMAC verification failed")
	}

	fmt.Printf("[READ DEBUG] HMAC verification SUCCESS\n")

	// Decrypt
	plaintext, err := decryptAES(session.sessionKey, messageData)
	if err != nil {
		fmt.Printf("[READ DEBUG] Decryption error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[READ DEBUG] Decrypted successfully, length: %d\n", len(plaintext))
	return plaintext, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// readUnencryptedMessage reads and decodes an unencrypted message (for key exchange)
func (session *ClientSession) readUnencryptedMessage() ([]byte, error) {
	line, err := session.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	return base64.StdEncoding.DecodeString(line)
}

// sendMessage encrypts, computes HMAC, and sends a message
func (session *ClientSession) sendMessage(msg *Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Encrypt
	ciphertext, err := encryptAES(session.sessionKey, jsonData)
	if err != nil {
		return err
	}

	// Compute HMAC
	signature := computeHMAC(ciphertext, session.sessionKey)

	// Combine ciphertext and signature
	packet := append(ciphertext, signature...)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(packet)

	// Send
	_, err = session.writer.WriteString(encoded + "\n")
	if err != nil {
		return err
	}

	return session.writer.Flush()
}

// sendUnencryptedMessage sends a base64-encoded but unencrypted message
func (session *ClientSession) sendUnencryptedMessage(msg *Message) error {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	_, err = session.writer.WriteString(encoded + "\n")
	if err != nil {
		return err
	}

	return session.writer.Flush()
}

// sendError sends an error message to the client
func (session *ClientSession) sendError(errorMsg string) {
	msg := &Message{
		Type:    "error",
		Content: errorMsg,
	}

	if session.authenticated && session.sessionKey != nil {
		session.sendMessage(msg)
	} else {
		session.sendUnencryptedMessage(msg)
	}
}

// ============================================
// 6. KEY EXCHANGE
// ============================================

// performKeyExchange performs the simplified key exchange with the server
func (s *SecureServer) performKeyExchange(session *ClientSession) ([]byte, error) {
	// Generate server's DH parameters (simplified)
	serverPrivate := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, serverPrivate)
	if err != nil {
		return nil, err
	}

	serverPublic := derivePublic(serverPrivate)

	// Send server public key
	keyExchangeMsg := &Message{
		Type:    "key_exchange",
		Content: base64.StdEncoding.EncodeToString(serverPublic),
	}
	session.sendUnencryptedMessage(keyExchangeMsg)

	// Receive client's public key
	clientMsg := &Message{}
	data, err := session.readUnencryptedMessage()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, clientMsg)
	if err != nil {
		return nil, err
	}

	clientPublicB64 := clientMsg.Content
	fmt.Printf("[KEY EXCHANGE DEBUG] Received client public B64 (first 40): %s\n", clientPublicB64[:min(40, len(clientPublicB64))])

	clientPublic, err := base64.StdEncoding.DecodeString(clientPublicB64)
	if err != nil {
		return nil, err
	}

	fmt.Printf("[KEY EXCHANGE DEBUG] Client public decoded (first 8): %v\n", clientPublic[:8])

	// Derive shared session key
	sessionKey := deriveSharedSecret(serverPrivate, clientPublic)

	fmt.Printf("[KEY EXCHANGE DEBUG] Server private (first 8): %v\n", serverPrivate[:8])
	fmt.Printf("[KEY EXCHANGE DEBUG] Client public (first 8): %v\n", clientPublic[:8])
	fmt.Printf("[KEY EXCHANGE DEBUG] Derived session key (first 8): %v\n", sessionKey[:8])

	fmt.Printf("[ENCRYPTION] Session key established\n")

	return sessionKey, nil
}

// ============================================
// 7. CLIENT HANDLER & ROOM MANAGEMENT
// ============================================

// HandleClient handles a single client connection
func (s *SecureServer) HandleClient(conn net.Conn) {
	defer conn.Close()

	session := &ClientSession{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		sessionID: generateSessionID(),
	}

	// Log connection
	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[CONNECTION] New client from %s\n", clientAddr)

	// Perform key exchange before authentication
	sessionKey, err := s.performKeyExchange(session)
	if err != nil {
		fmt.Printf("[ERROR] Key exchange failed for %s: %v\n", clientAddr, err)
		session.sendError("Key exchange failed")
		return
	}
	session.sessionKey = sessionKey

	// Authenticate with MFA
	if !s.handleAuthentication(session) {
		fmt.Printf("[AUTH FAILED] User authentication failed for %s\n", clientAddr)
		return
	}

	fmt.Printf("[SUCCESS] User '%s' authenticated and connected from %s\n", session.username, clientAddr)

	// Main message loop
	s.handleClientMessages(session)
}

// handleClientMessages handles incoming messages from a client
func (s *SecureServer) handleClientMessages(session *ClientSession) {
	defer func() {
		s.leaveAllRooms(session)
		fmt.Printf("[DISCONNECT] User '%s' disconnected\n", session.username)
	}()

	for {
		msg := &Message{}
		data, err := session.readEncryptedMessage()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("[ERROR] Failed to read message from '%s': %v\n", session.username, err)
			break
		}

		err = json.Unmarshal(data, msg)
		if err != nil {
			session.sendError("Invalid message format")
			continue
		}

		// Process message based on type
		switch msg.Type {
		case "join":
			s.handleJoinRoom(session, msg)
		case "chat":
			s.handleChatMessage(session, msg)
		case "leave":
			s.handleLeaveRoom(session, msg)
		case "list":
			s.handleListRooms(session)
		default:
			session.sendError("Unknown message type")
		}
	}
}

// handleJoinRoom processes a room join request with ACL
func (s *SecureServer) handleJoinRoom(session *ClientSession, msg *Message) {
	room := Room(msg.Room)

	// Verify room exists
	s.clientsMutex.RLock()
	_, roomExists := s.rooms[room]
	s.clientsMutex.RUnlock()

	if !roomExists {
		session.sendError(fmt.Sprintf("Room '%s' does not exist", room))
		return
	}

	// Check ACL (Authorization)
	if !s.canJoinRoom(session.role, room) {
		session.sendError(fmt.Sprintf("Access Denied: Your role '%s' cannot access '%s'", session.role, room))
		fmt.Printf("[ACL DENIED] User '%s' (role: %s) tried to access '%s'\n", session.username, session.role, room)
		return
	}

	// Remove user from previous room if they were in one
	s.clientsMutex.Lock()
	if session.currentRoom != "" {
		oldRoom := session.currentRoom
		delete(s.rooms[oldRoom], session.username)
		fmt.Printf("[ROOM] User '%s' left '%s' (automatic when joining '%s')\n", session.username, oldRoom, room)
	}
	
	// Add user to new room
	s.rooms[room][session.username] = session
	s.clientsMutex.Unlock()

	session.currentRoom = room

	// Notify client
	response := &Message{
		Type:    "join_success",
		Room:    string(room),
		Content: fmt.Sprintf("Joined %s", room),
	}
	session.sendMessage(response)

	// Broadcast join announcement
	s.broadcastToRoom(room, &Message{
		Type:      "system",
		Username:  "System",
		Room:      string(room),
		Content:   fmt.Sprintf("%s joined the room", session.username),
		Timestamp: time.Now().Format("15:04:05"),
	})

	fmt.Printf("[ROOM] User '%s' joined '%s'\n", session.username, room)
}

// handleChatMessage broadcasts a chat message to a room
func (s *SecureServer) handleChatMessage(session *ClientSession, msg *Message) {
	if session.currentRoom == "" {
		session.sendError("You are not in any room")
		return
	}

	msg.Username = session.username
	msg.Room = string(session.currentRoom)
	msg.Timestamp = time.Now().Format("15:04:05")

	s.broadcastToRoom(session.currentRoom, msg)
	fmt.Printf("[MSG] %s (%s): %s\n", session.username, session.currentRoom, msg.Content)
}

// handleLeaveRoom removes a user from a room
func (s *SecureServer) handleLeaveRoom(session *ClientSession, msg *Message) {
	if session.currentRoom == "" {
		session.sendError("You are not in any room")
		return
	}

	room := session.currentRoom

	s.clientsMutex.Lock()
	delete(s.rooms[room], session.username)
	s.clientsMutex.Unlock()

	response := &Message{
		Type:    "leave_success",
		Content: fmt.Sprintf("Left %s", room),
	}
	session.sendMessage(response)

	s.broadcastToRoom(room, &Message{
		Type:      "system",
		Username:  "System",
		Room:      string(room),
		Content:   fmt.Sprintf("%s left the room", session.username),
		Timestamp: time.Now().Format("15:04:05"),
	})

	session.currentRoom = ""
	fmt.Printf("[ROOM] User '%s' left '%s'\n", session.username, room)
}

// handleListRooms lists available rooms based on user permissions
func (s *SecureServer) handleListRooms(session *ClientSession) {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	var availableRooms []string
	for room := range s.rooms {
		if s.canJoinRoom(session.role, room) {
			availableRooms = append(availableRooms, string(room))
		}
	}

	response := &Message{
		Type:  "rooms_list",
		Rooms: availableRooms,
	}
	session.sendMessage(response)
}

// broadcastToRoom sends a message to all users in a room
func (s *SecureServer) broadcastToRoom(room Room, msg *Message) {
	s.clientsMutex.RLock()
	clients := s.rooms[room]
	s.clientsMutex.RUnlock()

	for _, client := range clients {
		err := client.sendMessage(msg)
		if err != nil {
			fmt.Printf("[ERROR] Failed to send message to '%s': %v\n", client.username, err)
		}
	}
}

// leaveAllRooms removes a user from all rooms
func (s *SecureServer) leaveAllRooms(session *ClientSession) {
	s.clientsMutex.Lock()
	defer s.clientsMutex.Unlock()

	for room := range s.rooms {
		if _, exists := s.rooms[room][session.username]; exists {
			delete(s.rooms[room], session.username)
		}
	}
}

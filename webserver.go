package main

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		return true
	},
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	tcpConn    net.Conn
	tcpReader  *bufio.Reader
	tcpWriter  *bufio.Writer
	ws         *websocket.Conn
	server     *SecureServer
	username   string
	mu         sync.Mutex
	done       chan bool
	encryptor  *ClientEncryptor
	isAuthed   bool
	sessionKey []byte
}

// ClientEncryptor handles encryption/decryption for web clients
type ClientEncryptor struct {
	sessionKey []byte
	hmacKey    []byte
}

// Message types for web protocol
type WebMessage struct {
	Type     string   `json:"type"`
	Content  string   `json:"content"`
	Error    string   `json:"error"`
	OTP      string   `json:"otp,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Room     string   `json:"room,omitempty"`
	Rooms    []string `json:"rooms,omitempty"`
}

func runHTTPServer(server *SecureServer, addr string) {
	// Serve static files from disk
	http.HandleFunc("/", serveIndex)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(w, r, server)
	})

	fmt.Printf("[WEB SERVER] Starting HTTP+WebSocket server on http://localhost%s\n", addr)
	fmt.Printf("[WEB SERVER] Open browser: http://localhost%s\n", addr)
	fmt.Println("")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Serve HTML from templates directory
	data, err := os.ReadFile("templates/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error loading page"))
		fmt.Printf("[ERROR] Failed to load index.html: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

/* OLD EMBEDDED HTML REMOVED - Now using templates/index.html from disk */

func dummyOld_RemoveThis() {
	/* THIS SECTION IS KEPT ONLY TO PREVENT PARSING ERRORS
		<!DOCTYPE html>
	<html lang="en">
	<head>
	    <meta charset="UTF-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1.0">
	    <title>Secure Multi-Room Chat</title>
	    <style>
	        * {
	            margin: 0;
	            padding: 0;
	            box-sizing: border-box;
	        }

	        :root {
	            --bg: #0f1720;
	            --fg: #e0e0e0;
	            --accent: #61dafb;
	            --error: #ff6b6b;
	            --success: #51cf66;
	            --input: #1a252f;
	        }

	        body {
	            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
	            background: linear-gradient(135deg, var(--bg) 0%, #1a2940 100%);
	            color: var(--fg);
	            height: 100vh;
	            overflow: hidden;
	        }

	        .container {
	            display: flex;
	            height: 100vh;
	        }

	        .sidebar {
	            width: 250px;
	            background: rgba(15, 23, 32, 0.95);
	            border-right: 2px solid var(--accent);
	            padding: 20px;
	            display: flex;
	            flex-direction: column;
	            gap: 20px;
	        }

	        .logo {
	            font-size: 24px;
	            font-weight: bold;
	            color: var(--accent);
	            text-align: center;
	            padding-bottom: 10px;
	            border-bottom: 2px solid var(--accent);
	        }

	        .auth-section {
	            display: flex;
	            flex-direction: column;
	            gap: 10px;
	        }

	        .auth-section input {
	            padding: 10px;
	            background: var(--input);
	            border: 1px solid var(--accent);
	            color: var(--fg);
	            border-radius: 4px;
	            font-size: 14px;
	        }

	        .auth-section input::placeholder {
	            color: #888;
	        }

	        .auth-section button {
	            padding: 10px;
	            background: var(--accent);
	            color: var(--bg);
	            border: none;
	            border-radius: 4px;
	            cursor: pointer;
	            font-weight: bold;
	            transition: background 0.3s;
	        }

	        .auth-section button:hover {
	            background: #4da8cc;
	        }

	        .otp-section {
	            display: none;
	            flex-direction: column;
	            gap: 10px;
	        }

	        .status {
	            padding: 10px;
	            background: var(--input);
	            border-left: 4px solid var(--accent);
	            border-radius: 4px;
	            font-size: 14px;
	        }

	        .status.error {
	            border-left-color: var(--error);
	            color: var(--error);
	        }

	        .status.success {
	            border-left-color: var(--success);
	            color: var(--success);
	        }

	        .rooms-section {
	            display: none;
	            flex-direction: column;
	            gap: 10px;
	        }

	        .rooms-section h3 {
	            color: var(--accent);
	            font-size: 14px;
	            text-transform: uppercase;
	            letter-spacing: 1px;
	        }

	        .room-list {
	            display: flex;
	            flex-direction: column;
	            gap: 5px;
	        }

	        .room-btn {
	            padding: 10px;
	            background: var(--input);
	            border: 1px solid #333;
	            color: var(--fg);
	            border-radius: 4px;
	            cursor: pointer;
	            transition: all 0.3s;
	            text-align: left;
	            font-size: 14px;
	        }

	        .room-btn:hover {
	            border-color: var(--accent);
	            background: rgba(97, 218, 251, 0.1);
	        }

	        .room-btn.active {
	            background: var(--accent);
	            color: var(--bg);
	            border-color: var(--accent);
	        }

	        .main-content {
	            flex: 1;
	            display: flex;
	            flex-direction: column;
	            background: var(--bg);
	        }

	        .header {
	            padding: 20px;
	            background: rgba(97, 218, 251, 0.1);
	            border-bottom: 2px solid var(--accent);
	            display: flex;
	            justify-content: space-between;
	            align-items: center;
	        }

	        .header h2 {
	            color: var(--accent);
	        }

	        .user-info {
	            font-size: 14px;
	            color: var(--fg);
	        }

	        .logout-btn {
	            padding: 8px 16px;
	            background: var(--error);
	            color: white;
	            border: none;
	            border-radius: 4px;
	            cursor: pointer;
	            font-size: 14px;
	        }

	        .logout-btn:hover {
	            background: #ff5252;
	        }

	        .messages {
	            flex: 1;
	            overflow-y: auto;
	            padding: 20px;
	            display: flex;
	            flex-direction: column;
	            gap: 10px;
	        }

	        .message {
	            padding: 12px;
	            background: rgba(97, 218, 251, 0.1);
	            border-left: 4px solid var(--accent);
	            border-radius: 4px;
	            word-wrap: break-word;
	        }

	        .message .time {
	            font-size: 12px;
	            color: #888;
	            margin-bottom: 4px;
	        }

	        .message .user {
	            color: var(--accent);
	            font-weight: bold;
	            margin-right: 8px;
	        }

	        .message .text {
	            color: var(--fg);
	        }

	        .system-message {
	            background: rgba(81, 207, 102, 0.1);
	            border-left-color: var(--success);
	            font-style: italic;
	            color: var(--success);
	        }

	        .error-message {
	            background: rgba(255, 107, 107, 0.1);
	            border-left-color: var(--error);
	            color: var(--error);
	        }

	        .input-area {
	            padding: 20px;
	            border-top: 2px solid var(--accent);
	            display: none;
	            gap: 10px;
	        }

	        .input-area.active {
	            display: flex;
	        }

	        .input-area input {
	            flex: 1;
	            padding: 12px;
	            background: var(--input);
	            border: 1px solid var(--accent);
	            color: var(--fg);
	            border-radius: 4px;
	            font-size: 14px;
	        }

	        .input-area button {
	            padding: 12px 24px;
	            background: var(--accent);
	            color: var(--bg);
	            border: none;
	            border-radius: 4px;
	            cursor: pointer;
	            font-weight: bold;
	            transition: background 0.3s;
	        }

	        .input-area button:hover {
	            background: #4da8cc;
	        }

	        ::-webkit-scrollbar {
	            width: 8px;
	        }

	        ::-webkit-scrollbar-track {
	            background: var(--bg);
	        }

	        ::-webkit-scrollbar-thumb {
	            background: var(--accent);
	            border-radius: 4px;
	        }

	        ::-webkit-scrollbar-thumb:hover {
	            background: #4da8cc;
	        }

	        .login-container {
	            display: flex;
	            justify-content: center;
	            align-items: center;
	            height: 100vh;
	            background: linear-gradient(135deg, var(--bg) 0%, #1a2940 100%);
	        }

	        .login-box {
	            background: rgba(15, 23, 32, 0.95);
	            padding: 40px;
	            border-radius: 8px;
	            border: 2px solid var(--accent);
	            min-width: 300px;
	            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
	        }

	        .login-box h1 {
	            color: var(--accent);
	            margin-bottom: 30px;
	            text-align: center;
	        }

	        .login-box input {
	            width: 100%;
	            padding: 12px;
	            margin-bottom: 15px;
	            background: var(--input);
	            border: 1px solid var(--accent);
	            color: var(--fg);
	            border-radius: 4px;
	            font-size: 14px;
	        }

	        .login-box button {
	            width: 100%;
	            padding: 12px;
	            background: var(--accent);
	            color: var(--bg);
	            border: none;
	            border-radius: 4px;
	            cursor: pointer;
	            font-weight: bold;
	            font-size: 16px;
	            transition: background 0.3s;
	        }

	        .login-box button:hover {
	            background: #4da8cc;
	        }

	        .error {
	            color: var(--error);
	            margin-top: 15px;
	            text-align: center;
	        }

	        .info {
	            color: #888;
	            font-size: 12px;
	            margin-top: 20px;
	            padding-top: 20px;
	            border-top: 1px solid #333;
	            line-height: 1.6;
	        }
	    </style>
	</head>
	<body>
	    <div id="loginContainer" class="login-container">
	        <div class="login-box">
	            <h1>🔐 Secure Chat</h1>
	            <input type="text" id="username" placeholder="Username" />
	            <input type="password" id="password" placeholder="Password" />
	            <button onclick="login()">Login</button>
	            <div id="loginError" class="error"></div>
	            <div class="info">
	                <strong>Test Credentials:</strong><br>
	                alice / password123 (Admin)<br>
	                bob / secure456 (Member)<br>
	                charlie / guest789 (Guest)
	            </div>
	        </div>
	    </div>

	    <div id="chatContainer" style="display:none;">
	        <div class="container">
	            <div class="sidebar">
	                <div class="logo">💬 Chatish</div>

	                <div id="otpSection" class="otp-section">
	                    <label style="font-size: 12px; color: var(--accent);">ENTER OTP:</label>
	                    <input type="text" id="otpInput" placeholder="6-digit OTP" maxlength="6" />
	                    <button onclick="submitOTP()">Verify OTP</button>
	                </div>

	                <div id="statusDiv" class="status" style="display:none;"></div>

	                <div id="roomsSection" class="rooms-section">
	                    <h3>Rooms</h3>
	                    <div class="room-list" id="roomList"></div>
	                </div>

	                <div style="margin-top: auto;">
	                    <div id="userInfoDiv" class="user-info"></div>
	                    <button class="logout-btn" onclick="logout()">Logout</button>
	                </div>
	            </div>

	            <div class="main-content">
	                <div class="header">
	                    <h2 id="currentRoom">Connecting...</h2>
	                </div>
	                <div class="messages" id="messagesList"></div>
	                <div class="input-area" id="inputArea">
	                    <input type="text" id="messageInput" placeholder="Type a message..." onkeypress="handleKeyPress(event)" />
	                    <button onclick="sendMessage()">Send</button>
	                </div>
	            </div>
	        </div>
	    </div>

	    <script>
	        let ws = null;
	        let username = '';
	        let currentRoom = null;
	        let sessionKey = null;
	        let isAuthenticated = false;

	        function showStatus(message, type = 'info') {
	            const statusDiv = document.getElementById('statusDiv');
	            statusDiv.textContent = message;
	            statusDiv.className = 'status ' + type;
	            statusDiv.style.display = 'block';
	            if (type !== 'error') {
	                setTimeout(() => statusDiv.style.display = 'none', 3000);
	            }
	        }

	        function login() {
	            const usernameInput = document.getElementById('username').value;
	            const passwordInput = document.getElementById('password').value;

	            if (!usernameInput || !passwordInput) {
	                document.getElementById('loginError').textContent = 'Please enter username and password';
	                return;
	            }

	            username = usernameInput;
	            connectWebSocket();

	            setTimeout(() => {
	                if (ws && ws.readyState === WebSocket.OPEN) {
	                    const msg = {
	                        type: 'auth',
	                        username: usernameInput,
	                        password: passwordInput
	                    };
	                    ws.send(JSON.stringify(msg));
	                }
	            }, 100);
	        }

	        function connectWebSocket() {
	            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
	            const url = protocol + '//' + window.location.host + '/ws';

	            ws = new WebSocket(url);

	            ws.onopen = () => {
	                console.log('Connected to server');
	            };

	            ws.onmessage = (event) => {
	                handleMessage(JSON.parse(event.data));
	            };

	            ws.onerror = (error) => {
	                document.getElementById('loginError').textContent = 'Connection error: ' + error;
	            };

	            ws.onclose = () => {
	                if (isAuthenticated) {
	                    showStatus('Disconnected from server', 'error');
	                }
	            };
	        }

	        function handleMessage(msg) {
	            if (msg.type === 'auth_challenge') {
	                // Show OTP section
	                document.getElementById('otpSection').style.display = 'flex';
	                showStatus('Enter the 6-digit OTP sent to the server console', 'success');
	            } else if (msg.type === 'auth_success') {
	                isAuthenticated = true;
	                sessionKey = msg.sessionKey;
	                document.getElementById('loginContainer').style.display = 'none';
	                document.getElementById('chatContainer').style.display = '';
	                document.getElementById('userInfoDiv').textContent = 'Logged in as: ' + username;
	                loadRooms();
	                showStatus('Authenticated! Select a room', 'success');
	            } else if (msg.type === 'auth_error') {
	                document.getElementById('loginError').textContent = 'Authentication failed: ' + msg.error;
	                ws.close();
	            } else if (msg.type === 'rooms_list') {
	                displayRooms(msg.rooms);
	            } else if (msg.type === 'join_success') {
	                currentRoom = msg.room;
	                document.getElementById('currentRoom').textContent = '#' + msg.room;
	                document.getElementById('inputArea').classList.add('active');
	                document.getElementById('messagesList').innerHTML = '';
	                showStatus('Joined room: ' + msg.room, 'success');
	                updateActiveRoom(msg.room);
	            } else if (msg.type === 'join_error') {
	                showStatus('Cannot join room: ' + msg.error, 'error');
	            } else if (msg.type === 'message') {
	                addMessage(msg.username, msg.content, false);
	            } else if (msg.type === 'system') {
	                addMessage('SYSTEM', msg.content, true);
	            } else if (msg.type === 'error') {
	                showStatus(msg.error, 'error');
	            }
	        }

	        function submitOTP() {
	            const otpInput = document.getElementById('otpInput').value;
	            if (otpInput.length !== 6 || isNaN(otpInput)) {
	                showStatus('OTP must be 6 digits', 'error');
	                return;
	            }

	            const msg = {
	                type: 'otp',
	                otp: otpInput
	            };
	            ws.send(JSON.stringify(msg));
	        }

	        function loadRooms() {
	            const msg = { type: 'list_rooms' };
	            ws.send(JSON.stringify(msg));
	        }

	        function displayRooms(rooms) {
	            const roomList = document.getElementById('roomList');
	            roomList.innerHTML = '';
	            if (rooms && rooms.length > 0) {
	                rooms.forEach(room => {
	                    const btn = document.createElement('button');
	                    btn.className = 'room-btn';
	                    btn.textContent = '# ' + room;
	                    btn.onclick = () => joinRoom(room);
	                    roomList.appendChild(btn);
	                });
	                document.getElementById('roomsSection').style.display = 'flex';
	            }
	        }

	        function updateActiveRoom(roomName) {
	            document.querySelectorAll('.room-btn').forEach(btn => {
	                if (btn.textContent === '# ' + roomName) {
	                    btn.classList.add('active');
	                } else {
	                    btn.classList.remove('active');
	                }
	            });
	        }

	        function joinRoom(roomName) {
	            const msg = {
	                type: 'join',
	                room: roomName
	            };
	            ws.send(JSON.stringify(msg));
	        }

	        function sendMessage() {
	            const input = document.getElementById('messageInput');
	            const text = input.value.trim();

	            if (!text || !currentRoom) {
	                return;
	            }

	            const msg = {
	                type: 'message',
	                content: text,
	                room: currentRoom
	            };
	            ws.send(JSON.stringify(msg));
	            input.value = '';
	        }

	        function handleKeyPress(event) {
	            if (event.key === 'Enter' && !event.shiftKey) {
	                event.preventDefault();
	                sendMessage();
	            }
	        }

	        function addMessage(user, text, isSystem) {
	            const messagesList = document.getElementById('messagesList');
	            const messageDiv = document.createElement('div');
	            messageDiv.className = 'message' + (isSystem ? ' system-message' : '');

	            const now = new Date();
	            const time = now.getHours().toString().padStart(2, '0') + ':' +
	                        now.getMinutes().toString().padStart(2, '0');

	            messageDiv.innerHTML = '<div class="time">' + time + '</div>' +
	                                  '<span class="user">' + user + ':</span>' +
	                                  '<span class="text">' + escapeHtml(text) + '</span>';

	            messagesList.appendChild(messageDiv);
	            messagesList.scrollTop = messagesList.scrollHeight;
	        }

	        function escapeHtml(text) {
	            const map = {
	                '&': '&amp;',
	                '<': '&lt;',
	                '>': '&gt;',
	                '"': '&quot;',
	                "'": '&#039;'
	            };
	            return text.replace(/[&<>"']/g, m => map[m]);
	        }

	        function logout() {
	            if (ws) {
	                ws.close();
	            }
	            isAuthenticated = false;
	            document.getElementById('chatContainer').style.display = 'none';
	            document.getElementById('loginContainer').style.display = '';
	            document.getElementById('username').value = '';
	            document.getElementById('password').value = '';
	            document.getElementById('loginError').textContent = '';
	        }
	    </script>
	</body>
	</html>
	    `
	*/
}

func handleWebSocket(w http.ResponseWriter, r *http.Request, server *SecureServer) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &WebSocketClient{
		ws:     ws,
		server: server,
		done:   make(chan bool),
	}

	// Handle WebSocket messages
	go client.handleWebSocketMessages()
}

func (wsc *WebSocketClient) handleWebSocketMessages() {
	defer wsc.ws.Close()

	for {
		var msg WebMessage
		if err := wsc.ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "auth":
			wsc.handleAuthentication(msg.Username, msg.Password)
		case "otp":
			wsc.handleOTP(msg.OTP)
		case "list":
			if !wsc.isAuthed || wsc.tcpConn == nil {
				wsc.sendError("Not authenticated")
				return
			}
			listMsg := &Message{Type: "list"}
			listJSON, _ := json.Marshal(listMsg)
			encryptedList, err := encryptAES(wsc.sessionKey, listJSON)
			if err != nil {
				wsc.sendError("Encryption failed: " + err.Error())
				return
			}
			listHMAC := computeHMAC(encryptedList, wsc.encryptor.hmacKey)
			fullMsg := append(encryptedList, listHMAC...)
			listEncoded := base64.StdEncoding.EncodeToString(fullMsg)
			wsc.tcpWriter.WriteString(listEncoded + "\n")
			wsc.tcpWriter.Flush()
		case "join":
			wsc.handleJoinRoom(msg.Room)
		case "message":
			wsc.handleSendMessage(msg.Content, msg.Room)
		}
	}
}

func (wsc *WebSocketClient) handleAuthentication(username, password string) {
	// Connect to TCP server for authentication
	tcpConn, err := net.Dial("tcp", TCP_SERVER_ADDR)
	if err != nil {
		wsc.sendError("Failed to connect to chat server")
		return
	}

	wsc.tcpConn = tcpConn
	wsc.tcpReader = bufio.NewReader(tcpConn)
	wsc.tcpWriter = bufio.NewWriter(tcpConn)
	wsc.username = username

	// Perform key exchange using line-based protocol
	// Step 1: Read server's public key (JSON message, base64 encoded, with newline)
	serverKeyLine, err := wsc.tcpReader.ReadString('\n')
	if err != nil {
		wsc.sendError("Failed to read server public key")
		tcpConn.Close()
		return
	}

	// Decode the server's key exchange message
	serverKeyLine = strings.TrimSpace(serverKeyLine)
	keyExchangeData, err := base64.StdEncoding.DecodeString(serverKeyLine)
	if err != nil {
		wsc.sendError("Invalid server key format")
		tcpConn.Close()
		return
	}

	keyExchangeMsg := &Message{}
	err = json.Unmarshal(keyExchangeData, keyExchangeMsg)
	if err != nil {
		wsc.sendError("Invalid key exchange message")
		tcpConn.Close()
		return
	}

	// Extract server's public key
	serverPublic, err := base64.StdEncoding.DecodeString(keyExchangeMsg.Content)
	if err != nil {
		wsc.sendError("Invalid server public key encoding")
		tcpConn.Close()
		return
	}

	// Generate client's DH parameters and send public key
	clientPrivate := make([]byte, 32)
	_, err = rand.Read(clientPrivate)
	if err != nil {
		wsc.sendError("Failed to generate client key")
		tcpConn.Close()
		return
	}

	clientPublic := derivePublic(clientPrivate)
	clientPublicB64 := base64.StdEncoding.EncodeToString(clientPublic)

	// Send client public key as JSON message
	clientKeyMsg := Message{
		Type:    "key_exchange",
		Content: clientPublicB64,
	}
	clientKeyJSON, _ := json.Marshal(clientKeyMsg)
	clientKeyEncoded := base64.StdEncoding.EncodeToString(clientKeyJSON)

	wsc.tcpWriter.WriteString(clientKeyEncoded + "\n")
	wsc.tcpWriter.Flush()

	// Derive shared secret (32 bytes from SHA256)
	sharedSecret := deriveSharedSecret(clientPrivate, serverPublic)
	wsc.sessionKey = sharedSecret

	wsc.encryptor = &ClientEncryptor{
		sessionKey: wsc.sessionKey,
		hmacKey:    wsc.sessionKey, // Use same key for HMAC as server does
	}

	// Send encrypted auth request
	authMsg := &Message{
		Type:     "auth",
		Username: username,
		Content:  password,
	}
	authJSON, _ := json.Marshal(authMsg)
	encryptedAuth, err := encryptAES(wsc.sessionKey, authJSON)
	if err != nil {
		wsc.sendError("Encryption failed: " + err.Error())
		tcpConn.Close()
		return
	}

	authHMAC := computeHMAC(encryptedAuth, wsc.sessionKey)
	fullMsg := append(encryptedAuth, authHMAC...)
	authEncoded := base64.StdEncoding.EncodeToString(fullMsg)

	wsc.tcpWriter.WriteString(authEncoded + "\n")
	wsc.tcpWriter.Flush()

	// Wait for OTP challenge (encrypted response)
	responseLine, err := wsc.tcpReader.ReadString('\n')
	if err != nil {
		wsc.sendError("No response from server")
		tcpConn.Close()
		return
	}

	responseLine = strings.TrimSpace(responseLine)
	responseData, err := base64.StdEncoding.DecodeString(responseLine)
	if err != nil {
		wsc.sendError("Invalid response format")
		tcpConn.Close()
		return
	}

	// Extract HMAC and verify
	if len(responseData) < 32 {
		wsc.sendError("Invalid response length")
		tcpConn.Close()
		return
	}

	messageData := responseData[:len(responseData)-32]
	signature := responseData[len(responseData)-32:]

	if !verifyHMAC(messageData, signature, wsc.sessionKey) {
		wsc.sendError("HMAC verification failed")
		tcpConn.Close()
		return
	}

	// Decrypt the response
	plaintext, err := decryptAES(wsc.sessionKey, messageData)
	if err != nil {
		wsc.sendError("Failed to decrypt response: " + err.Error())
		tcpConn.Close()
		return
	}

	responseMsg := &Message{}
	err = json.Unmarshal(plaintext, responseMsg)
	if err != nil {
		wsc.sendError("Failed to parse response")
		tcpConn.Close()
		return
	}

	if responseMsg.Type == "otp_required" {
		wsc.ws.WriteJSON(WebMessage{Type: "otp_required", Content: "Enter your 6-digit OTP"})
		// Listen for OTP submission
	} else {
		wsc.sendError("Authentication failed: " + responseMsg.Content)
		tcpConn.Close()
	}
}

func (wsc *WebSocketClient) handleOTP(otp string) {
	if wsc.tcpConn == nil {
		wsc.sendError("Not authenticated")
		return
	}

	otpMsg := &Message{
		Type:    "otp",
		Content: otp,
	}
	otpJSON, _ := json.Marshal(otpMsg)
	encryptedOTP, err := encryptAES(wsc.sessionKey, otpJSON)
	if err != nil {
		wsc.sendError("Encryption failed: " + err.Error())
		return
	}

	otpHMAC := computeHMAC(encryptedOTP, wsc.encryptor.hmacKey)
	fullMsg := append(encryptedOTP, otpHMAC...)
	otpEncoded := base64.StdEncoding.EncodeToString(fullMsg)

	wsc.tcpWriter.WriteString(otpEncoded + "\n")
	wsc.tcpWriter.Flush()

	// Read response
	responseLine, err := wsc.tcpReader.ReadString('\n')
	if err != nil {
		wsc.sendError("No response from server")
		return
	}

	responseLine = strings.TrimSpace(responseLine)
	responseData, err := base64.StdEncoding.DecodeString(responseLine)
	if err != nil {
		wsc.sendError("Invalid response format")
		return
	}

	// Extract HMAC and verify
	if len(responseData) < 32 {
		wsc.sendError("Invalid response length")
		return
	}

	messageData := responseData[:len(responseData)-32]
	signature := responseData[len(responseData)-32:]

	if !verifyHMAC(messageData, signature, wsc.sessionKey) {
		wsc.sendError("HMAC verification failed on OTP response")
		return
	}

	// Decrypt response
	decrypted, err := decryptAES(wsc.sessionKey, messageData)
	if err != nil {
		wsc.sendError("Failed to decrypt response")
		return
	}

	responseMsg := &Message{}
	err = json.Unmarshal(decrypted, responseMsg)
	if err != nil {
		wsc.sendError("Failed to parse response")
		return
	}

	if responseMsg.Type == "auth_success" {
		wsc.isAuthed = true
		wsc.ws.WriteJSON(map[string]interface{}{
			"type":       "auth_success",
			"sessionKey": base64.StdEncoding.EncodeToString(wsc.sessionKey),
		})

		// Start listening for messages
		go wsc.listenForServerMessages()
	} else {
		wsc.sendError("OTP verification failed: " + responseMsg.Content)
		wsc.tcpConn.Close()
		wsc.tcpConn = nil
	}
}

func (wsc *WebSocketClient) sendRoomsList() {
	// Get list of available rooms
	roomList := []string{"General", "VIP", "Admin"}

	// Send rooms properly
	response := map[string]interface{}{
		"type":  "rooms_list",
		"rooms": roomList,
	}
	wsc.ws.WriteJSON(response)
}

func (wsc *WebSocketClient) handleJoinRoom(roomName string) {
	if !wsc.isAuthed || wsc.tcpConn == nil {
		wsc.sendError("Not authenticated")
		return
	}

	log.Printf("[WEB] %s joining room: %s", wsc.username, roomName)

	joinMsg := &Message{
		Type: "join",
		Room: roomName,
	}
	joinJSON, _ := json.Marshal(joinMsg)
	encryptedJoin, err := encryptAES(wsc.sessionKey, joinJSON)
	if err != nil {
		wsc.sendError("Encryption failed: " + err.Error())
		return
	}

	joinHMAC := computeHMAC(encryptedJoin, wsc.encryptor.hmacKey)
	fullMsg := append(encryptedJoin, joinHMAC...)
	joinEncoded := base64.StdEncoding.EncodeToString(fullMsg)

	_, err = wsc.tcpWriter.WriteString(joinEncoded + "\n")
	if err != nil {
		wsc.sendError("Failed to send join request")
		return
	}
	err = wsc.tcpWriter.Flush()
	if err != nil {
		wsc.sendError("Failed to send join request")
		return
	}
	log.Printf("[WEB] Join request sent for %s to room %s", wsc.username, roomName)
	// Response will come through listenForServerMessages
}

func (wsc *WebSocketClient) handleSendMessage(content, room string) {
	if !wsc.isAuthed || wsc.tcpConn == nil {
		wsc.sendError("Not authenticated")
		return
	}

	log.Printf("[WEB] %s sending message to room %s: %s", wsc.username, room, content)

	sendMsg := &Message{
		Type:    "chat",
		Content: content,
		Room:    room,
	}
	sendJSON, _ := json.Marshal(sendMsg)
	encryptedSend, err := encryptAES(wsc.sessionKey, sendJSON)
	if err != nil {
		wsc.sendError("Encryption failed: " + err.Error())
		return
	}

	sendHMAC := computeHMAC(encryptedSend, wsc.encryptor.hmacKey)
	fullMsg := append(encryptedSend, sendHMAC...)
	sendEncoded := base64.StdEncoding.EncodeToString(fullMsg)

	wsc.tcpWriter.WriteString(sendEncoded + "\n")
	err = wsc.tcpWriter.Flush()
	if err != nil {
		log.Printf("[WEB] Error flushing message for %s: %v", wsc.username, err)
		wsc.sendError("Failed to send message")
		return
	}
	log.Printf("[WEB] Message sent successfully from %s", wsc.username)
}

func (wsc *WebSocketClient) listenForServerMessages() {
	for {
		line, err := wsc.tcpReader.ReadString('\n')
		if err != nil {
			log.Printf("[WEB] Listen error for %s: %v", wsc.username, err)
			break
		}

		line = strings.TrimSpace(line)
		messageData, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			log.Printf("[WEB] Failed to decode message for %s: %v", wsc.username, err)
			continue
		}

		// Verify HMAC and decrypt
		if len(messageData) > 32 {
			encryptedPart := messageData[:len(messageData)-32]
			signature := messageData[len(messageData)-32:]

			// Verify HMAC
			if !verifyHMAC(encryptedPart, signature, wsc.sessionKey) {
				log.Printf("[WEB] HMAC verification failed for %s", wsc.username)
				continue
			}

			decrypted, err := decryptAES(wsc.sessionKey, encryptedPart)
			if err != nil {
				log.Printf("[WEB] Decryption failed for %s: %v", wsc.username, err)
				continue
			}

			msg := &Message{}
			err = json.Unmarshal(decrypted, msg)
			if err != nil {
				log.Printf("[WEB] Failed to unmarshal message for %s: %v", wsc.username, err)
				continue
			}

			// Send to WebSocket client
			if msg.Type == "chat" {
				log.Printf("[WEB] Sending chat message to %s: %s", wsc.username, msg.Content)
				wsc.ws.WriteJSON(map[string]interface{}{
					"type":     "chat",
					"username": msg.Username,
					"content":  msg.Content,
					"room":     msg.Room,
				})
			} else if msg.Type == "system" {
				log.Printf("[WEB] Sending system message to %s: %s", wsc.username, msg.Content)
				wsc.ws.WriteJSON(map[string]interface{}{
					"type":    "system",
					"content": msg.Content,
				})
			} else if msg.Type == "join_success" {
				log.Printf("[WEB] Sending join_success to %s for room %s", wsc.username, msg.Room)
				wsc.ws.WriteJSON(map[string]interface{}{
					"type": "join_success",
					"room": msg.Room,
				})
			} else if msg.Type == "rooms_list" {
				log.Printf("[WEB] Sending rooms_list to %s", wsc.username)
				wsc.ws.WriteJSON(map[string]interface{}{
					"type":  "rooms_list",
					"rooms": msg.Rooms,
				})
			}
		}
	}
}

func (wsc *WebSocketClient) sendError(msg string) {
	wsc.ws.WriteJSON(WebMessage{
		Type:  "error",
		Error: msg,
	})
}

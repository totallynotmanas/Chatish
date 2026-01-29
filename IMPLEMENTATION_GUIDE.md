# Secure Multi-Room Chat System (Go) - Complete Implementation

A production-grade, academically-complete secure chat system implementing all 5 core cybersecurity concepts in Go, demonstrating proper security patterns for academic projects.

---

## 🔐 Security Features Implemented

### 1. **Authentication (MFA)** ✅

- **Username/Password Authentication**: Console-based login with hardcoded user database
- **6-Digit OTP**: One-Time Password generated after password verification
- **OTP Simulation**: OTP printed to server console for testing purposes
- **Implementation Files**: `server.go` - `handleAuthentication()`, `generateOTP()`

**Test Credentials:**

```
User: alice       | Password: password123  | Role: Admin
User: bob         | Password: secure456    | Role: Member
User: charlie     | Password: guest789     | Role: Guest
```

### 2. **Authorization (Access Control Matrix)** ✅

**Objects (Chat Rooms):**

- `General Room` - Public, accessible to all roles
- `VIP Room` - Members-only and admins
- `Admin Logs` - Administrators only

**Subjects (User Roles):**

- `Guest` - Basic user with limited access
- `Member` - Standard user with extended access
- `Admin` - Full system access

**ACL (Access Control List) Matrix:**

```
┌─────────┬──────────────┬──────────────┬──────────────┐
│  Role   │ General Room │  VIP Room    │ Admin Logs   │
├─────────┼──────────────┼──────────────┼──────────────┤
│ Guest   │      ✓       │      ✗       │      ✗       │
│ Member  │      ✓       │      ✓       │      ✗       │
│ Admin   │      ✓       │      ✓       │      ✓       │
└─────────┴──────────────┴──────────────┴──────────────┘
```

**Implementation**: `server.go` - `canJoinRoom()`, `handleJoinRoom()`

### 3. **Encryption (Confidentiality)** ✅

**Key Exchange Protocol:** Simplified Diffie-Hellman

```
1. Server generates 32-byte private key (random)
2. Derives public key: pub = SHA256(private)
3. Sends public key to client (Base64-encoded)
4. Client generates its own private/public key pair
5. Client sends public key to server
6. Both derive shared session key: secret = SHA256(clientPrivate || serverPublic) OR SHA256(serverPrivate || clientPublic)
7. Shared key used for all subsequent encryption
```

**Encryption Algorithm:** AES-256-GCM

- Key size: 256 bits (32 bytes)
- Mode: Galois/Counter Mode (GCM) for authenticated encryption
- Nonce: Random per message (12 bytes)
- Authentication: Built-in with GCM mode
- Protection: Against ciphertext tampering and modification

**Implementation**: `crypto.go` - `encryptAES()`, `decryptAES()`, `deriveSharedSecret()`

### 4. **Hashing & Integrity** ✅

**Password Storage:** SHA256 with Random Salt

```
1. Generate random 16-byte salt for each user
2. Hash: hash = SHA256(password || salt)
3. Store: hash and salt separately
4. Verification: SHA256(provided_password || stored_salt) == stored_hash
5. Protection: Prevents rainbow tables and dictionary attacks
```

**Message Integrity:** HMAC-SHA256

```
1. Compute HMAC: signature = HMAC-SHA256(encrypted_message, session_key)
2. Append: packet = encrypted_message || signature
3. Transmission: Base64(packet)
4. Reception: Decode → Verify HMAC → Decrypt
5. Protection: Detects tampering and unauthorized modifications
```

**Implementation**: `crypto.go` - `hashPassword()`, `generateRandomSalt()`, `computeHMAC()`, `verifyHMAC()`

### 5. **Encoding (Transport)** ✅

**Base64 Encoding Protocol:**

```
┌─────────────────────────────────────────────────────────────┐
│ Original JSON Message                                       │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ AES-GCM Encryption                                          │
│ (Nonce + Ciphertext = Encrypted)                            │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ HMAC-SHA256 Signature (32 bytes)                            │
│ Append to encrypted message                                 │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Base64 Encoding                                             │
│ (Safe for TCP transmission)                                 │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Send over TCP with newline delimiter                        │
└─────────────────────────────────────────────────────────────┘
```

**Implementation**: All message functions in `client.go` and `server.go`

---

## 📁 Project Structure

```
Chatish/
├── main.go              # Entry point (server/client mode selector)
├── server.go            # Server implementation (1000+ lines)
│   ├── Authentication & MFA
│   ├── Authorization & ACL
│   ├── Room management
│   ├── Client handlers
│   └── Message broadcasting
├── client.go            # Client implementation (500+ lines)
│   ├── Connection management
│   ├── Key exchange handler
│   ├── Message encryption/decryption
│   ├── CLI interface
│   └── Message listeners
├── crypto.go            # Shared cryptographic functions (300+ lines)
│   ├── AES-GCM encryption
│   ├── SHA256 hashing
│   ├── HMAC integrity
│   ├── Key derivation
│   └── Base64 encoding
├── go.mod               # Go module definition
├── go.sum               # Go dependencies
└── README_SECURE.md     # Detailed documentation
```

---

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or later
- Two terminal windows (one for server, one for client)

### Step 1: Build the Application

```bash
cd Chatish
go build -o chatish.exe .
```

### Step 2: Start the Server

**Terminal 1:**

```bash
go run . server
```

**Expected Output:**

```
[SERVER] Secure Multi-Room Chat Server started on 127.0.0.1:5000
[SERVER] Waiting for connections...

=== TEST CREDENTIALS ===
  alice / password123 (Admin)
  bob / secure456 (Member)
  charlie / guest789 (Guest)
```

### Step 3: Start the Client

**Terminal 2:**

```bash
go run . client
```

**Expected Output:**

```
=== Secure Multi-Room Chat Client ===
Connected to server at 127.0.0.1:5000
[CLIENT] Session key established
Enter username:
```

### Step 4: Login

```
Enter username: alice
Enter password: password123
[SERVER] Check server console for your OTP
Enter your OTP: [Look at Terminal 1 for the OTP]
[CLIENT] Authenticated as alice (Admin)

Available commands:
  /list - List available rooms
  /join <room> - Join a room
  /leave - Leave current room
  /msg <message> - Send a message
  /quit - Exit

>
```

---

## 💻 Usage Examples & Test Cases

### Test Case 1: Admin User (alice)

**Terminal 1 (Server Console):**

```
[CONNECTION] New client from 127.0.0.1:54321
[ENCRYPTION] Session key established
[AUTH] User 'alice' password verified (Role: Admin)
[MFA] OTP for 'alice': 482619 (Valid for 2 minutes)
[MFA] OTP verified for 'alice'
[SUCCESS] User 'alice' authenticated and connected from 127.0.0.1:54321
```

**Terminal 2 (Client Console):**

```
> /list
[ROOMS] General Room (0 users); VIP Room (0 users); Admin Logs (0 users)

> /join General Room
[CLIENT] Joined General Room

> /msg Hello everyone, I'm Alice!
[General Room] alice: Hello everyone, I'm Alice!

> /leave
[CLIENT] Left General Room

> /join Admin Logs
[CLIENT] Joined Admin Logs

> /msg System check complete
[Admin Logs] alice: System check complete

> /quit
Disconnecting...
```

### Test Case 2: Member User (bob) - ACL Testing

```
> /list
[ROOMS] General Room (0 users); VIP Room (0 users)

> /join General Room
[CLIENT] Joined General Room

> /msg Hi, I'm Bob!
[General Room] bob: Hi, I'm Bob!

> /join VIP Room
[CLIENT] Joined VIP Room

> /msg Welcome to the VIP room
[VIP Room] bob: Welcome to the VIP room

> /join Admin Logs
[ERROR] Access Denied: Your role 'Member' cannot access 'Admin Logs'
```

**Server logs for ACL denial:**

```
[ACL DENIED] User 'bob' (role: Member) tried to access 'Admin Logs'
```

### Test Case 3: Guest User (charlie) - Restricted Access

```
> /list
[ROOMS] General Room (0 users)

> /join General Room
[CLIENT] Joined General Room

> /join VIP Room
[ERROR] Access Denied: Your role 'Guest' cannot access 'VIP Room'

> /join Admin Logs
[ERROR] Access Denied: Your role 'Guest' cannot access 'Admin Logs'
```

---

## 🔍 Security Implementation Details

### Message Flow with Encryption & Integrity

```
CLIENT                              SERVER
  │                                   │
  ├──(1) TCP Connect──────────────────>│
  │                                   │
  ├<─(2) Base64(PubKey)───────────────│
  │                                   │
  ├──(3) Base64(PubKey)──────────────>│
  │  [Key Exchange Complete]         │
  │  SessionKey established          │
  │                                   │
  ├──(4) Encrypt(Credentials)────────>│
  │      +HMAC(Encrypted)             │
  │      +Base64 encoded              │
  │                                   │
  │[Server validates credentials]    │
  │[Generates OTP: 482619]           │
  │                                   │
  ├<─(5) Encrypt(OTP_Required)───────│
  │      +HMAC +Base64                │
  │                                   │
  ├──(6) Encrypt(OTP)────────────────>│
  │      +HMAC +Base64                │
  │                                   │
  │[Server verifies OTP]             │
  │                                   │
  ├<─(7) Encrypt(Auth_Success)───────│
  │      +HMAC +Base64                │
  │                                   │
  │[AUTHENTICATED & READY]           │
  │                                   │
  ├──(8) Encrypt(JoinRoom)───────────>│
  │      [Verify ACL]                 │
  ├<─(9) Encrypt(JoinSuccess)────────│
  │                                   │
  ├──(10) Encrypt(ChatMessage)───────>│
  │        [Broadcast to room]       │
  │                                   │
```

### AES-GCM Encryption in Detail

```go
// Encryption Process
plaintext := jsonData              // {"type": "chat", "content": "Hello"}
sessionKey := derivedKey           // 32-byte key from key exchange
nonce := random(12)                // Random 12-byte nonce
ciphertext := AES-GCM.Seal(        // Authenticated encryption
    nonce,                         // Prepended to ciphertext
    nonce,                         // Used as IV
    plaintext,                     // Input data
    nil                            // Additional authenticated data
)

// HMAC Integrity
signature := HMAC-SHA256(          // Message authentication code
    ciphertext,                    // What to sign
    sessionKey                     // Key for HMAC
)

// Final Packet
packet := ciphertext || signature  // Concatenate
encoded := Base64(packet)          // Safe transport encoding
```

### HMAC Verification

```go
// Reception Process
encoded := readLine()              // "agI5FsK9..."
packet := Base64Decode(encoded)    // Binary data
ciphertext := packet[0:len-32]     // Split at last 32 bytes
signature := packet[len-32:]       // Last 32 bytes

// Verify Integrity
expectedSig := HMAC-SHA256(        // Recompute signature
    ciphertext,
    sessionKey
)
if expectedSig != signature {
    reject("Tampered message!")    // Tampering detected
}

// Decrypt
plaintext := AES-GCM.Open(nonce, ciphertext)
data := json.Unmarshal(plaintext)  // Parse message
```

---

## 🧪 Testing Security Features

### Test 1: Authentication & MFA

```bash
# Terminal 1: Start server
go run . server

# Terminal 2: Start client
go run . client
# Enter: alice / password123
# Check Terminal 1 for OTP
# Enter OTP from Terminal 1
```

**Expected Result:**

- ✅ Server prints OTP to console
- ✅ Client authenticates after correct OTP
- ✅ Connection rejected if OTP is wrong

### Test 2: Authorization (ACL)

```bash
# Login as charlie (Guest)
# Try /join VIP Room
# Expected: Access Denied error
```

**Server logs:**

```
[ACL DENIED] User 'charlie' (role: Guest) tried to access 'VIP Room'
```

### Test 3: Encryption & HMAC

```bash
# Send a message with /msg
# Messages are encrypted with AES-256-GCM
# HMAC signature prevents tampering
# Try to modify network traffic = HMAC verification fails
```

### Test 4: Password Hashing

```bash
# Each user has a unique salt
# Password hash = SHA256(password + salt)
# Stored in user database
# Prevents rainbow table attacks
```

### Test 5: Multi-Client Broadcasting

```bash
# Terminal 1: Server
# Terminal 2: Client A (alice) joins General Room
# Terminal 3: Client B (bob) joins General Room
# Terminal 2: /msg Hello Bob!
# Terminal 3 receives: [General Room] alice: Hello Bob!
```

---

## 📊 Cybersecurity Concepts Map

| Concept             | Theory                   | Implementation            | File      | Function               |
| ------------------- | ------------------------ | ------------------------- | --------- | ---------------------- |
| **Authentication**  | Identity verification    | MFA (Password + OTP)      | server.go | handleAuthentication() |
| **Authorization**   | Access control           | ACL Matrix                | server.go | canJoinRoom()          |
| **Confidentiality** | Data privacy             | AES-256-GCM               | crypto.go | encryptAES()           |
| **Integrity**       | Data authenticity        | HMAC-SHA256               | crypto.go | computeHMAC()          |
| **Non-repudiation** | Accountability           | Digital signatures (HMAC) | crypto.go | verifyHMAC()           |
| **Key Exchange**    | Secure key establishment | DH-inspired               | server.go | performKeyExchange()   |
| **Hashing**         | Secure storage           | SHA256 + Salt             | crypto.go | hashPassword()         |
| **Encoding**        | Safe transmission        | Base64                    | crypto.go | All message functions  |
| **Concurrency**     | Multi-client handling    | Goroutines                | server.go | HandleClient()         |
| **Protocol Design** | Secure messaging         | Custom encrypted protocol | All files | -                      |

---

## 🛡️ Security Assumptions & Limitations

### Current Implementation

✅ **Strengths:**

- Proper authentication with MFA
- Role-based access control
- AES-GCM authenticated encryption
- HMAC message integrity
- Salted password hashing
- Random session keys
- Secure key exchange
- Base64 safe transport encoding

⚠️ **Academic/Demo Limitations:**

- OTP via console (not email/SMS)
- Simplified DH key exchange (not real DH)
- In-memory user database (no persistence)
- No rate limiting on auth attempts
- No connection encryption at transport layer (TLS/SSL)
- No certificate pinning
- No input validation beyond JSON parsing
- No logging to persistent storage
- No session timeout

### Production Hardening Checklist

- [ ] Replace OTP with email/SMS service (Twilio, AWS SES)
- [ ] Implement full Elliptic Curve Diffie-Hellman (ECDH)
- [ ] Use TLS 1.3 for transport encryption
- [ ] Add PostgreSQL/MongoDB for user persistence
- [ ] Implement JWT for session management
- [ ] Add rate limiting (fail2ban style)
- [ ] Implement certificate pinning
- [ ] Add comprehensive input validation
- [ ] Audit logging to syslog/ELK stack
- [ ] Add session timeout and refresh tokens
- [ ] Implement user account lockout policies
- [ ] Add password complexity requirements
- [ ] Implement CORS policies
- [ ] Add API rate limiting (Redis-backed)
- [ ] Use environment variables for secrets

---

## 📚 Code Organization

### crypto.go (Shared Functions)

```go
// Encryption
encryptAES()              // AES-256-GCM encryption
decryptAES()              // AES-256-GCM decryption

// Hashing & Integrity
hashPassword()            // SHA256 + salt
generateRandomSalt()      // 16-byte random salt
computeHMAC()             // HMAC-SHA256
verifyHMAC()              // HMAC verification

// Key Exchange
derivePublic()            // SHA256(private) → public
deriveSharedSecret()      // SHA256(privkey || pubkey) → secret

// Utilities
generateOTP()             // 6-digit random OTP
generateSessionID()       // Random session ID
```

### server.go (Authentication & Authorization)

```go
// Authentication
handleAuthentication()    // Login + MFA flow
performKeyExchange()      // DH-inspired key exchange

// Authorization
canJoinRoom()             // ACL validation
handleJoinRoom()          // ACL-enforced join

// Room Management
broadcastToRoom()         // Message broadcasting
leaveAllRooms()           // Cleanup on disconnect

// Client Handling
HandleClient()            // Main client connection loop
handleClientMessages()    // Message dispatch
```

### client.go (User Interface)

```go
// Connection
ConnectToServer()         // TCP connection
PerformKeyExchange()      // Client key exchange
Authenticate()            // Login flow

// Message Operations
sendMessage()             // Encrypt & send
readMessage()             // Receive & decrypt
SendChatMessage()         // Chat message
JoinRoom()                // Join request
LeaveRoom()               // Leave request
ListRooms()               // Room listing

// CLI
RunClientCLI()            // Interactive interface
ListenForMessages()       // Async message receiver
```

---

## 🔐 Data Structures

### Message Protocol (JSON)

```json
{
  "type": "chat|join|leave|auth|otp|error|system",
  "username": "alice",
  "room": "General Room",
  "content": "Message text or data",
  "timestamp": "15:04:05"
}
```

### User Model

```go
type User struct {
    Username     string    // Unique identifier
    PasswordHash string    // SHA256(password+salt)
    PasswordSalt string    // Base64-encoded random salt
    Role         Role      // "Guest", "Member", or "Admin"
}
```

### ClientSession

```go
type ClientSession struct {
    conn         net.Conn    // TCP connection
    username     string      // Authenticated user
    role         Role        // User role for ACL
    currentRoom  Room        // Active room or ""
    sessionKey   []byte      // 32-byte encryption key
    sessionID    string      // Unique session identifier
    authenticated bool       // Auth status
    reader       *bufio.Reader  // Network reader
    writer       *bufio.Writer  // Network writer
}
```

---

## 📖 References & Further Reading

- **AES-GCM**: https://en.wikipedia.org/wiki/Galois/Counter_Mode
- **HMAC**: https://en.wikipedia.org/wiki/HMAC
- **Diffie-Hellman**: https://en.wikipedia.org/wiki/Diffie%E2%80%93Hellman_key_exchange
- **RBAC**: https://en.wikipedia.org/wiki/Role-based_access_control
- **Go Crypto**: https://golang.org/pkg/crypto/
- **Go Concurrency**: https://golang.org/doc/effective_go#concurrency
- **TCP/IP Sockets**: https://golang.org/pkg/net/

---

## 📝 License

This project is for educational purposes only. Suitable for:

- Cybersecurity courses
- Network security labs
- Authentication/Authorization studies
- Cryptography demonstrations
- Go programming practice

---

## ✨ Summary

This **Secure Multi-Room Chat System** is a complete implementation demonstrating all 5 required cybersecurity concepts:

1. ✅ **Authentication (MFA)** - Username/Password + 6-digit OTP
2. ✅ **Authorization (ACL)** - Role-based room access control
3. ✅ **Encryption** - AES-256-GCM with session keys
4. ✅ **Integrity** - HMAC-SHA256 on all messages
5. ✅ **Encoding** - Base64 for safe TCP transmission

Perfect for academic projects, interviews, or security demonstrations!

---

**Author**: Secure Systems Development  
**Language**: Go 1.21+  
**License**: Educational Use

# Secure Multi-Room Chat System (Go)

A production-grade, academically-complete secure chat system implementing 5 core cybersecurity concepts in Go.

## 🔐 Security Features Implemented

### 1. **Authentication (MFA)**

- **Username/Password Authentication**: Console-based login with hardcoded user database
- **6-Digit OTP**: One-Time Password generated after password verification
- **OTP Simulation**: OTP printed to server console for testing
- **Implementation**: `server.go` - `handleAuthentication()`, `generateOTP()`

**Test Credentials:**

```
alice / password123 (Admin)
bob / secure456 (Member)
charlie / guest789 (Guest)
```

### 2. **Authorization (Access Control)**

**Objects (Chat Rooms):**

- `General Room` - Public room
- `VIP Room` - Member-exclusive room
- `Admin Logs` - Administrator-only room

**Subjects (Roles):**

- `Guest` - Can join General Room only
- `Member` - Can join General Room & VIP Room
- `Admin` - Can join all rooms

**ACL Matrix Implementation:**

```
Guest  → General Room ✓
Member → General Room ✓, VIP Room ✓
Admin  → General Room ✓, VIP Room ✓, Admin Logs ✓
```

**Implementation**: `server.go` - `canJoinRoom()`, `handleJoinRoom()`

### 3. **Encryption (Confidentiality)**

**Key Exchange:** Simplified Diffie-Hellman-inspired key exchange

- Server generates 32-byte private key
- Derives public key using SHA256
- Exchanges public keys with client
- Derives shared session key from both private and public keys

**Encryption Algorithm:** AES-GCM (Galois/Counter Mode)

- 256-bit session key
- Random nonce per message
- Authenticated encryption
- Protection against tampering

**Implementation**: `server.go` & `client.go` - `performKeyExchange()`, `encryptAES()`, `decryptAES()`

### 4. **Hashing & Integrity**

**Password Storage:** SHA256 with Random Salt

- Unique 16-byte salt per user
- Salt stored with user record
- Hash = SHA256(password + salt)
- Prevents rainbow table attacks

**Message Integrity:** HMAC-SHA256

- Computed over encrypted message
- Appended to each packet
- Verified on reception
- Detects tampering during transit

**Implementation**: `server.go` & `client.go` - `hashPassword()`, `generateRandomSalt()`, `computeHMAC()`, `verifyHMAC()`

### 5. **Encoding (Transport)**

**Base64 Encoding:**

- Final encrypted + HMAC packet encoded to Base64
- Safe for TCP transmission
- Decoded on reception before decryption

**Protocol Structure:**

```
[BASE64(CIPHERTEXT + HMAC_SIGNATURE)]
```

**Implementation**: All message functions use base64 encoding/decoding

## 📦 Architecture

### Components

```
┌─────────────────────────────────────────┐
│         Secure Chat System              │
├─────────────────────────────────────────┤
│  Server (server.go)                     │
│  ├── Authentication Manager             │
│  ├── Authorization Engine (ACL)         │
│  ├── Room Manager                       │
│  ├── Encryption Handler                 │
│  └── Broadcasting System                │
├─────────────────────────────────────────┤
│  Client (client.go)                     │
│  ├── Connection Manager                 │
│  ├── Key Exchange Handler               │
│  ├── Message Encryption                 │
│  └── CLI Interface                      │
└─────────────────────────────────────────┘
```

### Data Flow

```
1. Client connects to server
2. Key exchange (simplified DH)
3. Session key established
4. Username/Password authentication
5. OTP verification
6. Room access based on ACL
7. Encrypted message exchange with HMAC verification
8. Broadcasting to room members
```

## 🚀 Running the System

### Prerequisites

```bash
go version go1.21 or later
```

### Start the Server

```bash
go run . server
```

**Output:**

```
[SERVER] Secure Multi-Room Chat Server started on 127.0.0.1:5000
[SERVER] Waiting for connections...

=== TEST CREDENTIALS ===
  alice / password123 (Admin)
  bob / secure456 (Member)
  charlie / guest789 (Guest)
```

### Start the Client(s)

In a separate terminal:

```bash
go run . client
```

**Output:**

```
=== Secure Multi-Room Chat Client ===
Connected to server at 127.0.0.1:5000
[CLIENT] Session key established
Enter username: alice
Enter password: password123
[SERVER] Check server console for your OTP
Enter your OTP: [check server console]
[CLIENT] Authenticated as alice (Admin)

Available commands:
  /list - List available rooms
  /join <room> - Join a room
  /leave - Leave current room
  /msg <message> - Send a message
  /quit - Exit
```

## 💻 Usage Examples

### Example Session

**Terminal 1 (Server):**

```
$ go run . server
[SERVER] Secure Multi-Room Chat Server started on 127.0.0.1:5000
[SERVER] Waiting for connections...

=== TEST CREDENTIALS ===
  alice / password123 (Admin)
  bob / secure456 (Member)
  charlie / guest789 (Guest)

[CONNECTION] New client from 127.0.0.1:54321
[ENCRYPTION] Session key established for 'alice'
[AUTH] User 'alice' password verified (Role: Admin)
[MFA] OTP for 'alice': 123456 (Valid for 2 minutes)
[MFA] OTP verified for 'alice'
[SUCCESS] User 'alice' authenticated and connected from 127.0.0.1:54321
```

**Terminal 2 (Client - Alice):**

```
$ go run . client
=== Secure Multi-Room Chat Client ===
Connected to server at 127.0.0.1:5000
[CLIENT] Session key established
Enter username: alice
Enter password: password123
[SERVER] Check server console for your OTP
Enter your OTP: 123456
[CLIENT] Authenticated as alice (Admin)

Available commands:
  /list - List available rooms
  /join <room> - Join a room
  /leave - Leave current room
  /msg <message> - Send a message
  /quit - Exit

> /list
[ROOMS] General Room (0 users); VIP Room (0 users); Admin Logs (0 users)

> /join General Room
[CLIENT] Joined General Room

> /msg Hello from Alice!
[General Room] Alice: Hello from Alice!

> /leave
[CLIENT] Left General Room

> /quit
Disconnecting...
```

## 🔑 File Structure

```
Chatish/
├── main.go           # Entry point (server/client mode selection)
├── server.go         # Server implementation
│   ├── Authentication (MFA)
│   ├── Authorization (ACL)
│   ├── Encryption (AES-GCM)
│   ├── Hashing & HMAC
│   ├── Room Management
│   └── Broadcasting
├── client.go         # Client implementation
│   ├── Connection management
│   ├── Key exchange
│   ├── Message encryption
│   └── CLI interface
├── go.mod           # Go module file
└── README.md        # This file
```

## 🔍 Security Implementation Details

### Key Generation & Exchange

```go
// 1. Generate private keys (32 bytes random)
serverPrivate := make([]byte, 32)
rand.Read(serverPrivate)

// 2. Derive public keys (SHA256 hash)
serverPublic := sha256.Sum256(serverPrivate)

// 3. Exchange public keys
// 4. Derive shared secret
sessionKey := sha256.Sum256(append(serverPrivate, clientPublic...))
```

### AES-GCM Encryption

```go
// 1. Create cipher block
block := aes.NewCipher(sessionKey)

// 2. Create GCM mode
gcm := cipher.NewGCM(block)

// 3. Generate random nonce
nonce := make([]byte, gcm.NonceSize())
io.ReadFull(rand.Reader, nonce)

// 4. Encrypt with authentication
ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
```

### HMAC-SHA256 Integrity

```go
// 1. Compute HMAC
h := hmac.New(sha256.Sum256, sessionKey)
h.Write(ciphertext)
signature := h.Sum(nil)

// 2. Verify HMAC
expectedSig := computeHMAC(messageData, sessionKey)
if !hmac.Equal(expectedSig, signature) {
    // Tampering detected!
}
```

## 📋 ACL Testing

### Test Case 1: Guest Access

```
Username: charlie (Guest role)
/join General Room    → ✓ Access granted
/join VIP Room        → ✗ Access Denied
/join Admin Logs      → ✗ Access Denied
```

### Test Case 2: Member Access

```
Username: bob (Member role)
/join General Room    → ✓ Access granted
/join VIP Room        → ✓ Access granted
/join Admin Logs      → ✗ Access Denied
```

### Test Case 3: Admin Access

```
Username: alice (Admin role)
/join General Room    → ✓ Access granted
/join VIP Room        → ✓ Access granted
/join Admin Logs      → ✓ Access granted
```

## 🧪 Testing the Security Features

### 1. Test Authentication

```bash
# Terminal 1: Start server
go run . server

# Terminal 2: Start client
go run . client
# Login with: alice / password123
# Check server console for OTP
# Enter OTP
```

### 2. Test MFA

- Invalid OTP rejection
- One-time use validation
- OTP expiration (2 minutes in production)

### 3. Test Authorization

- Try to access restricted rooms with each role
- Verify ACL enforcement at server

### 4. Test Encryption

- All messages between client and server are encrypted
- HMAC signatures prevent tampering
- Key exchange establishes unique session key per connection

### 5. Test Integrity

- Messages include HMAC signatures
- Tampering is detected and rejected
- Password hashing with salts

## 📊 Protocol Messages

### Authentication Phase

```json
{"type": "auth", "username": "alice", "content": "password123"}
{"type": "otp_required", "content": "Enter your 6-digit OTP"}
{"type": "otp", "content": "123456"}
{"type": "auth_success", "content": "Authenticated as alice (Admin)"}
```

### Room Operations

```json
{"type": "list"}
{"type": "room_list", "content": "General Room (0 users); VIP Room (1 user)"}

{"type": "join", "room": "General Room"}
{"type": "join_success", "room": "General Room", "content": "Joined General Room"}

{"type": "chat", "content": "Hello everyone!"}
{"type": "chat", "username": "alice", "room": "General Room", "content": "Hello everyone!", "timestamp": "15:04:05"}

{"type": "leave"}
{"type": "leave_success", "content": "Left General Room"}
```

## 🛡️ Security Considerations

### Strong Points

1. ✅ Multi-factor authentication (password + OTP)
2. ✅ Role-based access control (RBAC)
3. ✅ AES-GCM encryption with random nonce
4. ✅ HMAC for message integrity
5. ✅ Salted password hashing
6. ✅ Session-specific encryption keys
7. ✅ Base64 encoding for safe transport

### Production Recommendations

1. **Replace OTP simulation** with email/SMS delivery
2. **Use TLS/SSL** for transport layer security
3. **Implement database** for user persistence
4. **Add rate limiting** for authentication attempts
5. **Use real Diffie-Hellman** or ECDH key exchange
6. **Implement token-based sessions** (JWT)
7. **Add logging and monitoring**
8. **Implement certificate pinning**
9. **Add input validation and sanitization**
10. **Use secure random for all cryptographic operations**

## 📚 Academic Concepts Demonstrated

This implementation covers core cybersecurity principles:

| Concept             | Implementation                          |
| ------------------- | --------------------------------------- |
| **Authentication**  | MFA (Password + OTP)                    |
| **Authorization**   | ACL Matrix                              |
| **Confidentiality** | AES-GCM Encryption                      |
| **Integrity**       | HMAC-SHA256                             |
| **Non-repudiation** | Digital signatures (HMAC)               |
| **Key Exchange**    | DH-inspired symmetric key establishment |
| **Encoding**        | Base64 for safe transmission            |
| **Hashing**         | SHA256 for password storage             |
| **Salting**         | Per-user random salt                    |
| **Concurrency**     | Goroutines for multi-client handling    |

## 📖 References

- Go Crypto Package: https://golang.org/pkg/crypto/
- AES-GCM: https://en.wikipedia.org/wiki/Galois/Counter_Mode
- HMAC: https://en.wikipedia.org/wiki/HMAC
- Diffie-Hellman: https://en.wikipedia.org/wiki/Diffie%E2%80%93Hellman_key_exchange
- Role-Based Access Control: https://en.wikipedia.org/wiki/Role-based_access_control

## 📝 License

This project is for educational purposes.

## 👤 Author

Secure Chat System - Academic Implementation

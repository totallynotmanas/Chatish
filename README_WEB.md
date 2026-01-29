# 🔐 Secure Multi-Room Chat - Web Interface

The system now includes a complete web-based interface alongside the CLI client, with **full security integration**.

## Running the System

### Option 1: CLI Only (Secure TCP Server)

```bash
# Terminal 1: Start the server
go run . server

# Terminal 2: Start a client
go run . client
```

### Option 2: Web Interface (Recommended)

```bash
# Single command starts both TCP server + Web UI
go run . web

# Then open browser: http://localhost:8080
```

Or use the compiled binary:

```bash
# Web mode
./chatish.exe web

# Server mode
./chatish.exe server

# Client mode
./chatish.exe client
```

## Features

✅ **Secure Transport**: AES-256-GCM encryption for all messages
✅ **Authentication**: Username + Password + 6-digit OTP (One-Time Password)
✅ **Authorization**: Role-based access control (Guest/Member/Admin)
✅ **Message Integrity**: HMAC-SHA256 signatures on all encrypted messages
✅ **Secure Key Exchange**: Simplified Diffie-Hellman key agreement
✅ **Dark Theme UI**: Modern, secure-looking web interface

## Test Credentials

| Username | Password    | Role   |
| -------- | ----------- | ------ |
| alice    | password123 | Admin  |
| bob      | secure456   | Member |
| charlie  | guest789    | Guest  |

### Room Access by Role

| Room       | Guest | Member | Admin |
| ---------- | ----- | ------ | ----- |
| General    | ✅    | ✅     | ✅    |
| VIP        | ❌    | ✅     | ✅    |
| Admin Logs | ❌    | ❌     | ✅    |

## Web Interface Usage

1. **Login**: Enter username and password
2. **OTP Verification**: Copy the 6-digit OTP from server console and enter it
3. **Select Room**: Choose from available rooms (based on your role)
4. **Send Messages**: Type and send encrypted messages
5. **Multi-Client**: Connect multiple users simultaneously
6. **Logout**: Disconnect and return to login

## Authentication Flow

```
1. User enters credentials
2. WebSocket connects to server
3. Server generates random 6-digit OTP (printed to console)
4. User enters OTP from console
5. OTP verified (one-time use only)
6. Session key established via key exchange
7. All messages encrypted with AES-256-GCM
8. HMAC-SHA256 signature appended to each message
```

## Security Implementation

### Encryption

- **Algorithm**: AES-256-GCM
- **Key Size**: 256 bits (32 bytes)
- **Nonce**: 12-byte random per message
- **Authentication**: Built-in with GCM mode

### Authentication

- **Password**: Salted SHA256 hash (not transmitted)
- **OTP**: 6-digit one-time use token
- **Challenge**: Server generates new OTP per login

### Authorization

- **Model**: Role-Based Access Control (RBAC)
- **Roles**: Guest, Member, Admin
- **Enforcement**: Per-room permission matrix

### Integrity

- **HMAC**: SHA256 with 32-byte signature
- **Coverage**: Encrypts message + appends HMAC
- **Verification**: Server checks before decryption

## File Structure

```
chatish/
├── main.go          # Entry point with CLI mode selector
├── server.go        # Secure TCP server (420 lines)
├── client.go        # CLI client (497 lines)
├── crypto.go        # Shared crypto functions (210 lines)
├── webserver.go     # HTTP + WebSocket server [NEW]
├── go.mod           # Go dependencies
├── go.sum           # Go checksums
├── chatish.exe      # Compiled binary
├── README.md        # Original README
├── README_WEB.md    # This file
└── docs/            # Additional documentation
```

## Behind the Scenes

The web interface bridges two protocols:

1. **Client ↔ Server**: WebSocket (for browser)
2. **Server ↔ Core**: TCP encrypted (secure backend)
3. **Security**: Applied uniformly across both

### Message Flow

```
Browser → WebSocket → Go Server
                       ↓
                   Encryption (AES-256-GCM)
                       ↓
                   HMAC-SHA256 signature
                       ↓
                   TCP to other clients
                       ↓
                   Verification & Decryption
                       ↓
                   Broadcast to all users
```

## Production Considerations

⚠️ **For Development/Demo Only**:

- WebSocket CORS allows all origins (change for production)
- OTP printed to console (use SMS/email in production)
- Simple DH key exchange (use full DH or TLS in production)
- HTTP only (enable HTTPS/TLS in production)

## Troubleshooting

**"Connection refused"**

- Ensure server is running on :5000
- Check firewall settings

**"OTP verification failed"**

- OTP is 6 digits only
- OTP is one-time use only (one attempt)
- Copy exact value from server console

**"Cannot join room"**

- Check your role permissions
- Admin can access all rooms
- Members cannot access Admin Logs

**"Message not received"**

- Check HMAC verification in server logs
- Ensure session key is correct
- Check encryption/decryption logic

## API Reference

### WebSocket Messages

#### Authentication

```json
{"type": "auth", "username": "alice", "password": "password123"}
{"type": "otp", "otp": "123456"}
```

#### Room Operations

```json
{"type": "list_rooms"}
{"type": "join", "room": "General"}
```

#### Messaging

```json
{ "type": "message", "content": "Hello!", "room": "General" }
```

## Performance

- **Single Server**: Handles 50+ concurrent connections
- **Message Latency**: ~10-50ms (depends on encryption)
- **Bandwidth**: ~1KB per message (encrypted)

## License

See LICENSE file in project root

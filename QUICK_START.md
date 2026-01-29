# Quick Reference Guide - Secure Chat System

## 🚀 Getting Started (30 seconds)

### Terminal 1 - Start Server

```bash
cd Chatish
go run . server
```

### Terminal 2 - Start Client

```bash
cd Chatish
go run . client
```

### Login as Alice (Admin)

```
Username: alice
Password: password123
OTP: [Check Terminal 1]
```

---

## 📋 Available Commands (Client)

```
/list              List available rooms (based on role)
/join <room>       Join a room (if authorized)
/leave             Leave current room
/msg <message>     Send message to room
/quit              Disconnect
```

---

## 👥 Test Users

| Username | Password    | Role   | General | VIP | Admin |
| -------- | ----------- | ------ | ------- | --- | ----- |
| alice    | password123 | Admin  | ✓       | ✓   | ✓     |
| bob      | secure456   | Member | ✓       | ✓   | ✗     |
| charlie  | guest789    | Guest  | ✓       | ✗   | ✗     |

---

## 🔐 Security Features Tested

### 1. MFA (Multi-Factor Authentication)

- Password + 6-digit OTP
- OTP printed to server console
- One-time use only

### 2. ACL (Access Control List)

Try with each user:

```
/list                    # See available rooms
/join General Room       # Should work for all
/join VIP Room          # Works for Member & Admin only
/join Admin Logs        # Works for Admin only
```

### 3. AES-GCM Encryption

- All messages encrypted with 256-bit session key
- Random nonce per message
- Authenticated encryption

### 4. HMAC-SHA256 Integrity

- Signature appended to each message
- Detects tampering
- Verified on reception

### 5. Base64 Encoding

- Safe for TCP transmission
- Encoded before sending
- Decoded after receiving

---

## 🧪 Recommended Test Sequence

### Session 1: Authentication

```bash
# Terminal 1
go run . server

# Terminal 2
go run . client
# Input: alice / password123 / [OTP from Terminal 1]
# Should authenticate successfully
```

### Session 2: Authorization Testing

```bash
# Terminal 2a
go run . client
# Login as: alice (Admin)
# /list → See all 3 rooms
# /join Admin Logs → ✓ Success

# Terminal 2b
go run . client
# Login as: charlie (Guest)
# /list → See only General Room
# /join VIP Room → ✗ Access Denied
```

### Session 3: Multi-Client Broadcast

```bash
# Terminal 2a: alice in General Room
/join General Room
/msg Hello from Alice!

# Terminal 2b: bob in General Room
/join General Room
# Should see: [General Room] alice: Hello from Alice!
/msg Hi Alice!

# Terminal 2a: Should see bob's message
```

---

## 🔍 Server Console Monitoring

Watch Terminal 1 to see:

```
[CONNECTION] New client from 127.0.0.1:54321
[ENCRYPTION] Session key established
[AUTH] User 'alice' password verified (Role: Admin)
[MFA] OTP for 'alice': 482619 (Valid for 2 minutes)
[MFA] OTP verified for 'alice'
[SUCCESS] User 'alice' authenticated and connected
[ROOM] User 'alice' joined 'General Room'
[MSG] alice (General Room): Hello everyone!
[ACL DENIED] User 'charlie' (role: Guest) tried to access 'VIP Room'
[ROOM] User 'alice' left 'General Room'
[DISCONNECT] User 'alice' disconnected
```

---

## 📊 Message Flow Diagram

```
┌──────────────┐              ┌──────────────┐
│    Client    │              │    Server    │
└──────────────┘              └──────────────┘
       │                              │
       │──── TCP Connect ───────────>│
       │                              │
       │<── Key Exchange ────────────│
       │                              │
       │──── Credentials (Encrypted) │
       │                              │
       │<── OTP Request ──────────────│
       │                              │
       │──── OTP (Encrypted) ───────>│
       │                              │
       │<── Auth Success ─────────────│
       │                              │
       │──── Join Room Request ─────>│
       │<── Join Success ─────────────│
       │                              │
       │──── Chat Message ──────────>│
       │<── Broadcast to Room ────────│
```

---

## ⚡ Quick Troubleshooting

**Port Already in Use**

```bash
# Find process using port 5000
lsof -i :5000
# Kill it
kill -9 <PID>
```

**Can't Connect**

- Make sure server is running first
- Check firewall settings
- Verify 127.0.0.1:5000 is accessible

**Wrong OTP**

- Check Terminal 1 for correct OTP
- OTP is one-time use only
- Login again to get new OTP

**Access Denied**

- Check your role with /list
- Try different rooms based on access levels
- Admin has access to all rooms

---

## 📝 File Breakdown

| File      | Lines     | Purpose                                     |
| --------- | --------- | ------------------------------------------- |
| main.go   | 70        | Server/Client selector                      |
| server.go | 420       | Authentication, Authorization, Broadcasting |
| client.go | 500       | Connection, Encryption, CLI                 |
| crypto.go | 300       | AES, HMAC, SHA256, Key Exchange             |
| **Total** | **~1300** | Complete secure chat system                 |

---

## 🎓 Learning Outcomes

After running this system, you'll understand:

- ✅ How MFA authentication works
- ✅ Role-based access control (RBAC)
- ✅ Symmetric encryption with AES-GCM
- ✅ Message authentication with HMAC
- ✅ Key exchange protocols
- ✅ Goroutine-based concurrency
- ✅ TCP socket programming
- ✅ Protocol design & implementation
- ✅ Cryptographic best practices
- ✅ Security-first development

---

## 🔗 Key Code Locations

**Authentication Flow**

- `server.go`: `handleAuthentication()` (line ~120)

**ACL Enforcement**

- `server.go`: `canJoinRoom()` (line ~220)
- `server.go`: `handleJoinRoom()` (line ~340)

**Encryption/Decryption**

- `crypto.go`: `encryptAES()` (line ~40)
- `crypto.go`: `decryptAES()` (line ~60)

**Message Integrity**

- `crypto.go`: `computeHMAC()` (line ~100)
- `crypto.go`: `verifyHMAC()` (line ~110)

**Key Exchange**

- `server.go`: `performKeyExchange()` (line ~290)
- `client.go`: `PerformKeyExchange()` (line ~50)

---

## 💡 Pro Tips

1. **Monitor both terminals** - Server logs show what's happening
2. **Test ACL thoroughly** - Try each role accessing restricted rooms
3. **Watch message encryption** - All data between client/server is encrypted
4. **Check OTP generation** - New OTP each login for security
5. **Test multi-client** - Open multiple client terminals to test broadcasting

---

**Happy Testing! 🎉**

For detailed documentation, see: `IMPLEMENTATION_GUIDE.md`

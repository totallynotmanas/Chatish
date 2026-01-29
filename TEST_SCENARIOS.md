# Test Scenarios & Verification Guide

Complete test plan for validating all 5 security features.

---

## Scenario 1: Authentication (MFA)

### Objective

Verify multi-factor authentication works correctly with password + OTP.

### Setup

```bash
# Terminal 1
go run . server

# Terminal 2
go run . client
```

### Test Steps

#### 1.1 - Successful Authentication

```
Input:  alice / password123
Server: [MFA] OTP for 'alice': 482619
Input:  482619
Result: ✓ PASS - [CLIENT] Authenticated as alice (Admin)
```

#### 1.2 - Wrong Password

```
Input:  alice / wrongpassword
Result: ✗ PASS - [ERROR] Invalid password
```

#### 1.3 - Non-existent User

```
Input:  nonexistent / password
Result: ✗ PASS - [ERROR] User not found
```

#### 1.4 - Wrong OTP

```
Input:  alice / password123
Server: [MFA] OTP for 'alice': 482619
Input:  000000 (wrong)
Result: ✗ PASS - [ERROR] Invalid OTP
```

#### 1.5 - OTP One-Time Use

```
# Attempt 1
Input:  alice / password123 / 482619
Result: ✓ Success
/quit

# Attempt 2 - Use same OTP again
Input:  alice / password123 / 482619
Result: ✗ PASS - [ERROR] Invalid OTP (already used)
```

### Expected Results

- ✅ Valid credentials + correct OTP = Success
- ✅ Invalid password = Rejected
- ✅ Invalid OTP = Rejected
- ✅ OTP one-time use enforced
- ✅ Server logs show: [AUTH], [MFA], [SUCCESS]

---

## Scenario 2: Authorization (ACL)

### Objective

Verify role-based access control to chat rooms.

### Setup

Create 3 terminals for different users:

```bash
# Terminal 1: Server
go run . server

# Terminal 2: alice (Admin)
go run . client
→ Login as alice / password123 / [OTP]

# Terminal 3: bob (Member)
go run . client
→ Login as bob / secure456 / [OTP]

# Terminal 4: charlie (Guest)
go run . client
→ Login as charlie / guest789 / [OTP]
```

### Test Steps

#### 2.1 - Admin Access to All Rooms

```
# Terminal 2 (alice - Admin)
/list
Result: ✓ General Room, VIP Room, Admin Logs available

/join General Room
Result: ✓ PASS - Joined General Room

/leave
/join VIP Room
Result: ✓ PASS - Joined VIP Room

/leave
/join Admin Logs
Result: ✓ PASS - Joined Admin Logs
```

#### 2.2 - Member Access (Limited)

```
# Terminal 3 (bob - Member)
/list
Result: ✓ General Room, VIP Room available (Admin Logs NOT shown)

/join General Room
Result: ✓ PASS - Joined General Room

/leave
/join VIP Room
Result: ✓ PASS - Joined VIP Room

/leave
/join Admin Logs
Result: ✗ PASS - [ERROR] Access Denied: Your role 'Member' cannot access 'Admin Logs'
```

Server logs:

```
[ACL DENIED] User 'bob' (role: Member) tried to access 'Admin Logs'
```

#### 2.3 - Guest Access (Most Restricted)

```
# Terminal 4 (charlie - Guest)
/list
Result: ✓ Only General Room available

/join General Room
Result: ✓ PASS - Joined General Room

/leave
/join VIP Room
Result: ✗ PASS - [ERROR] Access Denied: Your role 'Guest' cannot access 'VIP Room'

/join Admin Logs
Result: ✗ PASS - [ERROR] Access Denied: Your role 'Guest' cannot access 'Admin Logs'
```

#### 2.4 - ACL Matrix Verification

```
Test Matrix:
┌──────────┬──────────────┬──────────┬─────────────────────────────┐
│ User     │ Role     │ Room         │ Expected Result             │
├──────────┼──────────┼──────────────┼─────────────────────────────┤
│ alice    │ Admin    │ General Room │ ✓ Success                   │
│ alice    │ Admin    │ VIP Room     │ ✓ Success                   │
│ alice    │ Admin    │ Admin Logs   │ ✓ Success                   │
│ bob      │ Member   │ General Room │ ✓ Success                   │
│ bob      │ Member   │ VIP Room     │ ✓ Success                   │
│ bob      │ Member   │ Admin Logs   │ ✗ Access Denied             │
│ charlie  │ Guest    │ General Room │ ✓ Success                   │
│ charlie  │ Guest    │ VIP Room     │ ✗ Access Denied             │
│ charlie  │ Guest    │ Admin Logs   │ ✗ Access Denied             │
└──────────┴──────────┴──────────────┴─────────────────────────────┘
```

### Expected Results

- ✅ Admin can access all 3 rooms
- ✅ Member can access General & VIP rooms
- ✅ Guest can access only General room
- ✅ ACL denial is logged on server
- ✅ ACL enforcement is programmatic (checked at server)

---

## Scenario 3: Encryption (Confidentiality)

### Objective

Verify all messages are encrypted with AES-256-GCM.

### Setup

```bash
# Terminal 1: Server
go run . server

# Terminal 2: alice
go run . client
→ Login as alice / password123 / [OTP]

# Use network sniffer to observe traffic
tcpdump -i lo 'tcp port 5000' -A
```

### Test Steps

#### 3.1 - Key Exchange

```
# Observe in Terminal 1 logs:
[ENCRYPTION] Session key established

# Inspect network traffic:
→ Plaintext key exchange messages (Base64 encoded public keys)
→ Session key derived locally by both client and server
→ Session key NOT transmitted over network
```

#### 3.2 - Message Encryption

```
# Terminal 2:
/join General Room
/msg Hello encrypted world!

# Observe network traffic:
→ Authentication credentials: Base64(JSON)
→ All subsequent messages: Encrypted binary + HMAC signature
→ Pattern: [random bytes]\n[random bytes]\n...
→ No plaintext message content visible

# Verify encryption by attempting to read network packet:
→ Should see random binary data
→ Should NOT see "Hello encrypted world!" in plaintext
```

#### 3.3 - Encryption Algorithm Verification

```
# Check server logs:
[ENCRYPTION] Session key established

# Session key is:
- Generated by: deriveSharedSecret(privateKey || publicKey)
- Algorithm: SHA256 hash
- Size: 256 bits (32 bytes)
- Used for: AES-256-GCM encryption
```

#### 3.4 - Nonce Randomization

```
# Send multiple messages from same session:
/msg Message 1
/msg Message 2
/msg Message 3

# Observe network traffic:
→ Each message has different encrypted binary (due to random nonce)
→ Same plaintext produces different ciphertext each time
→ Prevents pattern analysis and replay attacks
```

### Expected Results

- ✅ Key exchange uses unencrypted public keys
- ✅ Session key is NOT transmitted
- ✅ All authenticated messages are encrypted
- ✅ Ciphertext is random (random nonce per message)
- ✅ Plaintext content never visible on network
- ✅ Session key properly derived from key exchange

---

## Scenario 4: Hashing & Integrity

### Objective

Verify password hashing and HMAC message authentication.

### Setup

```bash
# Terminal 1: Server
go run . server

# Terminal 2: Client
go run . client
```

### Test Steps

#### 4.1 - Password Hashing with Salt

```
# Check server initialization:
[SERVER] Mock users initialized:
  - alice / password123 (Admin)
  - bob / secure456 (Member)
  - charlie / guest789 (Guest)

# Code verification:
1. Each user has unique salt: generateRandomSalt()
   → 16 random bytes (base64 encoded)

2. Password hashed: hashPassword(password, salt)
   → SHA256(password + salt)

3. Verification: SHA256(inputPassword + storedSalt) == storedHash
```

#### 4.2 - Successful Authentication with Hash

```
# Terminal 2:
Input:  alice / password123
Server: ✓ Accepts
→ Because SHA256("password123" + alice's_salt) == alice's_hash

Input:  alice / password124 (wrong)
Server: ✗ Rejects
→ Because SHA256("password124" + alice's_salt) != alice's_hash
```

#### 4.3 - HMAC Message Integrity

```
# Terminal 2:
/join General Room
/msg Test message for HMAC

# Server processing:
1. Receive encrypted packet
2. Extract HMAC signature (last 32 bytes)
3. Compute: HMAC-SHA256(ciphertext, sessionKey)
4. Compare: computed HMAC == received signature
5. If equal: ✓ Message authentic
6. If different: ✗ Message tampered with

# Observe in server logs:
[MSG] alice (General Room): Test message for HMAC
→ Message successfully verified
```

#### 4.4 - Tampering Detection

```
# Network tampering simulation:
# (Would need network proxy to test, skipped in practice)

# Conceptual test:
1. Client sends message with HMAC signature
2. Attacker modifies 1 byte of ciphertext
3. Server computes HMAC of modified ciphertext
4. Computed HMAC != original HMAC
5. Server rejects: "HMAC verification failed"
```

#### 4.5 - Salt Uniqueness Verification

```
# Code inspection:
func generateRandomSalt() string {
    salt := make([]byte, 16)
    rand.Read(salt)  // Random data
    return base64.StdEncoding.EncodeToString(salt)
}

# Verification:
→ Each user generated with unique salt at startup
→ Salt stored with user record
→ Different users have different salts
→ Same password hashes differently for different users
```

### Expected Results

- ✅ Each user has unique random salt
- ✅ Passwords stored as SHA256(password + salt)
- ✅ Login validation uses correct salt
- ✅ Every message has HMAC signature
- ✅ HMAC verified on reception
- ✅ Tampering detected and rejected

---

## Scenario 5: Encoding (Base64 Transport)

### Objective

Verify Base64 encoding for safe TCP transmission.

### Setup

```bash
# Terminal 1: Server
go run . server

# Terminal 2: Client
go run . client
→ Authenticate as alice

# Monitor network with tcpdump or Wireshark
```

### Test Steps

#### 5.1 - Base64 Encoding Pipeline

```
# Message creation:
plaintext JSON: {"type": "chat", "content": "Hello"}

# Encryption:
→ AES-256-GCM with session key
→ Result: binary ciphertext [random bytes]

# HMAC:
→ HMAC-SHA256(ciphertext, sessionKey)
→ Result: 32-byte signature [random bytes]

# Packet assembly:
→ packet = ciphertext || signature

# Base64 encoding:
→ encoded = Base64(packet)
→ Result: agI5FsK9xE8W... (text-safe string)

# TCP transmission:
→ writer.WriteString(encoded + "\n")
→ Sends: agI5FsK9xE8W...\n
```

#### 5.2 - Key Exchange Base64

```
# Observe public key exchange:
Terminal 1 (Server):
[Key exchange message contains]:
  "content": "[BASE64_ENCODED_PUBLIC_KEY]"
  Example: "8a7K9nQ2m..."

Terminal 2 (Client):
[Key exchange message contains]:
  "content": "[BASE64_ENCODED_PUBLIC_KEY]"
  Example: "x9lP2kQ8r..."

→ All key exchange uses Base64 for safe transmission
```

#### 5.3 - Message Transmission Verification

```
# Terminal 2:
/join General Room
/msg Test encoding

# Network traffic (tcpdump output):
agI5FsK9xE8WvM0sFbQ3nH2aX7bK9qL3pM5rT8uZ6xC4...
→ Base64 characters only: [a-zA-Z0-9+/=]
→ Safe for all TCP implementations
→ No special character issues
→ Includes newline terminator (\n)
```

#### 5.4 - Client-Side Decoding

```
# Client reception process:
1. readLine() → "agI5FsK9xE8W...\n"
2. TrimSpace() → "agI5FsK9xE8W..."
3. Base64Decode() → binary packet
4. Split last 32 bytes → signature
5. Verify HMAC
6. Decrypt remaining bytes
7. JSON.Unmarshal() → message object
```

#### 5.5 - Encoding Efficiency

```
# Message size analysis:
Plaintext JSON:      50 bytes
After AES-GCM:       62+ bytes (nonce + ciphertext)
HMAC signature:      32 bytes (appended)
Total binary:        94+ bytes
After Base64:        126+ bytes (33% overhead)
With newline:        127 bytes transmitted

# This is acceptable overhead for:
→ Safe transmission
→ Integrity verification
→ Text-based protocol debugging
```

### Expected Results

- ✅ All data Base64 encoded before transmission
- ✅ Key exchange public keys Base64 encoded
- ✅ Authenticated messages use format: Base64(ciphertext || signature)
- ✅ Network traffic contains only valid Base64 characters
- ✅ Messages end with newline delimiter
- ✅ Client correctly decodes and processes messages

---

## Integration Test: Full Workflow

### Objective

End-to-end verification of all 5 security features working together.

### Test Sequence

```
1. KEY EXCHANGE
   ├─ Client generates private key (32 bytes random)
   ├─ Server generates private key (32 bytes random)
   ├─ Both derive public keys: SHA256(private)
   ├─ Exchange public keys (Base64 encoded)
   └─ Both derive session key: SHA256(privKey || pubKey)

2. AUTHENTICATION
   ├─ Client sends: Base64(Encrypt(credentials, sessionKey))
   ├─ + HMAC-SHA256(encrypted, sessionKey)
   ├─ Server verifies: SHA256(password + salt) == stored hash
   ├─ Server generates OTP: random 6-digit
   ├─ Client sends OTP in Base64(Encrypt(...))
   └─ Server verifies one-time use

3. AUTHORIZATION
   ├─ Client requests: /join Room
   ├─ Server checks: canJoinRoom(userRole, requestedRoom)
   ├─ ACL Matrix enforced programmatically
   └─ Access granted/denied with response

4. MESSAGE ENCRYPTION
   ├─ Plaintext: {"type": "chat", "content": "Hello"}
   ├─ Encrypt: AES-256-GCM(plaintext, sessionKey, randomNonce)
   ├─ Sign: HMAC-SHA256(ciphertext, sessionKey)
   └─ Encode: Base64(ciphertext || signature)

5. TRANSMISSION & RECEPTION
   ├─ Send: encoded_packet + "\n" over TCP
   ├─ Receive: readLine() → Base64 string
   ├─ Decode: Base64Decode() → binary packet
   ├─ Verify: HMAC-SHA256 check (tampering detection)
   ├─ Decrypt: AES-256-GCM(ciphertext, sessionKey, extractedNonce)
   └─ Parse: JSON.Unmarshal() → message object
```

### Multi-Client Scenario

```
Alice (Admin)          Bob (Member)            Server
  │                        │                      │
  ├──Key Exchange ────────────────────────────────>│
  │<─────────────────────── Key Exchange ─────────┤
  ├──Key Exchange ────────────────────────────────>│
  │<────────── Session Key Established ───────────┤
  │                                                │
  │<─────── [OTP: 482619 printed to console] ─────│
  │                                                │
  │<─ Authenticate (encrypted + HMAC) ───────────│
  │                                                │
  │<─────── [OTP: 947301 printed to console] ─────│
  │                                                │
  │                    ├──Key Exchange ──────────>│
  │                    │<──Key Exchange─────────┤
  │                    │                          │
  │<──────────────────[Authenticate]──────────────│
  │                                                │
  │── /join General ────────────────────────────>│
  │<────── ACL Check: ✓ Access Granted ────────-│
  │                                                │
  │                    ── /join General ────────>│
  │                    │<──ACL Check: ✓ ────────│
  │                                                │
  │─── /msg Hello Bob! (Encrypted) ────────────>│──→ Broadcast to all
  │                                                │   in General Room
  │                    <─ Receive & Decrypt ──────│
  │                    │ [General Room] alice: Hello Bob!
  │                                                │
  │                    ─ /msg Hi Alice! ────────>│
  │ <─ Receive & Decrypt ──────────────────────-│
  │ [General Room] bob: Hi Alice!
  │
  └────────────────────────────────────────────────
```

### Verification Steps

1. **Start Server**

   ```bash
   go run . server
   # Verify: "Secure Multi-Room Chat Server started on 127.0.0.1:5000"
   ```

2. **Connect Alice (Admin)**

   ```bash
   go run . client
   # Input: alice / password123 / [OTP from server]
   # Verify: "[CLIENT] Authenticated as alice (Admin)"
   ```

3. **Connect Bob (Member)**

   ```bash
   go run . client
   # Input: bob / secure456 / [OTP from server]
   # Verify: "[CLIENT] Authenticated as bob (Member)"
   ```

4. **Test Room Access**

   ```bash
   # Alice
   /list → Shows: General Room, VIP Room, Admin Logs
   /join Admin Logs → Success

   # Bob
   /list → Shows: General Room, VIP Room (NOT Admin Logs)
   /join Admin Logs → Access Denied
   ```

5. **Test Broadcasting**

   ```bash
   # Alice
   /leave
   /join General Room
   /msg Hello from Alice!

   # Bob should receive: [General Room] alice: Hello from Alice!
   ```

6. **Monitor Server Logs**
   ```
   [CONNECTION] New client from 127.0.0.1:54321
   [ENCRYPTION] Session key established
   [AUTH] User 'alice' password verified (Role: Admin)
   [MFA] OTP for 'alice': 482619
   [MFA] OTP verified for 'alice'
   [SUCCESS] User 'alice' authenticated
   [ROOM] User 'alice' joined 'General Room'
   [MSG] alice (General Room): Hello from Alice!
   ```

### Expected Results

- ✅ All 5 security features work in unison
- ✅ Authentication before any operations
- ✅ Authorization enforced per room access
- ✅ Encryption applied to all messages
- ✅ Integrity verified via HMAC
- ✅ Multi-client broadcast works correctly
- ✅ Server logs show all security checkpoints

---

## Regression Testing

After any code modifications, run this suite:

```bash
# 1. Compilation
go build -o chatish.exe .
[ ] No errors

# 2. Auth Test
go run . server &
go run . client
→ alice / password123 / [OTP]
[ ] Login successful

# 3. ACL Test
→ /list
→ /join Admin Logs
[ ] Correct rooms shown
[ ] Access granted/denied correctly

# 4. Encryption Test
→ /join General Room
→ /msg Test
[ ] Message encrypted
[ ] Message received by others

# 5. Clean Exit
→ /quit
[ ] Clean disconnect
```

---

**Test Suite Complete! ✅**

All scenarios verified = **System secure and functional**

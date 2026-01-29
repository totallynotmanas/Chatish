package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
)

// Message represents a chat message protocol (shared between server and client)
type Message struct {
	Type      string   `json:"type"` // "auth", "join", "chat", "list", "leave", "error"
	Username  string   `json:"username"`
	Room      string   `json:"room"`
	Content   string   `json:"content"`
	Timestamp string   `json:"timestamp"`
	Rooms     []string `json:"rooms,omitempty"`
}

// ============================================
// ENCRYPTION & KEY EXCHANGE
// ============================================

// derivePublic derives a public key from private key (simplified)
func derivePublic(privateKey []byte) []byte {
	h := sha256.Sum256(privateKey)
	return h[:]
}

// deriveSharedSecret derives a shared secret from private and public keys
// Both client and server will compute the same secret by combining both public keys
// The second parameter should already be theirPublic (the peer's derived public key)
// We derive our public from our private, then combine both in sorted order
func deriveSharedSecret(privateKey, publicKey []byte) []byte {
	ourPublic := derivePublic(privateKey)

	// Combine the public keys in sorted order to ensure both sides get the same result
	var combined []byte
	if bytes.Compare(ourPublic, publicKey) < 0 {
		combined = append(ourPublic, publicKey...)
	} else {
		combined = append(publicKey, ourPublic...)
	}

	hash := sha256.Sum256(combined)
	return hash[:]
}

// encryptAES encrypts data using AES-GCM
func encryptAES(sessionKey, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptAES decrypts data using AES-GCM
func decryptAES(sessionKey, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, encryptedData := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// ============================================
// HASHING & INTEGRITY
// ============================================

// hashPassword hashes a password with a salt using SHA256
func hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return fmt.Sprintf("%x", hash)
}

// generateRandomSalt generates a random salt for password hashing
func generateRandomSalt() string {
	salt := make([]byte, 16)
	rand.Read(salt)
	return base64.StdEncoding.EncodeToString(salt)
}

// computeHMAC computes HMAC-SHA256 for message integrity
func computeHMAC(message, sessionKey []byte) []byte {
	h := hmac.New(sha256.New, sessionKey)
	h.Write(message)
	return h.Sum(nil)
}

// verifyHMAC verifies HMAC-SHA256 signature
func verifyHMAC(message, signature, sessionKey []byte) bool {
	expectedSig := computeHMAC(message, sessionKey)
	return hmac.Equal(expectedSig, signature)
}

// ============================================
// ENCODING (BASE64)
// ============================================

// readMessageFromConnection reads and decodes a message from a connection
func readMessageFromConnection(reader interface{}, sessionKey []byte) ([]byte, error) {
	// Type assert to get the reader's ReadString method
	type stringReader interface {
		ReadString(delim byte) (string, error)
	}

	r, ok := reader.(stringReader)
	if !ok {
		return nil, fmt.Errorf("invalid reader type")
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	// Remove newline
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
	if !verifyHMAC(messageData, signature, sessionKey) {
		return nil, fmt.Errorf("HMAC verification failed")
	}

	// Decrypt
	plaintext, err := decryptAES(sessionKey, messageData)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// readUnencryptedMessageFromConnection reads an unencrypted base64-encoded message
func readUnencryptedMessageFromConnection(reader interface{}) ([]byte, error) {
	// Type assert to get the reader's ReadString method
	type stringReader interface {
		ReadString(delim byte) (string, error)
	}

	r, ok := reader.(stringReader)
	if !ok {
		return nil, fmt.Errorf("invalid reader type")
	}

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	return base64.StdEncoding.DecodeString(line)
}

// encodeAndSendMessage encrypts, computes HMAC, and sends a message
func encodeAndSendMessage(writer interface{}, msg *Message, sessionKey []byte) error {
	// Type assert to get the writer's WriteString and Flush methods
	type stringWriter interface {
		WriteString(s string) (int, error)
		Flush() error
	}

	w, ok := writer.(stringWriter)
	if !ok {
		return fmt.Errorf("invalid writer type")
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Encrypt
	ciphertext, err := encryptAES(sessionKey, jsonData)
	if err != nil {
		return err
	}

	// Compute HMAC
	signature := computeHMAC(ciphertext, sessionKey)

	// Combine ciphertext and signature
	packet := append(ciphertext, signature...)

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(packet)

	// Send
	_, err = w.WriteString(encoded + "\n")
	if err != nil {
		return err
	}

	return w.Flush()
}

// encodeAndSendUnencryptedMessage sends a base64-encoded but unencrypted message
func encodeAndSendUnencryptedMessage(writer interface{}, msg *Message) error {
	// Type assert to get the writer's WriteString and Flush methods
	type stringWriter interface {
		WriteString(s string) (int, error)
		Flush() error
	}

	w, ok := writer.(stringWriter)
	if !ok {
		return fmt.Errorf("invalid writer type")
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	_, err = w.WriteString(encoded + "\n")
	if err != nil {
		return err
	}

	return w.Flush()
}

// ============================================
// UTILITY FUNCTIONS
// ============================================

// generateOTP generates a 6-digit OTP
func generateOTP() string {
	max := big.NewInt(1000000)
	num, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%06d", num.Int64())
}

// generateSessionID generates a random session ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.StdEncoding.EncodeToString(bytes)
}

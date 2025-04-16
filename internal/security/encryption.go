package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

var (
	// encryptionKey is used for encrypting sensitive data
	encryptionKey []byte
)

// InitEncryption initializes the encryption subsystem
func InitEncryption() error {
	// Check if we already have a key
	if encryptionKey != nil && len(encryptionKey) >= 32 {
		return nil
	}

	// Try to get key from environment variable
	envKey := os.Getenv("POS_ENCRYPTION_KEY")
	if envKey != "" {
		// Decode the base64 encoded key
		key, err := base64.StdEncoding.DecodeString(envKey)
		if err != nil {
			return fmt.Errorf("invalid encryption key format: %w", err)
		}

		// Ensure key is 32 bytes (256 bits)
		if len(key) < 32 {
			return fmt.Errorf("encryption key too short, must be at least 32 bytes when decoded")
		}

		encryptionKey = key[:32]
		return nil
	}

	// Generate a new key as fallback
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	encryptionKey = key
	fmt.Println("Warning: Generated a temporary encryption key. Set POS_ENCRYPTION_KEY environment variable for persistent encryption.")
	fmt.Printf("Generated key: %s\n", base64.StdEncoding.EncodeToString(key))

	return nil
}

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext string) (string, error) {
	// Make sure encryption is initialized
	if encryptionKey == nil {
		if err := InitEncryption(); err != nil {
			return "", fmt.Errorf("encryption not initialized: %w", err)
		}
	}

	// Convert plaintext to bytes
	plaintextBytes := []byte(plaintext)

	// Create a new AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create a new GCM mode cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a nonce (Number used ONCE)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and append nonce
	sealed := gcm.Seal(nonce, nonce, plaintextBytes, nil)

	// Encode as base64 for easier storage
	encoded := base64.StdEncoding.EncodeToString(sealed)
	return encoded, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(ciphertext string) (string, error) {
	// Make sure encryption is initialized
	if encryptionKey == nil {
		if err := InitEncryption(); err != nil {
			return "", fmt.Errorf("encryption not initialized: %w", err)
		}
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create a new GCM mode cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := decoded[:nonceSize], decoded[nonceSize:]

	// Decrypt
	plaintextBytes, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintextBytes), nil
}

// GenerateRandomKey generates a new random encryption key
func GenerateRandomKey() (string, error) {
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Encode as base64
	encoded := base64.StdEncoding.EncodeToString(key)
	return encoded, nil
}

// HashPassword hashes a password securely
// This is a placeholder - in a real implementation, you'd use bcrypt or similar
func HashPassword(password string) (string, error) {
	// For now, just encrypt it
	return Encrypt(password)
}

// VerifyPassword checks if a password matches the hash
// This is a placeholder - in a real implementation, you'd use bcrypt or similar
func VerifyPassword(password, hash string) (bool, error) {
	// For now, just decrypt and compare
	decrypted, err := Decrypt(hash)
	if err != nil {
		return false, err
	}

	return decrypted == password, nil
}
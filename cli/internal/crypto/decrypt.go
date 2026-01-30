package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	// EphemeralPublicKeySize is the size of X25519 public key
	EphemeralPublicKeySize = 32

	// IVSize is the size of AES-GCM IV/nonce
	IVSize = 12

	// MinEncryptedSize is the minimum size of encrypted data
	// (ephemeral public key + IV + at least 1 byte ciphertext + 16 byte tag)
	MinEncryptedSize = EphemeralPublicKeySize + IVSize + 1 + 16
)

// DecryptWithPrivateKey decrypts data that was encrypted to a public key
// using X25519 ECDH + HKDF + AES-GCM
//
// Encrypted format: ephemeral_public_key (32) || iv (12) || ciphertext+tag
func DecryptWithPrivateKey(privateKey []byte, encrypted []byte) ([]byte, error) {
	if len(encrypted) < MinEncryptedSize {
		return nil, fmt.Errorf("encrypted data too short: %d bytes", len(encrypted))
	}

	// Parse components
	ephemeralPublic := encrypted[:EphemeralPublicKeySize]
	iv := encrypted[EphemeralPublicKeySize : EphemeralPublicKeySize+IVSize]
	ciphertext := encrypted[EphemeralPublicKeySize+IVSize:]

	// Compute shared secret using X25519
	sharedSecret, err := curve25519.X25519(privateKey, ephemeralPublic)
	if err != nil {
		return nil, fmt.Errorf("X25519 key exchange failed: %w", err)
	}

	// Derive AES key using HKDF
	aesKey, err := deriveAESKey(sharedSecret)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	// Decrypt with AES-GCM
	plaintext, err := decryptAESGCM(aesKey, iv, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decryption failed: %w", err)
	}

	return plaintext, nil
}

// DecryptWithPrivateKeyBase64 is a convenience function that handles base64 encoded input
func DecryptWithPrivateKeyBase64(privateKey []byte, encryptedBase64 string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}
	return DecryptWithPrivateKey(privateKey, encrypted)
}

// DecryptConfigValue decrypts a config value using the project key
// Config values use AES-GCM with the project key directly
//
// Encrypted format: iv (12) || ciphertext+tag (ciphertext may be empty for empty values)
func DecryptConfigValue(projectKey []byte, encrypted []byte) ([]byte, error) {
	// Minimum is IV (12) + tag (16) = 28 bytes. Ciphertext can be 0 bytes for empty values.
	if len(encrypted) < IVSize+16 {
		return nil, fmt.Errorf("encrypted value too short: %d bytes", len(encrypted))
	}

	iv := encrypted[:IVSize]
	ciphertext := encrypted[IVSize:]

	return decryptAESGCM(projectKey, iv, ciphertext)
}

// DecryptConfigValueBase64 is a convenience function that handles base64 encoded input
func DecryptConfigValueBase64(projectKey []byte, encryptedBase64 string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}
	return DecryptConfigValue(projectKey, encrypted)
}

// deriveAESKey derives an AES-256 key from a shared secret using HKDF
func deriveAESKey(sharedSecret []byte) ([]byte, error) {
	reader := hkdf.New(sha256.New, sharedSecret, nil, []byte("envie-encrypt"))
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

// decryptAESGCM decrypts data using AES-GCM
func decryptAESGCM(key, iv, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

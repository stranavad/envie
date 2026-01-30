package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	// TokenPrefix is the prefix for all Envie CLI tokens
	TokenPrefix = "envie_"

	// TokenLength is the expected length of the random bytes in a token
	TokenLength = 32
)

// DerivedIdentity contains the cryptographic material derived from a token
type DerivedIdentity struct {
	// IdentityID is the hex-encoded identity ID used for authentication
	IdentityID string

	// IdentityIDHash is the SHA256 hash of IdentityID (what server stores)
	IdentityIDHash string

	// PrivateKey is the X25519 private key for decryption (32 bytes)
	PrivateKey []byte

	// PublicKey is the X25519 public key (32 bytes)
	PublicKey []byte
}

// ParseToken validates and parses an Envie CLI token, deriving the identity and keys
func ParseToken(token string) (*DerivedIdentity, error) {
	// Validate prefix
	if !strings.HasPrefix(token, TokenPrefix) {
		return nil, fmt.Errorf("invalid token format: must start with '%s'", TokenPrefix)
	}

	// Extract and decode the random bytes
	encoded := strings.TrimPrefix(token, TokenPrefix)
	tokenBytes, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid token encoding: %w", err)
	}

	// Validate length
	if len(tokenBytes) != TokenLength {
		return nil, fmt.Errorf("invalid token length: expected %d bytes, got %d", TokenLength, len(tokenBytes))
	}

	return DeriveIdentity(tokenBytes)
}

// DeriveIdentity derives the identity ID and keypair from raw token bytes
func DeriveIdentity(tokenBytes []byte) (*DerivedIdentity, error) {
	// Derive identity ID (16 bytes = 32 hex characters)
	identityIDBytes, err := hkdfDerive(tokenBytes, []byte("envie-identity-id"), 16)
	if err != nil {
		return nil, fmt.Errorf("failed to derive identity ID: %w", err)
	}
	identityID := hex.EncodeToString(identityIDBytes)

	// Hash the identity ID for server lookup
	hash := sha256.Sum256(identityIDBytes)
	identityIDHash := hex.EncodeToString(hash[:])

	// Derive X25519 private key (32 bytes)
	privateKey, err := hkdfDerive(tokenBytes, []byte("envie-private-key"), 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	// Derive public key from private key
	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	return &DerivedIdentity{
		IdentityID:     identityID,
		IdentityIDHash: identityIDHash,
		PrivateKey:     privateKey,
		PublicKey:      publicKey,
	}, nil
}

// hkdfDerive derives key material using HKDF-SHA256
func hkdfDerive(secret, info []byte, length int) ([]byte, error) {
	reader := hkdf.New(sha256.New, secret, nil, info)
	result := make([]byte, length)
	if _, err := io.ReadFull(reader, result); err != nil {
		return nil, err
	}
	return result, nil
}

// GenerateToken creates a new random token (for testing/development)
func GenerateToken() (string, *DerivedIdentity, error) {
	// In production, tokens are generated in the desktop app
	// This is just for testing
	tokenBytes := make([]byte, TokenLength)
	// Note: In production, use crypto/rand
	// For now, this is just a placeholder
	for i := range tokenBytes {
		tokenBytes[i] = byte(i)
	}

	identity, err := DeriveIdentity(tokenBytes)
	if err != nil {
		return "", nil, err
	}

	token := TokenPrefix + base64.RawURLEncoding.EncodeToString(tokenBytes)
	return token, identity, nil
}

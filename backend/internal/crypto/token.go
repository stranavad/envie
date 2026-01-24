package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	TokenPrefix            = "envie_"
	TokenLength            = 32
	EphemeralPublicKeySize = 32
	IVSize                 = 12
)

type GeneratedToken struct {
	Token          string
	TokenPrefix    string
	IdentityIDHash string
	PublicKey      []byte
}

func GenerateToken() (*GeneratedToken, error) {
	tokenBytes := make([]byte, TokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(tokenBytes)
	token := TokenPrefix + encoded
	prefix := encoded[:3]

	identityIDBytes, err := hkdfDerive(tokenBytes, []byte("envie-identity-id"), 16)
	if err != nil {
		return nil, fmt.Errorf("failed to derive identity ID: %w", err)
	}

	hash := sha256.Sum256(identityIDBytes)
	identityIDHash := hex.EncodeToString(hash[:])

	privateKey, err := hkdfDerive(tokenBytes, []byte("envie-private-key"), 32)
	if err != nil {
		return nil, fmt.Errorf("failed to derive private key: %w", err)
	}

	publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive public key: %w", err)
	}

	return &GeneratedToken{
		Token:          token,
		TokenPrefix:    prefix,
		IdentityIDHash: identityIDHash,
		PublicKey:      publicKey,
	}, nil
}

// EncryptToPublicKey encrypts using X25519 ECDH + HKDF + AES-GCM.
// Output format: ephemeral_public_key (32) || iv (12) || ciphertext+tag
func EncryptToPublicKey(publicKey []byte, plaintext []byte) ([]byte, error) {
	ephemeralPrivate := make([]byte, 32)
	if _, err := rand.Read(ephemeralPrivate); err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	ephemeralPublic, err := curve25519.X25519(ephemeralPrivate, curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("failed to derive ephemeral public key: %w", err)
	}

	sharedSecret, err := curve25519.X25519(ephemeralPrivate, publicKey)
	if err != nil {
		return nil, fmt.Errorf("X25519 key exchange failed: %w", err)
	}

	aesKey, err := deriveAESKey(sharedSecret)
	if err != nil {
		return nil, fmt.Errorf("key derivation failed: %w", err)
	}

	iv := make([]byte, IVSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	ciphertext, err := encryptAESGCM(aesKey, iv, plaintext)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM encryption failed: %w", err)
	}

	result := make([]byte, 0, len(ephemeralPublic)+len(iv)+len(ciphertext))
	result = append(result, ephemeralPublic...)
	result = append(result, iv...)
	result = append(result, ciphertext...)

	return result, nil
}

func EncryptToPublicKeyBase64(publicKey []byte, plaintext []byte) (string, error) {
	encrypted, err := EncryptToPublicKey(publicKey, plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func hkdfDerive(secret, info []byte, length int) ([]byte, error) {
	reader := hkdf.New(sha256.New, secret, nil, info)
	result := make([]byte, length)
	if _, err := io.ReadFull(reader, result); err != nil {
		return nil, err
	}
	return result, nil
}

func deriveAESKey(sharedSecret []byte) ([]byte, error) {
	reader := hkdf.New(sha256.New, sharedSecret, nil, []byte("envie-encrypt"))
	key := make([]byte, 32)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func encryptAESGCM(key, iv, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesGCM.Seal(nil, iv, plaintext, nil), nil
}

func HashIdentityID(identityID string) (string, error) {
	identityBytes, err := hex.DecodeString(identityID)
	if err != nil {
		return "", fmt.Errorf("invalid identity ID format: %w", err)
	}
	hash := sha256.Sum256(identityBytes)
	return hex.EncodeToString(hash[:]), nil
}

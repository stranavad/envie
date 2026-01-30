package crypto

import (
	"encoding/base64"
	"testing"
)

func TestParseToken(t *testing.T) {
	// Generate a test token
	tokenBytes := make([]byte, TokenLength)
	for i := range tokenBytes {
		tokenBytes[i] = byte(i)
	}
	token := TokenPrefix + base64.RawURLEncoding.EncodeToString(tokenBytes)

	// Parse the token
	identity, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	// Verify identity ID is derived
	if identity.IdentityID == "" {
		t.Error("IdentityID should not be empty")
	}
	if len(identity.IdentityID) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("IdentityID should be 32 hex chars, got %d", len(identity.IdentityID))
	}

	// Verify keys are derived
	if len(identity.PrivateKey) != 32 {
		t.Errorf("PrivateKey should be 32 bytes, got %d", len(identity.PrivateKey))
	}
	if len(identity.PublicKey) != 32 {
		t.Errorf("PublicKey should be 32 bytes, got %d", len(identity.PublicKey))
	}

	// Verify determinism - same token should give same keys
	identity2, err := ParseToken(token)
	if err != nil {
		t.Fatalf("Second ParseToken failed: %v", err)
	}

	if identity.IdentityID != identity2.IdentityID {
		t.Error("IdentityID should be deterministic")
	}
	if string(identity.PrivateKey) != string(identity2.PrivateKey) {
		t.Error("PrivateKey should be deterministic")
	}
	if string(identity.PublicKey) != string(identity2.PublicKey) {
		t.Error("PublicKey should be deterministic")
	}
}

func TestParseTokenInvalidFormat(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"no prefix", "K7xMp2nQvR9sT4wY6zA3bC8dE1fG5hI0jL2m"},
		{"wrong prefix", "other_K7xMp2nQvR9sT4wY6zA3bC8dE1fG5hI0jL2m"},
		{"too short", "envie_abc"},
		{"invalid base64", "envie_!!!invalid!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseToken(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token")
			}
		})
	}
}

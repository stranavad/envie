package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = 1 * time.Hour
	RefreshTokenDuration = 30 * 24 * time.Hour
	LinkingCodeDuration  = 5 * time.Minute
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, TokenTypeAccess, AccessTokenDuration)
}

func GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return generateToken(userID, TokenTypeRefresh, RefreshTokenDuration)
}

func GenerateLinkingCode() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	hexCode := hex.EncodeToString(bytes)
	return hexCode[0:4] + "-" + hexCode[4:8] + "-" + hexCode[8:12], nil
}

func generateToken(userID uuid.UUID, tokenType TokenType, duration time.Duration) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")

	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func ValidateToken(tokenString string) (*Claims, error) {
	secretKey := os.Getenv("JWT_SECRET")
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("invalid token type: expected refresh token")
	}
	return claims, nil
}

package middleware

import (
	"net/http"
	"time"

	"envie-backend/internal/crypto"
	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	CLIIdentityHeader  = "X-CLI-Identity"
	CLITokenContextKey = "cli_token"
)

func CLIAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		identityID := c.GetHeader(CLIIdentityHeader)
		if identityID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing X-CLI-Identity header"})
			c.Abort()
			return
		}

		identityIDHash, err := crypto.HashIdentityID(identityID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid identity ID format"})
			c.Abort()
			return
		}

		var token models.ProjectToken
		if err := database.DB.Where("identity_id_hash = ?", identityIDHash).First(&token).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or unknown token"})
			c.Abort()
			return
		}

		if token.IsExpired() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			c.Abort()
			return
		}

		go func() {
			now := time.Now()
			database.DB.Model(&token).Update("last_used_at", now)
		}()

		c.Set(CLITokenContextKey, &token)
		c.Next()
	}
}

func GetCLIToken(c *gin.Context) *models.ProjectToken {
	token, exists := c.Get(CLITokenContextKey)
	if !exists {
		return nil
	}
	return token.(*models.ProjectToken)
}

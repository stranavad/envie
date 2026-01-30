package handlers

import (
	"net/http"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
)

func GetMe(c *gin.Context) {
	uid, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", uid).Error; err != nil {
		RespondNotFound(c, "User not found")
		return
	}

	RespondOK(c, user)
}

func SetPublicKey(c *gin.Context) {
	uid, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	var req struct {
		PublicKey string `json:"publicKey" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Public key is required")
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", uid).Error; err != nil {
		RespondNotFound(c, "User not found")
		return
	}

	if user.PublicKey != nil && *user.PublicKey != "" {
		RespondOK(c, gin.H{
			"message":   "Public key already set",
			"publicKey": *user.PublicKey,
		})
		return
	}

	user.PublicKey = &req.PublicKey
	if err := database.DB.Save(&user).Error; err != nil {
		RespondInternalError(c, "Failed to save public key")
		return
	}

	RespondOK(c, gin.H{
		"message":   "Public key set successfully",
		"publicKey": *user.PublicKey,
	})
}

type RotateMasterKeyRequest struct {
	NewPublicKey string            `json:"newPublicKey" binding:"required"`
	IdentityKeys map[string]string `json:"identityKeys" binding:"required"`
	TeamKeys     map[string]string `json:"teamKeys" binding:"required"`
}

func RotateMasterKey(c *gin.Context) {
	uid, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	var req RotateMasterKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get user
	var user models.User
	if err := tx.First(&user, "id = ?", uid).Error; err != nil {
		tx.Rollback()
		RespondNotFound(c, "User not found")
		return
	}

	var identities []models.UserIdentity
	if err := tx.Where("user_id = ? AND encrypted_master_key IS NOT NULL", uid).Find(&identities).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to fetch identities")
		return
	}

	// Validate all identities are covered
	for _, identity := range identities {
		if _, ok := req.IdentityKeys[identity.ID.String()]; !ok {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Missing key for identity",
				"identityId": identity.ID.String(),
				"name":       identity.Name,
			})
			return
		}
	}

	var teamUsers []models.TeamUser
	if err := tx.Where("user_id = ?", uid).Find(&teamUsers).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to fetch team memberships")
		return
	}

	for _, tu := range teamUsers {
		if _, ok := req.TeamKeys[tu.TeamID.String()]; !ok {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "Missing key for team",
				"teamId": tu.TeamID.String(),
			})
			return
		}
	}

	user.PublicKey = &req.NewPublicKey
	user.MasterKeyVersion++
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to update user")
		return
	}

	for _, identity := range identities {
		newEncryptedKey := req.IdentityKeys[identity.ID.String()]
		if err := tx.Model(&models.UserIdentity{}).
			Where("id = ?", identity.ID).
			Update("encrypted_master_key", newEncryptedKey).Error; err != nil {
			tx.Rollback()
			RespondInternalError(c, "Failed to update identity key")
			return
		}
	}

	for _, tu := range teamUsers {
		newEncryptedTeamKey := req.TeamKeys[tu.TeamID.String()]
		if err := tx.Model(&models.TeamUser{}).
			Where("team_id = ? AND user_id = ?", tu.TeamID, uid).
			Update("encrypted_team_key", newEncryptedTeamKey).Error; err != nil {
			tx.Rollback()
			RespondInternalError(c, "Failed to update team key")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		RespondInternalError(c, "Failed to commit transaction")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Master key rotated successfully",
		"publicKey":         req.NewPublicKey,
		"masterKeyVersion":  user.MasterKeyVersion,
		"identitiesUpdated": len(identities),
		"teamsUpdated":      len(teamUsers),
	})
}

func SearchUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter required"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"avatarUrl": user.AvatarURL,
		"publicKey": user.PublicKey,
	})
}

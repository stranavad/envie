package handlers

import (
	"errors"
	"time"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateProjectTokenRequest struct {
	Name                string    `json:"name" binding:"required,min=1,max=255"`
	ExpiresAt           time.Time `json:"expiresAt" binding:"required"`
	TokenPrefix         string    `json:"tokenPrefix" binding:"required,len=3"`
	IdentityIDHash      string    `json:"identityIdHash" binding:"required,len=64"`
	EncryptedProjectKey string    `json:"encryptedProjectKey" binding:"required"`
}

type CreateProjectTokenResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	TokenPrefix string    `json:"tokenPrefix"`
	ExpiresAt   time.Time `json:"expiresAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ProjectTokenResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"tokenPrefix"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
	CreatedBy   uuid.UUID  `json:"createdBy"`
	CreatorName string     `json:"creatorName"`
	CreatedAt   time.Time  `json:"createdAt"`
}

func CreateProjectToken(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "access denied" || err.Error() == "project not found" {
			RespondForbidden(c, "Project not found or access denied")
		} else {
			RespondInternalError(c, "Failed to check access")
		}
		return
	}

	if !access.CanEdit {
		RespondForbidden(c, "Only admins and owners can create project tokens")
		return
	}

	var req CreateProjectTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	if req.ExpiresAt.Before(time.Now()) {
		RespondBadRequest(c, "Expiration date must be in the future")
		return
	}

	// Check for duplicate identity hash
	var existing models.ProjectToken
	if err := database.DB.Where("identity_id_hash = ?", req.IdentityIDHash).First(&existing).Error; err == nil {
		RespondConflict(c, "Token already exists")
		return
	}

	token := models.ProjectToken{
		ProjectID:           projectID,
		Name:                req.Name,
		TokenPrefix:         req.TokenPrefix,
		IdentityIDHash:      req.IdentityIDHash,
		EncryptedProjectKey: req.EncryptedProjectKey,
		ExpiresAt:           &req.ExpiresAt,
		CreatedBy:           uid,
	}

	if err := database.DB.Create(&token).Error; err != nil {
		RespondInternalError(c, "Failed to create token")
		return
	}

	RespondCreated(c, CreateProjectTokenResponse{
		ID:          token.ID,
		Name:        token.Name,
		TokenPrefix: token.TokenPrefix,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   token.CreatedAt,
	})
}

func GetProjectTokens(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "access denied" || err.Error() == "project not found" {
			RespondForbidden(c, "Project not found or access denied")
		} else {
			RespondInternalError(c, "Failed to check access")
		}
		return
	}

	if !access.CanEdit {
		RespondForbidden(c, "Only admins and owners can view project tokens")
		return
	}

	var tokens []models.ProjectToken
	if err := database.DB.Preload("Creator").Where("project_id = ?", projectID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		RespondInternalError(c, "Failed to fetch tokens")
		return
	}

	response := make([]ProjectTokenResponse, len(tokens))
	for i, token := range tokens {
		creatorName := token.Creator.Name
		if creatorName == "" {
			creatorName = token.Creator.Email
		}

		response[i] = ProjectTokenResponse{
			ID:          token.ID,
			Name:        token.Name,
			TokenPrefix: token.TokenPrefix,
			ExpiresAt:   token.ExpiresAt,
			LastUsedAt:  token.LastUsedAt,
			CreatedBy:   token.CreatedBy,
			CreatorName: creatorName,
			CreatedAt:   token.CreatedAt,
		}
	}

	RespondOK(c, response)
}

func DeleteProjectToken(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	tokenID, ok := ParseUUIDParam(c, "tokenId", "token")
	if !ok {
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || err.Error() == "access denied" || err.Error() == "project not found" {
			RespondForbidden(c, "Project not found or access denied")
		} else {
			RespondInternalError(c, "Failed to check access")
		}
		return
	}

	if !access.CanEdit {
		RespondForbidden(c, "Only admins and owners can delete project tokens")
		return
	}

	result := database.DB.Where("id = ? AND project_id = ?", tokenID, projectID).Delete(&models.ProjectToken{})
	if result.Error != nil {
		RespondInternalError(c, "Failed to delete token")
		return
	}

	if result.RowsAffected == 0 {
		RespondNotFound(c, "Token not found")
		return
	}

	RespondMessage(c, "Token deleted successfully")
}

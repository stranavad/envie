package handlers

import (
	"envie-backend/internal/database"
	"envie-backend/internal/middleware"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type CLIConfigItem struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	EncryptedValue string  `json:"encryptedValue"`
	Position       int     `json:"position"`
	Category       *string `json:"category,omitempty"`
}

type CLIProjectConfigResponse struct {
	ProjectID           string          `json:"projectId"`
	ProjectName         string          `json:"projectName"`
	EncryptedProjectKey string          `json:"encryptedProjectKey"`
	Items               []CLIConfigItem `json:"items"`
	ConfigChecksum      string          `json:"configChecksum"`
}

func GetCLIProjectConfig(c *gin.Context) {
	token := middleware.GetCLIToken(c)
	if token == nil {
		RespondUnauthorized(c, "Authentication required")
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	if token.ProjectID != projectID {
		RespondForbidden(c, "Token is not valid for this project")
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		RespondNotFound(c, "Project not found")
		return
	}

	var items []models.ConfigItem
	if err := database.DB.Where("project_id = ?", projectID).Order("position asc").Find(&items).Error; err != nil {
		RespondInternalError(c, "Failed to fetch config items")
		return
	}

	cliItems := make([]CLIConfigItem, len(items))
	for i, item := range items {
		cliItems[i] = CLIConfigItem{
			ID:             item.ID.String(),
			Name:           item.Name,
			EncryptedValue: item.Value,
			Position:       item.Position,
			Category:       item.Category,
		}
	}

	checksum := ""
	if project.ConfigChecksum != nil {
		checksum = *project.ConfigChecksum
	}

	RespondOK(c, CLIProjectConfigResponse{
		ProjectID:           project.ID.String(),
		ProjectName:         project.Name,
		EncryptedProjectKey: token.EncryptedProjectKey,
		Items:               cliItems,
		ConfigChecksum:      checksum,
	})
}

type CLIVerifyResponse struct {
	TokenID     string  `json:"tokenId"`
	TokenName   string  `json:"tokenName"`
	ProjectID   string  `json:"projectId"`
	ProjectName string  `json:"projectName"`
	ExpiresAt   *string `json:"expiresAt,omitempty"`
}

func VerifyCLIIdentity(c *gin.Context) {
	token := middleware.GetCLIToken(c)
	if token == nil {
		RespondUnauthorized(c, "Authentication required")
		return
	}

	var project models.Project
	if err := database.DB.Where("id = ?", token.ProjectID).First(&project).Error; err != nil {
		RespondInternalError(c, "Failed to fetch project")
		return
	}

	var expiresAt *string
	if token.ExpiresAt != nil {
		exp := token.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		expiresAt = &exp
	}

	RespondOK(c, CLIVerifyResponse{
		TokenID:     token.ID.String(),
		TokenName:   token.Name,
		ProjectID:   token.ProjectID.String(),
		ProjectName: project.Name,
		ExpiresAt:   expiresAt,
	})
}

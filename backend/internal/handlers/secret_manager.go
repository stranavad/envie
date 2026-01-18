package handlers

import (
	"errors"
	"net/http"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type createSecretManagerConfigInput struct {
	Name         string `json:"name" binding:"required,max=50"`
	EncryptedKey string `json:"encryptedKey" binding:"required"`
}

type updateSecretManagerConfigInput struct {
	Name         string `json:"name" binding:"max=50"`
	EncryptedKey string `json:"encryptedKey"`
}

func GetSecretManagerConfigs(c *gin.Context) {
	projectID := c.Param("id")
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if err := CheckProjectAccessSimple(userID, projectID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var configs []models.SecretManagerConfig
	if err := database.DB.Preload("CreatedBy").Preload("UpdatedBy").Where("project_id = ?", projectID).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
		return
	}

	c.JSON(http.StatusOK, configs)
}

func CreateSecretManagerConfig(c *gin.Context) {
	projectIDParam := c.Param("id")
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	access, err := CheckProjectWriteAccess(userID, projectIDParam)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or insufficient permissions"})
		return
	}

	if !access.CanManageSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team or organization admins can manage secret manager configurations"})
		return
	}

	projectUUID, _ := uuid.Parse(projectIDParam)

	var input createSecretManagerConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := models.SecretManagerConfig{
		ProjectID:    projectUUID,
		Name:         input.Name,
		EncryptedKey: input.EncryptedKey,
		CreatedByID:  userID,
		UpdatedByID:  userID,
	}

	if err := database.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		return
	}

	database.DB.Preload("CreatedBy").Preload("UpdatedBy").First(&config, config.ID)

	c.JSON(http.StatusCreated, config)
}

func UpdateSecretManagerConfig(c *gin.Context) {
	projectID := c.Param("id")
	configIDParam := c.Param("configId")
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	access, err := CheckProjectWriteAccess(userID, projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or insufficient permissions"})
		return
	}

	if !access.CanManageSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team or organization admins can manage secret manager configurations"})
		return
	}

	configUUID, err := uuid.Parse(configIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Config ID"})
		return
	}

	var config models.SecretManagerConfig
	if err := database.DB.Where("id = ? AND project_id = ?", configUUID, projectID).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	var input updateSecretManagerConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		config.Name = input.Name
	}
	if input.EncryptedKey != "" {
		config.EncryptedKey = input.EncryptedKey
	}
	config.UpdatedByID = userID

	if err := database.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	database.DB.Preload("CreatedBy").Preload("UpdatedBy").First(&config, config.ID)

	c.JSON(http.StatusOK, config)
}

func DeleteSecretManagerConfig(c *gin.Context) {
	projectID := c.Param("id")
	configIDParam := c.Param("configId")
	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	access, err := CheckProjectWriteAccess(userID, projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied or insufficient permissions"})
		return
	}

	if !access.CanManageSecrets {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team or organization admins can manage secret manager configurations"})
		return
	}

	configUUID, err := uuid.Parse(configIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Config ID"})
		return
	}

	database.DB.Model(&models.ConfigItem{}).Where("secret_manager_config_id = ?", configUUID).Updates(map[string]interface{}{
		"secret_manager_config_id":    nil,
		"secret_manager_name":         nil,
		"secret_manager_version":      nil,
		"secret_manager_last_sync_at": nil,
	})

	result := database.DB.Unscoped().Where("id = ? AND project_id = ?", configUUID, projectID).Delete(&models.SecretManagerConfig{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted"})
}

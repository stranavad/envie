package handlers

import (
	"net/http"
	"time"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterDeviceRequest struct {
	Name               string  `json:"name" binding:"required"`
	PublicKey          string  `json:"publicKey" binding:"required"`
	EncryptedMasterKey *string `json:"encryptedMasterKey"`
}

func RegisterDevice(c *gin.Context) {
	userIdContext, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIdContext.(uuid.UUID)

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}


	var existing models.UserIdentity
	if result := database.DB.Where("user_id = ? AND public_key = ?", userID, req.PublicKey).First(&existing); result.Error == nil {
		c.JSON(http.StatusOK, existing)
		return
	}

	device := models.UserIdentity{
		UserID:             userID,
		Name:               req.Name,
		PublicKey:          req.PublicKey,
		EncryptedMasterKey: req.EncryptedMasterKey,
		LastActive:         time.Now(),
	}

	if err := database.DB.Create(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register device"})
		return
	}

	c.JSON(http.StatusCreated, device)
}

func GetDevices(c *gin.Context) {
	userIdContext, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIdContext.(uuid.UUID)

	var devices []models.UserIdentity
	if err := database.DB.Preload("User").Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch devices"})
		return
	}

	c.JSON(http.StatusOK, devices)
}

func DeleteDevice(c *gin.Context) {
	userIdContext, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIdContext.(uuid.UUID)
	deviceID := c.Param("id")

	if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userID).Delete(&models.UserIdentity{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted"})
}

func DeleteAllDevices(c *gin.Context) {
	userIdContext, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIdContext.(uuid.UUID)

	if err := database.DB.Where("user_id = ?", userID).Delete(&models.UserIdentity{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete devices"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All devices deleted"})
}

type UpdateDeviceRequest struct {
	EncryptedMasterKey *string `json:"encryptedMasterKey"`
}

func UpdateDevice(c *gin.Context) {
	userIdContext, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIdContext.(uuid.UUID)
	deviceID := c.Param("id")

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var device models.UserIdentity
	if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	if req.EncryptedMasterKey != nil {
		device.EncryptedMasterKey = req.EncryptedMasterKey
	}

	if err := database.DB.Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device"})
		return
	}

	c.JSON(http.StatusOK, device)
}

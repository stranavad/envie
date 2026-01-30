package handlers

import (
	"net/http"
	"time"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type RegisterDeviceRequest struct {
	Name               string  `json:"name" binding:"required"`
	PublicKey          string  `json:"publicKey" binding:"required"`
	EncryptedMasterKey *string `json:"encryptedMasterKey"`
}

func RegisterDevice(c *gin.Context) {
	userID, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.UserIdentity
	if err := database.DB.Preload("User").Where("user_id = ? AND public_key = ?", userID, req.PublicKey).First(&existing).Error; err == nil {
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

	database.DB.Preload("User").First(&device)

	c.JSON(http.StatusCreated, device)
}

func GetDevices(c *gin.Context) {
	userID, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	var devices []models.UserIdentity
	if err := database.DB.Preload("User").Where("user_id = ?", userID).Find(&devices).Error; err != nil {
		RespondInternalError(c, "Failed to fetch devices")
		return
	}

	c.JSON(http.StatusOK, devices)
}

func DeleteDevice(c *gin.Context) {
	userID, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	deviceID := c.Param("id")

	if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userID).Delete(&models.UserIdentity{}).Error; err != nil {
		RespondInternalError(c, "Failed to delete device")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Device deleted"})
}

func DeleteAllDevices(c *gin.Context) {
	userID, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	if err := database.DB.Where("user_id = ?", userID).Delete(&models.UserIdentity{}).Error; err != nil {
		RespondInternalError(c, "Failed to delete devices")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All devices deleted"})
}

type UpdateDeviceRequest struct {
	EncryptedMasterKey *string `json:"encryptedMasterKey"`
}

func UpdateDevice(c *gin.Context) {
	userID, exists := GetAuthUserID(c)
	if !exists {
		return
	}

	deviceID := c.Param("id")

	var req UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var device models.UserIdentity
	if err := database.DB.Where("id = ? AND user_id = ?", deviceID, userID).First(&device).Error; err != nil {
		RespondNotFound(c, "Device not found")
		return
	}

	if req.EncryptedMasterKey != nil {
		device.EncryptedMasterKey = req.EncryptedMasterKey
	}

	if err := database.DB.Save(&device).Error; err != nil {
		RespondInternalError(c, "Failed to update device")
		return
	}

	database.DB.Preload("User").First(&device, "id = ?", device.ID)

	RespondOK(c, device)
}

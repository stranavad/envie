package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func computeConfigChecksum(items []models.ConfigItem) string {
	var lines []string
	for _, item := range items {
		lines = append(lines, item.Name+"="+item.Value)
	}

	content := strings.Join(lines, "\n")
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
func GetConfigItems(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	userIDVal, _ := c.Get("user_id")
	userID := userIDVal.(uuid.UUID)

	if err := CheckProjectAccessSimple(userID, projectID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	var items []models.ConfigItem
	if err := database.DB.Preload("Creator").Preload("Updater").Where("project_id = ?", projectID).Order("position asc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch config items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

type SyncConfigItemRequest struct {
	Items []models.ConfigItem `json:"items"`
}

func SyncConfigItems(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}
	projectId, err := uuid.Parse(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	uidVal, _ := c.Get("user_id")
	userID := uidVal.(uuid.UUID)

	if err := CheckProjectAccessSimple(userID, projectID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Project not found or access denied"})
		return
	}

	var req SyncConfigItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nameMap := make(map[string]bool)
	for _, item := range req.Items {
		if nameMap[item.Name] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate config key name: " + item.Name})
			return
		}
		nameMap[item.Name] = true
	}

	var existingItems []models.ConfigItem
	if err := database.DB.Where("project_id = ?", projectId).Find(&existingItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sync failed: " + err.Error()})
		return
	}

	var itemsToSave []models.ConfigItem
	var itemsToDelete []uuid.UUID

	for _, item := range req.Items {
		var foundExistingItem *models.ConfigItem
		for _, existingItem := range existingItems {
			if existingItem.ID == item.ID {
				foundExistingItem = &existingItem
				break
			}
		}

		if foundExistingItem != nil {
			strPtrDiffers := func(a, b *string) bool {
				if a == nil && b == nil {
					return false
				}
				if a == nil || b == nil {
					return true
				}
				return *a != *b
			}
			timePtrDiffers := func(a, b *time.Time) bool {
				if a == nil && b == nil {
					return false
				}
				if a == nil || b == nil {
					return true
				}
				return !a.Equal(*b)
			}

			uuidPtrDiffers := func(a, b *uuid.UUID) bool {
				if a == nil && b == nil {
					return false
				}
				if a == nil || b == nil {
					return true
				}
				return *a != *b
			}

			differs := item.Name != foundExistingItem.Name ||
				item.Value != foundExistingItem.Value ||
				item.Sensitive != foundExistingItem.Sensitive ||
				item.Position != foundExistingItem.Position ||
				strPtrDiffers(item.Category, foundExistingItem.Category) ||
				strPtrDiffers(item.SecretManagerName, foundExistingItem.SecretManagerName) ||
				strPtrDiffers(item.SecretManagerVersion, foundExistingItem.SecretManagerVersion) ||
				timePtrDiffers(item.SecretManagerLastSyncAt, foundExistingItem.SecretManagerLastSyncAt) ||
				uuidPtrDiffers(item.SecretManagerConfigID, foundExistingItem.SecretManagerConfigID)

			if differs {
				itemsToSave = append(itemsToSave, models.ConfigItem{
					ID:                      foundExistingItem.ID,
					ProjectID:               foundExistingItem.ProjectID,
					Name:                    item.Name,
					Value:                   item.Value,
					Sensitive:               item.Sensitive,
					Position:                item.Position,
					Category:                item.Category,
					SecretManagerConfigID:   item.SecretManagerConfigID,
					SecretManagerName:       item.SecretManagerName,
					SecretManagerVersion:    item.SecretManagerVersion,
					SecretManagerLastSyncAt: item.SecretManagerLastSyncAt,
					CreatedBy:               foundExistingItem.CreatedBy,
					CreatedAt:               foundExistingItem.CreatedAt,
					UpdatedBy:               userID,
					UpdatedAt:               time.Now(),
				})
			}
		} else {
			itemsToSave = append(itemsToSave, models.ConfigItem{
				ProjectID:               projectId,
				Name:                    item.Name,
				Value:                   item.Value,
				Sensitive:               item.Sensitive,
				Position:                item.Position,
				Category:                item.Category,
				SecretManagerConfigID:   item.SecretManagerConfigID,
				SecretManagerName:       item.SecretManagerName,
				SecretManagerVersion:    item.SecretManagerVersion,
				SecretManagerLastSyncAt: item.SecretManagerLastSyncAt,
				CreatedBy:               userID,
				CreatedAt:               time.Now(),
				UpdatedBy:               userID,
				UpdatedAt:               time.Now(),
			})
		}
	}

	for _, existingItem := range existingItems {
		var foundItem *models.ConfigItem
		for _, item := range req.Items {
			if item.ID == existingItem.ID {
				foundItem = &existingItem
				break
			}
		}
		if foundItem == nil {
			itemsToDelete = append(itemsToDelete, existingItem.ID)
		}
	}

	err = database.DB.Transaction(func(tx *gorm.DB) error {

		if len(itemsToSave) > 0 {
			if err := tx.Save(&itemsToSave).Error; err != nil {
				return err
			}
		}

		if len(itemsToDelete) > 0 {
			if err := tx.Unscoped().Delete(&[]models.ConfigItem{}, itemsToDelete).Error; err != nil {
				return err
			}
		}

		var finalItems []models.ConfigItem
		if err := tx.Where("project_id = ?", projectId).Order("position asc").Find(&finalItems).Error; err != nil {
			return err
		}

		checksum := computeConfigChecksum(finalItems)
		if err := tx.Model(&models.Project{}).Where("id = ?", projectId).Update("config_checksum", checksum).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sync failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config synced successfully"})
}

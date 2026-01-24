package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TeamEncryptedKeyEntry - Project key encrypted for team
type TeamEncryptedKeyEntry struct {
	TeamID              string `json:"teamId"`
	EncryptedProjectKey string `json:"encryptedProjectKey"`
}

// ReEncryptedConfigItem - config item reencrypted with new key
type ReEncryptedConfigItem struct {
	ID    string `json:"id"`
	Value string `json:"value"` // Re-encrypted value
}

// ReEncryptedFileFEK - file key reencrypted with new key
type ReEncryptedFileFEK struct {
	ID           string `json:"id"`
	EncryptedFEK string `json:"encryptedFek"`
}

// Init request for rotation
type InitiateRotationRequest struct {
	TeamEncryptedKeys      []TeamEncryptedKeyEntry  `json:"teamEncryptedKeys" binding:"required"`
	ReEncryptedConfigItems []ReEncryptedConfigItem  `json:"reEncryptedConfigItems" binding:"required"`
	ReEncryptedFileFEKs    []ReEncryptedFileFEK     `json:"reEncryptedFileFEKs"`
}

func GetPendingRotation(c *gin.Context) {
	projectID := c.Param("id")
	uid, _ := c.Get("user_id")
	userID := uid.(uuid.UUID)

	access, err := GetUserProjectAccess(userID, uuid.MustParse(projectID))
	if err != nil || access == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var pending models.PendingKeyRotation
	err = database.DB.
		Preload("Initiator").
		Preload("Approvals").
		Preload("Approvals.User").
		Where("project_id = ? AND status = ?", projectID, "pending").
		First(&pending).Error

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"pending": nil})
		return
	}

	isStale, _ := checkRotationStaleness(&pending)
	if isStale {
		database.DB.Model(&pending).Update("status", "stale")
		c.JSON(http.StatusOK, gin.H{"pending": nil, "staleRotationExists": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending": pending})
}

func InitiateKeyRotation(c *gin.Context) {
	projectID := c.Param("id")
	uid, _ := c.Get("user_id")
	userID := uid.(uuid.UUID)

	access, err := GetUserProjectAccess(userID, uuid.MustParse(projectID))
	if err != nil || access == nil || !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project admins can rotate keys"})
		return
	}

	var existingPending models.PendingKeyRotation
	if err := database.DB.Where("project_id = ? AND status = ?", projectID, "pending").First(&existingPending).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "A key rotation is already pending for this project"})
		return
	}

	var req InitiateRotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	currentConfigItems, currentTeamIDs, currentSecretManagerConfigIDs, configItemsHash := getProjectSnapshot(uuid.MustParse(projectID))

	if err := validateConfigItemsComplete(req.ReEncryptedConfigItems, currentConfigItems); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateTeamsComplete(req.TeamEncryptedKeys, currentTeamIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newVersion := project.KeyVersion + 1

	requiredApprovals := getRequiredApprovals(uuid.MustParse(projectID), project.OrganizationID)

	teamKeysJSON, _ := json.Marshal(req.TeamEncryptedKeys)
	configsJSON, _ := json.Marshal(req.ReEncryptedConfigItems)
	configItemIDsJSON, _ := json.Marshal(extractConfigItemIDs(currentConfigItems))
	teamIDsJSON, _ := json.Marshal(currentTeamIDs)
	secretManagerConfigIDsJSON, _ := json.Marshal(currentSecretManagerConfigIDs)
	fileFEKsJSON, _ := json.Marshal(req.ReEncryptedFileFEKs)

	pending := models.PendingKeyRotation{
		ProjectID:                    uuid.MustParse(projectID),
		InitiatedBy:                  userID,
		NewVersion:                   newVersion,
		Status:                       "pending",
		RequiredApprovals:            requiredApprovals,
		ExpiresAt:                    time.Now().Add(24 * time.Hour), // 24 hour expiry
		TeamEncryptedKeys:            string(teamKeysJSON),
		EncryptedConfigsSnapshot:     string(configsJSON),
		EncryptedFileFEKsSnapshot:    string(fileFEKsJSON),
		SnapshotConfigItemIDs:        string(configItemIDsJSON),
		SnapshotTeamIDs:              string(teamIDsJSON),
		SnapshotSecretManagerConfIDs: string(secretManagerConfigIDsJSON),
		SnapshotConfigItemsHash:      configItemsHash,
	}

	var tokenCount int64
	database.DB.Model(&models.ProjectToken{}).Where("project_id = ?", projectID).Count(&tokenCount)

	if requiredApprovals == 0 {
		if err := commitRotation(&pending, &project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit rotation: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":              "Key rotation completed immediately (single admin)",
			"newVersion":           newVersion,
			"committed":            true,
			"tokensInvalidated":    tokenCount,
		})
		return
	}

	if err := database.DB.Create(&pending).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create pending rotation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Key rotation initiated, awaiting approval",
		"rotationId":           pending.ID,
		"requiredApprovals":    requiredApprovals,
		"expiresAt":            pending.ExpiresAt,
		"committed":            false,
		"tokensToBeInvalidated": tokenCount,
	})
}

func ApproveKeyRotation(c *gin.Context) {
	projectID := c.Param("id")
	rotationID := c.Param("rotationId")
	uid, _ := c.Get("user_id")
	userID := uid.(uuid.UUID)

	access, err := GetUserProjectAccess(userID, uuid.MustParse(projectID))
	if err != nil || access == nil || !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project admins can approve rotations"})
		return
	}

	var req struct {
		VerifiedDecryption bool `json:"verifiedDecryption"`
	}
	c.ShouldBindJSON(&req)

	var pending models.PendingKeyRotation
	if err := database.DB.Preload("Approvals").First(&pending, "id = ? AND project_id = ? AND status = ?", rotationID, projectID, "pending").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending rotation not found"})
		return
	}

	if pending.InitiatedBy == userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot approve your own key rotation"})
		return
	}

	for _, approval := range pending.Approvals {
		if approval.UserID == userID {
			c.JSON(http.StatusConflict, gin.H{"error": "You have already voted on this rotation"})
			return
		}
	}

	if time.Now().After(pending.ExpiresAt) {
		database.DB.Model(&pending).Update("status", "expired")
		c.JSON(http.StatusGone, gin.H{"error": "Rotation has expired"})
		return
	}

	isStale, reason := checkRotationStaleness(&pending)
	if isStale {
		database.DB.Model(&pending).Update("status", "stale")
		c.JSON(http.StatusConflict, gin.H{"error": "Rotation is stale: " + reason})
		return
	}

	approval := models.KeyRotationApproval{
		RotationID:         pending.ID,
		UserID:             userID,
		Approved:           true,
		VerifiedDecryption: req.VerifiedDecryption,
	}
	database.DB.Create(&approval)

	var approvalCount int64
	database.DB.Model(&models.KeyRotationApproval{}).
		Where("rotation_id = ? AND approved = ?", pending.ID, true).
		Count(&approvalCount)

	if int(approvalCount) >= pending.RequiredApprovals {
		isStale, reason := checkRotationStaleness(&pending)
		if isStale {
			database.DB.Model(&pending).Update("status", "stale")
			c.JSON(http.StatusConflict, gin.H{"error": "Rotation became stale: " + reason})
			return
		}

		var project models.Project
		database.DB.First(&project, "id = ?", projectID)

		if err := commitRotation(&pending, &project); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit rotation: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Rotation approved and committed",
			"newVersion": pending.NewVersion,
			"committed":  true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Approval recorded",
		"currentApprovals":  approvalCount,
		"requiredApprovals": pending.RequiredApprovals,
		"committed":         false,
	})
}

func RejectKeyRotation(c *gin.Context) {
	projectID := c.Param("id")
	rotationID := c.Param("rotationId")
	uid, _ := c.Get("user_id")
	userID := uid.(uuid.UUID)

	access, err := GetUserProjectAccess(userID, uuid.MustParse(projectID))
	if err != nil || access == nil || !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project admins can reject rotations"})
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	c.ShouldBindJSON(&req)

	var pending models.PendingKeyRotation
	if err := database.DB.First(&pending, "id = ? AND project_id = ? AND status = ?", rotationID, projectID, "pending").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending rotation not found"})
		return
	}

	rejection := models.KeyRotationApproval{
		RotationID: pending.ID,
		UserID:     userID,
		Approved:   false,
		Comment:    req.Comment,
	}
	database.DB.Create(&rejection)

	database.DB.Model(&pending).Update("status", "rejected")

	c.JSON(http.StatusOK, gin.H{"message": "Rotation rejected"})
}

func CancelKeyRotation(c *gin.Context) {
	projectID := c.Param("id")
	rotationID := c.Param("rotationId")
	uid, _ := c.Get("user_id")
	userID := uid.(uuid.UUID)

	var pending models.PendingKeyRotation
	if err := database.DB.First(&pending, "id = ? AND project_id = ? AND status = ?", rotationID, projectID, "pending").Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending rotation not found"})
		return
	}

	if pending.InitiatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the initiator can cancel a rotation"})
		return
	}

	database.DB.Model(&pending).Update("status", "cancelled")
	c.JSON(http.StatusOK, gin.H{"message": "Rotation cancelled"})
}

func GetUserPendingRotations(c *gin.Context) {
	uid, _ := c.Get("user_id")
	userID := uid.(uuid.UUID)

	projectIDs := getUserAccessibleProjectIDs(userID)

	if len(projectIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"pendingRotations": []models.PendingKeyRotation{}})
		return
	}

	var pendingRotations []models.PendingKeyRotation
	database.DB.
		Preload("Initiator").
		Preload("Project").
		Preload("Approvals").
		Preload("Approvals.User").
		Where("project_id IN ? AND status = ?", projectIDs, "pending").
		Find(&pendingRotations)

	var validRotations []models.PendingKeyRotation
	for i := range pendingRotations {
		isStale, _ := checkRotationStaleness(&pendingRotations[i])
		if isStale {
			database.DB.Model(&pendingRotations[i]).Update("status", "stale")
			continue
		}

		alreadyVoted := false
		for _, approval := range pendingRotations[i].Approvals {
			if approval.UserID == userID {
				alreadyVoted = true
				break
			}
		}
		if pendingRotations[i].InitiatedBy != userID && !alreadyVoted {
			validRotations = append(validRotations, pendingRotations[i])
		}
	}

	c.JSON(http.StatusOK, gin.H{"pendingRotations": validRotations})
}

func commitRotation(pending *models.PendingKeyRotation, project *models.Project) error {
	tx := database.DB.Begin()

	if err := tx.Model(project).Updates(map[string]any{
		"key_version": pending.NewVersion,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var reEncryptedItems []ReEncryptedConfigItem
	json.Unmarshal([]byte(pending.EncryptedConfigsSnapshot), &reEncryptedItems)

	for _, item := range reEncryptedItems {
		if err := tx.Model(&models.ConfigItem{}).
			Where("id = ?", item.ID).
			Update("value", item.Value).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	var teamKeys []TeamEncryptedKeyEntry
	json.Unmarshal([]byte(pending.TeamEncryptedKeys), &teamKeys)

	for _, tk := range teamKeys {
		if err := tx.Model(&models.TeamProject{}).
			Where("team_id = ? AND project_id = ?", tk.TeamID, project.ID).
			Update("encrypted_project_key", tk.EncryptedProjectKey).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if pending.EncryptedFileFEKsSnapshot != "" {
		var reEncryptedFileFEKs []ReEncryptedFileFEK
		json.Unmarshal([]byte(pending.EncryptedFileFEKsSnapshot), &reEncryptedFileFEKs)

		for _, fileFEK := range reEncryptedFileFEKs {
			if err := tx.Model(&models.ProjectFile{}).
				Where("id = ?", fileFEK.ID).
				Update("encrypted_fek", fileFEK.EncryptedFEK).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if pending.ID != uuid.Nil {
		if err := tx.Model(pending).Update("status", "approved").Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Where("project_id = ?", project.ID).Delete(&models.ProjectToken{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func getRequiredApprovals(projectID uuid.UUID, orgID uuid.UUID) int {
	var adminCount int64

	database.DB.Model(&models.OrganizationUser{}).
		Where("organization_id = ? AND (role = 'owner' OR role = 'Owner' OR role = 'admin')", orgID).
		Count(&adminCount)


	// Count team admins who aren't already org admins (we don't want to double-count them)
	var teamAdminCount int64
	database.DB.Raw(`
		SELECT COUNT(DISTINCT tu.user_id)
		FROM team_users tu
		JOIN team_projects tp ON tu.team_id = tp.team_id
		WHERE tp.project_id = ?
		AND (tu.role = 'owner' OR tu.role = 'admin')
		AND tu.user_id NOT IN (
			SELECT user_id FROM organization_users
			WHERE organization_id = ? AND (role = 'owner' OR role = 'Owner' OR role = 'admin')
		)
	`, projectID, orgID).Scan(&teamAdminCount)

	totalAdmins := int(adminCount + teamAdminCount)

	if totalAdmins <= 1 {
		return 0
	}

	// Todo add org policy here
	return 1
}

func getProjectSnapshot(projectID uuid.UUID) ([]models.ConfigItem, []string, []string, string) {
	var configItems []models.ConfigItem
	database.DB.Where("project_id = ?", projectID).Find(&configItems)

	var teamProjects []models.TeamProject
	database.DB.Where("project_id = ?", projectID).Find(&teamProjects)
	teamIDs := make([]string, len(teamProjects))
	for i, tp := range teamProjects {
		teamIDs[i] = tp.TeamID.String()
	}
	sort.Strings(teamIDs)

	var secretManagerConfigs []models.SecretManagerConfig
	database.DB.Where("project_id = ?", projectID).Find(&secretManagerConfigs)
	secretManagerConfigIDs := make([]string, len(secretManagerConfigs))
	for i, smc := range secretManagerConfigs {
		secretManagerConfigIDs[i] = smc.ID.String()
	}
	sort.Strings(secretManagerConfigIDs)

	configItemsHash := hashConfigItems(configItems)

	return configItems, teamIDs, secretManagerConfigIDs, configItemsHash
}

func extractConfigItemIDs(items []models.ConfigItem) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID.String()
	}
	sort.Strings(ids)
	return ids
}

func hashConfigItems(items []models.ConfigItem) string {
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID.String() < items[j].ID.String()
	})

	hasher := sha256.New()
	for _, item := range items {
		hasher.Write([]byte(item.ID.String()))
		hasher.Write([]byte(item.Value))
		hasher.Write([]byte(item.Name))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func validateConfigItemsComplete(requested []ReEncryptedConfigItem, current []models.ConfigItem) error {
	if len(requested) != len(current) {
		return &ValidationError{"Number of config items doesn't match. Expected " + string(rune(len(current))) + " but got " + string(rune(len(requested)))}
	}

	requestedIDs := make(map[string]bool)
	for _, item := range requested {
		requestedIDs[item.ID] = true
	}

	for _, item := range current {
		if !requestedIDs[item.ID.String()] {
			return &ValidationError{"Missing config item: " + item.Name}
		}
	}

	return nil
}

func validateTeamsComplete(requested []TeamEncryptedKeyEntry, currentTeamIDs []string) error {
	if len(requested) != len(currentTeamIDs) {
		return &ValidationError{"Number of teams doesn't match"}
	}

	requestedTeamIDs := make(map[string]bool)
	for _, entry := range requested {
		requestedTeamIDs[entry.TeamID] = true
	}

	for _, teamID := range currentTeamIDs {
		if !requestedTeamIDs[teamID] {
			return &ValidationError{"Missing encrypted key for team: " + teamID}
		}
	}

	return nil
}

func checkRotationStaleness(pending *models.PendingKeyRotation) (bool, string) {
	currentConfigItems, currentTeamIDs, currentSecretManagerConfigIDs, currentHash := getProjectSnapshot(pending.ProjectID)

	var snapshotConfigItemIDs []string
	json.Unmarshal([]byte(pending.SnapshotConfigItemIDs), &snapshotConfigItemIDs)
	currentConfigItemIDs := extractConfigItemIDs(currentConfigItems)

	if !stringSlicesEqual(snapshotConfigItemIDs, currentConfigItemIDs) {
		return true, "config items have changed"
	}

	var snapshotTeamIDs []string
	json.Unmarshal([]byte(pending.SnapshotTeamIDs), &snapshotTeamIDs)

	if !stringSlicesEqual(snapshotTeamIDs, currentTeamIDs) {
		return true, "teams with access have changed"
	}

	var snapshotSecretManagerConfigIDs []string
	json.Unmarshal([]byte(pending.SnapshotSecretManagerConfIDs), &snapshotSecretManagerConfigIDs)

	if !stringSlicesEqual(snapshotSecretManagerConfigIDs, currentSecretManagerConfigIDs) {
		return true, "secret manager configurations have changed"
	}

	if pending.SnapshotConfigItemsHash != currentHash {
		return true, "config item values have changed"
	}

	return false, ""
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getUserAccessibleProjectIDs(userID uuid.UUID) []uuid.UUID {
	var projectIDs []uuid.UUID

	// Projects user can access through their team memberships
	database.DB.Raw(`
		SELECT DISTINCT tp.project_id
		FROM team_projects tp
		JOIN team_users tu ON tp.team_id = tu.team_id
		WHERE tu.user_id = ?
	`, userID).Scan(&projectIDs)

	// Org admins/owners can see all projects in their org, even if not in a team
	var orgProjectIDs []uuid.UUID
	database.DB.Raw(`
		SELECT DISTINCT p.id
		FROM projects p
		WHERE p.organization_id IN (
			SELECT organization_id FROM organization_users
			WHERE user_id = ? AND (role = 'owner' OR role = 'Owner' OR role = 'admin')
		)
	`, userID).Scan(&orgProjectIDs)

	idSet := make(map[uuid.UUID]bool)
	for _, id := range projectIDs {
		idSet[id] = true
	}
	for _, id := range orgProjectIDs {
		idSet[id] = true
	}

	result := make([]uuid.UUID, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}

	return result
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

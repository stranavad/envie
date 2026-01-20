package handlers

import (
	"net/http"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateOrganizationRequest struct {
	Name                        string `json:"name" binding:"required"`
	EncryptedOrganizationKey    string `json:"encryptedOrganizationKey" binding:"required"`    // org master encrypted with user private key
	GeneralTeamEncryptedKey     string `json:"generalTeamEncryptedKey" binding:"required"`     // encrypted first team key
	GeneralTeamUserEncryptedKey string `json:"generalTeamUserEncryptedKey" binding:"required"` // encrypted first user to first team binding key
}

func CreateOrganization(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	org := models.Organization{
		Name: req.Name,
	}

	if err := tx.Create(&org).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	orgUser := models.OrganizationUser{
		OrganizationID:           org.ID,
		UserID:                   uid,
		Role:                     "owner",
		EncryptedOrganizationKey: &req.EncryptedOrganizationKey,
	}

	if err := tx.Create(&orgUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to organization"})
		return
	}

	generalTeam := models.Team{
		OrganizationID: org.ID,
		Name:           "General",
		EncryptedKey:   req.GeneralTeamEncryptedKey,
	}

	if err := tx.Create(&generalTeam).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create general team"})
		return
	}

	teamUser := models.TeamUser{
		TeamID:           generalTeam.ID,
		UserID:           uid,
		Role:             "owner",
		EncryptedTeamKey: req.GeneralTeamUserEncryptedKey,
	}

	if err := tx.Create(&teamUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to general team"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, org)
}

func GetOrganizations(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var orgs []models.Organization

	err := database.DB.Model(&models.Organization{}).
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("organization_users.user_id = ?", uid).
		Preload("Teams").
		Find(&orgs).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	type OrganizationResponse struct {
		models.Organization
		Role string `json:"role"`
		ProjectCount int64 `json:"projectCount"`
		MemberCount  int64 `json:"memberCount"`
	}

	response := []OrganizationResponse{}
	for _, org := range orgs {
		var role string
		database.DB.Model(&models.OrganizationUser{}).
			Select("role").
			Where("organization_id = ? AND user_id = ?", org.ID, uid).
			Scan(&role)

		var memberCount int64
		database.DB.Model(&models.OrganizationUser{}).Where("organization_id = ?", org.ID).Count(&memberCount)

		var projectCount int64
		database.DB.Model(&models.Project{}).Where("organization_id = ?", org.ID).Count(&projectCount)

		response = append(response, OrganizationResponse{
			Organization: org,
			Role:         role,
			MemberCount:  memberCount,
			ProjectCount: projectCount,
		})
	}

	c.JSON(http.StatusOK, response)
}

func GetOrganization(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	orgIDStr := c.Param("id")
	orgID, err := uuid.Parse(orgIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var org models.Organization
	err = database.DB.Model(&models.Organization{}).
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("organizations.id = ? AND organization_users.user_id = ?", orgID, uid).
		First(&org).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	var orgUser models.OrganizationUser
	database.DB.Where("organization_id = ? AND user_id = ?", org.ID, uid).First(&orgUser)

	c.JSON(http.StatusOK, gin.H{
		"organization":             org,
		"role":                     orgUser.Role,
		"encryptedOrganizationKey": orgUser.EncryptedOrganizationKey,
	})
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
}

func UpdateOrganization(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	orgIDStr := c.Param("id")
	orgID, err := uuid.Parse(orgIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, uid).First(&orgUser).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if orgUser.Role != "owner" && orgUser.Role != "Owner" && orgUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only organization owners and admins can update organization settings"})
		return
	}

	if err := database.DB.Model(&models.Organization{}).Where("id = ?", orgID).Update("name", req.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Organization updated"})
}

func GetOrganizationUsers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	orgIDStr := c.Param("id")
	orgID, err := uuid.Parse(orgIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var requester models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, uid).First(&requester).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var users []models.User
	if err := database.DB.Model(&models.User{}).
		Joins("JOIN organization_users ON organization_users.user_id = users.id").
		Where("organization_users.organization_id = ?", orgID).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization users"})
		return
	}

	type OrgUserResponse struct {
		models.User
		Role string `json:"role"`
	}

	response := []OrgUserResponse{}
	for _, u := range users {
		var role string
		database.DB.Model(&models.OrganizationUser{}).
			Select("role").
			Where("organization_id = ? AND user_id = ?", orgID, u.ID).
			Scan(&role)

		response = append(response, OrgUserResponse{
			User: u,
			Role: role,
		})
	}

	c.JSON(http.StatusOK, response)
}

type AddOrganizationMemberRequest struct {
	UserID                   uuid.UUID `json:"userId" binding:"required"`
	Role                     string    `json:"role"`
	EncryptedOrganizationKey *string   `json:"encryptedOrganizationKey"`
}

func AddOrganizationMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	requesterUID := userID.(uuid.UUID)
	orgIDStr := c.Param("id")
	orgID, err := uuid.Parse(orgIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var req AddOrganizationMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	validRoles := map[string]bool{"owner": true, "admin": true, "member": true}
	if !validRoles[req.Role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be owner, admin, or member"})
		return
	}

	var requesterOrgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, requesterUID).First(&requesterOrgUser).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if requesterOrgUser.Role != "owner" && requesterOrgUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only organization owners and admins can add members"})
		return
	}

	if req.Role == "owner" && requesterOrgUser.Role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only organization owners can add other owners"})
		return
	}

	// Verify target user exists and has public key
	var targetUser models.User
	if err := database.DB.First(&targetUser, "id = ?", req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target user not found"})
		return
	}

	if targetUser.PublicKey == nil || *targetUser.PublicKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user has not set up encryption keys"})
		return
	}

	var existingMembership models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, req.UserID).First(&existingMembership).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this organization"})
		return
	}

	if (req.Role == "admin" || req.Role == "owner") && (req.EncryptedOrganizationKey == nil || *req.EncryptedOrganizationKey == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "encryptedOrganizationKey is required for admin and owner roles"})
		return
	}

	orgUser := models.OrganizationUser{
		OrganizationID:           orgID,
		UserID:                   req.UserID,
		Role:                     req.Role,
		EncryptedOrganizationKey: req.EncryptedOrganizationKey,
	}

	if err := database.DB.Create(&orgUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to organization"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Member added successfully",
		"userId":  req.UserID,
		"role":    req.Role,
	})
}

package handlers

import (
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
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	tx := database.DB.Begin()

	org := models.Organization{
		Name: req.Name,
	}

	if err := tx.Create(&org).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to create organization")
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
		RespondInternalError(c, "Failed to add user to organization")
		return
	}

	generalTeam := models.Team{
		OrganizationID: org.ID,
		Name:           "General",
		EncryptedKey:   req.GeneralTeamEncryptedKey,
	}

	if err := tx.Create(&generalTeam).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to create general team")
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
		RespondInternalError(c, "Failed to add user to general team")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to commit transaction")
		return
	}

	RespondCreated(c, org)
}

func GetOrganizations(c *gin.Context) {
	userID, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	type OrganizationResponse struct {
		models.Organization
		Role         string `json:"role"`
		ProjectCount int64  `json:"projectCount"`
		MemberCount  int64  `json:"memberCount"`
	}

	var response []OrganizationResponse
	if err := database.DB.Raw(`
		SELECT
			organizations.*, organization_users.role as role,
			COUNT(DISTINCT(projects.id)) as project_count,
			COUNT(DISTINCT(organization_users.user_id)) as member_count
		FROM organizations
			  LEFT JOIN organization_users ON organization_users.organization_id = organizations.id
			  LEFT JOIN projects ON projects.organization_id = organizations.id
		WHERE organization_users.user_id = ?
		GROUP BY organizations.id, organization_users.role
	`, userID).Scan(&response).Error; err != nil {
		RespondInternalError(c, "Failed to load organizations")
		return
	}

	RespondOK(c, response)
}

func GetOrganization(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	type OrgWithUserInfo struct {
		models.Organization
		Role                     string
		EncryptedOrganizationKey *string
	}
	var result OrgWithUserInfo
	err := database.DB.Model(&models.Organization{}).
		Select("organizations.*, organization_users.role, organization_users.encrypted_organization_key").
		Joins("JOIN organization_users ON organization_users.organization_id = organizations.id").
		Where("organizations.id = ? AND organization_users.user_id = ?", orgID, uid).
		Scan(&result).Error

	if err != nil || result.ID == uuid.Nil {
		RespondNotFound(c, "Organization not found")
		return
	}

	RespondOK(c, gin.H{
		"organization":             result.Organization,
		"role":                     result.Role,
		"encryptedOrganizationKey": result.EncryptedOrganizationKey,
	})
}

type UpdateOrganizationRequest struct {
	Name string `json:"name" binding:"required"`
}

func UpdateOrganization(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	_, ok = RequireOrgAdmin(c, uid, orgID)
	if !ok {
		return
	}

	if err := database.DB.Model(&models.Organization{}).Where("id = ?", orgID).Update("name", req.Name).Error; err != nil {
		RespondInternalError(c, "Failed to update organization")
		return
	}

	RespondMessage(c, "Organization updated")
}

func GetOrganizationUsers(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	_, ok = RequireOrgMembership(c, uid, orgID)
	if !ok {
		return
	}

	// Single query to get users with their roles
	type UserWithRole struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		AvatarURL string    `json:"avatarUrl"`
		PublicKey *string   `json:"publicKey"`
		CreatedAt string    `json:"createdAt"`
		UpdatedAt string    `json:"updatedAt"`
		Role      string    `json:"role"`
	}

	var users []UserWithRole
	if err := database.DB.Model(&models.User{}).
		Select("users.id, users.name, users.email, users.avatar_url, users.public_key, users.created_at, users.updated_at, organization_users.role").
		Joins("JOIN organization_users ON organization_users.user_id = users.id").
		Where("organization_users.organization_id = ?", orgID).
		Scan(&users).Error; err != nil {
		RespondInternalError(c, "Failed to fetch organization users")
		return
	}

	RespondOK(c, users)
}

type AddOrganizationMemberRequest struct {
	UserID                   uuid.UUID `json:"userId" binding:"required"`
	Role                     string    `json:"role"`
	EncryptedOrganizationKey *string   `json:"encryptedOrganizationKey"`
}

func AddOrganizationMember(c *gin.Context) {
	requesterUID, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	var req AddOrganizationMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	if !IsValidRole(req.Role) {
		RespondBadRequest(c, "Invalid role. Must be owner, admin, or member")
		return
	}

	requesterOrgUser, ok := RequireOrgAdmin(c, requesterUID, orgID)
	if !ok {
		return
	}

	if req.Role == "owner" && !IsOwner(requesterOrgUser.Role) {
		RespondForbidden(c, "Only organization owners can add other owners")
		return
	}

	// Verify target user exists and has public key
	var targetUser models.User
	if err := database.DB.First(&targetUser, "id = ?", req.UserID).Error; err != nil {
		RespondNotFound(c, "Target user not found")
		return
	}

	if targetUser.PublicKey == nil || *targetUser.PublicKey == "" {
		RespondBadRequest(c, "Target user has not set up encryption keys")
		return
	}

	var existingMembership models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, req.UserID).First(&existingMembership).Error; err == nil {
		RespondConflict(c, "User is already a member of this organization")
		return
	}

	if (req.Role == "admin" || req.Role == "owner") && (req.EncryptedOrganizationKey == nil || *req.EncryptedOrganizationKey == "") {
		RespondBadRequest(c, "encryptedOrganizationKey is required for admin and owner roles")
		return
	}

	orgUser := models.OrganizationUser{
		OrganizationID:           orgID,
		UserID:                   req.UserID,
		Role:                     req.Role,
		EncryptedOrganizationKey: req.EncryptedOrganizationKey,
	}

	if err := database.DB.Create(&orgUser).Error; err != nil {
		RespondInternalError(c, "Failed to add member to organization")
		return
	}

	RespondCreated(c, gin.H{
		"message": "Member added successfully",
		"userId":  req.UserID,
		"role":    req.Role,
	})
}

type UpdateOrganizationMemberRequest struct {
	Role                     string  `json:"role" binding:"required"`
	EncryptedOrganizationKey *string `json:"encryptedOrganizationKey"`
}

func UpdateOrganizationMember(c *gin.Context) {
	requesterUID, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	targetUserID, ok := ParseUUIDParam(c, "userId", "user")
	if !ok {
		return
	}

	var req UpdateOrganizationMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	if !IsValidRole(req.Role) {
		RespondBadRequest(c, "Invalid role. Must be owner, admin, or member")
		return
	}

	requesterOrgUser, ok := RequireOrgAdmin(c, requesterUID, orgID)
	if !ok {
		return
	}

	var targetOrgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, targetUserID).First(&targetOrgUser).Error; err != nil {
		RespondNotFound(c, "Member not found")
		return
	}

	if (IsOwner(targetOrgUser.Role) || req.Role == "owner") && !IsOwner(requesterOrgUser.Role) {
		RespondForbidden(c, "Only organization owners can modify owner roles")
		return
	}

	if requesterUID == targetUserID && IsOwner(targetOrgUser.Role) && req.Role != "owner" {
		var ownerCount int64
		database.DB.Model(&models.OrganizationUser{}).Where("organization_id = ? AND role = ?", orgID, "owner").Count(&ownerCount)
		if ownerCount <= 1 {
			RespondBadRequest(c, "Cannot demote the last owner")
			return
		}
	}

	if (req.Role == "admin" || req.Role == "owner") && targetOrgUser.Role == "member" {
		if req.EncryptedOrganizationKey == nil || *req.EncryptedOrganizationKey == "" {
			RespondBadRequest(c, "encryptedOrganizationKey is required when promoting to admin or owner")
			return
		}
	}

	updates := map[string]any{"role": req.Role}
	if req.Role == "member" {
		updates["encrypted_organization_key"] = nil
	} else if req.EncryptedOrganizationKey != nil {
		updates["encrypted_organization_key"] = *req.EncryptedOrganizationKey
	}

	if err := database.DB.Model(&targetOrgUser).Updates(updates).Error; err != nil {
		RespondInternalError(c, "Failed to update member")
		return
	}

	RespondOK(c, gin.H{
		"message": "Member updated successfully",
		"userId":  targetUserID,
		"role":    req.Role,
	})
}

func RemoveOrganizationMember(c *gin.Context) {
	requesterUID, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	targetUserID, ok := ParseUUIDParam(c, "userId", "user")
	if !ok {
		return
	}

	requesterOrgUser, ok := RequireOrgAdmin(c, requesterUID, orgID)
	if !ok {
		return
	}

	var targetOrgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, targetUserID).First(&targetOrgUser).Error; err != nil {
		RespondNotFound(c, "Member not found")
		return
	}

	if IsOwner(targetOrgUser.Role) && !IsOwner(requesterOrgUser.Role) {
		RespondForbidden(c, "Only organization owners can remove other owners")
		return
	}

	if requesterUID == targetUserID && IsOwner(targetOrgUser.Role) {
		var ownerCount int64
		database.DB.Model(&models.OrganizationUser{}).Where("organization_id = ? AND role = ?", orgID, "owner").Count(&ownerCount)
		if ownerCount <= 1 {
			RespondBadRequest(c, "Cannot remove the last owner")
			return
		}
	}

	tx := database.DB.Begin()

	if err := tx.Where("user_id = ? AND team_id IN (SELECT id FROM teams WHERE organization_id = ?)", targetUserID, orgID).Delete(&models.TeamUser{}).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to remove member from teams")
		return
	}

	if err := tx.Delete(&targetOrgUser).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to remove member")
		return
	}

	if err := tx.Commit().Error; err != nil {
		RespondInternalError(c, "Failed to commit transaction")
		return
	}

	RespondOK(c, gin.H{
		"message": "Member removed successfully",
		"userId":  targetUserID,
	})
}

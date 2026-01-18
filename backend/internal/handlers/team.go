package handlers

import (
	"net/http"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateTeamRequest struct {
	Name             string    `json:"name" binding:"required"`
	OrganizationID   uuid.UUID `json:"organizationId" binding:"required"`
	EncryptedKey     string    `json:"encryptedKey" binding:"required"`     // encrypted with org master key
	UserEncryptedKey string    `json:"userEncryptedKey" binding:"required"` // encrypted with user PK
}

func CreateTeam(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var orgUser models.OrganizationUser
	if err := tx.Where("organization_id = ? AND user_id = ?", req.OrganizationID, uid).First(&orgUser).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this organization"})
		return
	}

	// 2. Create Team
	team := models.Team{
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
		EncryptedKey:   req.EncryptedKey,
	}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(http.StatusCreated, team)
}

func GetTeams(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	orgIDStr := c.Query("organizationId")

	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organizationId query parameter required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	var orgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, uid).First(&orgUser).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var teams []models.Team
	if err := database.DB.Where("organization_id = ?", orgID).Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	type TeamResponse struct {
		models.Team
		MemberCount  int64 `json:"memberCount"`
		ProjectCount int64 `json:"projectCount"`
		PreviewUsers     []models.User `json:"previewUsers"`
		UserEncryptedKey string        `json:"userEncryptedKey"`
	}

	response := []TeamResponse{}
	for _, team := range teams {
		var memberCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ?", team.ID).Count(&memberCount)

		var projectCount int64
		database.DB.Model(&models.TeamProject{}).Where("team_id = ?", team.ID).Count(&projectCount)

		var previewUsers []models.User
		// Get first 5 users
		database.DB.Model(&models.User{}).
			Joins("JOIN team_users ON team_users.user_id = users.id").
			Where("team_users.team_id = ?", team.ID).
			Limit(5).
			Find(&previewUsers)

		var userEncryptedKey string
		database.DB.Model(&models.TeamUser{}).
			Select("encrypted_team_key").
			Where("team_id = ? AND user_id = ?", team.ID, uid).
			Scan(&userEncryptedKey)

		response = append(response, TeamResponse{
			Team:             team,
			MemberCount:      memberCount,
			ProjectCount:     projectCount,
			PreviewUsers:     previewUsers,
			UserEncryptedKey: userEncryptedKey,
		})
	}

	c.JSON(http.StatusOK, response)
}

func GetTeamMembers(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	teamIDStr := c.Param("id")

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	var orgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", team.OrganizationID, uid).First(&orgUser).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	type MemberResponse struct {
		UserID    uuid.UUID `json:"userId"`
		Name      string    `json:"name"`
		Email     string    `json:"email"`
		AvatarUrl string    `json:"avatarUrl"`
		Role      string    `json:"role"`
		JoinedAt  string    `json:"joinedAt"`
	}

	var members []MemberResponse
	database.DB.Model(&models.TeamUser{}).
		Select("team_users.user_id, users.name, users.email, users.avatar_url, team_users.role, team_users.created_at as joined_at").
		Joins("JOIN users ON users.id = team_users.user_id").
		Where("team_users.team_id = ?", teamID).
		Scan(&members)

	c.JSON(http.StatusOK, members)
}

type AddTeamMemberRequest struct {
	UserID           uuid.UUID `json:"userId" binding:"required"`
	EncryptedTeamKey string    `json:"encryptedTeamKey" binding:"required"`
	Role             string    `json:"role"`
}

// AddTeamMember adds a user to a team
func AddTeamMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	teamIDStr := c.Param("id")

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	canManage, err := canManageTeam(uid, teamID, team.OrganizationID)
	if err != nil || !canManage {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add members to this team"})
		return
	}

	var targetOrgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", team.OrganizationID, req.UserID).First(&targetOrgUser).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a member of this organization"})
		return
	}

	var existingMember models.TeamUser
	if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, req.UserID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this team"})
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	teamUser := models.TeamUser{
		TeamID:           teamID,
		UserID:           req.UserID,
		EncryptedTeamKey: req.EncryptedTeamKey,
		Role:             role,
	}

	if err := database.DB.Create(&teamUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member to team"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Member added successfully"})
}

type UpdateTeamMemberRequest struct {
	Role string `json:"role" binding:"required"`
}

func UpdateTeamMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	teamIDStr := c.Param("id")
	memberIDStr := c.Param("userId")

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	canManage, err := canManageTeam(uid, teamID, team.OrganizationID)
	if err != nil || !canManage {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to manage this team"})
		return
	}

	if memberID == uid && req.Role != "owner" {
		var ownerCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ? AND role = ?", teamID, "owner").Count(&ownerCount)
		if ownerCount <= 1 {
			var currentMember models.TeamUser
			if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, uid).First(&currentMember).Error; err == nil {
				if currentMember.Role == "owner" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot demote yourself as you are the only team owner"})
					return
				}
			}
		}
	}

	result := database.DB.Model(&models.TeamUser{}).
		Where("team_id = ? AND user_id = ?", teamID, memberID).
		Update("role", req.Role)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member role updated successfully"})
}

func RemoveTeamMember(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	teamIDStr := c.Param("id")
	memberIDStr := c.Param("userId")

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if memberID != uid {
		canManage, err := canManageTeam(uid, teamID, team.OrganizationID)
		if err != nil || !canManage {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to remove members from this team"})
			return
		}
	}

	var memberToRemove models.TeamUser
	if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, memberID).First(&memberToRemove).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	if memberToRemove.Role == "owner" {
		var ownerCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ? AND role = ?", teamID, "owner").Count(&ownerCount)
		if ownerCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the only team owner. Transfer ownership first."})
			return
		}
	}

	result := database.DB.Where("team_id = ? AND user_id = ?", teamID, memberID).Delete(&models.TeamUser{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

type UpdateMyTeamKeyRequest struct {
	EncryptedTeamKey string `json:"encryptedTeamKey" binding:"required"`
}

func UpdateMyTeamKey(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	teamIDStr := c.Param("id")

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req UpdateMyTeamKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Model(&models.TeamUser{}).
		Where("team_id = ? AND user_id = ?", teamID, uid).
		Update("encrypted_team_key", req.EncryptedTeamKey)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not a member of this team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Team key updated successfully"})
}

func GetMyTeams(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	type TeamWithKey struct {
		TeamID           uuid.UUID `json:"teamId"`
		TeamName         string    `json:"teamName"`
		OrganizationID   uuid.UUID `json:"organizationId"`
		EncryptedTeamKey string    `json:"encryptedTeamKey"`
		EncryptedKey     string    `json:"encryptedKey"`
	}

	var teams []TeamWithKey
	database.DB.Model(&models.TeamUser{}).
		Select("team_users.team_id, teams.name as team_name, teams.organization_id, team_users.encrypted_team_key, teams.encrypted_key").
		Joins("JOIN teams ON teams.id = team_users.team_id").
		Where("team_users.user_id = ?", uid).
		Scan(&teams)

	c.JSON(http.StatusOK, teams)
}

func canManageTeam(userID uuid.UUID, teamID uuid.UUID, orgID uuid.UUID) (bool, error) {
	var teamUser models.TeamUser
	if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, userID).First(&teamUser).Error; err == nil {
		if teamUser.Role == "owner" || teamUser.Role == "admin" {
			return true, nil
		}
	}

	var orgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&orgUser).Error; err == nil {
		if orgUser.Role == "owner" || orgUser.Role == "admin" {
			return true, nil
		}
	}

	return false, nil
}

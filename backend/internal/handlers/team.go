package handlers

import (
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
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	tx := database.DB.Begin()

	var orgUser models.OrganizationUser
	if err := tx.Where("organization_id = ? AND user_id = ?", req.OrganizationID, uid).First(&orgUser).Error; err != nil {
		tx.Rollback()
		RespondForbidden(c, "You are not a member of this organization")
		return
	}

	team := models.Team{
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
		EncryptedKey:   req.EncryptedKey,
	}

	if err := tx.Create(&team).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to create team")
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to commit transaction")
		return
	}

	RespondCreated(c, team)
}

func GetTeams(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	orgID, ok := ParseUUIDQuery(c, "organizationId", "organization")
	if !ok {
		return
	}

	_, ok = RequireOrgMembership(c, uid, orgID)
	if !ok {
		return
	}

	var teams []models.Team
	if err := database.DB.Where("organization_id = ?", orgID).Find(&teams).Error; err != nil {
		RespondInternalError(c, "Failed to fetch teams")
		return
	}

	if len(teams) == 0 {
		RespondOK(c, []any{})
		return
	}

	teamIDs := make([]uuid.UUID, len(teams))
	for i, t := range teams {
		teamIDs[i] = t.ID
	}

	type CountResult struct {
		TeamID uuid.UUID
		Count  int64
	}
	var memberCounts []CountResult
	database.DB.Model(&models.TeamUser{}).
		Select("team_id, COUNT(*) as count").
		Where("team_id IN ?", teamIDs).
		Group("team_id").
		Scan(&memberCounts)

	memberCountMap := make(map[uuid.UUID]int64)
	for _, mc := range memberCounts {
		memberCountMap[mc.TeamID] = mc.Count
	}

	var projectCounts []CountResult
	database.DB.Model(&models.TeamProject{}).
		Select("team_id, COUNT(*) as count").
		Where("team_id IN ?", teamIDs).
		Group("team_id").
		Scan(&projectCounts)

	projectCountMap := make(map[uuid.UUID]int64)
	for _, pc := range projectCounts {
		projectCountMap[pc.TeamID] = pc.Count
	}

	type TeamKeyResult struct {
		TeamID           uuid.UUID
		EncryptedTeamKey string
	}
	var teamKeys []TeamKeyResult
	database.DB.Model(&models.TeamUser{}).
		Select("team_id, encrypted_team_key").
		Where("team_id IN ? AND user_id = ?", teamIDs, uid).
		Scan(&teamKeys)

	teamKeyMap := make(map[uuid.UUID]string)
	for _, tk := range teamKeys {
		teamKeyMap[tk.TeamID] = tk.EncryptedTeamKey
	}

	// Fetch all team users for the requested teams
	type TeamUserResult struct {
		TeamID    uuid.UUID
		ID        uuid.UUID
		Name      string
		Email     string
		AvatarURL string `gorm:"column:avatar_url"`
	}
	var teamUserResults []TeamUserResult
	database.DB.Model(&models.TeamUser{}).
		Select("team_users.team_id, users.id, users.name, users.email, users.avatar_url").
		Joins("JOIN users ON users.id = team_users.user_id").
		Where("team_users.team_id IN ?", teamIDs).
		Order("team_users.created_at").
		Scan(&teamUserResults)

	// Group users by team
	teamUsersMap := make(map[uuid.UUID][]models.User)
	for _, tu := range teamUserResults {
		teamUsersMap[tu.TeamID] = append(teamUsersMap[tu.TeamID], models.User{
			ID:        tu.ID,
			Name:      tu.Name,
			Email:     tu.Email,
			AvatarURL: tu.AvatarURL,
		})
	}

	// Build response
	type TeamResponse struct {
		models.Team
		MemberCount      int64         `json:"memberCount"`
		ProjectCount     int64         `json:"projectCount"`
		Users            []models.User `json:"users"`
		UserEncryptedKey string        `json:"userEncryptedKey"`
	}

	response := make([]TeamResponse, len(teams))
	for i, team := range teams {
		users := teamUsersMap[team.ID]
		if users == nil {
			users = []models.User{}
		}
		response[i] = TeamResponse{
			Team:             team,
			MemberCount:      memberCountMap[team.ID],
			ProjectCount:     projectCountMap[team.ID],
			Users:            users,
			UserEncryptedKey: teamKeyMap[team.ID],
		}
	}

	RespondOK(c, response)
}

func GetTeamMembers(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	teamID, ok := ParseUUIDParam(c, "id", "team")
	if !ok {
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		RespondNotFound(c, "Team not found")
		return
	}

	_, ok = RequireOrgMembership(c, uid, team.OrganizationID)
	if !ok {
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

	RespondOK(c, members)
}

type AddTeamMemberRequest struct {
	UserID           uuid.UUID `json:"userId" binding:"required"`
	EncryptedTeamKey string    `json:"encryptedTeamKey" binding:"required"`
	Role             string    `json:"role"`
}

// AddTeamMember adds a user to a team
func AddTeamMember(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	teamID, ok := ParseUUIDParam(c, "id", "team")
	if !ok {
		return
	}

	var req AddTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		RespondNotFound(c, "Team not found")
		return
	}

	canManage, err := canManageTeam(uid, teamID, team.OrganizationID)
	if err != nil || !canManage {
		RespondForbidden(c, "You don't have permission to add members to this team")
		return
	}

	var targetOrgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", team.OrganizationID, req.UserID).First(&targetOrgUser).Error; err != nil {
		RespondBadRequest(c, "User is not a member of this organization")
		return
	}

	var existingMember models.TeamUser
	if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, req.UserID).First(&existingMember).Error; err == nil {
		RespondConflict(c, "User is already a member of this team")
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
		RespondInternalError(c, "Failed to add member to team")
		return
	}

	RespondCreated(c, gin.H{"message": "Member added successfully"})
}

type UpdateTeamMemberRequest struct {
	Role string `json:"role" binding:"required"`
}

func UpdateTeamMember(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	teamID, ok := ParseUUIDParam(c, "id", "team")
	if !ok {
		return
	}

	memberID, ok := ParseUUIDParam(c, "userId", "user")
	if !ok {
		return
	}

	var req UpdateTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		RespondNotFound(c, "Team not found")
		return
	}

	canManage, err := canManageTeam(uid, teamID, team.OrganizationID)
	if err != nil || !canManage {
		RespondForbidden(c, "You don't have permission to manage this team")
		return
	}

	if memberID == uid && req.Role != "owner" {
		var ownerCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ? AND role = ?", teamID, "owner").Count(&ownerCount)
		if ownerCount <= 1 {
			var currentMember models.TeamUser
			if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, uid).First(&currentMember).Error; err == nil {
				if currentMember.Role == "owner" {
					RespondBadRequest(c, "Cannot demote yourself as you are the only team owner")
					return
				}
			}
		}
	}

	result := database.DB.Model(&models.TeamUser{}).
		Where("team_id = ? AND user_id = ?", teamID, memberID).
		Update("role", req.Role)

	if result.RowsAffected == 0 {
		RespondNotFound(c, "Team member not found")
		return
	}

	RespondMessage(c, "Member role updated successfully")
}

func RemoveTeamMember(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	teamID, ok := ParseUUIDParam(c, "id", "team")
	if !ok {
		return
	}

	memberID, ok := ParseUUIDParam(c, "userId", "user")
	if !ok {
		return
	}

	var team models.Team
	if err := database.DB.First(&team, "id = ?", teamID).Error; err != nil {
		RespondNotFound(c, "Team not found")
		return
	}

	if memberID != uid {
		canManage, err := canManageTeam(uid, teamID, team.OrganizationID)
		if err != nil || !canManage {
			RespondForbidden(c, "You don't have permission to remove members from this team")
			return
		}
	}

	var memberToRemove models.TeamUser
	if err := database.DB.Where("team_id = ? AND user_id = ?", teamID, memberID).First(&memberToRemove).Error; err != nil {
		RespondNotFound(c, "Team member not found")
		return
	}

	if memberToRemove.Role == "owner" {
		var ownerCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ? AND role = ?", teamID, "owner").Count(&ownerCount)
		if ownerCount <= 1 {
			RespondBadRequest(c, "Cannot remove the only team owner. Transfer ownership first.")
			return
		}
	}

	result := database.DB.Where("team_id = ? AND user_id = ?", teamID, memberID).Delete(&models.TeamUser{})
	if result.RowsAffected == 0 {
		RespondNotFound(c, "Team member not found")
		return
	}

	RespondMessage(c, "Member removed successfully")
}

type UpdateMyTeamKeyRequest struct {
	EncryptedTeamKey string `json:"encryptedTeamKey" binding:"required"`
}

func UpdateMyTeamKey(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	teamID, ok := ParseUUIDParam(c, "id", "team")
	if !ok {
		return
	}

	var req UpdateMyTeamKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	result := database.DB.Model(&models.TeamUser{}).
		Where("team_id = ? AND user_id = ?", teamID, uid).
		Update("encrypted_team_key", req.EncryptedTeamKey)

	if result.RowsAffected == 0 {
		RespondNotFound(c, "You are not a member of this team")
		return
	}

	RespondMessage(c, "Team key updated successfully")
}

func GetMyTeams(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

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

	RespondOK(c, teams)
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

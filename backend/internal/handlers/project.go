package handlers

import (
	"errors"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateProjectRequest struct {
	Name           string    `json:"name" binding:"required"`
	EncryptedKey   string    `json:"encryptedKey" binding:"required"`
	OrganizationID uuid.UUID `json:"organizationId" binding:"required"`
	TeamID         uuid.UUID `json:"teamId" binding:"required"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" binding:"required"`
}

type ProjectResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	OrganizationID      uuid.UUID `json:"organizationId"`
	OrganizationName    string    `json:"organizationName"`
	CreatedAt           string    `json:"createdAt"`
	UpdatedAt           string    `json:"updatedAt"`
	EncryptedProjectKey string    `json:"encryptedProjectKey"`
	EncryptedTeamKey    string    `json:"encryptedTeamKey,omitempty"`
	TeamID              uuid.UUID `json:"teamId"`
	TeamName            string    `json:"teamName"`
	TeamRole            string    `json:"teamRole,omitempty"`
	OrgRole             string    `json:"orgRole,omitempty"`
	CanEdit             bool      `json:"canEdit"`
	CanDelete           bool      `json:"canDelete"`
	KeyVersion          int       `json:"keyVersion"`
	ConfigChecksum      string    `json:"configChecksum,omitempty"`
}

type ProjectListItem struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	OrganizationID   uuid.UUID `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	KeyVersion       int       `json:"keyVersion"`
	ConfigChecksum   string    `json:"configChecksum,omitempty"`
	CreatedAt        string    `json:"createdAt"`
	UpdatedAt        string    `json:"updatedAt"`
}

type projectWithOrg struct {
	models.Project
	Organization models.Organization `gorm:"embedded;embeddedPrefix:org_"`
}

func mapProjectsToListItems(results []projectWithOrg) []ProjectListItem {
	projects := make([]ProjectListItem, 0, len(results))
	for _, r := range results {
		configChecksum := ""
		if r.ConfigChecksum != nil {
			configChecksum = *r.ConfigChecksum
		}

		projects = append(projects, ProjectListItem{
			ID:               r.ID,
			Name:             r.Name,
			OrganizationID:   r.OrganizationID,
			OrganizationName: r.Organization.Name,
			KeyVersion:       r.KeyVersion,
			ConfigChecksum:   configChecksum,
			CreatedAt:        r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:        r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return projects
}

func CreateProject(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	var orgUser models.OrganizationUser
	if err := database.DB.Where("user_id = ? AND organization_id = ?", uid, req.OrganizationID).First(&orgUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondForbidden(c, "You don't have access to this organization")
		} else {
			RespondInternalError(c, "Internal error when checking access")
		}
		return
	}

	var team models.Team
	if err := database.DB.Where("id = ? AND organization_id = ?", req.TeamID, req.OrganizationID).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			RespondNotFound(c, "Requested team not found in organization")
		} else {
			RespondInternalError(c, "Internal error when checking team access")
		}
		return
	}

	canCreate, err := CanUserCreateProjectInTeam(uid, req.TeamID, req.OrganizationID)
	if err != nil {
		RespondInternalError(c, "Internal error when checking permissions")
		return
	}

	if !canCreate {
		RespondForbidden(c, "You don't have permissions to create projects in this team")
		return
	}

	tx := database.DB.Begin()

	projectData := models.Project{
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
	}

	if err := tx.Create(&projectData).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to create project")
		return
	}

	teamProjectData := models.TeamProject{
		TeamID:              req.TeamID,
		ProjectID:           projectData.ID,
		EncryptedProjectKey: req.EncryptedKey,
	}

	if err := tx.Create(&teamProjectData).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed adding project to team")
		return
	}

	if err := tx.Commit().Error; err != nil {
		RespondInternalError(c, "Failed creating project")
		return
	}

	RespondCreated(c, gin.H{
		"id":             projectData.ID,
		"name":           projectData.Name,
		"organizationId": projectData.OrganizationID,
	})
}

func GetProjects(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	var results []projectWithOrg
	err := database.DB.Raw(`
		SELECT projects.*, organizations.id as org_id, organizations.name as org_name
		FROM projects
		JOIN organizations ON organizations.id = projects.organization_id
		JOIN team_projects ON team_projects.project_id = projects.id
		JOIN team_users ON team_users.team_id = team_projects.team_id
		WHERE team_users.user_id = ?

		UNION

		SELECT projects.*, organizations.id as org_id, organizations.name as org_name
		FROM projects
		JOIN organizations ON organizations.id = projects.organization_id
		JOIN organization_users ON organization_users.organization_id = projects.organization_id
		WHERE organization_users.user_id = ?
		AND (organization_users.role = 'admin' OR organization_users.role = 'owner')

		ORDER BY updated_at DESC
	`, uid, uid).Scan(&results).Error

	if err != nil {
		RespondInternalError(c, "Failed to fetch projects")
		return
	}

	RespondOK(c, mapProjectsToListItems(results))
}

func GetOrganizationProjects(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}
	orgID, ok := ParseUUIDParam(c, "id", "organization")
	if !ok {
		return
	}

	var results []projectWithOrg
	err := database.DB.Raw(`
		SELECT projects.*, organizations.id as org_id, organizations.name as org_name
		FROM projects
		JOIN organizations ON organizations.id = projects.organization_id
		JOIN team_projects ON team_projects.project_id = projects.id
		JOIN team_users ON team_users.team_id = team_projects.team_id
		WHERE team_users.user_id = ? AND projects.organization_id = ?

		UNION

		SELECT projects.*, organizations.id as org_id, organizations.name as org_name
		FROM projects
		JOIN organizations ON organizations.id = projects.organization_id
		JOIN organization_users ON organization_users.organization_id = projects.organization_id
		WHERE organization_users.user_id = ? AND projects.organization_id = ?
		AND (organization_users.role = 'admin' OR organization_users.role = 'owner')

		ORDER BY updated_at DESC
	`, uid, orgID, uid, orgID).Scan(&results).Error

	if err != nil {
		RespondInternalError(c, "Failed to fetch projects")
		return
	}

	RespondOK(c, mapProjectsToListItems(results))
}

func GetProject(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if err.Error() == "project not found" {
			RespondNotFound(c, "Project not found")
		} else if err.Error() == "access denied" {
			RespondForbidden(c, "Access denied")
		} else {
			RespondInternalError(c, "Failed to check access")
		}
		return
	}

	var org models.Organization
	orgName := ""
	if err := database.DB.Where("id = ?", access.Project.OrganizationID).First(&org).Error; err == nil {
		orgName = org.Name
	}

	configChecksum := ""
	if access.Project.ConfigChecksum != nil {
		configChecksum = *access.Project.ConfigChecksum
	}

	response := ProjectResponse{
		ID:                  access.Project.ID,
		Name:                access.Project.Name,
		OrganizationID:      access.Project.OrganizationID,
		OrganizationName:    orgName,
		CreatedAt:           access.Project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           access.Project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		EncryptedProjectKey: access.EncryptedProjectKey,
		EncryptedTeamKey:    access.EncryptedTeamKey,
		TeamRole:            access.TeamRole,
		OrgRole:             access.OrgRole,
		CanEdit:             access.CanEdit,
		CanDelete:           access.CanDelete,
		KeyVersion:          access.Project.KeyVersion,
		ConfigChecksum:      configChecksum,
	}

	if access.Team != nil {
		response.TeamID = access.Team.ID
		response.TeamName = access.Team.Name
	}

	RespondOK(c, response)
}

func UpdateProject(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			RespondForbidden(c, "Access denied")
		} else {
			RespondInternalError(c, "Failed to verify access")
		}
		return
	}

	if !access.CanEdit {
		RespondForbidden(c, "You don't have permission to edit this project")
		return
	}

	if err := database.DB.Model(&models.Project{}).Where("id = ?", projectID).Update("name", req.Name).Error; err != nil {
		RespondInternalError(c, "Failed to update project")
		return
	}

	RespondMessage(c, "Project updated")
}

func DeleteProject(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			RespondForbidden(c, "Access denied")
		} else {
			RespondInternalError(c, "Failed to verify access")
		}
		return
	}

	if !access.CanDelete {
		RespondForbidden(c, "Only team owners or organization owners can delete projects")
		return
	}

	tx := database.DB.Begin()

	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&models.TeamProject{}).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to delete project associations")
		return
	}

	if err := tx.Unscoped().Delete(&models.Project{}, "id = ?", projectID).Error; err != nil {
		tx.Rollback()
		RespondInternalError(c, "Failed to delete project")
		return
	}

	if err := tx.Commit().Error; err != nil {
		RespondInternalError(c, "Failed to delete project")
		return
	}

	RespondMessage(c, "Project deleted")
}

type TeamWithUsers struct {
	ID    uuid.UUID      `json:"id"`
	Name  string         `json:"name"`
	Users []TeamUserInfo `json:"users"`
}

type TeamUserInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	Role      string    `json:"role"`
}

type OrganizationAdmin struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	OrgRole   string    `json:"role"`
}

type AvailableTeam struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type ProjectAccessResponse struct {
	Teams              []TeamWithUsers     `json:"teams"`
	OrganizationAdmins []OrganizationAdmin `json:"organizationAdmins"`
	AvailableTeams     []AvailableTeam     `json:"availableTeams"`
}

func GetProjectTeams(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		RespondForbidden(c, "Access denied")
		return
	}

	orgID := access.Project.OrganizationID

	// Query 1: Get teams with their users
	type teamUserRow struct {
		models.Team
		Role          *string    `gorm:"column:role"`
		UserID        *uuid.UUID `gorm:"column:user_id"`
		UserName      *string    `gorm:"column:user_name"`
		UserEmail     *string    `gorm:"column:user_email"`
		UserAvatarURL *string    `gorm:"column:user_avatar_url"`
	}
	var rows []teamUserRow
	if err := database.DB.Raw(`
		SELECT teams.*, team_users.role, users.id as user_id, users.name as user_name,
		       users.email as user_email, users.avatar_url as user_avatar_url
		FROM teams
		LEFT JOIN team_users ON team_users.team_id = teams.id
		LEFT JOIN users ON team_users.user_id = users.id
		WHERE teams.id IN (
			SELECT team_id FROM team_projects WHERE project_id = ?
		)
	`, projectID).Scan(&rows).Error; err != nil {
		RespondInternalError(c, "Failed to fetch teams")
		return
	}

	teamsMap := make(map[uuid.UUID]*TeamWithUsers)
	for _, row := range rows {
		if _, exists := teamsMap[row.Team.ID]; !exists {
			teamsMap[row.Team.ID] = &TeamWithUsers{
				ID:    row.Team.ID,
				Name:  row.Team.Name,
				Users: []TeamUserInfo{},
			}
		}
		if row.UserID != nil {
			userName := ""
			if row.UserName != nil {
				userName = *row.UserName
			}
			userEmail := ""
			if row.UserEmail != nil {
				userEmail = *row.UserEmail
			}
			userAvatarURL := ""
			if row.UserAvatarURL != nil {
				userAvatarURL = *row.UserAvatarURL
			}
			teamsMap[row.Team.ID].Users = append(teamsMap[row.Team.ID].Users, TeamUserInfo{
				ID:        *row.UserID,
				Name:      userName,
				Email:     userEmail,
				AvatarURL: userAvatarURL,
				Role:      *row.Role,
			})
		}
	}

	// Query 2: Get available teams (org teams not assigned to this project)
	teamIDs := make([]uuid.UUID, 0, len(teamsMap))
	for id := range teamsMap {
		teamIDs = append(teamIDs, id)
	}

	var availableTeams []AvailableTeam
	if len(teamIDs) > 0 {
		database.DB.Model(&models.Team{}).
			Select("id, name").
			Where("organization_id = ? AND id NOT IN ?", orgID, teamIDs).
			Scan(&availableTeams)
	} else {
		database.DB.Model(&models.Team{}).
			Select("id, name").
			Where("organization_id = ?", orgID).
			Scan(&availableTeams)
	}
	if availableTeams == nil {
		availableTeams = []AvailableTeam{}
	}

	// Query 3: Get org admins/owners who are NOT in any team with project access
	type orgAdminRow struct {
		OrgRole string `gorm:"column:org_role"`
		models.User
	}
	var adminRows []orgAdminRow
	database.DB.Raw(`
		SELECT organization_users.role as org_role, users.*
		FROM organization_users
		JOIN users ON users.id = organization_users.user_id
		WHERE organization_users.organization_id = ?
		AND (organization_users.role = 'admin' OR organization_users.role = 'owner')
	`, orgID).Scan(&adminRows)

	orgAdmins := make([]OrganizationAdmin, len(adminRows))
	for i, row := range adminRows {
		orgAdmins[i] = OrganizationAdmin{
			ID:        row.User.ID,
			Name:      row.User.Name,
			Email:     row.User.Email,
			AvatarURL: row.User.AvatarURL,
			OrgRole:   row.OrgRole,
		}
	}

	// Build response
	teamsResponse := make([]TeamWithUsers, 0, len(teamsMap))
	for _, t := range teamsMap {
		teamsResponse = append(teamsResponse, *t)
	}

	RespondOK(c, ProjectAccessResponse{
		Teams:              teamsResponse,
		OrganizationAdmins: orgAdmins,
		AvailableTeams:     availableTeams,
	})
}

type AddTeamToProjectRequest struct {
	TeamID              uuid.UUID `json:"teamId" binding:"required"`
	EncryptedProjectKey string    `json:"encryptedProjectKey" binding:"required"`
}

func AddTeamToProject(c *gin.Context) {
	uid, ok := GetAuthUserID(c)
	if !ok {
		return
	}

	projectID, ok := ParseUUIDParam(c, "id", "project")
	if !ok {
		return
	}

	var req AddTeamToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, err.Error())
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		RespondForbidden(c, "Access denied")
		return
	}

	if !access.CanEdit {
		RespondForbidden(c, "You don't have permission to modify this project")
		return
	}

	var team models.Team
	if err := database.DB.Where("id = ? AND organization_id = ?", req.TeamID, access.Project.OrganizationID).First(&team).Error; err != nil {
		RespondBadRequest(c, "Team not found in this organization")
		return
	}

	var existing models.TeamProject
	if err := database.DB.Where("team_id = ? AND project_id = ?", req.TeamID, projectID).First(&existing).Error; err == nil {
		RespondConflict(c, "Team already has access to this project")
		return
	}

	teamProject := models.TeamProject{
		TeamID:              req.TeamID,
		ProjectID:           projectID,
		EncryptedProjectKey: req.EncryptedProjectKey,
	}

	if err := database.DB.Create(&teamProject).Error; err != nil {
		RespondInternalError(c, "Failed to add team to project")
		return
	}

	RespondCreated(c, gin.H{"message": "Team added to project"})
}

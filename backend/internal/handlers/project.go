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
	CanEdit    bool `json:"canEdit"`
	CanDelete  bool `json:"canDelete"`
	KeyVersion int  `json:"keyVersion"`
}

type ProjectListItem struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	OrganizationID      uuid.UUID `json:"organizationId"`
	OrganizationName    string    `json:"organizationName"`
	TeamID              uuid.UUID `json:"teamId"`
	TeamName            string    `json:"teamName"`
	EncryptedProjectKey string    `json:"encryptedProjectKey"`
	EncryptedTeamKey    string    `json:"encryptedTeamKey,omitempty"`
	KeyVersion          int       `json:"keyVersion"`
	CreatedAt           string    `json:"createdAt"`
	UpdatedAt           string    `json:"updatedAt"`
}

func CreateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orgUser models.OrganizationUser
	if err := database.DB.Where("user_id = ? AND organization_id = ?", uid, req.OrganizationID).First(&orgUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this organization"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error when checking access"})
		}
		return
	}

	var team models.Team
	if err := database.DB.Where("id = ? AND organization_id = ?", req.TeamID, req.OrganizationID).First(&team).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Requested team not found in organization"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal error when checking team access"})
		}
		return
	}

	canCreate, err := CanUserCreateProjectInTeam(uid, req.TeamID, req.OrganizationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error when checking permissions"})
		return
	}

	if !canCreate {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permissions to create projects in this team"})
		return
	}

	tx := database.DB.Begin()

	projectData := models.Project{
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
	}

	if err := tx.Create(&projectData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	teamProjectData := models.TeamProject{
		TeamID:              req.TeamID,
		ProjectID:           projectData.ID,
		EncryptedProjectKey: req.EncryptedKey,
	}

	if err := tx.Create(&teamProjectData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed adding project to team"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed creating project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             projectData.ID,
		"name":           projectData.Name,
		"organizationId": projectData.OrganizationID,
	})
}

func GetProjects(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var teamUsers []models.TeamUser
	if err := database.DB.Where("user_id = ?", uid).Find(&teamUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team memberships"})
		return
	}

	userTeamKeys := make(map[uuid.UUID]string)
	var userTeamIDs []uuid.UUID
	for _, tu := range teamUsers {
		userTeamIDs = append(userTeamIDs, tu.TeamID)
		userTeamKeys[tu.TeamID] = tu.EncryptedTeamKey
	}

	var orgUsers []models.OrganizationUser
	if err := database.DB.Where("user_id = ? AND (role = 'owner' OR role = 'Owner' OR role = 'admin')", uid).Find(&orgUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization memberships"})
		return
	}

	var adminOrgIDs []uuid.UUID
	for _, ou := range orgUsers {
		adminOrgIDs = append(adminOrgIDs, ou.OrganizationID)
	}

	var projectIDs []uuid.UUID

	if len(userTeamIDs) > 0 {
		var teamProjectIDs []uuid.UUID
		database.DB.Model(&models.TeamProject{}).
			Where("team_id IN ?", userTeamIDs).
			Distinct().
			Pluck("project_id", &teamProjectIDs)
		projectIDs = append(projectIDs, teamProjectIDs...)
	}

	if len(adminOrgIDs) > 0 {
		var orgProjectIDs []uuid.UUID
		database.DB.Model(&models.Project{}).
			Where("organization_id IN ?", adminOrgIDs).
			Pluck("id", &orgProjectIDs)
		projectIDs = append(projectIDs, orgProjectIDs...)
	}

	projectIDSet := make(map[uuid.UUID]bool)
	for _, id := range projectIDs {
		projectIDSet[id] = true
	}

	if len(projectIDSet) == 0 {
		c.JSON(http.StatusOK, []ProjectListItem{})
		return
	}

	uniqueProjectIDs := make([]uuid.UUID, 0, len(projectIDSet))
	for id := range projectIDSet {
		uniqueProjectIDs = append(uniqueProjectIDs, id)
	}

	var dbProjects []models.Project
	if err := database.DB.Where("id IN ?", uniqueProjectIDs).Order("updated_at DESC").Find(&dbProjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	var orgIDs []uuid.UUID
	for _, p := range dbProjects {
		orgIDs = append(orgIDs, p.OrganizationID)
	}

	var organizations []models.Organization
	database.DB.Where("id IN ?", orgIDs).Find(&organizations)
	orgMap := make(map[uuid.UUID]models.Organization)
	for _, org := range organizations {
		orgMap[org.ID] = org
	}

	var teamProjects []models.TeamProject
	database.DB.Where("project_id IN ?", uniqueProjectIDs).Find(&teamProjects)

	var allTeamIDs []uuid.UUID
	for _, tp := range teamProjects {
		allTeamIDs = append(allTeamIDs, tp.TeamID)
	}

	var teams []models.Team
	if len(allTeamIDs) > 0 {
		database.DB.Where("id IN ?", allTeamIDs).Find(&teams)
	}
	teamMap := make(map[uuid.UUID]models.Team)
	for _, t := range teams {
		teamMap[t.ID] = t
	}

	type teamInfo struct {
		TeamID              uuid.UUID
		TeamName            string
		EncryptedProjectKey string
		EncryptedTeamKey    string
	}
	projectTeamInfo := make(map[uuid.UUID]teamInfo)

	for _, tp := range teamProjects {
		team := teamMap[tp.TeamID]
		info := teamInfo{
			TeamID:              tp.TeamID,
			TeamName:            team.Name,
			EncryptedProjectKey: tp.EncryptedProjectKey,
			EncryptedTeamKey:    userTeamKeys[tp.TeamID],
		}

		existing, exists := projectTeamInfo[tp.ProjectID]
		if !exists {
			projectTeamInfo[tp.ProjectID] = info
		} else if info.EncryptedTeamKey != "" && existing.EncryptedTeamKey == "" {
			// Prefer team where user has membership (has encrypted team key)
			projectTeamInfo[tp.ProjectID] = info
		}
	}

	projects := make([]ProjectListItem, 0, len(dbProjects))
	for _, p := range dbProjects {
		org := orgMap[p.OrganizationID]
		ti := projectTeamInfo[p.ID]

		projects = append(projects, ProjectListItem{
			ID:                  p.ID,
			Name:                p.Name,
			OrganizationID:      p.OrganizationID,
			OrganizationName:    org.Name,
			TeamID:              ti.TeamID,
			TeamName:            ti.TeamName,
			EncryptedProjectKey: ti.EncryptedProjectKey,
			EncryptedTeamKey:    ti.EncryptedTeamKey,
			KeyVersion:          p.KeyVersion,
			CreatedAt:           p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:           p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	c.JSON(http.StatusOK, projects)
}

func GetProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if err.Error() == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check access"})
		}
		return
	}

	var org models.Organization
	orgName := ""
	if err := database.DB.Where("id = ?", access.Project.OrganizationID).First(&org).Error; err == nil {
		orgName = org.Name
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
		CanEdit:    access.CanEdit,
		CanDelete:  access.CanDelete,
		KeyVersion: access.Project.KeyVersion,
	}

	if access.Team != nil {
		response.TeamID = access.Team.ID
		response.TeamName = access.Team.Name
	}

	c.JSON(http.StatusOK, response)
}

func UpdateProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify access"})
		}
		return
	}

	if !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this project"})
		return
	}

	if err := database.DB.Model(&models.Project{}).Where("id = ?", projectID).Update("name", req.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project updated"})
}

func DeleteProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify access"})
		}
		return
	}

	if !access.CanDelete {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owners or organization owners can delete projects"})
		return
	}

	tx := database.DB.Begin()

	if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&models.TeamProject{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project associations"})
		return
	}

	if err := tx.Unscoped().Delete(&models.Project{}, "id = ?", projectID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}

type TeamWithUsers struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	MemberCount  int64          `json:"memberCount"`
	ProjectCount int64          `json:"projectCount"`
	Users        []TeamUserInfo `json:"users"`
}

type TeamUserInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	Role      string    `json:"role"`
}

type OrgUserInfo struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatarUrl"`
	Role      string    `json:"role"`
}

type ProjectAccessResponse struct {
	Teams               []TeamWithUsers `json:"teams"`
	OrganizationAdmins  []OrgUserInfo   `json:"organizationAdmins"`
	AvailableTeams      []TeamWithUsers `json:"availableTeams"`
}

func GetProjectTeams(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var teamProjects []models.TeamProject
	if err := database.DB.Where("project_id = ?", projectID).Find(&teamProjects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project teams"})
		return
	}

	teams := []TeamWithUsers{}
	teamIDsWithAccess := make(map[uuid.UUID]bool)

	for _, tp := range teamProjects {
		teamIDsWithAccess[tp.TeamID] = true

		var team models.Team
		if err := database.DB.Where("id = ?", tp.TeamID).First(&team).Error; err != nil {
			continue
		}

		var memberCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ?", team.ID).Count(&memberCount)

		var projectCount int64
		database.DB.Model(&models.TeamProject{}).Where("team_id = ?", team.ID).Count(&projectCount)

		var teamUsers []models.TeamUser
		database.DB.Where("team_id = ?", team.ID).Find(&teamUsers)

		users := []TeamUserInfo{}
		for _, tu := range teamUsers {
			var user models.User
			if err := database.DB.Where("id = ?", tu.UserID).First(&user).Error; err == nil {
				users = append(users, TeamUserInfo{
					ID:        user.ID,
					Name:      user.Name,
					Email:     user.Email,
					AvatarURL: user.AvatarURL,
					Role:      tu.Role,
				})
			}
		}

		teams = append(teams, TeamWithUsers{
			ID:           team.ID,
			Name:         team.Name,
			MemberCount:  memberCount,
			ProjectCount: projectCount,
			Users:        users,
		})
	}

	var orgUsers []models.OrganizationUser
	database.DB.Where("organization_id = ? AND (role = 'owner' OR role = 'Owner' OR role = 'admin')", access.Project.OrganizationID).Find(&orgUsers)

	orgAdmins := []OrgUserInfo{}
	for _, ou := range orgUsers {
		var user models.User
		if err := database.DB.Where("id = ?", ou.UserID).First(&user).Error; err == nil {
			orgAdmins = append(orgAdmins, OrgUserInfo{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				AvatarURL: user.AvatarURL,
				Role:      ou.Role,
			})
		}
	}

	var allOrgTeams []models.Team
	database.DB.Where("organization_id = ?", access.Project.OrganizationID).Find(&allOrgTeams)

	availableTeams := []TeamWithUsers{}
	for _, team := range allOrgTeams {
		if teamIDsWithAccess[team.ID] {
			continue
		}

		var memberCount int64
		database.DB.Model(&models.TeamUser{}).Where("team_id = ?", team.ID).Count(&memberCount)

		var projectCount int64
		database.DB.Model(&models.TeamProject{}).Where("team_id = ?", team.ID).Count(&projectCount)

		availableTeams = append(availableTeams, TeamWithUsers{
			ID:           team.ID,
			Name:         team.Name,
			MemberCount:  memberCount,
			ProjectCount: projectCount,
			Users:        []TeamUserInfo{},
		})
	}

	c.JSON(http.StatusOK, ProjectAccessResponse{
		Teams:              teams,
		OrganizationAdmins: orgAdmins,
		AvailableTeams:     availableTeams,
	})
}

type AddTeamToProjectRequest struct {
	TeamID              uuid.UUID `json:"teamId" binding:"required"`
	EncryptedProjectKey string    `json:"encryptedProjectKey" binding:"required"`
}

func AddTeamToProject(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	projectIDStr := c.Param("id")
	projectID, err := uuid.Parse(projectIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req AddTeamToProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	access, err := GetUserProjectAccess(uid, projectID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if !access.CanEdit {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this project"})
		return
	}

	var team models.Team
	if err := database.DB.Where("id = ? AND organization_id = ?", req.TeamID, access.Project.OrganizationID).First(&team).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team not found in this organization"})
		return
	}

	var existing models.TeamProject
	if err := database.DB.Where("team_id = ? AND project_id = ?", req.TeamID, projectID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Team already has access to this project"})
		return
	}

	teamProject := models.TeamProject{
		TeamID:              req.TeamID,
		ProjectID:           projectID,
		EncryptedProjectKey: req.EncryptedProjectKey,
	}

	if err := database.DB.Create(&teamProject).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add team to project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Team added to project"})
}

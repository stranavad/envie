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
	CanEdit        bool   `json:"canEdit"`
	CanDelete      bool   `json:"canDelete"`
	KeyVersion     int    `json:"keyVersion"`
	ConfigChecksum string `json:"configChecksum,omitempty"`
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
	ConfigChecksum      string    `json:"configChecksum,omitempty"`
	CreatedAt           string    `json:"createdAt"`
	UpdatedAt           string    `json:"updatedAt"`
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

	var teamUsers []models.TeamUser
	if err := database.DB.Where("user_id = ?", uid).Find(&teamUsers).Error; err != nil {
		RespondInternalError(c, "Failed to fetch team memberships")
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
		RespondInternalError(c, "Failed to fetch organization memberships")
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
		RespondOK(c, []ProjectListItem{})
		return
	}

	uniqueProjectIDs := make([]uuid.UUID, 0, len(projectIDSet))
	for id := range projectIDSet {
		uniqueProjectIDs = append(uniqueProjectIDs, id)
	}

	var dbProjects []models.Project
	if err := database.DB.Where("id IN ?", uniqueProjectIDs).Order("updated_at DESC").Find(&dbProjects).Error; err != nil {
		RespondInternalError(c, "Failed to fetch projects")
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

		configChecksum := ""
		if p.ConfigChecksum != nil {
			configChecksum = *p.ConfigChecksum
		}

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
			ConfigChecksum:      configChecksum,
			CreatedAt:           p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:           p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	RespondOK(c, projects)
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
		CanEdit:        access.CanEdit,
		CanDelete:      access.CanDelete,
		KeyVersion:     access.Project.KeyVersion,
		ConfigChecksum: configChecksum,
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

	var allOrgTeams []models.Team
	database.DB.Where("organization_id = ?", access.Project.OrganizationID).Find(&allOrgTeams)

	if len(allOrgTeams) == 0 {
		RespondOK(c, ProjectAccessResponse{
			Teams:              []TeamWithUsers{},
			OrganizationAdmins: []OrgUserInfo{},
			AvailableTeams:     []TeamWithUsers{},
		})
		return
	}

	var teamProjects []models.TeamProject
	database.DB.Where("project_id = ?", projectID).Find(&teamProjects)

	teamIDsWithAccess := make(map[uuid.UUID]bool)
	for _, tp := range teamProjects {
		teamIDsWithAccess[tp.TeamID] = true
	}

	allTeamIDs := make([]uuid.UUID, len(allOrgTeams))
	for i, t := range allOrgTeams {
		allTeamIDs[i] = t.ID
	}

	type CountResult struct {
		TeamID uuid.UUID
		Count  int64
	}
	var memberCounts []CountResult
	database.DB.Model(&models.TeamUser{}).
		Select("team_id, COUNT(*) as count").
		Where("team_id IN ?", allTeamIDs).
		Group("team_id").
		Scan(&memberCounts)

	memberCountMap := make(map[uuid.UUID]int64)
	for _, mc := range memberCounts {
		memberCountMap[mc.TeamID] = mc.Count
	}

	var projectCounts []CountResult
	database.DB.Model(&models.TeamProject{}).
		Select("team_id, COUNT(*) as count").
		Where("team_id IN ?", allTeamIDs).
		Group("team_id").
		Scan(&projectCounts)

	projectCountMap := make(map[uuid.UUID]int64)
	for _, pc := range projectCounts {
		projectCountMap[pc.TeamID] = pc.Count
	}

	accessTeamIDs := make([]uuid.UUID, 0, len(teamIDsWithAccess))
	for teamID := range teamIDsWithAccess {
		accessTeamIDs = append(accessTeamIDs, teamID)
	}

	teamUsersMap := make(map[uuid.UUID][]TeamUserInfo)
	if len(accessTeamIDs) > 0 {
		type TeamUserWithInfo struct {
			TeamID    uuid.UUID
			UserID    uuid.UUID
			Name      string
			Email     string
			AvatarURL string `gorm:"column:avatar_url"`
			Role      string
		}
		var teamUsersWithInfo []TeamUserWithInfo
		database.DB.Model(&models.TeamUser{}).
			Select("team_users.team_id, users.id as user_id, users.name, users.email, users.avatar_url, team_users.role").
			Joins("JOIN users ON users.id = team_users.user_id").
			Where("team_users.team_id IN ?", accessTeamIDs).
			Scan(&teamUsersWithInfo)

		for _, tu := range teamUsersWithInfo {
			teamUsersMap[tu.TeamID] = append(teamUsersMap[tu.TeamID], TeamUserInfo{
				ID:        tu.UserID,
				Name:      tu.Name,
				Email:     tu.Email,
				AvatarURL: tu.AvatarURL,
				Role:      tu.Role,
			})
		}
	}

	teams := []TeamWithUsers{}
	teamMap := make(map[uuid.UUID]models.Team)
	for _, team := range allOrgTeams {
		teamMap[team.ID] = team
	}

	for teamID := range teamIDsWithAccess {
		team, exists := teamMap[teamID]
		if !exists {
			continue
		}
		users := teamUsersMap[teamID]
		if users == nil {
			users = []TeamUserInfo{}
		}
		teams = append(teams, TeamWithUsers{
			ID:           team.ID,
			Name:         team.Name,
			MemberCount:  memberCountMap[team.ID],
			ProjectCount: projectCountMap[team.ID],
			Users:        users,
		})
	}

	// Build available teams (no access yet)
	availableTeams := []TeamWithUsers{}
	for _, team := range allOrgTeams {
		if teamIDsWithAccess[team.ID] {
			continue
		}
		availableTeams = append(availableTeams, TeamWithUsers{
			ID:           team.ID,
			Name:         team.Name,
			MemberCount:  memberCountMap[team.ID],
			ProjectCount: projectCountMap[team.ID],
			Users:        []TeamUserInfo{},
		})
	}

	type OrgAdminWithInfo struct {
		UserID    uuid.UUID `gorm:"column:user_id"`
		Name      string
		Email     string
		AvatarURL string `gorm:"column:avatar_url"`
		Role      string
	}
	var orgAdminsWithInfo []OrgAdminWithInfo
	database.DB.Model(&models.OrganizationUser{}).
		Select("organization_users.user_id, users.name, users.email, users.avatar_url, organization_users.role").
		Joins("JOIN users ON users.id = organization_users.user_id").
		Where("organization_users.organization_id = ? AND (organization_users.role = 'owner' OR organization_users.role = 'Owner' OR organization_users.role = 'admin')", access.Project.OrganizationID).
		Scan(&orgAdminsWithInfo)

	orgAdmins := make([]OrgUserInfo, len(orgAdminsWithInfo))
	for i, oa := range orgAdminsWithInfo {
		orgAdmins[i] = OrgUserInfo{
			ID:        oa.UserID,
			Name:      oa.Name,
			Email:     oa.Email,
			AvatarURL: oa.AvatarURL,
			Role:      oa.Role,
		}
	}

	RespondOK(c, ProjectAccessResponse{
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

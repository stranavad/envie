package handlers

import (
	"errors"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectAccess struct {
	Project             *models.Project
	Team                *models.Team
	TeamProject         *models.TeamProject
	TeamRole            string
	OrgRole             string
	CanEdit             bool
	CanDelete           bool
	CanManageSecrets    bool
	EncryptedProjectKey string
	EncryptedTeamKey    string
}

func GetUserProjectAccess(userID uuid.UUID, projectID uuid.UUID) (*ProjectAccess, error) {
	var project models.Project
	if err := database.DB.Where("id = ?", projectID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, err
	}

	access := &ProjectAccess{
		Project: &project,
	}

	orgRole, err := GetUserOrgRole(userID, project.OrganizationID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	access.OrgRole = orgRole

	var teamProject models.TeamProject
	var teamUser models.TeamUser

	err = database.DB.
		Joins("JOIN team_users ON team_users.team_id = team_projects.team_id").
		Where("team_projects.project_id = ? AND team_users.user_id = ?", projectID, userID).
		First(&teamProject).Error

	if err == nil {
		access.TeamProject = &teamProject
		access.EncryptedProjectKey = teamProject.EncryptedProjectKey

		var team models.Team
		if err := database.DB.Where("id = ?", teamProject.TeamID).First(&team).Error; err == nil {
			access.Team = &team
		}

		if err := database.DB.Where("team_id = ? AND user_id = ?", teamProject.TeamID, userID).First(&teamUser).Error; err == nil {
			access.TeamRole = teamUser.Role
			access.EncryptedTeamKey = teamUser.EncryptedTeamKey
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if access.TeamProject == nil && (orgRole == "owner" || orgRole == "admin") {
		if err := database.DB.Where("project_id = ?", projectID).First(&teamProject).Error; err == nil {
			access.TeamProject = &teamProject
			access.EncryptedProjectKey = teamProject.EncryptedProjectKey

			var team models.Team
			if err := database.DB.Where("id = ?", teamProject.TeamID).First(&team).Error; err == nil {
				access.Team = &team
			}
		}
	}

	if access.TeamProject == nil && access.OrgRole == "" {
		return nil, errors.New("access denied")
	}
	if access.TeamProject == nil && access.OrgRole == "member" {
		return nil, errors.New("access denied")
	}

	access.CanEdit = access.TeamRole == "owner" || access.TeamRole == "admin" ||
		access.OrgRole == "owner" || access.OrgRole == "admin"

	access.CanDelete = access.TeamRole == "owner" || access.OrgRole == "owner"

	access.CanManageSecrets = access.CanEdit

	return access, nil
}

func GetUserOrgRole(userID uuid.UUID, orgID uuid.UUID) (string, error) {
	var orgUser models.OrganizationUser
	err := database.DB.Where("user_id = ? AND organization_id = ?", userID, orgID).First(&orgUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}

	role := orgUser.Role
	if role == "Owner" {
		role = "owner"
	}
	return role, nil
}

func GetUserTeamRole(userID uuid.UUID, teamID uuid.UUID) (string, error) {
	var teamUser models.TeamUser
	err := database.DB.Where("user_id = ? AND team_id = ?", userID, teamID).First(&teamUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return teamUser.Role, nil
}

func IsUserOrgOwnerOrAdmin(userID uuid.UUID, orgID uuid.UUID) (bool, string, error) {
	role, err := GetUserOrgRole(userID, orgID)
	if err != nil {
		return false, "", err
	}
	isOwnerOrAdmin := role == "owner" || role == "admin"
	return isOwnerOrAdmin, role, nil
}

func CanUserCreateProjectInTeam(userID uuid.UUID, teamID uuid.UUID, orgID uuid.UUID) (bool, error) {
	isOrgOwnerOrAdmin, _, err := IsUserOrgOwnerOrAdmin(userID, orgID)
	if err != nil {
		return false, err
	}
	if isOrgOwnerOrAdmin {
		return true, nil
	}

	teamRole, err := GetUserTeamRole(userID, teamID)
	if err != nil {
		return false, err
	}

	return teamRole == "owner" || teamRole == "admin", nil
}

func CheckProjectAccessSimple(userID uuid.UUID, projectIDStr string) error {
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return errors.New("invalid project ID")
	}

	_, err = GetUserProjectAccess(userID, projectID)
	return err
}

func CheckProjectWriteAccess(userID uuid.UUID, projectIDStr string) (*ProjectAccess, error) {
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, errors.New("invalid project ID")
	}

	access, err := GetUserProjectAccess(userID, projectID)
	if err != nil {
		return nil, err
	}

	if !access.CanEdit {
		return nil, errors.New("insufficient permissions to edit project")
	}

	return access, nil
}

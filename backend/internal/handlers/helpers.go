package handlers

import (
	"net/http"

	"envie-backend/internal/database"
	"envie-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAuthUserID extracts the authenticated user's ID from the context.
// Returns the user ID and a boolean indicating success.
// If unsuccessful, it sends an error response automatically.
func GetAuthUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.UUID{}, false
	}
	uid, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return uuid.UUID{}, false
	}
	return uid, true
}

// ParseUUIDParam parses a UUID from a route parameter.
// Returns the parsed UUID and a boolean indicating success.
// If unsuccessful, it sends an error response automatically.
func ParseUUIDParam(c *gin.Context, param string, entityName string) (uuid.UUID, bool) {
	idStr := c.Param(param)
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": entityName + " ID required"})
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + entityName + " ID"})
		return uuid.UUID{}, false
	}
	return id, true
}

// ParseUUIDQuery parses a UUID from a query parameter.
// Returns the parsed UUID and a boolean indicating success.
// If unsuccessful, it sends an error response automatically.
func ParseUUIDQuery(c *gin.Context, param string, entityName string) (uuid.UUID, bool) {
	idStr := c.Query(param)
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": entityName + " ID query parameter required"})
		return uuid.UUID{}, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + entityName + " ID"})
		return uuid.UUID{}, false
	}
	return id, true
}

// RequireOrgMembership checks if the user is a member of the organization.
// Returns the OrganizationUser and a boolean indicating success.
// If unsuccessful, it sends an error response automatically.
func RequireOrgMembership(c *gin.Context, userID, orgID uuid.UUID) (*models.OrganizationUser, bool) {
	var orgUser models.OrganizationUser
	if err := database.DB.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&orgUser).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return nil, false
	}
	return &orgUser, true
}

// RequireOrgAdmin checks if the user is an admin or owner of the organization.
// Returns the OrganizationUser and a boolean indicating success.
// If unsuccessful, it sends an error response automatically.
func RequireOrgAdmin(c *gin.Context, userID, orgID uuid.UUID) (*models.OrganizationUser, bool) {
	orgUser, ok := RequireOrgMembership(c, userID, orgID)
	if !ok {
		return nil, false
	}
	if !IsAdminOrOwner(orgUser.Role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only organization owners and admins can perform this action"})
		return nil, false
	}
	return orgUser, true
}

// RequireOrgOwner checks if the user is an owner of the organization.
// Returns the OrganizationUser and a boolean indicating success.
// If unsuccessful, it sends an error response automatically.
func RequireOrgOwner(c *gin.Context, userID, orgID uuid.UUID) (*models.OrganizationUser, bool) {
	orgUser, ok := RequireOrgMembership(c, userID, orgID)
	if !ok {
		return nil, false
	}
	if !IsOwner(orgUser.Role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only organization owners can perform this action"})
		return nil, false
	}
	return orgUser, true
}

// IsAdminOrOwner checks if a role is admin or owner (case-insensitive for owner).
func IsAdminOrOwner(role string) bool {
	return role == "owner" || role == "Owner" || role == "admin"
}

// IsOwner checks if a role is owner (case-insensitive).
func IsOwner(role string) bool {
	return role == "owner" || role == "Owner"
}

// RespondError sends a JSON error response with the given status and message.
func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// RespondBadRequest is a shorthand for 400 Bad Request errors.
func RespondBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": message})
}

// RespondForbidden is a shorthand for 403 Forbidden errors.
func RespondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{"error": message})
}

// RespondNotFound is a shorthand for 404 Not Found errors.
func RespondNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, gin.H{"error": message})
}

// RespondConflict is a shorthand for 409 Conflict errors.
func RespondConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{"error": message})
}

// RespondInternalError is a shorthand for 500 Internal Server Error.
func RespondInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": message})
}

// RespondOK sends a JSON response with 200 OK status.
func RespondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

// RespondCreated sends a JSON response with 201 Created status.
func RespondCreated(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, data)
}

// RespondMessage sends a simple message response with 200 OK status.
func RespondMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// ValidRoles is a map of valid organization/team roles.
var ValidRoles = map[string]bool{"owner": true, "admin": true, "member": true}

// IsValidRole checks if the given role is valid.
func IsValidRole(role string) bool {
	return ValidRoles[role]
}

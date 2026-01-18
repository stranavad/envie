package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamUserRoleT int

const (
	TeamMember TeamUserRoleT = iota
	TeamAdmin
	TeamOwner
)

var TeamUserRole = map[TeamUserRoleT]string {
	TeamMember: "member",
	TeamAdmin: "admin",
	TeamOwner: "owner",
}

type Team struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;index;not null" json:"organizationId"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	EncryptedKey   string    `gorm:"type:text" json:"encryptedKey"` // encrypted with org master key

	TeamUsers []TeamUser `json:"users"`
	Organization Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type TeamUser struct {
	TeamID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"teamId"`
	UserID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	EncryptedTeamKey string    `gorm:"type:text;not null" json:"encryptedTeamKey"` // encrypted with user mk
	Role             string    `gorm:"size:50;default:'member'" json:"role"`

	Team Team `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"team"`
	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TeamProject struct {
	TeamID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"teamId"`
	ProjectID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"projectId"`
	EncryptedProjectKey string    `gorm:"type:text;not null" json:"encryptedProjectKey"` // encrypted with decrypted team key

	Team    Team    `gorm:"foreignKey:TeamID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"team"`
	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"project"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

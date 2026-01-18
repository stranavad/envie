package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Org user role
type OrgUserRoleT int

const (
	OrgMember OrgUserRoleT = iota
	OrgAdmin
	OrgOwner
)

var OrgUserRole = map[OrgUserRoleT]string {
	OrgMember: "member",
	OrgAdmin: "admin",
	OrgOwner: "Owner",
}


type Organization struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name string    `gorm:"size:255;not null" json:"name"`

	Teams []Team             `json:"teams,omitempty"`
	Users []OrganizationUser `json:"users,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

type OrganizationUser struct {
	OrganizationID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"organizationId"`
	UserID                   uuid.UUID `gorm:"type:uuid;primaryKey" json:"userId"`
	Role                     string    `gorm:"size:50;default:'member'" json:"role"`      // 'owner', 'admin', 'member'
	EncryptedOrganizationKey *string   `gorm:"type:text" json:"encryptedOrganizationKey"` // only owner + admin have this, encrypted org master key with their pk

	Organization Organization `gorm:"foreignKey:OrganizationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"organization"`
	User         User         `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

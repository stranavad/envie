package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index;not null" json:"projectId"`
	Name      string    `gorm:"size:255;not null" json:"name"`

	TokenPrefix         string `gorm:"size:10;not null" json:"tokenPrefix"`          // first 3 chars after "envie_"
	IdentityIDHash      string `gorm:"size:64;uniqueIndex;not null" json:"-"`        // SHA256 of derived identity ID
	EncryptedProjectKey string `gorm:"type:text;not null" json:"-"`                  // project key encrypted to token's public key

	ExpiresAt  *time.Time `gorm:"index" json:"expiresAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"createdBy"`
	Creator   User      `gorm:"foreignKey:CreatedBy" json:"creator"`

	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *ProjectToken) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

func (t *ProjectToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

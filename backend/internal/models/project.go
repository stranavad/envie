package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	OrganizationID uuid.UUID `gorm:"type:uuid;index" json:"organizationId"`

	KeyVersion     int     `gorm:"default:1" json:"keyVersion"`
	ConfigChecksum *string `gorm:"size:64" json:"configChecksum"`

	CreatedAt            time.Time             `json:"createdAt"`
	UpdatedAt            time.Time             `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt        `gorm:"index" json:"deletedAt"`
	SecretManagerConfigs []SecretManagerConfig `json:"secretManagerConfigs"`
}

func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

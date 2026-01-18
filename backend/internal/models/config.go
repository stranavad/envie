package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfigItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID  uuid.UUID `gorm:"type:uuid;index;not null" json:"projectId"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	Value      string    `gorm:"type:text;not null" json:"value"`
	Sensitive bool    `gorm:"default:false" json:"sensitive"`
	Position  int     `gorm:"default:0" json:"position"`
	Category  *string `gorm:"size:255" json:"category"`

	CreatedBy uuid.UUID `gorm:"type:uuid" json:"createdBy"`
	UpdatedBy uuid.UUID `gorm:"type:uuid" json:"updatedBy"`

	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Creator User    `gorm:"foreignKey:CreatedBy" json:"creator"`
	Updater User    `gorm:"foreignKey:UpdatedBy" json:"updater"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`

	SecretManagerConfigID   *uuid.UUID          `gorm:"type:uuid;index" json:"secretManagerConfigId"`
	SecretManagerConfig     SecretManagerConfig `gorm:"foreignKey:SecretManagerConfigID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	SecretManagerName       *string             `json:"secretManagerName"`
	SecretManagerLastSyncAt *time.Time          `json:"secretManagerLastSyncAt"`
	SecretManagerVersion    *string             `json:"secretManagerVersion"`
}

func (c *ConfigItem) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}

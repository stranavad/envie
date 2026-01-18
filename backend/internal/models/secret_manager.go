package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SecretManagerConfig struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Name         string `gorm:"size:50;not null" json:"name"`
	EncryptedKey string `gorm:"type:text;not null" json:"encryptedKey"`

	ProjectID uuid.UUID `gorm:"type:uuid;not null" json:"projectId"`
	Project   Project   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	CreatedByID uuid.UUID `gorm:"type:uuid" json:"createdById"`
	CreatedBy   User      `json:"createdBy"`

	UpdatedByID uuid.UUID `gorm:"type:uuid" json:"updatedById"`
	UpdatedBy   User      `json:"updatedBy"`
}

func (s *SecretManagerConfig) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectFile struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID    uuid.UUID `gorm:"type:uuid;index;not null" json:"projectId"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	SizeBytes    int64     `gorm:"not null" json:"sizeBytes"`
	MimeType     string    `gorm:"size:100" json:"mimeType"`
	S3Key        string    `gorm:"size:500;not null" json:"s3Key"`
	EncryptedFEK string    `gorm:"type:text;not null" json:"encryptedFek"`
	Checksum     string    `gorm:"size:64" json:"checksum"`

	UploadedBy   uuid.UUID `gorm:"type:uuid;not null" json:"uploadedBy"`
	UploadedUser User      `gorm:"foreignKey:UploadedBy" json:"uploadedUser"`

	Project Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

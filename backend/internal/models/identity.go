package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserIdentity struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID             uuid.UUID `gorm:"type:uuid;index;not null" json:"userId"`
	Name               string    `gorm:"size:255;not null" json:"name"`
	PublicKey          string    `gorm:"type:text;not null" json:"publicKey"`
	EncryptedMasterKey *string   `gorm:"type:text" json:"encryptedMasterKey"` // null -> pending approval
	LastActive         time.Time `json:"lastActive"`

	User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

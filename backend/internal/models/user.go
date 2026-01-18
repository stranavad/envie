package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"size:255" json:"name"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	AvatarURL string         `gorm:"size:1024" json:"avatarUrl"`
	GithubID  int64          `gorm:"uniqueIndex" json:"githubId"`
	PublicKey string         `gorm:"type:text;not null" json:"publicKey"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

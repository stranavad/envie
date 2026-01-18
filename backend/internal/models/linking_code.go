package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LinkingCode struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code            string     `gorm:"size:32;uniqueIndex;not null" json:"code"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	DevicePublicKey string     `gorm:"type:text" json:"devicePublicKey"`
	ExpiresAt       time.Time  `gorm:"not null" json:"expiresAt"`
	UsedAt          *time.Time `json:"usedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}

func (lc *LinkingCode) BeforeCreate(tx *gorm.DB) (err error) {
	if lc.ID == uuid.Nil {
		lc.ID = uuid.New()
	}
	return
}

func (lc *LinkingCode) IsValid() bool {
	return lc.UsedAt == nil && time.Now().Before(lc.ExpiresAt)
}

type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Token     string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null" json:"userId"`
	User      User       `gorm:"foreignKey:UserID" json:"-"`
	DeviceID  *uuid.UUID `gorm:"type:uuid" json:"deviceId"`
	FamilyID  uuid.UUID  `gorm:"type:uuid;not null" json:"familyId"`
	ExpiresAt time.Time  `gorm:"not null" json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt"` // null -> active
	CreatedAt time.Time  `json:"createdAt"`
}

func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	if rt.FamilyID == uuid.Nil {
		rt.FamilyID = uuid.New()
	}
	return
}

func (rt *RefreshToken) IsValid() bool {
	return rt.RevokedAt == nil && time.Now().Before(rt.ExpiresAt)
}

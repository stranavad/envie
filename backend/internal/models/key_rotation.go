package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PendingKeyRotation struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID         uuid.UUID `gorm:"type:uuid;index;not null" json:"projectId"`
	InitiatedBy       uuid.UUID `gorm:"type:uuid;not null" json:"initiatedBy"`
	NewVersion        int       `gorm:"not null" json:"newVersion"`
	Status            string    `gorm:"size:50;default:'pending'" json:"status"` // pending, approved, rejected, expired, stale
	RequiredApprovals int       `gorm:"default:1" json:"requiredApprovals"`
	ExpiresAt         time.Time `json:"expiresAt"`

	EncryptedConfigsSnapshot string `gorm:"type:text" json:"encryptedConfigsSnapshot"`

	TeamEncryptedKeys string `gorm:"type:text" json:"teamEncryptedKeys"`

	EncryptedFileFEKsSnapshot string `gorm:"type:text" json:"encryptedFileFEKsSnapshot"`

	SnapshotConfigItemIDs        string `gorm:"type:text" json:"snapshotConfigItemIds"`
	SnapshotTeamIDs              string `gorm:"type:text" json:"snapshotTeamIds"`
	SnapshotSecretManagerConfIDs string `gorm:"type:text" json:"snapshotSecretManagerConfIds"`
	SnapshotConfigItemsHash      string `gorm:"type:text" json:"snapshotConfigItemsHash"`

	Project   Project `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Initiator User    `gorm:"foreignKey:InitiatedBy" json:"initiator"`

	Approvals []KeyRotationApproval `gorm:"foreignKey:RotationID" json:"approvals"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (p *PendingKeyRotation) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}

type KeyRotationApproval struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RotationID uuid.UUID `gorm:"type:uuid;index;not null" json:"rotationId"`
	UserID     uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	Approved   bool      `gorm:"not null" json:"approved"`
	Comment    string    `gorm:"type:text" json:"comment"`

	VerifiedDecryption bool `gorm:"default:false" json:"verifiedDecryption"`

	Rotation PendingKeyRotation `gorm:"foreignKey:RotationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	User     User               `gorm:"foreignKey:UserID" json:"user"`

	CreatedAt time.Time `json:"createdAt"`
}

func (k *KeyRotationApproval) BeforeCreate(tx *gorm.DB) (err error) {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return
}

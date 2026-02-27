package models

import "time"

type DocumentStatus string

const (
	DocumentStatusActive  DocumentStatus = "active"
	DocumentStatusDeleted DocumentStatus = "deleted"
)

type Document struct {
	ID         string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string         `gorm:"type:uuid;not null;index" json:"userId"`
	FileName   string         `gorm:"not null" json:"fileName"`
	FileSize   int64          `gorm:"not null" json:"fileSize"`
	MimeType   string         `gorm:"not null" json:"mimeType"`
	StorageKey string         `gorm:"not null" json:"storageKey"`
	UploadedAt time.Time      `gorm:"autoCreateTime" json:"uploadedAt"`
	ExpiresAt  *time.Time     `json:"expiresAt"`
	Status        DocumentStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	SecretKeyHash string         `gorm:"" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

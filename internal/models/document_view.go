package models

import "time"

type DocumentView struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID string    `gorm:"type:uuid;not null;index" json:"documentId"`
	UserID     string    `gorm:"type:uuid;not null;index" json:"userId"`
	IPAddress  string    `gorm:"type:varchar(45)" json:"ipAddress"`
	ViewedAt   time.Time `gorm:"autoCreateTime" json:"viewedAt"`

	Document Document `gorm:"foreignKey:DocumentID" json:"-"`
	User     User     `gorm:"foreignKey:UserID" json:"-"`
}

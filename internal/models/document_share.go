package models

import "time"

type DocumentShare struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_doc_user_share" json:"documentId"`
	UserID     string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_doc_user_share" json:"userId"`
	SharedBy   string    `gorm:"type:uuid;not null" json:"sharedBy"`
	SharedAt   time.Time `gorm:"autoCreateTime" json:"sharedAt"`

	Document Document `gorm:"foreignKey:DocumentID" json:"-"`
	User     User     `gorm:"foreignKey:UserID" json:"-"`
	Sharer   User     `gorm:"foreignKey:SharedBy" json:"-"`
}

type DocumentGroupShare struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DocumentID string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_doc_group_share" json:"documentId"`
	GroupID    string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_doc_group_share" json:"groupId"`
	SharedBy   string    `gorm:"type:uuid;not null" json:"sharedBy"`
	SharedAt   time.Time `gorm:"autoCreateTime" json:"sharedAt"`

	Document Document `gorm:"foreignKey:DocumentID" json:"-"`
	Group    Group    `gorm:"foreignKey:GroupID" json:"-"`
	Sharer   User     `gorm:"foreignKey:SharedBy" json:"-"`
}

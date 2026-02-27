package models

import "time"

type Group struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `gorm:"type:uuid;not null;index" json:"createdBy"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`

	Creator User `gorm:"foreignKey:CreatedBy" json:"-"`
}

type GroupMember struct {
	ID      string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	GroupID string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_group_user" json:"groupId"`
	UserID  string    `gorm:"type:uuid;not null;index;uniqueIndex:idx_group_user" json:"userId"`
	AddedAt time.Time `gorm:"autoCreateTime" json:"addedAt"`

	Group Group `gorm:"foreignKey:GroupID" json:"-"`
	User  User  `gorm:"foreignKey:UserID" json:"-"`
}

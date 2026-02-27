package models

import "time"

type Role string

const (
	RoleClient Role = "client"
	RoleStaff  Role = "staff"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FirstName    string     `gorm:"not null" json:"firstName"`
	LastName     string     `gorm:"not null" json:"lastName"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Role         Role       `gorm:"type:varchar(20);not null;default:'client'" json:"role"`
	MFAEnabled   bool       `gorm:"default:false" json:"mfaEnabled"`
	MFASecret    string     `gorm:"" json:"-"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	LastLogin    *time.Time `json:"lastLogin"`
}

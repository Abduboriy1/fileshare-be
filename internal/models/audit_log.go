package models

import "time"

type AuditAction string

const (
	AuditActionUpload      AuditAction = "UPLOAD"
	AuditActionDownload    AuditAction = "DOWNLOAD"
	AuditActionDelete      AuditAction = "DELETE"
	AuditActionLogin       AuditAction = "LOGIN"
	AuditActionFailedLogin AuditAction = "FAILED_LOGIN"
	AuditActionView        AuditAction = "VIEW"
	AuditActionShare       AuditAction = "SHARE"
	AuditActionGroupCreate AuditAction = "GROUP_CREATE"
)

type AuditLog struct {
	ID         string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     string      `gorm:"type:uuid;not null;index" json:"userId"`
	Action     AuditAction `gorm:"type:varchar(20);not null" json:"action"`
	DocumentID *string     `gorm:"type:uuid" json:"documentId"`
	IPAddress  string      `gorm:"type:varchar(45)" json:"ipAddress"`
	Timestamp  time.Time   `gorm:"autoCreateTime" json:"timestamp"`
}

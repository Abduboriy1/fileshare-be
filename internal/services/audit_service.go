package services

import (
	"context"
	"fileshare-be/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuditService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewAuditService(db *gorm.DB, logger *zap.Logger) *AuditService {
	return &AuditService{db: db, logger: logger}
}

func (s *AuditService) Log(ctx context.Context, userID string, action models.AuditAction, documentID *string, ip string) {
	entry := models.AuditLog{
		UserID:     userID,
		Action:     action,
		DocumentID: documentID,
		IPAddress:  ip,
	}

	if err := s.db.WithContext(ctx).Create(&entry).Error; err != nil {
		s.logger.Error("failed to create audit log",
			zap.String("user_id", userID),
			zap.String("action", string(action)),
			zap.Error(err),
		)
	}
}

package services

import (
	"context"
	"fileshare-be/internal/models"
	"fmt"
	"math"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ViewService struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewViewService(db *gorm.DB, logger *zap.Logger) *ViewService {
	return &ViewService{db: db, logger: logger}
}

func (s *ViewService) RecordView(ctx context.Context, userID string, docID string, ip string) {
	view := models.DocumentView{
		DocumentID: docID,
		UserID:     userID,
		IPAddress:  ip,
	}

	if err := s.db.WithContext(ctx).Create(&view).Error; err != nil {
		s.logger.Error("failed to record document view", zap.String("docID", docID), zap.Error(err))
	}
}

func (s *ViewService) GetDocumentViews(ctx context.Context, requesterID string, requesterRole models.Role, docID string, page, pageSize int) (*models.PaginatedResponse[models.DocumentViewResponse], error) {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return nil, ErrDocumentNotFound
	}

	if requesterRole == models.RoleClient && doc.UserID != requesterID {
		return nil, ErrForbidden
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&models.DocumentView{}).Where("document_id = ?", docID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("counting views: %w", err)
	}

	var views []models.DocumentView
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Preload("User").Where("document_id = ?", docID).
		Order("viewed_at DESC").Offset(offset).Limit(pageSize).Find(&views).Error; err != nil {
		return nil, fmt.Errorf("listing views: %w", err)
	}

	responses := make([]models.DocumentViewResponse, len(views))
	for i, v := range views {
		responses[i] = models.DocumentViewResponse{
			ID:         v.ID,
			DocumentID: v.DocumentID,
			UserID:     v.UserID,
			FirstName:  v.User.FirstName,
			LastName:   v.User.LastName,
			Email:      v.User.Email,
			IPAddress:  v.IPAddress,
			ViewedAt:   v.ViewedAt,
		}
	}

	return &models.PaginatedResponse[models.DocumentViewResponse]{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

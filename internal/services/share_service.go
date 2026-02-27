package services

import (
	"context"
	"fileshare-be/internal/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ShareService struct {
	db     *gorm.DB
	audit  *AuditService
	logger *zap.Logger
}

func NewShareService(db *gorm.DB, audit *AuditService, logger *zap.Logger) *ShareService {
	return &ShareService{db: db, audit: audit, logger: logger}
}

func (s *ShareService) ShareWithUser(ctx context.Context, ownerID string, docID string, req models.ShareDocumentRequest, ip string) (*models.DocumentShareResponse, error) {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return nil, ErrDocumentNotFound
	}

	if doc.UserID != ownerID {
		return nil, ErrForbidden
	}

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "email = ?", req.UserEmail).Error; err != nil {
		return nil, ErrUserNotFound
	}

	share := models.DocumentShare{
		DocumentID: docID,
		UserID:     user.ID,
		SharedBy:   ownerID,
	}

	if err := s.db.WithContext(ctx).Create(&share).Error; err != nil {
		return nil, ErrAlreadyShared
	}

	s.audit.Log(ctx, ownerID, models.AuditActionShare, &docID, ip)

	return &models.DocumentShareResponse{
		ID:         share.ID,
		DocumentID: share.DocumentID,
		UserID:     user.ID,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Email:      user.Email,
		SharedAt:   share.SharedAt,
	}, nil
}

func (s *ShareService) ShareWithGroup(ctx context.Context, ownerID string, docID string, req models.ShareDocumentWithGroupRequest, ip string) (*models.DocumentGroupShareResponse, error) {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return nil, ErrDocumentNotFound
	}

	if doc.UserID != ownerID {
		return nil, ErrForbidden
	}

	var group models.Group
	if err := s.db.WithContext(ctx).First(&group, "id = ?", req.GroupID).Error; err != nil {
		return nil, ErrGroupNotFound
	}

	share := models.DocumentGroupShare{
		DocumentID: docID,
		GroupID:    group.ID,
		SharedBy:   ownerID,
	}

	if err := s.db.WithContext(ctx).Create(&share).Error; err != nil {
		return nil, ErrAlreadyShared
	}

	s.audit.Log(ctx, ownerID, models.AuditActionShare, &docID, ip)

	return &models.DocumentGroupShareResponse{
		ID:         share.ID,
		DocumentID: share.DocumentID,
		GroupID:    group.ID,
		GroupName:  group.Name,
		SharedAt:   share.SharedAt,
	}, nil
}

func (s *ShareService) RevokeUserShare(ctx context.Context, ownerID string, docID string, shareID string) error {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return ErrDocumentNotFound
	}

	if doc.UserID != ownerID {
		return ErrForbidden
	}

	result := s.db.WithContext(ctx).Where("id = ? AND document_id = ?", shareID, docID).Delete(&models.DocumentShare{})
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (s *ShareService) RevokeGroupShare(ctx context.Context, ownerID string, docID string, shareID string) error {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return ErrDocumentNotFound
	}

	if doc.UserID != ownerID {
		return ErrForbidden
	}

	result := s.db.WithContext(ctx).Where("id = ? AND document_id = ?", shareID, docID).Delete(&models.DocumentGroupShare{})
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (s *ShareService) ListDocumentShares(ctx context.Context, ownerID string, docID string) ([]models.DocumentShareResponse, error) {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return nil, ErrDocumentNotFound
	}

	if doc.UserID != ownerID {
		return nil, ErrForbidden
	}

	var shares []models.DocumentShare
	if err := s.db.WithContext(ctx).Preload("User").Where("document_id = ?", docID).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("listing shares: %w", err)
	}

	responses := make([]models.DocumentShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = models.DocumentShareResponse{
			ID:         share.ID,
			DocumentID: share.DocumentID,
			UserID:     share.UserID,
			FirstName:  share.User.FirstName,
			LastName:   share.User.LastName,
			Email:      share.User.Email,
			SharedAt:   share.SharedAt,
		}
	}

	return responses, nil
}

func (s *ShareService) ListDocumentGroupShares(ctx context.Context, ownerID string, docID string) ([]models.DocumentGroupShareResponse, error) {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return nil, ErrDocumentNotFound
	}

	if doc.UserID != ownerID {
		return nil, ErrForbidden
	}

	var shares []models.DocumentGroupShare
	if err := s.db.WithContext(ctx).Preload("Group").Where("document_id = ?", docID).Find(&shares).Error; err != nil {
		return nil, fmt.Errorf("listing group shares: %w", err)
	}

	responses := make([]models.DocumentGroupShareResponse, len(shares))
	for i, share := range shares {
		responses[i] = models.DocumentGroupShareResponse{
			ID:         share.ID,
			DocumentID: share.DocumentID,
			GroupID:    share.GroupID,
			GroupName:  share.Group.Name,
			SharedAt:   share.SharedAt,
		}
	}

	return responses, nil
}

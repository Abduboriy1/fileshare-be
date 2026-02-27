package services

import (
	"context"
	"fileshare-be/internal/models"
	"fileshare-be/pkg/crypto"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DocumentService struct {
	db            *gorm.DB
	s3Client      *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	appKey        string
	audit         *AuditService
	viewService   *ViewService
	logger        *zap.Logger
}

func NewDocumentService(db *gorm.DB, s3Client *s3.Client, presignClient *s3.PresignClient, bucket string, appKey string, audit *AuditService, viewService *ViewService, logger *zap.Logger) *DocumentService {
	return &DocumentService{
		db:            db,
		s3Client:      s3Client,
		presignClient: presignClient,
		bucket:        bucket,
		appKey:        appKey,
		audit:         audit,
		viewService:   viewService,
		logger:        logger,
	}
}

func (s *DocumentService) InitiateUpload(ctx context.Context, userID string, req models.UploadRequest, ip string) (*models.UploadResponse, error) {
	if req.FileName == "" || req.FileSize <= 0 || req.MimeType == "" {
		return nil, ErrInvalidInput
	}

	storageKey := uuid.New().String()

	expiresAt := time.Now().Add(90 * 24 * time.Hour)
	doc := models.Document{
		UserID:     userID,
		FileName:   req.FileName,
		FileSize:   req.FileSize,
		MimeType:   req.MimeType,
		StorageKey: storageKey,
		ExpiresAt:  &expiresAt,
		Status:     models.DocumentStatusActive,
	}

	if req.SecretKey != "" {
		encrypted, err := crypto.Encrypt(req.SecretKey, s.appKey)
		if err != nil {
			return nil, fmt.Errorf("encrypting secret key: %w", err)
		}
		doc.SecretKeyHash = encrypted
	}

	if err := s.db.WithContext(ctx).Create(&doc).Error; err != nil {
		return nil, fmt.Errorf("creating document record: %w", err)
	}

	presignResult, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(storageKey),
		ContentType: aws.String(req.MimeType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	s.audit.Log(ctx, userID, models.AuditActionUpload, &doc.ID, ip)

	return &models.UploadResponse{
		DocumentID: doc.ID,
		UploadURL:  presignResult.URL,
	}, nil
}

func (s *DocumentService) ListDocuments(ctx context.Context, userID string, role models.Role, page, pageSize int) (*models.PaginatedResponse[models.DocumentResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&models.Document{}).Where("status = ?", models.DocumentStatusActive)
	if role == models.RoleClient {
		query = query.Where(
			"user_id = ? OR id IN (?) OR id IN (?)",
			userID,
			s.db.Model(&models.DocumentShare{}).Select("document_id").Where("user_id = ?", userID),
			s.db.Model(&models.DocumentGroupShare{}).Select("document_id").
				Where("group_id IN (?)",
					s.db.Model(&models.GroupMember{}).Select("group_id").Where("user_id = ?", userID),
				),
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("counting documents: %w", err)
	}

	var docs []models.Document
	offset := (page - 1) * pageSize
	if err := query.Order("uploaded_at DESC").Offset(offset).Limit(pageSize).Find(&docs).Error; err != nil {
		return nil, fmt.Errorf("listing documents: %w", err)
	}

	responses := make([]models.DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(&doc, userID, s.appKey)
	}

	return &models.PaginatedResponse[models.DocumentResponse]{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

func (s *DocumentService) GetDownloadURL(ctx context.Context, userID string, role models.Role, docID string, ip string) (*models.DownloadResponse, error) {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return nil, ErrDocumentNotFound
	}

	// Access check: owner, shared user, group member, or staff/admin
	if role == models.RoleClient && doc.UserID != userID {
		if !s.hasAccess(ctx, userID, docID) {
			return nil, ErrForbidden
		}
	}

	presignResult, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(doc.StorageKey),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("generating download URL: %w", err)
	}

	s.audit.Log(ctx, userID, models.AuditActionDownload, &doc.ID, ip)
	s.viewService.RecordView(ctx, userID, doc.ID, ip)

	return &models.DownloadResponse{
		DownloadURL: presignResult.URL,
		FileName:    doc.FileName,
	}, nil
}

func (s *DocumentService) DeleteDocument(ctx context.Context, userID string, role models.Role, docID string, ip string) error {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return ErrDocumentNotFound
	}

	// Allow owner, staff, or admin
	if doc.UserID != userID && role == models.RoleClient {
		return ErrForbidden
	}

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(doc.StorageKey),
	})
	if err != nil {
		s.logger.Error("failed to delete S3 object", zap.String("key", doc.StorageKey), zap.Error(err))
	}

	if err := s.db.WithContext(ctx).Model(&doc).Update("status", models.DocumentStatusDeleted).Error; err != nil {
		return fmt.Errorf("marking document deleted: %w", err)
	}

	s.audit.Log(ctx, userID, models.AuditActionDelete, &doc.ID, ip)

	return nil
}

func (s *DocumentService) CleanupExpired(ctx context.Context) error {
	var docs []models.Document
	if err := s.db.WithContext(ctx).Where("expires_at < ? AND status = ?", time.Now(), models.DocumentStatusActive).Find(&docs).Error; err != nil {
		return fmt.Errorf("querying expired documents: %w", err)
	}

	systemUser := "system"
	for _, doc := range docs {
		_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(doc.StorageKey),
		})
		if err != nil {
			s.logger.Error("failed to delete expired S3 object", zap.String("key", doc.StorageKey), zap.Error(err))
		}

		s.db.WithContext(ctx).Model(&doc).Update("status", models.DocumentStatusDeleted)
		s.audit.Log(ctx, systemUser, models.AuditActionDelete, &doc.ID, "system")
	}

	if len(docs) > 0 {
		s.logger.Info("cleaned up expired documents", zap.Int("count", len(docs)))
	}

	return nil
}

func (s *DocumentService) SetSecretKey(ctx context.Context, userID string, docID string, req models.SetSecretKeyRequest) error {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return ErrDocumentNotFound
	}

	if doc.UserID != userID {
		return ErrForbidden
	}

	if req.SecretKey == "" {
		return ErrInvalidInput
	}

	encrypted, err := crypto.Encrypt(req.SecretKey, s.appKey)
	if err != nil {
		return fmt.Errorf("encrypting secret key: %w", err)
	}

	return s.db.WithContext(ctx).Model(&doc).Update("secret_key_hash", encrypted).Error
}

func (s *DocumentService) RemoveSecretKey(ctx context.Context, userID string, docID string) error {
	var doc models.Document
	if err := s.db.WithContext(ctx).First(&doc, "id = ? AND status = ?", docID, models.DocumentStatusActive).Error; err != nil {
		return ErrDocumentNotFound
	}

	if doc.UserID != userID {
		return ErrForbidden
	}

	return s.db.WithContext(ctx).Model(&doc).Update("secret_key_hash", "").Error
}

// hasAccess checks if a user has access via direct share or group share.
func (s *DocumentService) hasAccess(ctx context.Context, userID string, docID string) bool {
	var count int64

	// Check direct share
	s.db.WithContext(ctx).Model(&models.DocumentShare{}).
		Where("document_id = ? AND user_id = ?", docID, userID).
		Count(&count)
	if count > 0 {
		return true
	}

	// Check group share
	s.db.WithContext(ctx).Model(&models.DocumentGroupShare{}).
		Where("document_id = ? AND group_id IN (?)",
			docID,
			s.db.Model(&models.GroupMember{}).Select("group_id").Where("user_id = ?", userID),
		).Count(&count)

	return count > 0
}

func toDocumentResponse(d *models.Document, requestingUserID string, appKey string) models.DocumentResponse {
	resp := models.DocumentResponse{
		ID:           d.ID,
		UserID:       d.UserID,
		FileName:     d.FileName,
		FileSize:     d.FileSize,
		MimeType:     d.MimeType,
		UploadedAt:   d.UploadedAt,
		ExpiresAt:    d.ExpiresAt,
		Status:       d.Status,
		HasSecretKey: d.SecretKeyHash != "",
		IsOwner:      d.UserID == requestingUserID,
	}
	if d.SecretKeyHash != "" {
		if plaintext, err := crypto.Decrypt(d.SecretKeyHash, appKey); err == nil {
			resp.SecretKey = plaintext
		}
	}
	return resp
}

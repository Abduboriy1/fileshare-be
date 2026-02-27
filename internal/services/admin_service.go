package services

import (
	"context"
	"fileshare-be/internal/models"
	"fmt"
	"math"

	"gorm.io/gorm"
)

type AdminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) *AdminService {
	return &AdminService{db: db}
}

func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) (*models.PaginatedResponse[models.UserResponse], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}

	var users []models.User
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	responses := make([]models.UserResponse, len(users))
	for i, u := range users {
		responses[i] = toUserResponse(&u)
	}

	return &models.PaginatedResponse[models.UserResponse]{
		Data:       responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

func (s *AdminService) GetAuditLogs(ctx context.Context, page, pageSize int) (*models.PaginatedResponse[models.AuditLog], error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&models.AuditLog{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("counting audit logs: %w", err)
	}

	var logs []models.AuditLog
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Order("timestamp DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("listing audit logs: %w", err)
	}

	return &models.PaginatedResponse[models.AuditLog]{
		Data:       logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
	}, nil
}

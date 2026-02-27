package services

import (
	"context"
	"fileshare-be/internal/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GroupService struct {
	db     *gorm.DB
	audit  *AuditService
	logger *zap.Logger
}

func NewGroupService(db *gorm.DB, audit *AuditService, logger *zap.Logger) *GroupService {
	return &GroupService{db: db, audit: audit, logger: logger}
}

func (s *GroupService) CreateGroup(ctx context.Context, userID string, req models.CreateGroupRequest, ip string) (*models.GroupResponse, error) {
	if req.Name == "" {
		return nil, ErrInvalidInput
	}

	group := models.Group{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := s.db.WithContext(ctx).Create(&group).Error; err != nil {
		return nil, fmt.Errorf("creating group: %w", err)
	}

	s.audit.Log(ctx, userID, models.AuditActionGroupCreate, nil, ip)

	return &models.GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedBy:   group.CreatedBy,
		CreatedAt:   group.CreatedAt,
		MemberCount: 0,
	}, nil
}

func (s *GroupService) ListGroups(ctx context.Context, userID string) ([]models.GroupResponse, error) {
	var groups []models.Group

	err := s.db.WithContext(ctx).
		Where("created_by = ? OR id IN (?)",
			userID,
			s.db.Model(&models.GroupMember{}).Select("group_id").Where("user_id = ?", userID),
		).
		Find(&groups).Error
	if err != nil {
		return nil, fmt.Errorf("listing groups: %w", err)
	}

	if len(groups) == 0 {
		return []models.GroupResponse{}, nil
	}

	groupIDs := make([]string, len(groups))
	for i, g := range groups {
		groupIDs[i] = g.ID
	}

	type countResult struct {
		GroupID string
		Count   int
	}
	var counts []countResult
	s.db.WithContext(ctx).
		Model(&models.GroupMember{}).
		Select("group_id, count(*) as count").
		Where("group_id IN ?", groupIDs).
		Group("group_id").
		Scan(&counts)

	countMap := make(map[string]int)
	for _, c := range counts {
		countMap[c.GroupID] = c.Count
	}

	responses := make([]models.GroupResponse, len(groups))
	for i, g := range groups {
		responses[i] = models.GroupResponse{
			ID:          g.ID,
			Name:        g.Name,
			Description: g.Description,
			CreatedBy:   g.CreatedBy,
			CreatedAt:   g.CreatedAt,
			MemberCount: countMap[g.ID],
		}
	}

	return responses, nil
}

func (s *GroupService) GetGroup(ctx context.Context, userID string, groupID string) (*models.GroupResponse, []models.GroupMemberResponse, error) {
	var group models.Group
	if err := s.db.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		return nil, nil, ErrGroupNotFound
	}

	// Check user is creator or member
	if group.CreatedBy != userID {
		var count int64
		s.db.WithContext(ctx).Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
		if count == 0 {
			return nil, nil, ErrForbidden
		}
	}

	var members []models.GroupMember
	s.db.WithContext(ctx).Preload("User").Where("group_id = ?", groupID).Find(&members)

	memberResponses := make([]models.GroupMemberResponse, len(members))
	for i, m := range members {
		memberResponses[i] = models.GroupMemberResponse{
			ID:        m.ID,
			UserID:    m.UserID,
			FirstName: m.User.FirstName,
			LastName:  m.User.LastName,
			Email:     m.User.Email,
			AddedAt:   m.AddedAt,
		}
	}

	return &models.GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatedBy:   group.CreatedBy,
		CreatedAt:   group.CreatedAt,
		MemberCount: len(members),
	}, memberResponses, nil
}

func (s *GroupService) AddMember(ctx context.Context, userID string, groupID string, req models.AddGroupMemberRequest) (*models.GroupMemberResponse, error) {
	var group models.Group
	if err := s.db.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		return nil, ErrGroupNotFound
	}

	if group.CreatedBy != userID {
		return nil, ErrNotGroupOwner
	}

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "email = ?", req.UserEmail).Error; err != nil {
		return nil, ErrUserNotFound
	}

	member := models.GroupMember{
		GroupID: groupID,
		UserID:  user.ID,
	}

	if err := s.db.WithContext(ctx).Create(&member).Error; err != nil {
		return nil, ErrAlreadyMember
	}

	return &models.GroupMemberResponse{
		ID:        member.ID,
		UserID:    user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		AddedAt:   member.AddedAt,
	}, nil
}

func (s *GroupService) RemoveMember(ctx context.Context, userID string, groupID string, memberUserID string) error {
	var group models.Group
	if err := s.db.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		return ErrGroupNotFound
	}

	if group.CreatedBy != userID {
		return ErrNotGroupOwner
	}

	result := s.db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, memberUserID).Delete(&models.GroupMember{})
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, userID string, groupID string) error {
	var group models.Group
	if err := s.db.WithContext(ctx).First(&group, "id = ?", groupID).Error; err != nil {
		return ErrGroupNotFound
	}

	if group.CreatedBy != userID {
		return ErrNotGroupOwner
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error; err != nil {
			return err
		}
		if err := tx.Where("group_id = ?", groupID).Delete(&models.DocumentGroupShare{}).Error; err != nil {
			return err
		}
		return tx.Delete(&group).Error
	})
}

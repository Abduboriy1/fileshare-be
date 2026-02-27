package services

import (
	"context"
	"fileshare-be/internal/auth"
	"fileshare-be/internal/models"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db       *gorm.DB
	jwt      *auth.JWTManager
	audit    *AuditService
	logger   *zap.Logger
}

func NewAuthService(db *gorm.DB, jwt *auth.JWTManager, audit *AuditService, logger *zap.Logger) *AuthService {
	return &AuthService{db: db, jwt: jwt, audit: audit, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest, ip string) (*models.AuthResponse, string, error) {
	if req.Email == "" || req.Password == "" {
		return nil, "", ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hashing password: %w", err)
	}

	user := models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         models.RoleClient,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, "", ErrEmailTaken
		}
		return nil, "", fmt.Errorf("creating user: %w", err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generating refresh token: %w", err)
	}

	s.audit.Log(ctx, user.ID, models.AuditActionLogin, nil, ip)

	return &models.AuthResponse{
		AccessToken: accessToken,
		User:        toUserResponse(&user),
	}, refreshToken, nil
}

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest, ip string) (*models.AuthResponse, string, error) {
	if req.Email == "" || req.Password == "" {
		return nil, "", ErrInvalidInput
	}

	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		s.audit.Log(ctx, "", models.AuditActionFailedLogin, nil, ip)
		return nil, "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		s.audit.Log(ctx, user.ID, models.AuditActionFailedLogin, nil, ip)
		return nil, "", ErrInvalidCredentials
	}

	if user.MFAEnabled {
		if req.TOTPCode == "" {
			return nil, "", ErrMFARequired
		}
		if !totp.Validate(req.TOTPCode, user.MFASecret) {
			s.audit.Log(ctx, user.ID, models.AuditActionFailedLogin, nil, ip)
			return nil, "", ErrInvalidTOTP
		}
	}

	now := time.Now()
	s.db.WithContext(ctx).Model(&user).Update("last_login", now)
	user.LastLogin = &now

	accessToken, err := s.jwt.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", fmt.Errorf("generating access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generating refresh token: %w", err)
	}

	s.audit.Log(ctx, user.ID, models.AuditActionLogin, nil, ip)

	return &models.AuthResponse{
		AccessToken: accessToken,
		User:        toUserResponse(&user),
	}, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, tokenStr string, ip string) (*models.RefreshResponse, string, error) {
	claims, err := s.jwt.ValidateRefreshToken(tokenStr)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", claims.UserID).Error; err != nil {
		return nil, "", ErrUserNotFound
	}

	accessToken, err := s.jwt.GenerateAccessToken(&user)
	if err != nil {
		return nil, "", fmt.Errorf("generating access token: %w", err)
	}

	newRefreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("generating refresh token: %w", err)
	}

	return &models.RefreshResponse{
		AccessToken: accessToken,
	}, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID string, ip string) error {
	s.audit.Log(ctx, userID, models.AuditActionLogin, nil, ip)
	return nil
}

func toUserResponse(u *models.User) models.UserResponse {
	return models.UserResponse{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Role:       u.Role,
		MFAEnabled: u.MFAEnabled,
		CreatedAt:  u.CreatedAt,
		LastLogin:  u.LastLogin,
	}
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "UNIQUE constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fileshare-be/internal/models"
	"fmt"
	"image/png"

	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MFAService struct {
	db     *gorm.DB
	audit  *AuditService
	logger *zap.Logger
}

func NewMFAService(db *gorm.DB, audit *AuditService, logger *zap.Logger) *MFAService {
	return &MFAService{db: db, audit: audit, logger: logger}
}

func (s *MFAService) SetupMFA(ctx context.Context, userID string) (*models.MFASetupResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Fileshare",
		AccountName: user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("generating TOTP key: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&user).Update("mfa_secret", key.Secret()).Error; err != nil {
		return nil, fmt.Errorf("storing MFA secret: %w", err)
	}

	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("generating QR image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encoding QR image: %w", err)
	}

	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &models.MFASetupResponse{
		Secret: key.Secret(),
		QRCode: "data:image/png;base64," + qrBase64,
	}, nil
}

func (s *MFAService) VerifyAndEnableMFA(ctx context.Context, userID string, code string, ip string) error {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return ErrUserNotFound
	}

	if user.MFASecret == "" {
		return ErrInvalidInput
	}

	if !totp.Validate(code, user.MFASecret) {
		return ErrInvalidTOTP
	}

	if err := s.db.WithContext(ctx).Model(&user).Update("mfa_enabled", true).Error; err != nil {
		return fmt.Errorf("enabling MFA: %w", err)
	}

	return nil
}

func (s *MFAService) DisableMFA(ctx context.Context, userID string, ip string) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).
		Updates(map[string]interface{}{"mfa_enabled": false, "mfa_secret": ""})

	if result.Error != nil {
		return fmt.Errorf("disabling MFA: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

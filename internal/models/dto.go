package models

import "time"

// Auth DTOs

type RegisterRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totpCode,omitempty"`
}

type AuthResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
}

type UserResponse struct {
	ID        string     `json:"id"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Email     string     `json:"email"`
	Role       Role       `json:"role"`
	MFAEnabled bool       `json:"mfaEnabled"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastLogin  *time.Time `json:"lastLogin"`
}

// MFA DTOs

type MFASetupResponse struct {
	Secret string `json:"secret"`
	QRCode string `json:"qrCode"`
}

type MFAVerifyRequest struct {
	Code string `json:"code"`
}

// Document DTOs

type UploadRequest struct {
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize"`
	MimeType  string `json:"mimeType"`
	SecretKey string `json:"secretKey,omitempty"`
}

type UploadResponse struct {
	DocumentID string `json:"documentId"`
	UploadURL  string `json:"uploadUrl"`
	Method     string `json:"method"`
}

type DocumentResponse struct {
	ID           string         `json:"id"`
	UserID       string         `json:"userId"`
	FileName     string         `json:"fileName"`
	FileSize     int64          `json:"fileSize"`
	MimeType     string         `json:"mimeType"`
	UploadedAt   time.Time      `json:"uploadedAt"`
	ExpiresAt    *time.Time     `json:"expiresAt"`
	Status       DocumentStatus `json:"status"`
	HasSecretKey bool           `json:"hasSecretKey"`
	SecretKey    string         `json:"secretKey"`
	IsOwner      bool           `json:"isOwner"`
}

type DownloadResponse struct {
	DownloadURL string `json:"downloadUrl"`
	FileName    string `json:"fileName"`
}

// Secret Key DTOs

type SetSecretKeyRequest struct {
	SecretKey string `json:"secretKey"`
}

type VerifySecretKeyRequest struct {
	SecretKey string `json:"secretKey"`
}

// Share DTOs

type ShareDocumentRequest struct {
	UserEmail string `json:"userEmail"`
}

type ShareDocumentWithGroupRequest struct {
	GroupID string `json:"groupId"`
}

type DocumentShareResponse struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	UserID     string    `json:"userId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	SharedAt   time.Time `json:"sharedAt"`
}

type DocumentGroupShareResponse struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	GroupID    string    `json:"groupId"`
	GroupName  string    `json:"groupName"`
	SharedAt   time.Time `json:"sharedAt"`
}

// Group DTOs

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddGroupMemberRequest struct {
	UserEmail string `json:"userEmail"`
}

type GroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	MemberCount int       `json:"memberCount"`
}

type GroupMemberResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	AddedAt   time.Time `json:"addedAt"`
}

// View DTOs

type DocumentViewResponse struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"documentId"`
	UserID     string    `json:"userId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	IPAddress  string    `json:"ipAddress"`
	ViewedAt   time.Time `json:"viewedAt"`
}

// Generic DTOs

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

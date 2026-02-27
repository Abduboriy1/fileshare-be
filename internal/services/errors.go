package services

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailTaken         = errors.New("email already taken")
	ErrMFARequired        = errors.New("MFA verification required")
	ErrInvalidTOTP        = errors.New("invalid TOTP code")
	ErrDocumentNotFound   = errors.New("document not found")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUserNotFound       = errors.New("user not found")
	ErrGroupNotFound      = errors.New("group not found")
	ErrAlreadyShared      = errors.New("already shared")
	ErrAlreadyMember      = errors.New("already a member")
	ErrSecretKeyRequired  = errors.New("secret key required")
	ErrInvalidSecretKey   = errors.New("invalid secret key")
	ErrNotGroupOwner      = errors.New("not group owner")
)

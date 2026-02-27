package handlers

import (
	"errors"
	"fileshare-be/internal/middleware"
	"fileshare-be/internal/models"
	"fileshare-be/internal/services"
	"net/http"
)

type AuthHandler struct {
	authService *services.AuthService
	mfaService  *services.MFAService
}

func NewAuthHandler(authService *services.AuthService, mfaService *services.MFAService) *AuthHandler {
	return &AuthHandler{authService: authService, mfaService: mfaService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, refreshToken, err := h.authService.Register(r.Context(), req, getIPAddress(r))
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	setRefreshTokenCookie(w, refreshToken)
	respondJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, refreshToken, err := h.authService.Login(r.Context(), req, getIPAddress(r))
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	setRefreshTokenCookie(w, refreshToken)
	respondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		respondError(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	resp, newRefreshToken, err := h.authService.RefreshToken(r.Context(), cookie.Value, getIPAddress(r))
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	setRefreshTokenCookie(w, newRefreshToken)
	respondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	_ = h.authService.Logout(r.Context(), claims.UserID, getIPAddress(r))

	clearRefreshTokenCookie(w)
	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.mfaService.SetupMFA(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to setup MFA")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.MFAVerifyRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.mfaService.VerifyAndEnableMFA(r.Context(), claims.UserID, req.Code, getIPAddress(r)); err != nil {
		if errors.Is(err, services.ErrInvalidTOTP) {
			respondError(w, http.StatusBadRequest, "invalid TOTP code")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to verify MFA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "MFA enabled"})
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, services.ErrEmailTaken):
		respondError(w, http.StatusConflict, "email already taken")
	case errors.Is(err, services.ErrMFARequired):
		respondError(w, http.StatusForbidden, "MFA verification required")
	case errors.Is(err, services.ErrInvalidTOTP):
		respondError(w, http.StatusUnauthorized, "invalid TOTP code")
	case errors.Is(err, services.ErrInvalidInput):
		respondError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, services.ErrUserNotFound):
		respondError(w, http.StatusUnauthorized, "invalid credentials")
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

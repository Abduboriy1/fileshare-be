package handlers

import (
	"errors"
	"fileshare-be/internal/middleware"
	"fileshare-be/internal/models"
	"fileshare-be/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ShareHandler struct {
	shareService *services.ShareService
}

func NewShareHandler(shareService *services.ShareService) *ShareHandler {
	return &ShareHandler{shareService: shareService}
}

func (h *ShareHandler) ShareWithUser(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	var req models.ShareDocumentRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.shareService.ShareWithUser(r.Context(), claims.UserID, docID, req, getIPAddress(r))
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *ShareHandler) ShareWithGroup(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	var req models.ShareDocumentWithGroupRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.shareService.ShareWithGroup(r.Context(), claims.UserID, docID, req, getIPAddress(r))
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *ShareHandler) ListDocumentShares(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	shares, err := h.shareService.ListDocumentShares(r.Context(), claims.UserID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, shares)
}

func (h *ShareHandler) ListDocumentGroupShares(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	shares, err := h.shareService.ListDocumentGroupShares(r.Context(), claims.UserID, docID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, shares)
}

func (h *ShareHandler) RevokeUserShare(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")
	shareID := chi.URLParam(r, "shareId")

	if err := h.shareService.RevokeUserShare(r.Context(), claims.UserID, docID, shareID); err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "share revoked"})
}

func (h *ShareHandler) RevokeGroupShare(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")
	shareID := chi.URLParam(r, "shareId")

	if err := h.shareService.RevokeGroupShare(r.Context(), claims.UserID, docID, shareID); err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "group share revoked"})
}

func (h *ShareHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrDocumentNotFound):
		respondError(w, http.StatusNotFound, "document not found")
	case errors.Is(err, services.ErrForbidden):
		respondError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, services.ErrUserNotFound):
		respondError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, services.ErrGroupNotFound):
		respondError(w, http.StatusNotFound, "group not found")
	case errors.Is(err, services.ErrAlreadyShared):
		respondError(w, http.StatusConflict, "already shared")
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

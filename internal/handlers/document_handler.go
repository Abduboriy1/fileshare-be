package handlers

import (
	"errors"
	"fileshare-be/internal/middleware"
	"fileshare-be/internal/models"
	"fileshare-be/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type DocumentHandler struct {
	docService  *services.DocumentService
	viewService *services.ViewService
}

func NewDocumentHandler(docService *services.DocumentService, viewService *services.ViewService) *DocumentHandler {
	return &DocumentHandler{docService: docService, viewService: viewService}
}

func (h *DocumentHandler) InitiateUpload(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.UploadRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.docService.InitiateUpload(r.Context(), claims.UserID, req, getIPAddress(r))
	if err != nil {
		h.handleDocError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *DocumentHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	resp, err := h.docService.ListDocuments(r.Context(), claims.UserID, claims.Role, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DocumentHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	resp, err := h.docService.GetDownloadURL(r.Context(), claims.UserID, claims.Role, docID, getIPAddress(r))
	if err != nil {
		h.handleDocError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")
	if err := h.docService.DeleteDocument(r.Context(), claims.UserID, claims.Role, docID, getIPAddress(r)); err != nil {
		h.handleDocError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "document deleted"})
}

func (h *DocumentHandler) SetSecretKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	var req models.SetSecretKeyRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.docService.SetSecretKey(r.Context(), claims.UserID, docID, req); err != nil {
		h.handleDocError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "secret key set"})
}

func (h *DocumentHandler) RemoveSecretKey(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")

	if err := h.docService.RemoveSecretKey(r.Context(), claims.UserID, docID); err != nil {
		h.handleDocError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "secret key removed"})
}

func (h *DocumentHandler) GetDocumentViews(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	docID := chi.URLParam(r, "id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))

	resp, err := h.viewService.GetDocumentViews(r.Context(), claims.UserID, claims.Role, docID, page, pageSize)
	if err != nil {
		h.handleDocError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

func (h *DocumentHandler) handleDocError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrDocumentNotFound):
		respondError(w, http.StatusNotFound, "document not found")
	case errors.Is(err, services.ErrForbidden):
		respondError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, services.ErrInvalidInput):
		respondError(w, http.StatusBadRequest, "invalid input")
	case errors.Is(err, services.ErrSecretKeyRequired):
		respondError(w, http.StatusForbidden, "secret key required")
	case errors.Is(err, services.ErrInvalidSecretKey):
		respondError(w, http.StatusForbidden, "invalid secret key")
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

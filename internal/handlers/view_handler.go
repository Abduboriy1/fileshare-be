package handlers

import (
	"errors"
	"fileshare-be/internal/middleware"
	"fileshare-be/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ViewHandler struct {
	viewService *services.ViewService
}

func NewViewHandler(viewService *services.ViewService) *ViewHandler {
	return &ViewHandler{viewService: viewService}
}

func (h *ViewHandler) GetDocumentViews(w http.ResponseWriter, r *http.Request) {
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
		switch {
		case errors.Is(err, services.ErrDocumentNotFound):
			respondError(w, http.StatusNotFound, "document not found")
		case errors.Is(err, services.ErrForbidden):
			respondError(w, http.StatusForbidden, "forbidden")
		default:
			respondError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

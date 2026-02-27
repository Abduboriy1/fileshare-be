package handlers

import (
	"errors"
	"fileshare-be/internal/middleware"
	"fileshare-be/internal/models"
	"fileshare-be/internal/services"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req models.CreateGroupRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.groupService.CreateGroup(r.Context(), claims.UserID, req, getIPAddress(r))
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groups, err := h.groupService.ListGroups(r.Context(), claims.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, groups)
}

func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "id")
	group, members, err := h.groupService.GetGroup(r.Context(), claims.UserID, groupID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"group":   group,
		"members": members,
	})
}

func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "id")
	if err := h.groupService.DeleteGroup(r.Context(), claims.UserID, groupID); err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "group deleted"})
}

func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "id")

	var req models.AddGroupMemberRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.groupService.AddMember(r.Context(), claims.UserID, groupID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID := chi.URLParam(r, "id")
	memberUserID := chi.URLParam(r, "userId")

	if err := h.groupService.RemoveMember(r.Context(), claims.UserID, groupID, memberUserID); err != nil {
		h.handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *GroupHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrGroupNotFound):
		respondError(w, http.StatusNotFound, "group not found")
	case errors.Is(err, services.ErrNotGroupOwner):
		respondError(w, http.StatusForbidden, "only group creator can perform this action")
	case errors.Is(err, services.ErrForbidden):
		respondError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, services.ErrUserNotFound):
		respondError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, services.ErrAlreadyMember):
		respondError(w, http.StatusConflict, "user is already a member")
	case errors.Is(err, services.ErrInvalidInput):
		respondError(w, http.StatusBadRequest, "invalid input")
	default:
		respondError(w, http.StatusInternalServerError, "internal server error")
	}
}

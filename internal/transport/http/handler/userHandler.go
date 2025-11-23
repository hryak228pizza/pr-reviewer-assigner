package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/services"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type GetReviewResponse struct {
	UserID       string      `json:"user_id"`
	PullRequests interface{} `json:"pull_requests"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON body")
		return
	}

	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)

	if err != nil {
		slog.Error("Failed to set user active status", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "user_id query parameter is required")
		return
	}

	prs, err := h.userService.GetReviews(r.Context(), userID)

	if err != nil {
		slog.Error("Failed to get reviews for user", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	resp := GetReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

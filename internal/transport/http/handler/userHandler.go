// Package handler processes incoming http requests
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

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type GetReviewResponse struct {
	UserID       string      `json:"user_id"`
	PullRequests interface{} `json:"pull_requests"`
}

// NewUserHandler initializes new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// SetIsActive updates the active status of a user
func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON body")
		return
	}

	// call service to update status
	user, err := h.userService.SetIsActive(r.Context(), req.UserID, req.IsActive)

	if err != nil {
		slog.Error("Failed to set user active status", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// GetReviews retrieves all reviews assigned to a specific user
func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	// extract user id from query parameters
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "user_id query parameter is required")
		return
	}

	// fetch reviews from service
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

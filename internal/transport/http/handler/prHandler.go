// Package handler processes incoming http requests
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/services"
)

type PRHandler struct {
	prService services.PRService
}

// NewPRHandler creates a new instance of prhandler
func NewPRHandler(prService services.PRService) *PRHandler {
	return &PRHandler{prService: prService}
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldReviewerID string `json:"old_reviewer_id"`
}

type PRIDRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// CreatePR processes request to create a new pull request
func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	// decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON body")
		return
	}

	// call service to create pr and assign reviewers
	pr, err := h.prService.Create(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)

	if err != nil {
		slog.Error("Failed to create PR", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{"pr": pr})
}

// MergePR handles request to mark a pr as merged
func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req PRIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON body")
		return
	}

	// update pr status in service
	pr, err := h.prService.Merge(r.Context(), req.PullRequestID)

	if err != nil {
		slog.Error("Failed to merge PR", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	respondWithJSON(w, http.StatusOK, pr)
}

// ReassignReviewer processes request to change a reviewer
func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON body")
		return
	}

	// find new reviewer and replace the old one
	pr, newReviewerID, err := h.prService.Reassign(r.Context(), req.PullRequestID, req.OldReviewerID)

	if err != nil {
		slog.Error("Failed to reassign reviewer", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	resp := struct {
		PullRequest   *entity.PullRequest `json:"pull_request"`
		NewReviewerID string              `json:"replaced_by"`
	}{
		PullRequest:   pr,
		NewReviewerID: newReviewerID,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

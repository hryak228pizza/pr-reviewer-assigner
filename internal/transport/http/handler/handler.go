// Package handler processes incoming http requests
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
)

// ErrorResponse standardizes error response structure
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// MapDomainErrorToHTTPCode translates domain errors to http status codes
func MapDomainErrorToHTTPCode(err error) (int, string, string) {
	if errors.Is(err, entity.ErrNotFound) {
		return http.StatusNotFound, "NOT_FOUND", err.Error()
	}
	if errors.Is(err, entity.ErrTeamExists) {
		return http.StatusConflict, "TEAM_EXISTS", err.Error()
	}
	if errors.Is(err, entity.ErrPRMerged) {
		return http.StatusConflict, "PR_MERGED", err.Error()
	}
	if errors.Is(err, entity.ErrNotAssigned) {
		return http.StatusConflict, "NOT_ASSIGNED", err.Error()
	}
	if errors.Is(err, entity.ErrNoCandidate) {
		return http.StatusNotFound, "NO_CANDIDATE", err.Error()
	}
	if errors.Is(err, entity.ErrPRExists) {
		return http.StatusConflict, "PR_EXISTS", err.Error()
	}
	return http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error"
}

// respondWithError sends a json formatted error response
func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := ErrorResponse{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{
			Code:    code,
			Message: message,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// respondWithJSON sends a successful json response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

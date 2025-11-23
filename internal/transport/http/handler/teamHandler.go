// package handler processes incoming http requests
package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/services"
)

type AddTeamRequest struct {
	TeamName string `json:"team_name"`
	Members  []struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	} `json:"members"`
}

type TeamHandler struct {
	teamService services.TeamService
}

// NewTeamHandler creates a new team handler instance
func NewTeamHandler(service services.TeamService) *TeamHandler {
	return &TeamHandler{teamService: service}
}

// AddTeam processes request to create a team and its members
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req AddTeamRequest

	// parse json body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid JSON body")
		return
	}

	// map request data to domain entities
	teamEntity := &entity.Team{Name: req.TeamName}
	var userEntities []*entity.User
	for _, m := range req.Members {
		userEntities = append(userEntities, &entity.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	// delegate creation to service
	err := h.teamService.CreateTeamWithUsers(r.Context(), teamEntity, userEntities)

	if err != nil {
		slog.Error("Failed to create team", "error", err)
		status, code, msg := MapDomainErrorToHTTPCode(err)
		respondWithError(w, status, code, msg)
		return
	}

	resp := struct {
		Team *entity.Team `json:"team"`
	}{
		Team: teamEntity,
	}
	respondWithJSON(w, http.StatusCreated, resp)
}

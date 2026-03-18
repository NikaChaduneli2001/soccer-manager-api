package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

// GetTeam returns the authenticated user's team (GET /api/v1/team).
func (c *Controller) GetTeam(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	team, err := c.Service.GetTeam(userID)
	if err != nil {
		if err == service.ErrTeamNotFound {
			response.Error(w, http.StatusNotFound, "team not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get team")
		return
	}
	response.JSON(w, http.StatusOK, team)
}

// UpdateTeam updates name and country (PUT /api/v1/team).
func (c *Controller) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req models.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.Service.UpdateTeam(userID, req.Name, req.Country); err != nil {
		if err == service.ErrTeamNotFound {
			response.Error(w, http.StatusNotFound, "team not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update team")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CreateTeam creates a team with 20 players for the authenticated user (POST /api/v1/team).
func (c *Controller) CreateTeam(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	team, err := c.Service.CreateTeam(userID)
	if err != nil {
		if err == service.ErrTeamAlreadyExists {
			response.Error(w, http.StatusConflict, "user already has a team")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create team")
		return
	}
	response.JSON(w, http.StatusCreated, team)
}

// GetTeamPlayers returns all players of the authenticated user's team (GET /api/v1/team/players).
func (c *Controller) GetTeamPlayers(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	players, err := c.Service.GetTeamPlayers(userID)
	if err != nil {
		if err == service.ErrTeamNotFound {
			response.Error(w, http.StatusNotFound, "team not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to get players")
		return
	}
	response.JSON(w, http.StatusOK, players)
}

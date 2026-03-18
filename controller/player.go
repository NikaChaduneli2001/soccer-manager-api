package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

// CreatePlayer adds a player to the authenticated user's team (POST /api/v1/players).
func (c *Controller) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req models.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	player, err := c.Service.CreatePlayer(userID, req.FirstName, req.LastName, req.Country, req.Position)
	if err != nil {
		switch err {
		case service.ErrTeamNotFound:
			response.Error(w, http.StatusNotFound, "team not found")
			return
		case service.ErrInvalidPosition:
			response.Error(w, http.StatusBadRequest, "invalid position: must be goalkeeper, defender, midfielder, or attacker")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create player")
		return
	}
	response.JSON(w, http.StatusCreated, player)
}

// UpdatePlayer updates first name, last name, country of a player (PUT /api/v1/players/:id).
func (c *Controller) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	playerID, err := parsePlayerID(r.URL.Path)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid player id")
		return
	}
	var req models.UpdatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.Service.UpdatePlayer(userID, playerID, req.FirstName, req.LastName, req.Country); err != nil {
		switch err {
		case service.ErrTeamNotFound, service.ErrPlayerNotFound:
			response.Error(w, http.StatusNotFound, "not found")
			return
		case service.ErrNotYourPlayer:
			response.Error(w, http.StatusForbidden, "player does not belong to your team")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update player")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// parsePlayerID extracts player id from path like /api/v1/players/123
func parsePlayerID(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 5 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(parts[4], 10, 64)
}

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
// @Summary      Get my team
// @Description  Returns the authenticated user's team. Total value is the sum of all player market values.
// @Tags         team
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  models.Team
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /team [get]
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
// @Summary      Update team
// @Description  Update team name and country (editable fields)
// @Tags         team
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.UpdateTeamRequest  true  "Team name and country"
// @Success      200   {object}  map[string]string
// @Failure      401   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Router       /team [put]
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

// CreateTeam creates a team with 20 players (POST /api/v1/team). user_id, name, country, budget come from request body.
// @Summary      Create team
// @Description  Create a team with 20 players (3 GK, 6 DEF, 6 MID, 5 ATT), $1M each. Send user_id, name, country, budget in body. Budget defaults to $5M if omitted or zero.
// @Tags         team
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateTeamRequest  true  "user_id, name, country, budget"
// @Success      201   {object}  models.Team
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401  {object}  models.ErrorResponse
// @Failure      409  {object}  models.ErrorResponse
// @Router       /team [post]
func (c *Controller) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID <= 0 {
		response.Error(w, http.StatusBadRequest, "user_id is required and must be positive")
		return
	}
	team, err := c.Service.CreateTeam(req.UserID, req.Name, req.Country, req.Budget)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.Error(w, http.StatusNotFound, "user not found")
			return
		case service.ErrTeamAlreadyExists:
			response.Error(w, http.StatusConflict, "user already has a team")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to create team")
		return
	}
	response.JSON(w, http.StatusCreated, team)
}

// GetTeamPlayers returns all players of the authenticated user's team (GET /api/v1/team/players).
// @Summary      Get team players
// @Description  Returns all players belonging to the authenticated user's team
// @Tags         team
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   models.Player
// @Failure      401  {object}  models.ErrorResponse
// @Failure      404  {object}  models.ErrorResponse
// @Router       /team/players [get]
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

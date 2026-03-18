package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

// ListPlayerOnTransfer puts a player on the transfer list (POST /api/v1/transfer/list).
// @Summary      List player on transfer market
// @Description  Put one of your players on the transfer list with an asking price. Player must belong to your team and not already be listed.
// @Tags         transfer
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.ListPlayerRequest  true  "Player ID and asking price"
// @Success      201   {object}  map[string]string
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      403   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Failure      409   {object}  models.ErrorResponse
// @Router       /transfer/list [post]
func (c *Controller) ListPlayerOnTransfer(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req models.ListPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.Service.ListPlayerOnTransfer(userID, req.PlayerID, req.AskingPrice); err != nil {
		switch err {
		case service.ErrTeamNotFound, service.ErrPlayerNotFound:
			response.Error(w, http.StatusNotFound, "not found")
			return
		case service.ErrNotYourPlayer:
			response.Error(w, http.StatusForbidden, "player does not belong to your team")
			return
		case service.ErrAlreadyListed:
			response.Error(w, http.StatusConflict, "player is already on transfer list")
			return
		}
		if err.Error() == "asking price must be positive" {
			response.Error(w, http.StatusBadRequest, "asking price must be positive")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to list player")
		return
	}
	response.JSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// GetTransferList returns all players on the market (GET /api/v1/transfer/list).
// @Summary      Get transfer market list
// @Description  Returns all players currently on the transfer list with asking prices
// @Tags         transfer
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   models.TransferMarketItem
// @Failure      401  {object}  models.ErrorResponse
// @Router       /transfer/list [get]
func (c *Controller) GetTransferList(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := c.Service.GetTransferList()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to get transfer list")
		return
	}
	response.JSON(w, http.StatusOK, items)
}

// BuyPlayer purchases a player from the transfer list (POST /api/v1/transfer/buy).
// @Summary      Buy player
// @Description  Purchase a player from the transfer list at the asking price. Budgets are updated; player's market value increases 10–100% on transfer.
// @Tags         transfer
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.BuyPlayerRequest  true  "Listing ID"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Router       /transfer/buy [post]
func (c *Controller) BuyPlayer(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req models.BuyPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := c.Service.BuyPlayer(userID, req.ListingID); err != nil {
		switch err {
		case service.ErrListingNotFound:
			response.Error(w, http.StatusNotFound, "listing not found")
			return
		case service.ErrTeamNotFound:
			response.Error(w, http.StatusNotFound, "team not found")
			return
		case service.ErrCannotBuyOwnPlayer:
			response.Error(w, http.StatusBadRequest, "cannot buy your own player")
			return
		case service.ErrInsufficientBudget:
			response.Error(w, http.StatusBadRequest, "insufficient budget")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to complete transfer")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

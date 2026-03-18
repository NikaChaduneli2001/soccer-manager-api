package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

// Signup creates an account, a team, and 20 players; returns user and JWT.
// @Summary      Register
// @Description  Create account with email/password. Automatically creates a team of 20 players (3 GK, 6 DEF, 6 MID, 5 ATT) and $5M budget.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.SignupRequest  true  "Signup payload"
// @Success      201   {object}  models.AuthResponse
// @Failure      400   {object}  models.ErrorResponse
// @Failure      409   {object}  models.ErrorResponse
// @Router       /signup [post]
func (c *Controller) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := c.Service.Signup(req.Email, req.Password, req.FullName, req.Age)
	if err != nil {
		switch err {
		case service.ErrEmailInvalid:
			response.Error(w, http.StatusBadRequest, "invalid email")
			return
		case service.ErrEmailExists:
			response.Error(w, http.StatusConflict, "email already registered")
			return
		case service.ErrPasswordShort:
			response.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
			return
		case service.ErrAgeInvalid:
			response.Error(w, http.StatusBadRequest, "age must be between 0 and 150")
			return
		}
		response.Error(w, http.StatusInternalServerError, "registration failed")
		return
	}
	response.JSON(w, http.StatusCreated, resp)
}

// Login validates credentials and returns a JWT.
// @Summary      Login
// @Description  Authenticate with email and password, receive JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.LoginRequest  true  "Login payload"
// @Success      200   {object}  models.AuthResponse
// @Failure      401   {object}  models.ErrorResponse
// @Router       /login [post]
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := c.Service.Login(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Error(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		response.Error(w, http.StatusInternalServerError, "login failed")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

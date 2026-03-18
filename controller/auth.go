package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

// Signup handles POST /signup: create account, team, and 20 players; return user + JWT.
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

// Login handles POST /login: validate credentials and return JWT.
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

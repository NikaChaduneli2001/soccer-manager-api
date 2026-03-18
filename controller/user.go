package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/models"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

// CreateUser creates a user account only (no team, no players). Team can be created later via POST /team.
// @Summary      Create user
// @Description  Create a user account with email, password, fullname, age. Does not create a team; use POST /team to create a team for this user.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      models.SignupRequest  true  "email, password, fullname, age"
// @Success      201   {object}  models.User
// @Failure      400   {object}  models.ErrorResponse
// @Failure      401   {object}  models.ErrorResponse
// @Failure      409   {object}  models.ErrorResponse
// @Router       /users [post]
func (c *Controller) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := c.Service.CreateUser(req.Email, req.Password, req.FullName, req.Age)
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
		response.Error(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	response.JSON(w, http.StatusCreated, user)
}

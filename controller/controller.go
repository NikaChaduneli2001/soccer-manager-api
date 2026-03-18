package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/pkg/auth"
	"github.com/nika/soccer-manager-api/pkg/response"
	"github.com/nika/soccer-manager-api/service"
)

type Controller struct {
	Service *service.Service
}

func NewController(svc *service.Service) *Controller {
	return &Controller{Service: svc}
}

// Health returns service health status.
// @Summary      Health check
// @Description  Returns API health status
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (c *Controller) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = w.Write([]byte(`{"error":"method not allowed"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Me returns the current authenticated user ID.
// @Summary      Current user
// @Description  Returns the authenticated user ID from the JWT
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  map[string]int64
// @Failure      401  {object}  models.ErrorResponse
// @Router       /me [get]
func (c *Controller) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	userID := auth.UserIDFromContext(r.Context())
	if userID == 0 {
		response.Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	response.JSON(w, http.StatusOK, map[string]int64{"user_id": userID})
}

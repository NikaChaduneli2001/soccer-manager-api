package controller

import (
	"encoding/json"
	"net/http"

	"github.com/nika/soccer-manager-api/service"
)

type Controller struct {
	Service *service.Service
}

func NewController(svc *service.Service) *Controller {
	return &Controller{Service: svc}
}

// Health returns service health status.
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

package handler

import (
	"net/http"

	"github.com/nika/soccer-manager-api/controller"
)

type Handler struct {
	Controller *controller.Controller
}

func NewHandler(ctrl *controller.Controller) *Handler {
	return &Handler{Controller: ctrl}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Controller.Health)
	return mux
}

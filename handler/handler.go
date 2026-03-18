package handler

import (
	"net/http"

	"github.com/nika/soccer-manager-api/controller"
	"github.com/nika/soccer-manager-api/middleware"
)

type Handler struct {
	Controller *controller.Controller
}

func NewHandler(ctrl *controller.Controller) *Handler {
	return &Handler{Controller: ctrl}
}

func (h *Handler) Router(jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/health", middleware.Chain(middleware.Method(http.MethodGet))(http.HandlerFunc(h.Controller.Health)))

	mux.Handle("/api/v1/signup", middleware.Chain(middleware.Method(http.MethodPost))(http.HandlerFunc(h.Controller.Signup)))

	mux.Handle("/api/v1/login", middleware.Chain(middleware.Method(http.MethodPost))(http.HandlerFunc(h.Controller.Login)))

	protected := middleware.Chain(middleware.JWT(jwtSecret), middleware.Method(http.MethodGet))(http.HandlerFunc(h.Controller.Me))
	mux.Handle("/api/v1/me", protected)
	return mux
}

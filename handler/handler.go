package handler

import (
	"net/http"

	"github.com/nika/soccer-manager-api/controller"
	"github.com/nika/soccer-manager-api/middleware"
	"github.com/nika/soccer-manager-api/pkg/response"
)

type Handler struct {
	Controller *controller.Controller
}

func NewHandler(ctrl *controller.Controller) *Handler {
	return &Handler{Controller: ctrl}
}

func (h *Handler) Router(jwtSecret string) http.Handler {
	mux := http.NewServeMux()
	jwt := middleware.JWT(jwtSecret)

	mux.Handle("/api/v1/health", middleware.Chain(middleware.Method(http.MethodGet))(http.HandlerFunc(h.Controller.Health)))
	mux.Handle("/api/v1/signup", middleware.Chain(middleware.Method(http.MethodPost))(http.HandlerFunc(h.Controller.Signup)))
	mux.Handle("/api/v1/login", middleware.Chain(middleware.Method(http.MethodPost))(http.HandlerFunc(h.Controller.Login)))
	mux.Handle("/api/v1/me", middleware.Chain(jwt, middleware.Method(http.MethodGet))(http.HandlerFunc(h.Controller.Me)))

	mux.Handle("/api/v1/team", middleware.Chain(jwt)(http.HandlerFunc(h.teamHandler)))
	mux.Handle("/api/v1/team/players", middleware.Chain(jwt, middleware.Method(http.MethodGet))(http.HandlerFunc(h.Controller.GetTeamPlayers)))

	mux.Handle("/api/v1/players", middleware.Chain(jwt, middleware.Method(http.MethodPost))(http.HandlerFunc(h.Controller.CreatePlayer)))
	mux.Handle("/api/v1/players/", middleware.Chain(jwt, middleware.Method(http.MethodPut))(http.HandlerFunc(h.Controller.UpdatePlayer)))

	mux.Handle("/api/v1/transfer/list", middleware.Chain(jwt)(http.HandlerFunc(h.transferListHandler)))
	mux.Handle("/api/v1/transfer/buy", middleware.Chain(jwt, middleware.Method(http.MethodPost))(http.HandlerFunc(h.Controller.BuyPlayer)))

	return mux
}

func (h *Handler) teamHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Controller.GetTeam(w, r)
	case http.MethodPost:
		h.Controller.CreateTeam(w, r)
	case http.MethodPut:
		h.Controller.UpdateTeam(w, r)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) transferListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Controller.GetTransferList(w, r)
	case http.MethodPost:
		h.Controller.ListPlayerOnTransfer(w, r)
	default:
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

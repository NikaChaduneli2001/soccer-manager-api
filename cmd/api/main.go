// Package main Soccer Manager API
// @title           Soccer Manager API
// @version         1.0
// @description     RESTful API for fantasy football manager. Create teams, manage players, and trade on the transfer market.
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @license.name    MIT
// @host            localhost:3000
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nika/soccer-manager-api/config"
	"github.com/nika/soccer-manager-api/controller"
	_ "github.com/nika/soccer-manager-api/docs"
	"github.com/nika/soccer-manager-api/handler"
	"github.com/nika/soccer-manager-api/pkg/migrate"
	"github.com/nika/soccer-manager-api/repository"
	"github.com/nika/soccer-manager-api/service"
)

func main() {
	cfg := config.Load()

	var repo *repository.DB
	dsn := cfg.Database.DSN()
	if db, err := repository.NewDB(dsn); err != nil {
		log.Printf("database not available (run with docker-compose for full stack): %v", err)
	} else {
		repo = db
		defer repo.Close()
		if err := migrate.Run(repo.DB); err != nil {
			log.Printf("migrations: %v", err)
		}
	}

	svc := service.NewService(repo, cfg.App.JWTSecret, cfg.App.JWTExpireHours)
	ctrl := controller.NewController(svc)
	h := handler.NewHandler(ctrl)
	mux := h.Router(cfg.App.JWTSecret)

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	log.Println("server stopped")
}

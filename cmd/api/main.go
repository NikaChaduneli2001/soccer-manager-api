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
	"github.com/nika/soccer-manager-api/handler"
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
	}

	svc := service.NewService(repo)
	ctrl := controller.NewController(svc)
	h := handler.NewHandler(ctrl)
	mux := h.Router()

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

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/config"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/services"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/infrastructure/db/postgres"
	repoImpl "github.com/hryak228pizza/pr-reviewer-assigner/internal/infrastructure/db/repository"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/transport/http/handler"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/transport/http/router"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	ctx := context.Background()

	cfg := config.Load()
	log.Info("Starting service", "env", cfg.Env, "address", cfg.HTTPServer.Address)

	pg, err := postgres.New(ctx, cfg.PG.URL)
	if err != nil {
		log.Error("Failed to init postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close()
	log.Info("Successfully connected to PostgreSQL")

	trm := postgres.NewTransactionManager(pg.Pool)

	teamRepo := repoImpl.NewTeamRepository(trm)
	userRepo := repoImpl.NewUserRepository(trm)
	prRepo := repoImpl.NewPRRepository(trm)

	assigner := services.NewAssigner()

	teamService := services.NewTeamUseCase(teamRepo, trm)
	userService := services.NewUserUseCase(userRepo, prRepo)
	prService := services.NewPRUseCase(prRepo, userRepo, trm, assigner)

	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService)
	prHandler := handler.NewPRHandler(prService)

	r := router.NewRouter(teamHandler, userHandler, prHandler)

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	log.Info("HTTP server starting", "addr", cfg.HTTPServer.Address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("HTTP server failed to start", "error", err)
		return
	}
}

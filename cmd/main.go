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
	// init structured logger
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	ctx := context.Background()

	// load configuration from .env and environment variables
	cfg := config.Load()
	log.Info("Starting service", "env", cfg.Env, "address", cfg.HTTPServer.Address)

	// init postgres connection pool with retry logic
	pg, err := postgres.New(ctx, cfg.PG.URL)
	if err != nil {
		log.Error("Failed to init postgres", "error", err)
		os.Exit(1)
	}
	defer pg.Close() // close pool when main exits
	log.Info("Successfully connected to PostgreSQL")

	// init transaction manager (implements the Transactor interface)
	trm := postgres.NewTransactionManager(pg.Pool)

	// init concrete repository implementations (DAL)
	teamRepo := repoImpl.NewTeamRepository(trm)
	userRepo := repoImpl.NewUserRepository(trm)
	prRepo := repoImpl.NewPRRepository(trm)

	// init domain services and use cases (business logic)
	assigner := services.NewAssigner()

	// use cases are injected with required repositories and the transactor
	teamService := services.NewTeamUseCase(teamRepo, trm)
	// userService doesn't require trm if SetIsActive is not transactional
	userService := services.NewUserUseCase(userRepo, prRepo)
	prService := services.NewPRUseCase(prRepo, userRepo, trm, assigner)

	// init http handlers (transport layer)
	teamHandler := handler.NewTeamHandler(teamService)
	userHandler := handler.NewUserHandler(userService)
	prHandler := handler.NewPRHandler(prService)

	// init chi router with handlers and middleware
	r := router.NewRouter(teamHandler, userHandler, prHandler)

	// configure http server
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// start http server and block
	log.Info("HTTP server starting", "addr", cfg.HTTPServer.Address)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("HTTP server failed to start", "error", err)
		return
	}
}

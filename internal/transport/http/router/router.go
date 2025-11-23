// Package router defines api routes and middleware
package router

import (
	"net/http"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/transport/http/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter initializes and configures the http router
func NewRouter(teamHandler *handler.TeamHandler, userHandler *handler.UserHandler, prHandler *handler.PRHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Route("/team", func(r chi.Router) {
		r.Post("/add", teamHandler.AddTeam)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/setIsActive", userHandler.SetIsActive)
		r.Get("/getReview", userHandler.GetReviews)
	})

	r.Route("/pullRequest", func(r chi.Router) {
		r.Post("/create", prHandler.CreatePR)
		r.Post("/merge", prHandler.MergePR)
		r.Post("/reassign", prHandler.ReassignReviewer)
	})

	return r
}

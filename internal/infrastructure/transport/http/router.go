package http

import (
	"log/slog"

	"github.com/virogg/pr-assignment-service/internal/application/services/pr_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/stats_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/team_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/user_service"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	log *slog.Logger,
	prService *pr_service.PRService,
	statsService *stats_service.StatsService,
	teamService *team_service.TeamService,
	userService *user_service.UserService,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", handlers.Healthcheck)
	r.Get("/ping", handlers.Ping)

	if teamService != nil {
		teamHandler := handlers.NewTeamHandler(teamService, log)
		r.Post("/team/add", teamHandler.CreateTeam)
		r.Get("/team/get", teamHandler.GetTeam)
	}

	if userService != nil {
		userHandler := handlers.NewUserHandler(userService, log)
		r.Post("/users/setIsActive", userHandler.SetUserActive)
		r.Get("/users/getReview", userHandler.GetUserReviews)
		r.Post("/teams/deactivate", userHandler.DeactivateUsers)
	}

	if prService != nil {
		prHandler := handlers.NewPRHandler(prService, log)
		r.Post("/pullRequest/create", prHandler.CreatePR)
		r.Post("/pullRequest/merge", prHandler.MergePR)
		r.Post("/pullRequest/reassign", prHandler.ReassignPR)
	}

	if statsService != nil {
		statsHandler := handlers.NewStatsHandler(statsService, log)
		r.Get("/stats", statsHandler.GetStatistics)
	}

	return r
}

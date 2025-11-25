package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	pkg "github.com/virogg/pr-assignment-service/pkg/http"
)

//go:generate mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_stats_service.go -package=mocks ./internal/infrastructure/transport/http/handlers statsService
type statsService interface {
	GetUserStatistics(ctx context.Context) ([]entities.UserStats, error)
	GetPRStatistics(ctx context.Context) (*entities.PRStats, error)
	GetTeamStatistics(ctx context.Context) ([]entities.TeamStats, error)
}

type StatsHandler struct {
	statsService statsService
	log          *slog.Logger
}

func NewStatsHandler(statsService statsService, log *slog.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		log:          log,
	}
}

func (h *StatsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	userStats, err := h.statsService.GetUserStatistics(r.Context())
	if err != nil {
		h.log.Error("failed to get user statistics", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	prStats, err := h.statsService.GetPRStatistics(r.Context())
	if err != nil {
		h.log.Error("failed to get PR statistics", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	teamStats, err := h.statsService.GetTeamStatistics(r.Context())
	if err != nil {
		h.log.Error("failed to get team statistics", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"users": userStats,
		"prs":   prStats,
		"teams": teamStats,
	})
}

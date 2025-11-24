package stats_service

import (
	"context"
	"log/slog"

	"github.com/virogg/pr-assignment-service/internal/domain/entities"
)

type statsRepo interface {
	GetUserStatistics(ctx context.Context) ([]entities.UserStats, error)
	GetPRStatistics(ctx context.Context) (*entities.PRStats, error)
	GetTeamStatistics(ctx context.Context) ([]entities.TeamStats, error)
}

type StatsService struct {
	statsRepo statsRepo
	log       *slog.Logger
}

func New(statsRepo statsRepo, log *slog.Logger) *StatsService {
	return &StatsService{
		statsRepo: statsRepo,
		log:       log,
	}
}

func (s *StatsService) GetUserStatistics(ctx context.Context) ([]entities.UserStats, error) {
	stats, err := s.statsRepo.GetUserStatistics(ctx)
	if err != nil {
		s.log.Error("failed to get user statistics", slog.Any("error", err))
		return nil, err
	}

	return stats, nil
}

func (s *StatsService) GetPRStatistics(ctx context.Context) (*entities.PRStats, error) {
	stats, err := s.statsRepo.GetPRStatistics(ctx)
	if err != nil {
		s.log.Error("failed to get PR statistics", slog.Any("error", err))
		return nil, err
	}

	return stats, nil
}

func (s *StatsService) GetTeamStatistics(ctx context.Context) ([]entities.TeamStats, error) {
	stats, err := s.statsRepo.GetTeamStatistics(ctx)
	if err != nil {
		s.log.Error("failed to get team statistics", slog.Any("error", err))
		return nil, err
	}

	return stats, nil
}

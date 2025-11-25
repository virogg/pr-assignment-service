package stats_service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/virogg/pr-assignment-service/internal/application/services/stats_service/mocks"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStatsService_GetUserStatistics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatsRepo := mocks.NewMockstatsRepo(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockStatsRepo, log)

	ctx := context.Background()
	stats := []entities.UserStats{
		{ID: "user-1", Username: "john", TotalAssignments: 5, OpenAssignments: 2, CompletedAssignments: 3},
		{ID: "user-2", Username: "jane", TotalAssignments: 7, OpenAssignments: 3, CompletedAssignments: 4},
	}

	mockStatsRepo.EXPECT().
		GetUserStatistics(ctx).
		Return(stats, nil)

	result, err := service.GetUserStatistics(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "user-1", result[0].ID)
}

func TestStatsService_GetUserStatistics_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatsRepo := mocks.NewMockstatsRepo(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockStatsRepo, log)

	ctx := context.Background()

	mockStatsRepo.EXPECT().
		GetUserStatistics(ctx).
		Return(nil, errors.New("db error"))

	_, err := service.GetUserStatistics(ctx)
	assert.Error(t, err)
}

func TestStatsService_GetPRStatistics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatsRepo := mocks.NewMockstatsRepo(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockStatsRepo, log)

	ctx := context.Background()
	stats := &entities.PRStats{
		TotalPRs:          15,
		OpenPRs:           10,
		MergedPRs:         5,
		AvgReviewersPerPR: 1.8,
	}

	mockStatsRepo.EXPECT().
		GetPRStatistics(ctx).
		Return(stats, nil)

	result, err := service.GetPRStatistics(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 10, result.OpenPRs)
	assert.Equal(t, 5, result.MergedPRs)
}

func TestStatsService_GetPRStatistics_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatsRepo := mocks.NewMockstatsRepo(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockStatsRepo, log)

	ctx := context.Background()

	mockStatsRepo.EXPECT().
		GetPRStatistics(ctx).
		Return(nil, errors.New("db error"))

	_, err := service.GetPRStatistics(ctx)
	assert.Error(t, err)
}

func TestStatsService_GetTeamStatistics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatsRepo := mocks.NewMockstatsRepo(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockStatsRepo, log)

	ctx := context.Background()
	stats := []entities.TeamStats{
		{TeamName: "backend", TotalMembers: 5, ActiveMembers: 3, TotalPRs: 10},
		{TeamName: "frontend", TotalMembers: 4, ActiveMembers: 4, TotalPRs: 8},
	}

	mockStatsRepo.EXPECT().
		GetTeamStatistics(ctx).
		Return(stats, nil)

	result, err := service.GetTeamStatistics(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "backend", result[0].TeamName)
}

func TestStatsService_GetTeamStatistics_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStatsRepo := mocks.NewMockstatsRepo(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockStatsRepo, log)

	ctx := context.Background()

	mockStatsRepo.EXPECT().
		GetTeamStatistics(ctx).
		Return(nil, errors.New("db error"))

	_, err := service.GetTeamStatistics(ctx)
	assert.Error(t, err)
}

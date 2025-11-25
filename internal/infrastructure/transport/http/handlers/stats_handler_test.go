package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http/handlers/mocks"

	"go.uber.org/mock/gomock"
)

func TestStatsHandler_GetStatistics_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockstatsService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewStatsHandler(mockService, log)

	userStats := []entities.UserStats{
		{
			ID:                   "user-1",
			Username:             "john",
			TotalAssignments:     5,
			OpenAssignments:      2,
			CompletedAssignments: 3,
		},
		{
			ID:                   "user-2",
			Username:             "jane",
			TotalAssignments:     3,
			OpenAssignments:      1,
			CompletedAssignments: 2,
		},
	}

	prStats := &entities.PRStats{
		TotalPRs:          10,
		OpenPRs:           4,
		MergedPRs:         6,
		AvgReviewersPerPR: 2.5,
	}

	teamStats := []entities.TeamStats{
		{
			TeamName:      "backend",
			TotalMembers:  5,
			ActiveMembers: 4,
			TotalPRs:      8,
		},
		{
			TeamName:      "frontend",
			TotalMembers:  3,
			ActiveMembers: 3,
			TotalPRs:      2,
		},
	}

	mockService.EXPECT().
		GetUserStatistics(gomock.Any()).
		Return(userStats, nil)

	mockService.EXPECT().
		GetPRStatistics(gomock.Any()).
		Return(prStats, nil)

	mockService.EXPECT().
		GetTeamStatistics(gomock.Any()).
		Return(teamStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetStatistics(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	_, ok := resp["users"]
	assert.True(t, ok)

	_, ok = resp["prs"]
	assert.True(t, ok)

	_, ok = resp["teams"]
	assert.True(t, ok)

	users, ok := resp["users"].([]interface{})
	require.True(t, ok)
	assert.Len(t, users, 2)

	teams, ok := resp["teams"].([]interface{})
	require.True(t, ok)
	assert.Len(t, teams, 2)
}

func TestStatsHandler_GetStatistics_UserStatsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockstatsService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewStatsHandler(mockService, log)

	mockService.EXPECT().
		GetUserStatistics(gomock.Any()).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetStatistics(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatsHandler_GetStatistics_PRStatsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockstatsService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewStatsHandler(mockService, log)

	var userStats []entities.UserStats

	mockService.EXPECT().
		GetUserStatistics(gomock.Any()).
		Return(userStats, nil)

	mockService.EXPECT().
		GetPRStatistics(gomock.Any()).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetStatistics(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatsHandler_GetStatistics_TeamStatsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockstatsService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewStatsHandler(mockService, log)

	var userStats []entities.UserStats
	prStats := entities.PRStats{}

	mockService.EXPECT().
		GetUserStatistics(gomock.Any()).
		Return(userStats, nil)

	mockService.EXPECT().
		GetPRStatistics(gomock.Any()).
		Return(&prStats, nil)

	mockService.EXPECT().
		GetTeamStatistics(gomock.Any()).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetStatistics(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStatsHandler_GetStatistics_EmptyStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockstatsService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewStatsHandler(mockService, log)

	var userStats []entities.UserStats
	var teamStats []entities.TeamStats
	prStats := entities.PRStats{}

	mockService.EXPECT().
		GetUserStatistics(gomock.Any()).
		Return(userStats, nil)

	mockService.EXPECT().
		GetPRStatistics(gomock.Any()).
		Return(&prStats, nil)

	mockService.EXPECT().
		GetTeamStatistics(gomock.Any()).
		Return(teamStats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetStatistics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	_, ok := resp["users"]
	assert.True(t, ok)

	_, ok = resp["prs"]
	assert.True(t, ok)

	_, ok = resp["teams"]
	assert.True(t, ok)
}

package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http/handlers/mocks"

	"go.uber.org/mock/gomock"
)

func TestTeamHandler_CreateTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	users := []*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
		entities.NewUser("user-2", "jane", "backend", true),
	}
	team := entities.NewTeam(1, "backend", users)

	mockService.EXPECT().
		CreateTeam(gomock.Any(), gomock.Any()).
		Return(team, nil)

	reqBody := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "user-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "user-2", Username: "jane", TeamName: "backend", IsActive: true},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/team", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTeam(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp dto.TeamResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "backend", resp.Team.TeamName)
	assert.Len(t, resp.Team.Members, 2)
}

func TestTeamHandler_CreateTeam_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	req := httptest.NewRequest(http.MethodPost, "/api/team", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTeam(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_CreateTeam_MissingTeamName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	reqBody := dto.CreateTeamRequest{
		TeamName: "",
		Members:  []dto.UserDTO{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/team", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTeam(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_CreateTeam_TeamAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	mockService.EXPECT().
		CreateTeam(gomock.Any(), gomock.Any()).
		Return(nil, domainerr.ErrTeamExists)

	reqBody := dto.CreateTeamRequest{
		TeamName: "backend",
		Members:  []dto.UserDTO{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/team", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTeam(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_CreateTeam_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	mockService.EXPECT().
		CreateTeam(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("internal error"))

	reqBody := dto.CreateTeamRequest{
		TeamName: "backend",
		Members:  []dto.UserDTO{},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/team", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateTeam(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTeamHandler_GetTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	users := []*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
		entities.NewUser("user-2", "jane", "backend", true),
	}
	team := entities.NewTeam(1, "backend", users)

	mockService.EXPECT().
		GetTeam(gomock.Any(), "backend").
		Return(team, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/team?team_name=backend", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetTeam(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp dto.TeamDTO
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "backend", resp.TeamName)
	assert.Len(t, resp.Members, 2)
}

func TestTeamHandler_GetTeam_MissingTeamName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	req := httptest.NewRequest(http.MethodGet, "/api/team", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetTeam(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_GetTeam_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	mockService.EXPECT().
		GetTeam(gomock.Any(), "non-existent").
		Return(nil, domainerr.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/team?team_name=non-existent", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetTeam(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestTeamHandler_GetTeam_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockteamService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewTeamHandler(mockService, log)

	mockService.EXPECT().
		GetTeam(gomock.Any(), "backend").
		Return(nil, errors.New("internal error"))

	req := httptest.NewRequest(http.MethodGet, "/api/team?team_name=backend", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetTeam(w, req)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

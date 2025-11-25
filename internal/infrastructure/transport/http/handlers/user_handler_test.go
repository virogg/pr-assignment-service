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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http/handlers/mocks"

	"go.uber.org/mock/gomock"
)

func TestUserHandler_SetUserActive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	user := entities.NewUser("user-1", "john", "backend", false)

	mockService.EXPECT().
		SetUserActive(gomock.Any(), "user-1", false).
		Return(user, nil)

	reqBody := dto.SetUserActiveRequest{
		UserID:   "user-1",
		IsActive: false,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/active", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetUserActive(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp dto.UserResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "user-1", resp.User.ID)
}

func TestUserHandler_SetUserActive_InvalidRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/active", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetUserActive(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_SetUserActive_MissingUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	reqBody := dto.SetUserActiveRequest{
		UserID:   "",
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/active", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetUserActive(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_SetUserActive_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	mockService.EXPECT().
		SetUserActive(gomock.Any(), "non-existent", true).
		Return(nil, domainerr.ErrUserNotFound)

	reqBody := dto.SetUserActiveRequest{
		UserID:   "non-existent",
		IsActive: true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/active", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.SetUserActive(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_GetUserReviews_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	prs := []*entities.PullRequest{
		entities.NewPullRequest("pr-1", "Test PR 1", "author-1", vo.PRStatusOpen, time.Now(), "user-1"),
		entities.NewPullRequest("pr-2", "Test PR 2", "author-2", vo.PRStatusOpen, time.Now(), "user-1"),
	}

	mockService.EXPECT().
		GetUserReviews(gomock.Any(), "user-1").
		Return(prs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/user/reviews?user_id=user-1", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetUserReviews(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp dto.UserReviewsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "user-1", resp.UserID)
	assert.Len(t, resp.PullRequests, 2)
}

func TestUserHandler_GetUserReviews_MissingUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	req := httptest.NewRequest(http.MethodGet, "/api/user/reviews", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetUserReviews(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetUserReviews_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	mockService.EXPECT().
		GetUserReviews(gomock.Any(), "non-existent").
		Return(nil, domainerr.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/user/reviews?user_id=non-existent", http.NoBody)
	w := httptest.NewRecorder()

	handler.GetUserReviews(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_DeactivateUsers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	mockService.EXPECT().
		DeactivateTeamUsers(gomock.Any(), "backend").
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/deactivate?team_name=backend", http.NoBody)
	w := httptest.NewRecorder()

	handler.DeactivateUsers(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	_, ok := resp["message"]
	assert.True(t, ok)
}

func TestUserHandler_DeactivateUsers_MissingTeamName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/deactivate", http.NoBody)
	w := httptest.NewRecorder()

	handler.DeactivateUsers(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_DeactivateUsers_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	mockService.EXPECT().
		DeactivateTeamUsers(gomock.Any(), "non-existent").
		Return(domainerr.ErrTeamNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/user/deactivate?team_name=non-existent", http.NoBody)
	w := httptest.NewRecorder()

	handler.DeactivateUsers(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_DeactivateUsers_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockuserService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewUserHandler(mockService, log)

	mockService.EXPECT().
		DeactivateTeamUsers(gomock.Any(), "backend").
		Return(errors.New("internal error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/user/deactivate?team_name=backend", http.NoBody)
	w := httptest.NewRecorder()

	handler.DeactivateUsers(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

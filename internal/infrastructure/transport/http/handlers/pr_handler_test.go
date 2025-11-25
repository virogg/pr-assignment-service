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

	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/transport/http/handlers/mocks"

	"go.uber.org/mock/gomock"
)

func TestPRHandler_CreatePR_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockprService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewPRHandler(mockService, log)

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")

	mockService.EXPECT().
		CreatePR(gomock.Any(), "pr-1", "Test PR", "author-1").
		Return(pr, nil)

	reqBody := dto.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "author-1",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePR(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp dto.PRResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Equal(t, "pr-1", resp.PR.ID)
}

func TestPRHandler_CreatePR_Failure_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	handler := NewPRHandler(mockService, log)

	body := []byte("invalid json")

	req := httptest.NewRequest(http.MethodPost, "/api/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePR(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPRHandler_CreatePR_Failure_MissingRequiredField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	handler := NewPRHandler(mockService, log)

	reqBody := dto.CreatePRRequest{
		PullRequestID: "pr-1",
		AuthorID:      "",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePR(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPRHandler_CreatePR_Failure_AuthorNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	mockService.EXPECT().
		CreatePR(gomock.Any(), "pr-1", "Test PR", "author-1").
		Return(nil, domainerr.ErrAuthorNotFound)

	handler := NewPRHandler(mockService, log)

	reqBody := dto.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "author-1",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreatePR(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPRHandler_MergePR_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockprService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewPRHandler(mockService, log)

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusMerged, time.Now(), "rev-1")

	mockService.EXPECT().
		MergePR(gomock.Any(), "pr-1").
		Return(pr, nil)

	reqBody := dto.MergePRRequest{PullRequestID: "pr-1"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MergePR(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp dto.PRResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Equal(t, vo.PRStatusMerged.String(), resp.PR.Status)
}

func TestPRHandler_MergePR_Failure_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	handler := NewPRHandler(mockService, log)

	body := []byte("invalid json")

	req := httptest.NewRequest(http.MethodPost, "/api/pr/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MergePR(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPRHandler_MergePR_Failure_MissingRequiredField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	handler := NewPRHandler(mockService, log)

	reqBody := dto.MergePRRequest{
		PullRequestID: "",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MergePR(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPRHandler_MergePR_Failure_PRNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	mockService.EXPECT().
		MergePR(gomock.Any(), "pr-999").
		Return(nil, domainerr.ErrPRNotFound)

	handler := NewPRHandler(mockService, log)

	reqBody := dto.MergePRRequest{
		PullRequestID: "pr-999",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MergePR(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPRHandler_ReassignPR_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockprService(ctrl)
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handler := NewPRHandler(mockService, log)

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "new-rev")

	mockService.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "old-rev").
		Return(pr, "new-rev", nil)

	reqBody := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-rev",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPR(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp dto.ReassignPRResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Equal(t, "new-rev", resp.ReplacedBy)
}

func TestPRHandler_ReassignPR_Failure_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	handler := NewPRHandler(mockService, log)

	body := []byte("invalid json")
	req := httptest.NewRequest(http.MethodPost, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPR(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPRHandler_ReassignPR_Failure_MissingRequiredField(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	handler := NewPRHandler(mockService, log)

	reqBody := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     "",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPR(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPRHandler_ReassignPR_Failure_PRAlreadyMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	mockService.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "old-rev").
		Return(nil, "", domainerr.ErrPRMerged)

	handler := NewPRHandler(mockService, log)

	reqBody := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-rev",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPR(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestPRHandler_ReassignPR_Failure_ReviewerNotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	mockService.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "old-rev").
		Return(nil, "", domainerr.ErrReviewerNotAssigned)

	handler := NewPRHandler(mockService, log)

	reqBody := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-rev",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPR(w, req)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestPRHandler_ReassignPR_Failure_GenericError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockService := mocks.NewMockprService(ctrl)
	mockService.EXPECT().
		ReassignReviewer(gomock.Any(), "pr-1", "old-rev").
		Return(nil, "", errors.New("database error"))

	handler := NewPRHandler(mockService, log)

	reqBody := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     "old-rev",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ReassignPR(w, req)
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

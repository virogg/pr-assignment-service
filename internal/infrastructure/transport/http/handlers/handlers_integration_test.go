//go:build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/application/services/pr_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/stats_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/team_service"
	"github.com/virogg/pr-assignment-service/internal/application/services/user_service"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres/test_helpers"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	trm "github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRHandler_CreateAndMergePR(t *testing.T) {
	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	c := trmpgx.DefaultCtxGetter
	trManager := trm.Must(trmpgx.NewDefaultFactory(testDB.Pool))

	teamRepo := postgres.NewTeamPostgresRepository(testDB.Pool, c)
	userRepo := postgres.NewUserPostgresRepository(testDB.Pool, c)
	prRepo := postgres.NewPRPostgresRepository(testDB.Pool, c)

	prSvc := pr_service.New(userRepo, teamRepo, prRepo, trManager, log)
	prHandler := NewPRHandler(prSvc, log)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	author := entities.NewUser("author-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	reviewer := entities.NewUser("reviewer-1", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, reviewer, team.ID))

	createReq := dto.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "author-1",
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/api/pr", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	prHandler.CreatePR(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp dto.PRResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&createResp))
	assert.Equal(t, "pr-1", createResp.PR.ID)
	assert.Equal(t, vo.PRStatusOpen.String(), createResp.PR.Status)

	mergeReq := dto.MergePRRequest{PullRequestID: "pr-1"}
	body, _ = json.Marshal(mergeReq)

	req = httptest.NewRequest(http.MethodPatch, "/api/pr/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	prHandler.MergePR(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var mergeResp dto.PRResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&mergeResp))

	assert.Equal(t, vo.PRStatusMerged.String(), mergeResp.PR.Status)

	pr, err := prRepo.GetByID(ctx, "pr-1")
	require.NoError(t, err)
	assert.Equal(t, vo.PRStatusMerged, pr.Status)
}

func TestPRHandler_ReassignReviewer(t *testing.T) {
	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	c := trmpgx.DefaultCtxGetter
	trManager := trm.Must(trmpgx.NewDefaultFactory(testDB.Pool))

	teamRepo := postgres.NewTeamPostgresRepository(testDB.Pool, c)
	userRepo := postgres.NewUserPostgresRepository(testDB.Pool, c)
	prRepo := postgres.NewPRPostgresRepository(testDB.Pool, c)

	prSvc := pr_service.New(userRepo, teamRepo, prRepo, trManager, log)
	prHandler := NewPRHandler(prSvc, log)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	author := entities.NewUser("author-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	reviewer1 := entities.NewUser("reviewer-1", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, reviewer1, team.ID))

	reviewer2 := entities.NewUser("reviewer-2", "bob", "backend", true)
	require.NoError(t, userRepo.Create(ctx, reviewer2, team.ID))

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	require.NoError(t, prRepo.Create(ctx, pr))

	reassignReq := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     "reviewer-1",
	}
	body, _ := json.Marshal(reassignReq)

	req := httptest.NewRequest(http.MethodPatch, "/api/pr/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	prHandler.ReassignPR(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	prDB, err := prRepo.GetByID(ctx, "pr-1")
	require.NoError(t, err)
	assert.Len(t, prDB.ReviewerIDs, 1)
	assert.Equal(t, "reviewer-2", prDB.ReviewerIDs[0])
}

func TestTeamHandler_CreateAndGetTeam(t *testing.T) {
	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	c := trmpgx.DefaultCtxGetter
	trManager := trm.Must(trmpgx.NewDefaultFactory(testDB.Pool))

	teamRepo := postgres.NewTeamPostgresRepository(testDB.Pool, c)
	userRepo := postgres.NewUserPostgresRepository(testDB.Pool, c)

	teamSvc := team_service.New(userRepo, teamRepo, trManager, log)
	teamHandler := NewTeamHandler(teamSvc, log)

	users := []dto.UserDTO{
		{ID: "user-1", Username: "john", IsActive: true},
		{ID: "user-2", Username: "jane", IsActive: true},
	}

	createReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members:  users,
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest(http.MethodPost, "/api/team", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	teamHandler.CreateTeam(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/team?team_name=backend", nil)
	w = httptest.NewRecorder()

	teamHandler.GetTeam(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var getResp dto.TeamDTO
	require.NoError(t, json.NewDecoder(w.Body).Decode(&getResp))
	assert.Equal(t, "backend", getResp.TeamName)
	assert.Len(t, getResp.Members, 2)

	teamDB, err := teamRepo.GetWithMembers(ctx, "backend")
	require.NoError(t, err)
	assert.Len(t, teamDB.Users, 2)
}

func TestUserHandler_SetActiveAndGetReviews(t *testing.T) {
	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	c := trmpgx.DefaultCtxGetter
	trManager := trm.Must(trmpgx.NewDefaultFactory(testDB.Pool))

	teamRepo := postgres.NewTeamPostgresRepository(testDB.Pool, c)
	userRepo := postgres.NewUserPostgresRepository(testDB.Pool, c)
	prRepo := postgres.NewPRPostgresRepository(testDB.Pool, c)

	userSvc := user_service.New(userRepo, teamRepo, prRepo, trManager, log)
	userHandler := NewUserHandler(userSvc, log)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user, team.ID))

	author := entities.NewUser("author-1", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "user-1")
	require.NoError(t, prRepo.Create(ctx, pr))

	setActiveReq := dto.SetUserActiveRequest{
		UserID:   "user-1",
		IsActive: false,
	}
	body, _ := json.Marshal(setActiveReq)

	req := httptest.NewRequest(http.MethodPatch, "/api/user/active", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	userHandler.SetUserActive(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	userDB, err := userRepo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	require.False(t, userDB.IsActive)

	req = httptest.NewRequest(http.MethodGet, "/api/user/reviews?user_id=user-1", nil)
	w = httptest.NewRecorder()

	userHandler.GetUserReviews(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var reviewsResp dto.UserReviewsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&reviewsResp))
	assert.Len(t, reviewsResp.PullRequests, 1)
	assert.Equal(t, "pr-1", reviewsResp.PullRequests[0].PullRequestID)
}

func TestStatsHandler_GetStatistics(t *testing.T) {
	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	c := trmpgx.DefaultCtxGetter

	teamRepo := postgres.NewTeamPostgresRepository(testDB.Pool, c)
	userRepo := postgres.NewUserPostgresRepository(testDB.Pool, c)
	prRepo := postgres.NewPRPostgresRepository(testDB.Pool, c)
	statsRepo := postgres.NewStatsPostgresRepository(testDB.Pool, c)

	statsSvc := stats_service.New(statsRepo, log)
	statsHandler := NewStatsHandler(statsSvc, log)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	author := entities.NewUser("author-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	reviewer := entities.NewUser("reviewer-1", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, reviewer, team.ID))

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	require.NoError(t, prRepo.Create(ctx, pr))

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()

	statsHandler.GetStatistics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var statsResp map[string]interface{}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&statsResp))

	_, ok := statsResp["users"]
	assert.True(t, ok)

	_, ok = statsResp["prs"]
	assert.True(t, ok)

	_, ok = statsResp["teams"]
	assert.True(t, ok)
}

//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres/test_helpers"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
)

func TestPRRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	if err := teamRepo.Create(ctx, team); err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	reviewer := entities.NewUser("reviewer-1", "jane", "backend", true)

	require.NoError(t, userRepo.Create(ctx, author, team.ID))
	require.NoError(t, userRepo.Create(ctx, reviewer, team.ID))

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")

	err := repo.Create(ctx, pr)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, "pr-1")
	require.NoError(t, err)
	assert.Equal(t, "pr-1", retrieved.ID)
	assert.Equal(t, "Test PR", retrieved.Name)
	assert.Equal(t, "author-1", retrieved.AuthorID)
	assert.Equal(t, "author-1", retrieved.AuthorID)
	assert.Equal(t, vo.PRStatusOpen, retrieved.Status)
	assert.Equal(t, vo.PRStatusOpen, retrieved.Status)
	assert.Equal(t, vo.PRStatusOpen, retrieved.Status)
	assert.Len(t, retrieved.ReviewerIDs, 1)
}

func TestPRRepository_Create_Duplicate(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), nil...)
	require.NoError(t, repo.Create(ctx, pr))

	err := repo.Create(ctx, pr)
	assert.ErrorIs(t, err, domainerr.ErrPRExists)
}

func TestPRRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	_, err := repo.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, domainerr.ErrPRNotFound)
}

func TestPRRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), nil...)
	require.NoError(t, repo.Create(ctx, pr))

	require.NoError(t, repo.UpdateStatus(ctx, "pr-1", vo.PRStatusMerged))

	retrieved, err := repo.GetByID(ctx, "pr-1")
	require.NoError(t, err)
	assert.Equal(t, vo.PRStatusMerged, retrieved.Status)
	assert.NotNil(t, retrieved.MergedAt)
}

func TestPRRepository_AddReviewers(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	reviewer1 := entities.NewUser("reviewer-1", "jane", "backend", true)
	reviewer2 := entities.NewUser("reviewer-2", "bob", "backend", true)

	for _, user := range []*entities.User{author, reviewer1, reviewer2} {
		if err := userRepo.Create(ctx, user, team.ID); err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), nil...)
	require.NoError(t, repo.Create(ctx, pr))

	err := repo.AddReviewers(ctx, "pr-1", []string{"reviewer-1", "reviewer-2"})
	require.NoError(t, err)

	reviewers, err := repo.GetReviewers(ctx, "pr-1")
	require.NoError(t, err)
	assert.Len(t, reviewers, 2)
}

func TestPRRepository_RemoveReviewer(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	reviewer := entities.NewUser("reviewer-1", "jane", "backend", true)

	for _, user := range []*entities.User{author, reviewer} {
		require.NoError(t, userRepo.Create(ctx, user, team.ID))
	}

	pr := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	require.NoError(t, repo.Create(ctx, pr))

	err := repo.RemoveReviewer(ctx, "pr-1", "reviewer-1")
	require.NoError(t, err)

	reviewers, err := repo.GetReviewers(ctx, "pr-1")
	require.NoError(t, err)
	assert.Empty(t, reviewers)
}

func TestPRRepository_GetByReviewer(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	reviewer := entities.NewUser("reviewer-1", "jane", "backend", true)

	for _, user := range []*entities.User{author, reviewer} {
		require.NoError(t, userRepo.Create(ctx, user, team.ID))
	}

	pr1 := entities.NewPullRequest("pr-1", "Test PR 1", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	pr2 := entities.NewPullRequest("pr-2", "Test PR 2", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")

	require.NoError(t, repo.Create(ctx, pr1))
	require.NoError(t, repo.Create(ctx, pr2))

	prs, err := repo.GetByReviewer(ctx, "reviewer-1")
	require.NoError(t, err)
	assert.Len(t, prs, 2)
}

func TestPRRepository_GetAll(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	repo := NewPRPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	author := entities.NewUser("author-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, author, team.ID))

	pr1 := entities.NewPullRequest("pr-1", "Test PR 1", "author-1", vo.PRStatusOpen, time.Now(), nil...)
	pr2 := entities.NewPullRequest("pr-2", "Test PR 2", "author-1", vo.PRStatusMerged, time.Now(), nil...)

	require.NoError(t, repo.Create(ctx, pr1))
	require.NoError(t, repo.Create(ctx, pr2))
	require.NoError(t, repo.UpdateStatus(ctx, "pr-2", vo.PRStatusMerged))

	allPRs, err := repo.GetAll(ctx, nil, nil)
	require.NoError(t, err)
	assert.Len(t, allPRs, 2)

	openStatus := vo.PRStatusOpen
	openPRs, err := repo.GetAll(ctx, &openStatus, nil)
	require.NoError(t, err)
	assert.Len(t, openPRs, 1)

	authorID := "author-1"
	authorPRs, err := repo.GetAll(ctx, nil, &authorID)
	require.NoError(t, err)
	assert.Len(t, authorPRs, 2)
}

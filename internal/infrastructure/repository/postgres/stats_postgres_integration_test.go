//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres/test_helpers"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
)

func TestStatsRepository_GetUserStatistics(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	statsRepo := NewStatsPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	prRepo := NewPRPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	author := entities.NewUser("author-1", "john", "backend", true)
	reviewer1 := entities.NewUser("reviewer-1", "jane", "backend", true)
	reviewer2 := entities.NewUser("reviewer-2", "bob", "backend", true)

	for _, user := range []*entities.User{author, reviewer1, reviewer2} {
		require.NoError(t, userRepo.Create(ctx, user, team.ID))
	}

	pr1 := entities.NewPullRequest("pr-1", "Test PR 1", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	pr2 := entities.NewPullRequest("pr-2", "Test PR 2", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	pr3 := entities.NewPullRequest("pr-3", "Test PR 3", "author-1", vo.PRStatusMerged, time.Now(), []string{"reviewer-1", "reviewer-2"}...)

	for _, pr := range []*entities.PullRequest{pr1, pr2, pr3} {
		require.NoError(t, prRepo.Create(ctx, pr))
	}

	require.NoError(t, prRepo.UpdateStatus(ctx, "pr-3", vo.PRStatusMerged))

	stats, err := statsRepo.GetUserStatistics(ctx)
	require.NoError(t, err)
	assert.Len(t, stats, 3)

	var reviewer1Stats *entities.UserStats
	for i := range stats {
		if stats[i].ID == "reviewer-1" {
			reviewer1Stats = &stats[i]
			break
		}
	}

	require.NotNil(t, reviewer1Stats)
	assert.Equal(t, 3, reviewer1Stats.TotalAssignments)
	assert.Equal(t, 2, reviewer1Stats.OpenAssignments)
	assert.Equal(t, 1, reviewer1Stats.CompletedAssignments)

	var reviewer2Stats *entities.UserStats
	for i := range stats {
		if stats[i].ID == "reviewer-2" {
			reviewer2Stats = &stats[i]
			break
		}
	}

	require.NotNil(t, reviewer2Stats)
	assert.Equal(t, 1, reviewer2Stats.TotalAssignments)
	assert.Zero(t, reviewer2Stats.OpenAssignments)
	assert.Equal(t, 1, reviewer2Stats.CompletedAssignments)
}

func TestStatsRepository_GetUserStatistics_NoData(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	statsRepo := NewStatsPostgresRepository(testDB.Pool, getter)

	stats, err := statsRepo.GetUserStatistics(ctx)
	require.NoError(t, err)
	assert.Empty(t, stats)
}

func TestStatsRepository_GetPRStatistics(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	statsRepo := NewStatsPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	prRepo := NewPRPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	author := entities.NewUser("author-1", "john", "backend", true)
	reviewer := entities.NewUser("reviewer-1", "jane", "backend", true)

	for _, user := range []*entities.User{author, reviewer} {
		require.NoError(t, userRepo.Create(ctx, user, team.ID))
	}

	pr1 := entities.NewPullRequest("pr-1", "Test PR 1", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	pr2 := entities.NewPullRequest("pr-2", "Test PR 2", "author-1", vo.PRStatusOpen, time.Now(), "reviewer-1")
	pr3 := entities.NewPullRequest("pr-3", "Test PR 3", "author-1", vo.PRStatusMerged, time.Now(), []string{"reviewer-1", "author-1"}...)

	for _, pr := range []*entities.PullRequest{pr1, pr2, pr3} {
		require.NoError(t, prRepo.Create(ctx, pr))
	}

	require.NoError(t, prRepo.UpdateStatus(ctx, "pr-3", vo.PRStatusMerged))

	stats, err := statsRepo.GetPRStatistics(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, stats.TotalPRs)
	assert.Equal(t, 2, stats.OpenPRs)
	assert.Equal(t, 1, stats.MergedPRs)

	expectedAvg := 4.0 / 3.0
	assert.InDelta(t, expectedAvg, stats.AvgReviewersPerPR, 0.01)
}

func TestStatsRepository_GetPRStatistics_NoData(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	statsRepo := NewStatsPostgresRepository(testDB.Pool, getter)

	stats, err := statsRepo.GetPRStatistics(ctx)
	require.NoError(t, err)
	assert.Zero(t, stats.TotalPRs)
	assert.Zero(t, stats.OpenPRs)
	assert.Zero(t, stats.MergedPRs)
	assert.Zero(t, stats.AvgReviewersPerPR)
}

func TestStatsRepository_GetTeamStatistics(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	statsRepo := NewStatsPostgresRepository(testDB.Pool, getter)

	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	prRepo := NewPRPostgresRepository(testDB.Pool, getter)

	team1 := entities.NewTeam(0, "backend", nil)
	team2 := entities.NewTeam(0, "frontend", nil)

	require.NoError(t, teamRepo.Create(ctx, team1))
	require.NoError(t, teamRepo.Create(ctx, team2))

	//users
	user1 := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user1, team1.ID))

	user2 := entities.NewUser("user-2", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user2, team1.ID))

	user3 := entities.NewUser("user-3", "bob", "backend", false)
	require.NoError(t, userRepo.Create(ctx, user3, team1.ID))

	user4 := entities.NewUser("user-4", "alice", "frontend", true)
	require.NoError(t, userRepo.Create(ctx, user4, team2.ID))

	//prs
	pr1 := entities.NewPullRequest("pr-1", "Test PR 1", "user-1", vo.PRStatusOpen, time.Now(), nil...)
	require.NoError(t, prRepo.Create(ctx, pr1))

	pr2 := entities.NewPullRequest("pr-2", "Test PR 2", "user-1", vo.PRStatusOpen, time.Now(), nil...)
	require.NoError(t, prRepo.Create(ctx, pr2))

	pr3 := entities.NewPullRequest("pr-3", "Test PR 3", "user-4", vo.PRStatusOpen, time.Now(), nil...)
	require.NoError(t, prRepo.Create(ctx, pr3))

	stats, err := statsRepo.GetTeamStatistics(ctx)
	require.NoError(t, err)
	assert.Len(t, stats, 2)

	var backendStats *entities.TeamStats
	for i := range stats {
		if stats[i].TeamName == "backend" {
			backendStats = &stats[i]
			break
		}
	}

	require.NotNil(t, backendStats)
	assert.Equal(t, 3, backendStats.TotalMembers)
	assert.Equal(t, 2, backendStats.ActiveMembers)
	assert.Equal(t, 2, backendStats.TotalPRs)

	var frontendStats *entities.TeamStats
	for i := range stats {
		if stats[i].TeamName == "frontend" {
			frontendStats = &stats[i]
			break
		}
	}

	require.NotNil(t, frontendStats)
	assert.Equal(t, 1, frontendStats.TotalMembers)
	assert.Equal(t, 1, frontendStats.ActiveMembers)
	assert.Equal(t, 1, frontendStats.TotalPRs)
}

func TestStatsRepository_GetTeamStatistics_NoData(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	statsRepo := NewStatsPostgresRepository(testDB.Pool, getter)

	stats, err := statsRepo.GetTeamStatistics(ctx)
	require.NoError(t, err)
	assert.Empty(t, stats)
}

//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	"github.com/virogg/pr-assignment-service/internal/infrastructure/repository/postgres/test_helpers"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
)

func TestTeamRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	err := teamRepo.Create(ctx, team)
	require.NoError(t, err)
	assert.NotZero(t, team.ID)

	retrieved, err := teamRepo.GetByName(ctx, "backend")
	require.NoError(t, err)
	assert.Equal(t, "backend", retrieved.Name)
	assert.Equal(t, team.ID, retrieved.ID)
}

func TestTeamRepository_Create_Duplicate(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	team2 := entities.NewTeam(0, "backend", nil)
	err := teamRepo.Create(ctx, team2)
	assert.ErrorIs(t, err, domainerr.ErrTeamExists)
}

func TestTeamRepository_GetByName_NotFound(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	_, err := teamRepo.GetByName(ctx, "non-existent")
	assert.ErrorIs(t, err, domainerr.ErrTeamNotFound)
}

func TestTeamRepository_GetByID(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	retrieved, err := teamRepo.GetByID(ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, "backend", retrieved.Name)
	assert.Equal(t, team.ID, retrieved.ID)
}

func TestTeamRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	_, err := teamRepo.GetByID(ctx, 999)
	assert.ErrorIs(t, err, domainerr.ErrTeamNotFound)
}

func TestTeamRepository_GetWithMembers(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user1 := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user1, team.ID))

	user2 := entities.NewUser("user-2", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user2, team.ID))

	user3 := entities.NewUser("user-3", "bob", "backend", false)
	require.NoError(t, userRepo.Create(ctx, user3, team.ID))

	teamWithMembers, err := teamRepo.GetWithMembers(ctx, "backend")
	require.NoError(t, err)
	assert.Equal(t, "backend", teamWithMembers.Name)
	assert.Len(t, teamWithMembers.Users, 3)

	if len(teamWithMembers.Users) >= 3 {
		assert.Equal(t, "bob", teamWithMembers.Users[0].Username)
		assert.Equal(t, "jane", teamWithMembers.Users[1].Username)
		assert.Equal(t, "john", teamWithMembers.Users[2].Username)
	}
}

func TestTeamRepository_GetWithMembers_NoMembers(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	teamWithMembers, err := teamRepo.GetWithMembers(ctx, "backend")
	require.NoError(t, err)
	assert.Empty(t, teamWithMembers.Users)
}

func TestTeamRepository_TeamExists(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	exists, err := teamRepo.TeamExists(ctx, "backend")
	require.NoError(t, err)
	require.True(t, exists)

	exists, err = teamRepo.TeamExists(ctx, "frontend")
	require.NoError(t, err)
	require.False(t, exists)
}

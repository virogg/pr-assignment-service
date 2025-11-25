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

func TestUserRepository_Create(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user := entities.NewUser("user-1", "john", "backend", true)
	err := userRepo.Create(ctx, user, team.ID)
	require.NoError(t, err)

	retrieved, err := userRepo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "user-1", retrieved.ID)
	assert.Equal(t, "john", retrieved.Username)
	assert.Equal(t, "backend", retrieved.TeamName)
	assert.True(t, retrieved.IsActive)
}

func TestUserRepository_Create_UpsertBehavior(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user, team.ID))

	updatedUser := entities.NewUser("user-1", "john_updated", "backend", false)
	err := userRepo.Create(ctx, updatedUser, team.ID)
	require.NoError(t, err)

	retrieved, err := userRepo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "john_updated", retrieved.Username)
	assert.False(t, retrieved.IsActive)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)

	_, err := userRepo.GetByID(ctx, "non-existent")
	assert.ErrorIs(t, err, domainerr.ErrUserNotFound)
}

func TestUserRepository_GetByTeamID(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user1 := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user1, team.ID))

	user2 := entities.NewUser("user-2", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user2, team.ID))

	user3 := entities.NewUser("user-3", "bob", "backend", false)
	require.NoError(t, userRepo.Create(ctx, user3, team.ID))

	users, err := userRepo.GetByTeamID(ctx, "1")
	require.NoError(t, err)
	assert.Len(t, users, 3)

	if len(users) >= 3 {
		assert.Equal(t, "bob", users[0].Username)
		assert.Equal(t, "jane", users[1].Username)
		assert.Equal(t, "john", users[2].Username)
	}
}

func TestUserRepository_SetActive(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user, team.ID))

	err := userRepo.SetActive(ctx, "user-1", false)
	require.NoError(t, err)

	retrieved, err := userRepo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.False(t, retrieved.IsActive)

	err = userRepo.SetActive(ctx, "user-1", true)
	require.NoError(t, err)

	retrieved, err = userRepo.GetByID(ctx, "user-1")
	require.NoError(t, err)
	assert.True(t, retrieved.IsActive)
}

func TestUserRepository_SetActive_NotFound(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)

	err := userRepo.SetActive(ctx, "non-existent", false)
	assert.ErrorIs(t, err, domainerr.ErrUserNotFound)
}

func TestUserRepository_GetActiveUsersInTeam(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user1 := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user1, team.ID))

	user2 := entities.NewUser("user-2", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user2, team.ID))

	user3 := entities.NewUser("user-3", "bob", "backend", false)
	require.NoError(t, userRepo.Create(ctx, user3, team.ID))

	activeUsers, err := userRepo.GetActiveUsersInTeam(ctx, team.ID, nil)
	require.NoError(t, err)
	assert.Len(t, activeUsers, 2)

	activeUsersExcluding, err := userRepo.GetActiveUsersInTeam(ctx, team.ID, []string{"user-1"})
	require.NoError(t, err)
	assert.Len(t, activeUsersExcluding, 1)
	assert.Equal(t, "user-2", activeUsersExcluding[0].ID)
}

func TestUserRepository_SetUsersInactive(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)
	teamRepo := NewTeamPostgresRepository(testDB.Pool, getter)

	team := entities.NewTeam(0, "backend", nil)
	require.NoError(t, teamRepo.Create(ctx, team))

	user1 := entities.NewUser("user-1", "john", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user1, team.ID))

	user2 := entities.NewUser("user-2", "jane", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user2, team.ID))

	user3 := entities.NewUser("user-3", "bob", "backend", true)
	require.NoError(t, userRepo.Create(ctx, user3, team.ID))

	err := userRepo.SetUsersInactive(ctx, []string{"user-1", "user-2"})
	require.NoError(t, err)

	user1Retrieved, _ := userRepo.GetByID(ctx, "user-1")
	assert.False(t, user1Retrieved.IsActive)

	user2Retrieved, _ := userRepo.GetByID(ctx, "user-2")
	assert.False(t, user2Retrieved.IsActive)

	user3Retrieved, _ := userRepo.GetByID(ctx, "user-3")
	assert.True(t, user3Retrieved.IsActive)
}

func TestUserRepository_SetUsersInactive_EmptyList(t *testing.T) {
	t.Parallel()

	testDB := test_helpers.SetupTestDatabase(t)
	defer testDB.Cleanup(t)

	ctx := context.Background()
	getter := trmpgx.DefaultCtxGetter
	userRepo := NewUserPostgresRepository(testDB.Pool, getter)

	err := userRepo.SetUsersInactive(ctx, []string{})
	require.NoError(t, err)

	err = userRepo.SetUsersInactive(ctx, nil)
	require.NoError(t, err)
}

package team_service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/virogg/pr-assignment-service/internal/application/services/team_service/mocks"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newSimpleTrManager() *manager.Manager {
	// tx behavior is tested in integration tests
	return nil
}

func TestTeamService_CreateTeam_Failure_NilTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserCreator := mocks.NewMockuserCreator(ctrl)
	mockTeamRepo := mocks.NewMockteamRepo(ctrl)

	service := New(mockUserCreator, mockTeamRepo, trManager, log)
	_, err := service.CreateTeam(ctx, nil)
	assert.ErrorIs(t, err, domainerr.ErrInvalidInput)
}

func TestTeamService_CreateTeam_Failure_TeamAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserCreator := mocks.NewMockuserCreator(ctrl)
	mockTeamRepo := mocks.NewMockteamRepo(ctrl)

	mockTeamRepo.
		EXPECT().
		TeamExists(ctx, "backend").
		Return(true, nil)

	service := New(mockUserCreator, mockTeamRepo, trManager, log)
	team := entities.NewTeam(0, "backend", []*entities.User{})
	_, err := service.CreateTeam(ctx, team)
	assert.ErrorIs(t, err, domainerr.ErrTeamExists)
}

func TestTeamService_GetTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCreator := mocks.NewMockuserCreator(ctrl)
	mockTeamRepo := mocks.NewMockteamRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserCreator, mockTeamRepo, trManager, log)

	ctx := context.Background()
	team := entities.NewTeam(1, "backend", []*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
	})

	mockTeamRepo.EXPECT().
		GetWithMembers(ctx, "backend").
		Return(team, nil)

	result, err := service.GetTeam(ctx, "backend")
	assert.NoError(t, err)
	assert.Equal(t, "backend", result.Name)
	assert.Len(t, result.Users, 1)
}

func TestTeamService_GetTeam_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserCreator := mocks.NewMockuserCreator(ctrl)
	mockTeamRepo := mocks.NewMockteamRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserCreator, mockTeamRepo, trManager, log)

	ctx := context.Background()

	mockTeamRepo.EXPECT().
		GetWithMembers(ctx, "backend").
		Return(nil, domainerr.ErrTeamNotFound)

	_, err := service.GetTeam(ctx, "backend")
	assert.ErrorIs(t, err, domainerr.ErrTeamNotFound)
}

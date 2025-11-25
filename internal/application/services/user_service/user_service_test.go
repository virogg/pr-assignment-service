package user_service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/virogg/pr-assignment-service/internal/application/services/user_service/mocks"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newSimpleTrManager() *manager.Manager {
	// tx behavior is tested in integration tests
	return nil
}

func TestUserService_SetUserActive_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()
	user := entities.NewUser("user-1", "john", "backend", false)
	updatedUser := entities.NewUser("user-1", "john", "backend", true)

	mockUserGS.EXPECT().
		GetByID(ctx, "user-1").
		Return(user, nil)

	mockUserGS.EXPECT().
		SetActive(ctx, "user-1", true).
		Return(nil)

	mockUserGS.EXPECT().
		GetByID(ctx, "user-1").
		Return(updatedUser, nil)

	result, err := service.SetUserActive(ctx, "user-1", true)
	assert.NoError(t, err)
	assert.True(t, result.IsActive)
}

func TestUserService_SetUserActive_Failure_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	mockUserGS.
		EXPECT().
		GetByID(ctx, "user-1").
		Return(nil, domainerr.ErrUserNotFound)

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)
	_, err := service.SetUserActive(ctx, "user-1", true)
	assert.Error(t, err)
}

func TestUserService_SetUserActive_Failure_UpdateStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	user := entities.NewUser("user-1", "john", "backend", false)
	mockUserGS.
		EXPECT().
		GetByID(ctx, "user-1").
		Return(user, nil)

	mockUserGS.
		EXPECT().
		SetActive(ctx, "user-1", true).
		Return(errors.New("db error"))

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)
	_, err := service.SetUserActive(ctx, "user-1", true)
	assert.Error(t, err)
}

func TestUserService_GetUserReviews_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()
	user := entities.NewUser("user-1", "john", "backend", true)
	prs := []*entities.PullRequest{
		entities.NewPullRequest("pr-1", "Test PR 1", "author-1", vo.PRStatusOpen, time.Now(), "user-1"),
		entities.NewPullRequest("pr-2", "Test PR 2", "author-2", vo.PRStatusOpen, time.Now(), "user-1"),
	}

	mockUserGS.EXPECT().
		GetByID(ctx, "user-1").
		Return(user, nil)

	mockPRRepo.EXPECT().
		GetByReviewer(ctx, "user-1").
		Return(prs, nil)

	result, err := service.GetUserReviews(ctx, "user-1")
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUserService_GetUserReviews_Failure_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	mockUserGS.
		EXPECT().
		GetByID(ctx, "user-1").
		Return(nil, domainerr.ErrUserNotFound)

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)
	_, err := service.GetUserReviews(ctx, "user-1")
	assert.Error(t, err)
}

func TestUserService_GetUserReviews_Failure_GetPRs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	user := entities.NewUser("user-1", "john", "backend", true)
	mockUserGS.
		EXPECT().
		GetByID(ctx, "user-1").
		Return(user, nil)

	mockPRRepo.
		EXPECT().
		GetByReviewer(ctx, "user-1").
		Return(nil, errors.New("db error"))

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)
	_, err := service.GetUserReviews(ctx, "user-1")
	assert.Error(t, err)
}

func TestUserService_DeactivateTeamUsers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()
	team := entities.NewTeam(1, "backend", []*entities.User{})

	mockTeamGetter.EXPECT().
		GetByName(ctx, "backend").
		Return(team, nil)

	mockUserGS.EXPECT().
		GetActiveUsersInTeam(ctx, int64(1), []string{}).
		Return([]*entities.User{}, nil)

	err := service.DeactivateTeamUsers(ctx, "backend")
	assert.NoError(t, err)
}

func TestUserService_DeactivateTeamUsers_Failure_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	mockTeamGetter.
		EXPECT().
		GetByName(ctx, "backend").
		Return(nil, domainerr.ErrTeamNotFound)

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)
	err := service.DeactivateTeamUsers(ctx, "backend")
	assert.Error(t, err)
}

func TestUserService_DeactivateTeamUsers_Failure_GetActiveUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGS := mocks.NewMockuserGetterSetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	team := entities.NewTeam(1, "backend", []*entities.User{})
	mockTeamGetter.
		EXPECT().
		GetByName(ctx, "backend").
		Return(team, nil)

	mockUserGS.
		EXPECT().
		GetActiveUsersInTeam(ctx, int64(1), []string{}).
		Return(nil, errors.New("db error"))

	service := New(mockUserGS, mockTeamGetter, mockPRRepo, trManager, log)
	err := service.DeactivateTeamUsers(ctx, "backend")
	assert.Error(t, err)
}

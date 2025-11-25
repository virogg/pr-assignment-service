package pr_service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/virogg/pr-assignment-service/internal/application/services/pr_service/mocks"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"go.uber.org/mock/gomock"
)

func newSimpleTrManager() *manager.Manager {
	// tx behavior is tested in integration tests
	return nil
}

func TestPRService_GetPR_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()
	expectedPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")

	mockPRRepo.EXPECT().
		GetByID(ctx, "pr-1").
		Return(expectedPR, nil)

	pr, err := service.GetPR(ctx, "pr-1")

	assert.NoError(t, err)
	assert.Equal(t, "pr-1", pr.ID)
}

func TestPRService_GetPR_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()

	mockPRRepo.EXPECT().
		GetByID(ctx, "pr-999").
		Return(nil, domainerr.ErrPRNotFound)

	_, err := service.GetPR(ctx, "pr-999")
	assert.ErrorIs(t, err, domainerr.ErrPRNotFound)
}

func TestPRService_CreatePR_Failure_AuthorNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	mockUserGetter.
		EXPECT().
		GetByID(ctx, "author-1").
		Return(nil, domainerr.ErrUserNotFound)

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)
	_, err := service.CreatePR(ctx, "pr-1", "Test PR", "author-1")

	assert.ErrorIs(t, err, domainerr.ErrAuthorNotFound)
}

func TestPRService_CreatePR_Failure_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	author := entities.NewUser("author-1", "john", "backend", true)
	mockUserGetter.
		EXPECT().
		GetByID(ctx, "author-1").
		Return(author, nil)

	mockTeamGetter.
		EXPECT().
		GetByName(ctx, "backend").
		Return(nil, domainerr.ErrTeamNotFound)

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)
	_, err := service.CreatePR(ctx, "pr-1", "Test PR", "author-1")

	assert.ErrorIs(t, err, domainerr.ErrTeamNotFound)
}

func TestPRService_MergePR_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()
	openPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")
	mergedPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusMerged, time.Now(), "rev-1")

	mockPRRepo.EXPECT().
		GetByID(ctx, "pr-1").
		Return(openPR, nil)

	mockPRRepo.EXPECT().
		UpdateStatus(ctx, "pr-1", vo.PRStatusMerged).
		Return(nil)

	mockPRRepo.EXPECT().
		GetByID(ctx, "pr-1").
		Return(mergedPR, nil)

	pr, err := service.MergePR(ctx, "pr-1")

	assert.NoError(t, err)
	assert.Equal(t, vo.PRStatusMerged, pr.Status)
}

func TestPRService_MergePR_AlreadyMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)

	ctx := context.Background()
	mergedPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusMerged, time.Now(), "rev-1")

	mockPRRepo.EXPECT().
		GetByID(ctx, "pr-1").
		Return(mergedPR, nil)

	pr, err := service.MergePR(ctx, "pr-1")
	assert.NoError(t, err)
	assert.Equal(t, vo.PRStatusMerged, pr.Status)
}

func TestPRService_ReassignReviewer_Failure_PRAlreadyMerged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	mergedPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusMerged, time.Now(), "old-rev")
	mockPRRepo.
		EXPECT().
		GetByID(ctx, "pr-1").
		Return(mergedPR, nil)

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)
	_, _, err := service.ReassignReviewer(ctx, "pr-1", "old-rev")
	assert.ErrorIs(t, err, domainerr.ErrPRMerged)
}

func TestPRService_ReassignReviewer_Failure_ReviewerNotAssigned(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	openPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "other-rev")
	mockPRRepo.
		EXPECT().
		GetByID(ctx, "pr-1").
		Return(openPR, nil)

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)
	_, _, err := service.ReassignReviewer(ctx, "pr-1", "old-rev")
	assert.ErrorIs(t, err, domainerr.ErrReviewerNotAssigned)
}

func TestPRService_ReassignReviewer_Failure_NoActiveUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)

	openPR := entities.NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "old-rev")
	oldReviewer := entities.NewUser("old-rev", "old", "backend", true)
	team := entities.NewTeam(1, "backend", []*entities.User{})

	mockPRRepo.
		EXPECT().
		GetByID(ctx, "pr-1").
		Return(openPR, nil)

	mockUserGetter.
		EXPECT().
		GetByID(ctx, "old-rev").
		Return(oldReviewer, nil)

	mockTeamGetter.
		EXPECT().
		GetByName(ctx, "backend").
		Return(team, nil)

	mockUserGetter.
		EXPECT().
		GetActiveUsersInTeam(ctx, int64(1), []string{"old-rev", "author-1"}).
		Return([]*entities.User{}, nil)

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)
	_, _, err := service.ReassignReviewer(ctx, "pr-1", "old-rev")
	assert.ErrorIs(t, err, domainerr.ErrNoActiveUsers)
}

func TestPRService_selectRandom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := mocks.NewMockuserGetter(ctrl)
	mockTeamGetter := mocks.NewMockteamGetter(ctrl)
	mockPRRepo := mocks.NewMockprRepo(ctrl)
	trManager := newSimpleTrManager()
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	service := New(mockUserGetter, mockTeamGetter, mockPRRepo, trManager, log)

	f := func(candidates []*entities.User, maxCount, wantLen int) {
		t.Helper()
		result := service.selectRandom(candidates, maxCount)
		assert.Len(t, result, wantLen)
	}

	// empty
	f([]*entities.User{}, 2, 0)

	// < maxCount
	f([]*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
	}, 2, 1)

	// > maxCount
	f([]*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
		entities.NewUser("user-2", "jane", "backend", true),
		entities.NewUser("user-3", "bob", "backend", true),
	}, 2, 2)

	// = maxCount
	f([]*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
		entities.NewUser("user-2", "jane", "backend", true),
	}, 2, 2)
}

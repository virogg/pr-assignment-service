package pr_service

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
)

type userGetter interface {
	GetActiveUsersInTeam(ctx context.Context, id int64, excludeUsers []string) ([]*entities.User, error)
	GetByID(ctx context.Context, id string) (*entities.User, error)
}

type teamGetter interface {
	GetByName(ctx context.Context, name string) (*entities.Team, error)
}

type prRepo interface {
	Create(ctx context.Context, pr *entities.PullRequest) error
	GetByID(ctx context.Context, id string) (*entities.PullRequest, error)
	UpdateStatus(ctx context.Context, id string, status vo.PRStatus) error
	AddReviewers(ctx context.Context, id string, reviewerIDs []string) error
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error
}

type PRService struct {
	userGetter userGetter
	teamGetter teamGetter
	prRepo     prRepo
	trManager  *manager.Manager
	log        *slog.Logger
	rng        *rand.Rand
}

func New(userGetter userGetter, teamGetter teamGetter, prRepo prRepo, trManager *manager.Manager, log *slog.Logger) *PRService {
	return &PRService{
		userGetter: userGetter,
		teamGetter: teamGetter,
		prRepo:     prRepo,
		trManager:  trManager,
		log:        log,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*entities.PullRequest, error) {
	author, err := s.userGetter.GetByID(ctx, authorID)
	if err != nil {
		s.log.Error("failed to get author",
			slog.String("author_id", authorID),
			slog.Any("error", err),
		)
		return nil, domainerr.ErrAuthorNotFound
	}

	team, err := s.teamGetter.GetByName(ctx, author.TeamName)
	if err != nil {
		s.log.Error("failed to get author's team",
			slog.String("team_name", author.TeamName),
			slog.Any("error", err),
		)
		return nil, err
	}

	activeCandidates, err := s.userGetter.GetActiveUsersInTeam(ctx, team.ID, []string{author.ID})
	if err != nil {
		s.log.Error("failed to get active team members",
			slog.Int64("team_id", team.ID),
			slog.Any("error", err),
		)
		return nil, err
	}

	reviewers := s.selectRandom(activeCandidates, entities.MaxReviewers)
	reviewerIDs := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		reviewerIDs = append(reviewerIDs, r.ID)
	}

	pr := entities.NewPullRequest(prID, prName, author.ID, vo.PRStatusOpen, reviewerIDs, time.Now())

	err = s.trManager.Do(ctx, func(ctx context.Context) error {
		if err := s.prRepo.Create(ctx, pr); err != nil {
			s.log.Error("failed to create PR",
				slog.String("pr_id", pr.ID),
				slog.Any("error", err),
			)
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	s.log.Info("PR created successfully",
		slog.String("pr_id", pr.ID),
		slog.String("author_id", authorID),
		slog.Group("reviewers", reviewerIDs),
	)

	return pr, nil
}

func (s *PRService) GetPR(ctx context.Context, prID string) (*entities.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get PR",
			slog.String("pr_id", prID),
			slog.Any("error", err),
		)
		return nil, err
	}

	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*entities.PullRequest, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get PR",
			slog.String("pr_id", prID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if pr.IsMerged() {
		s.log.Info("PR already merged", slog.String("pr_id", prID))
		return pr, nil
	}

	if err := s.prRepo.UpdateStatus(ctx, prID, vo.PRStatusMerged); err != nil {
		s.log.Error("failed to merge PR",
			slog.String("pr_id", prID),
			slog.Any("error", err),
		)
		return nil, err
	}

	pr, err = s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get merged PR",
			slog.String("pr_id", prID),
			slog.Any("error", err),
		)
		return nil, err
	}

	s.log.Info("PR merged successfully", slog.String("pr_id", prID))

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldID string) (*entities.PullRequest, string, error) {
	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get PR",
			slog.String("pr_id", prID),
			slog.Any("error", err),
		)
		return nil, "", err
	}

	if pr.IsMerged() {
		return nil, "", domainerr.ErrPRMerged
	}

	if !pr.HasReviewer(oldID) {
		return nil, "", domainerr.ErrReviewerNotAssigned
	}

	oldReviewer, err := s.userGetter.GetByID(ctx, oldID)
	if err != nil {
		s.log.Error("failed to get old reviewer",
			slog.String("reviewer_id", oldID),
			slog.Any("error", err),
		)
		return nil, "", err
	}

	team, err := s.teamGetter.GetByName(ctx, oldReviewer.TeamName)
	if err != nil {
		s.log.Error("failed to get reviewer's team",
			slog.String("team_name", oldReviewer.TeamName),
			slog.Any("error", err),
		)
		return nil, "", err
	}

	excludeUserIDs := append(pr.ReviewerIDs, pr.AuthorID)
	activeCandidates, err := s.userGetter.GetActiveUsersInTeam(ctx, team.ID, excludeUserIDs)
	if err != nil {
		s.log.Error("failed to get active candidates",
			slog.Int64("team_id", team.ID),
			slog.Any("error", err),
		)
		return nil, "", err
	}

	if len(activeCandidates) == 0 {
		return nil, "", domainerr.ErrNoActiveUsers
	}

	newReviewer := s.selectRandom(activeCandidates, 1)[0]

	err = s.trManager.Do(ctx, func(ctx context.Context) error {
		if err := s.prRepo.RemoveReviewer(ctx, prID, oldID); err != nil {
			s.log.Error("failed to remove old reviewer",
				slog.String("pr_id", prID),
				slog.String("old_reviewer_id", oldID),
				slog.Any("error", err),
			)
			return err
		}

		if err := s.prRepo.AddReviewers(ctx, prID, []string{newReviewer.ID}); err != nil {
			s.log.Error("failed to add new reviewer",
				slog.String("pr_id", prID),
				slog.String("new_reviewer_id", newReviewer.ID),
				slog.Any("error", err),
			)
			return err
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	pr, err = s.prRepo.GetByID(ctx, prID)
	if err != nil {
		s.log.Error("failed to get updated PR",
			slog.String("pr_id", prID),
			slog.Any("error", err),
		)
		return nil, "", err
	}

	s.log.Info("reviewer reassigned",
		slog.String("pr_id", prID),
		slog.String("old_reviewer_id", oldReviewer.ID),
		slog.String("new_reviewer_id", newReviewer.ID),
	)

	return pr, newReviewer.ID, nil
}

func (s *PRService) selectRandom(candidates []*entities.User, maxCount int) []*entities.User {
	if len(candidates) == 0 {
		return []*entities.User{}
	}

	count := maxCount
	if len(candidates) < count {
		count = len(candidates)
	}

	shuffled := make([]*entities.User, len(candidates))
	copy(shuffled, candidates)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := s.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:count]
}

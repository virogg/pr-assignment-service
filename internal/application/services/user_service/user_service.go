package user_service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
)

type userGetterSetter interface {
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetActiveUsersInTeam(ctx context.Context, id int64, excludeUserIDs []string) ([]*entities.User, error)

	SetActive(ctx context.Context, id string, isActive bool) error
	SetUsersInactive(ctx context.Context, userIDs []string) error
}

type teamGetter interface {
	GetByName(ctx context.Context, name string) (*entities.Team, error)
}

type prRepo interface {
	AddReviewers(ctx context.Context, id string, reviewerIDs []string) error
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error
	GetByReviewer(ctx context.Context, id string) ([]*entities.PullRequest, error)
	GetOpenByReviewers(ctx context.Context, reviewerIDs []string) ([]*entities.PullRequest, error)
}

type UserService struct {
	userGetterSetter userGetterSetter
	teamGetter       teamGetter
	prRepo           prRepo
	trManager        *manager.Manager
	log              *slog.Logger
}

func New(userRepo userGetterSetter, teamRepo teamGetter, prRepo prRepo, trManager *manager.Manager, log *slog.Logger) *UserService {
	return &UserService{
		userGetterSetter: userRepo,
		teamGetter:       teamRepo,
		prRepo:           prRepo,
		trManager:        trManager,
		log:              log,
	}
}

func (s *UserService) SetUserActive(ctx context.Context, id string, isActive bool) (*entities.User, error) {
	if _, err := s.userGetterSetter.GetByID(ctx, id); err != nil {
		s.log.Error("failed to get user",
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	if err := s.userGetterSetter.SetActive(ctx, id, isActive); err != nil {
		s.log.Error("failed to update user status",
			slog.String("user_id", id),
			slog.Bool("is_active", isActive),
			slog.Any("error", err),
		)
		return nil, err
	}

	user, err := s.userGetterSetter.GetByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get updated user",
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	s.log.Info("user status updated",
		slog.String("user_id", id),
		slog.Bool("is_active", isActive),
	)

	return user, nil
}

func (s *UserService) GetUserReviews(ctx context.Context, id string) ([]*entities.PullRequest, error) {
	if _, err := s.userGetterSetter.GetByID(ctx, id); err != nil {
		s.log.Error("failed to get user",
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	prs, err := s.prRepo.GetByReviewer(ctx, id)
	if err != nil {
		s.log.Error("failed to get user reviews",
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	return prs, nil
}

func (s *UserService) DeactivateTeamUsers(ctx context.Context, teamName string) error {
	team, err := s.teamGetter.GetByName(ctx, teamName)
	if err != nil {
		s.log.Error("failed to get team",
			slog.String("team_name", teamName),
			slog.Any("error", err),
		)
		return err
	}

	activeUsers, err := s.userGetterSetter.GetActiveUsersInTeam(ctx, team.ID, []string{})
	if err != nil {
		s.log.Error("failed to get active users in team",
			slog.String("team_name", teamName),
			slog.Any("error", err),
		)
		return err
	}

	if len(activeUsers) == 0 {
		s.log.Info("no active users to deactivate",
			slog.String("team_name", teamName),
		)
		return nil
	}

	userIDs := make([]string, 0, len(activeUsers))
	for _, user := range activeUsers {
		userIDs = append(userIDs, user.ID)
	}

	err = s.trManager.Do(ctx, func(ctx context.Context) error {
		openPRs, err := s.prRepo.GetOpenByReviewers(ctx, userIDs)
		if err != nil {
			s.log.Error("failed to get open PRs",
				slog.Any("user_ids", userIDs),
				slog.Any("error", err),
			)
			return fmt.Errorf("failed to get open PRs: %w", err)
		}

		if err := s.userGetterSetter.SetUsersInactive(ctx, userIDs); err != nil {
			s.log.Error("failed to deactivate users",
				slog.Any("user_ids", userIDs),
				slog.Any("error", err),
			)
			return fmt.Errorf("failed to deactivate users: %w", err)
		}

		for _, pr := range openPRs {
			for _, reviewerID := range pr.ReviewerIDs {
				isDeactivated := false
				for _, uid := range userIDs {
					if reviewerID == uid {
						isDeactivated = true
						break
					}
				}

				if !isDeactivated {
					continue
				}

				activeUsers, err := s.userGetterSetter.GetActiveUsersInTeam(ctx, team.ID, append(pr.ReviewerIDs, pr.AuthorID))
				if err != nil {
					s.log.Error("failed to get active users",
						slog.Int64("team_id", team.ID),
						slog.Any("error", err),
					)
					return err
				}

				if len(activeUsers) == 0 {
					s.log.Warn("no active candidates for reassignment",
						slog.String("pr_id", pr.ID),
						slog.String("reviewer_id", reviewerID),
					)
					if err := s.prRepo.RemoveReviewer(ctx, pr.ID, reviewerID); err != nil {
						s.log.Error("failed to remove reviewer",
							slog.String("pr_id", pr.ID),
							slog.String("reviewer_id", reviewerID),
							slog.Any("error", err),
						)
						return err
					}
					continue
				}

				newReviewerID := activeUsers[0].ID

				if err := s.prRepo.RemoveReviewer(ctx, pr.ID, reviewerID); err != nil {
					s.log.Error("failed to remove old reviewer",
						slog.String("pr_id", pr.ID),
						slog.String("reviewer_id", reviewerID),
						slog.Any("error", err),
					)
					return err
				}

				if err := s.prRepo.AddReviewers(ctx, pr.ID, []string{newReviewerID}); err != nil {
					s.log.Error("failed to add new reviewer",
						slog.String("pr_id", pr.ID),
						slog.String("new_reviewer_id", newReviewerID),
						slog.Any("error", err),
					)
					return err
				}

				s.log.Info("reviewer reassigned",
					slog.String("pr_id", pr.ID),
					slog.String("old_reviewer_id", reviewerID),
					slog.String("new_reviewer_id", newReviewerID),
				)
			}
		}

		s.log.Info("deactivation completed",
			slog.String("team_name", teamName),
			slog.Int("users_count", len(userIDs)),
			slog.Int("affected_prs", len(openPRs)),
		)

		return nil
	})

	return err
}

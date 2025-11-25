package team_service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
)

//go:generate mockgen -destination=internal/application/services/team_service/mocks/mock_user_creator.go -package=mocks ./internal/application/services/team_service userCreator
type userCreator interface {
	Create(ctx context.Context, user *entities.User, teamID int64) error
}

//go:generate mockgen -destination=internal/application/services/team_service/mocks/mock_team_repo.go -package=mocks ./internal/application/services/team_service teamRepo
type teamRepo interface {
	Create(ctx context.Context, team *entities.Team) error
	TeamExists(ctx context.Context, name string) (bool, error)
	GetWithMembers(ctx context.Context, name string) (*entities.Team, error)
}

type TeamService struct {
	userCreator userCreator
	teamRepo    teamRepo
	trManager   *manager.Manager
	log         *slog.Logger
}

func New(userCreator userCreator, teamRepo teamRepo, trManager *manager.Manager, log *slog.Logger) *TeamService {
	return &TeamService{
		userCreator: userCreator,
		teamRepo:    teamRepo,
		trManager:   trManager,
		log:         log,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team *entities.Team) (*entities.Team, error) {
	if team == nil {
		return nil, domainerr.ErrInvalidInput
	}

	exists, err := s.teamRepo.TeamExists(ctx, team.Name)
	if err != nil {
		s.log.Error("failed to check team existence",
			slog.String("team_name", team.Name),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}

	if exists {
		return nil, domainerr.ErrTeamExists
	}

	var createdTeam *entities.Team
	err = s.trManager.Do(ctx, func(ctx context.Context) error {
		if err := s.teamRepo.Create(ctx, team); err != nil {
			s.log.Error("failed to create team",
				slog.String("team_name", team.Name),
				slog.Any("error", err),
			)
			return err
		}

		for _, user := range team.Users {
			if err := s.userCreator.Create(ctx, user, team.ID); err != nil {
				s.log.Error("failed to create user",
					slog.String("user_id", user.ID),
					slog.Any("error", err),
				)
				return fmt.Errorf("failed to create user %s: %w", user.ID, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	createdTeam, err = s.teamRepo.GetWithMembers(ctx, team.Name)
	if err != nil {
		return nil, err
	}

	s.log.Info("team created successfully",
		slog.String("team_name", team.Name),
		slog.Int("members_count", len(team.Users)),
	)

	return createdTeam, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*entities.Team, error) {
	team, err := s.teamRepo.GetWithMembers(ctx, teamName)
	if err != nil {
		s.log.Error("failed to get team",
			slog.String("team_name", teamName),
			slog.Any("error", err),
		)
		return nil, err
	}

	return team, nil
}

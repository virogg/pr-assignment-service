package mappers

import (
	"time"

	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
)

func ToTeamDTO(team *entities.Team) dto.TeamDTO {
	if team == nil {
		return dto.TeamDTO{}
	}

	members := make([]dto.UserDTO, 0, len(team.Users))
	for _, m := range team.Users {
		dto := dto.UserDTO{
			ID:       m.ID,
			Username: m.Username,
			TeamName: m.TeamName,
			IsActive: m.IsActive,
		}
		members = append(members, dto)
	}

	return dto.TeamDTO{
		TeamName: team.Name,
		Members:  members,
	}
}

func ToUserDTO(user *entities.User) dto.UserDTO {
	if user == nil {
		return dto.UserDTO{}
	}

	return dto.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		TeamName: user.TeamName,
		IsActive: user.IsActive,
	}
}

func ToPRDTO(pr *entities.PullRequest) dto.PullRequestDTO {
	if pr == nil {
		return dto.PullRequestDTO{}
	}

	dto := dto.PullRequestDTO{
		ID:        pr.ID,
		Name:      pr.Name,
		AuthorID:  pr.AuthorID,
		Status:    pr.Status.String(),
		Reviewers: pr.ReviewerIDs,
	}

	if !pr.CreatedAt.IsZero() {
		createdAt := pr.CreatedAt.Format(time.RFC3339)
		dto.CreatedAt = &createdAt
	}

	if pr.MergedAt != nil {
		mergedAt := pr.MergedAt.Format(time.RFC3339)
		dto.MergedAt = &mergedAt
	}

	return dto
}

func ToPRShortDTO(pr *entities.PullRequest) dto.PullRequestShortDTO {
	if pr == nil {
		return dto.PullRequestShortDTO{}
	}

	return dto.PullRequestShortDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          pr.Status.String(),
	}
}

func ToPRShortDTOs(prs []*entities.PullRequest) []dto.PullRequestShortDTO {
	dtos := make([]dto.PullRequestShortDTO, 0, len(prs))
	for _, pr := range prs {
		dtos = append(dtos, ToPRShortDTO(pr))
	}
	return dtos
}

func ToEntityTeam(dto *dto.CreateTeamRequest) *entities.Team {
	if dto == nil {
		return nil
	}

	members := make([]*entities.User, 0, len(dto.Members))
	for _, m := range dto.Members {
		user := entities.NewUser(m.ID, m.Username, dto.TeamName, m.IsActive)
		members = append(members, user)
	}

	return entities.NewTeam(0, dto.TeamName, members)
}

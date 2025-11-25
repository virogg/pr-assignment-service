package mappers

import (
	"testing"
	"time"

	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"

	"github.com/stretchr/testify/assert"
)

func TestToUserDTO(t *testing.T) {
	user := entities.NewUser("user-1", "john_doe", "backend", true)
	result := ToUserDTO(user)

	f := func(got, want string) {
		t.Helper()
		assert.Equal(t, want, got)
	}

	f(result.ID, "user-1")
	f(result.Username, "john_doe")
	f(result.TeamName, "backend")

	assert.True(t, result.IsActive)
}

func TestToUserDTO_Nil(t *testing.T) {
	result := ToUserDTO(nil)

	assert.Empty(t, result.ID)
	assert.Empty(t, result.Username)
	assert.Empty(t, result.TeamName)
}

func TestToTeamDTO(t *testing.T) {
	members := []*entities.User{
		entities.NewUser("user-1", "john", "backend", true),
		entities.NewUser("user-2", "jane", "backend", false),
	}
	team := entities.NewTeam(1, "backend", members)

	result := ToTeamDTO(team)

	assert.Equal(t, "backend", result.TeamName)
	assert.Len(t, result.Members, 2)
	assert.Equal(t, "user-1", result.Members[0].ID)
	assert.Equal(t, "user-2", result.Members[1].ID)
}

func TestToTeamDTO_Nil(t *testing.T) {
	result := ToTeamDTO(nil)

	assert.Empty(t, result.TeamName)
	assert.Empty(t, result.Members)
}

func TestToPRDTO(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	pr := entities.NewPullRequest("pr-1", "Fix bug", "author-1", vo.PRStatusOpen, createdAt, []string{"rev-1", "rev-2"}...)

	result := ToPRDTO(pr)

	f := func(got, want string) {
		t.Helper()
		assert.Equal(t, want, got)
	}

	f(result.ID, "pr-1")
	f(result.Name, "Fix bug")
	f(result.AuthorID, "author-1")
	f(result.Status, vo.PRStatusOpen.String())

	assert.Len(t, result.Reviewers, 2)
	assert.Equal(t, "rev-1", result.Reviewers[0])
	assert.NotNil(t, result.CreatedAt)

	expectedCreatedAt := createdAt.Format(time.RFC3339)
	assert.Equal(t, expectedCreatedAt, *result.CreatedAt)
	assert.Nil(t, result.MergedAt)
}

func TestToPRDTO_Merged(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	pr := entities.NewPullRequest("pr-1", "Fix bug", "author-1", vo.PRStatusOpen, createdAt, "rev-1")
	pr.Merge()

	result := ToPRDTO(pr)

	assert.Equal(t, vo.PRStatusMerged.String(), result.Status)
	assert.NotNil(t, result.MergedAt)
}

func TestToPRDTO_Nil(t *testing.T) {
	result := ToPRDTO(nil)

	assert.Empty(t, result.ID)
	assert.Empty(t, result.Name)
	assert.Empty(t, result.AuthorID)
}

func TestToPRShortDTO(t *testing.T) {
	createdAt := time.Now()
	pr := entities.NewPullRequest("pr-1", "Fix bug", "author-1", vo.PRStatusOpen, createdAt, "rev-1")

	result := ToPRShortDTO(pr)

	f := func(got, want string) {
		t.Helper()
		assert.Equal(t, want, got)
	}

	f(result.PullRequestID, "pr-1")
	f(result.PullRequestName, "Fix bug")
	f(result.AuthorID, "author-1")
	f(result.Status, "OPEN")
}

func TestToPRShortDTO_Nil(t *testing.T) {
	result := ToPRShortDTO(nil)

	assert.Empty(t, result.PullRequestID)
	assert.Empty(t, result.PullRequestName)
}

func TestToPRShortDTOs(t *testing.T) {
	createdAt := time.Now()
	prs := []*entities.PullRequest{
		entities.NewPullRequest("pr-1", "Fix bug 1", "author-1", vo.PRStatusOpen, createdAt, "rev-1"),
		entities.NewPullRequest("pr-2", "Fix bug 2", "author-2", vo.PRStatusMerged, createdAt, "rev-2"),
	}

	result := ToPRShortDTOs(prs)

	assert.Len(t, result, 2)

	assert.Equal(t, "pr-1", result[0].PullRequestID)
	assert.Equal(t, "pr-2", result[1].PullRequestID)
	assert.Equal(t, vo.PRStatusOpen.String(), result[0].Status)
	assert.Equal(t, vo.PRStatusMerged.String(), result[1].Status)
}

func TestToEntityTeam(t *testing.T) {
	req := &dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "user-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "user-2", Username: "jane", TeamName: "backend", IsActive: false},
		},
	}

	result := ToEntityTeam(req)

	assert.Equal(t, "backend", result.Name)
	assert.Len(t, result.Users, 2)

	f := func(user *entities.User, wantID, wantUsername, wantTeam string, wantActive bool) {
		t.Helper()
		assert.Equal(t, wantID, user.ID)
		assert.Equal(t, wantUsername, user.Username)
		assert.Equal(t, wantTeam, user.TeamName)
		assert.Equal(t, wantActive, user.IsActive)
	}

	f(result.Users[0], "user-1", "john", "backend", true)
	f(result.Users[1], "user-2", "jane", "backend", false)
}

func TestToEntityTeam_Nil(t *testing.T) {
	result := ToEntityTeam(nil)

	assert.Nil(t, result)
}

//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/virogg/pr-assignment-service/internal/application/dto"

	"github.com/stretchr/testify/suite"
)

type UserE2ETestSuite struct {
	suite.Suite
	server *TestServer
	client *http.Client
}

func (s *UserE2ETestSuite) SetupTest() {
	s.server = SetupTestServer(s.T())
	s.client = &http.Client{}
}

func (s *UserE2ETestSuite) TearDownTest() {
	s.server.Cleanup(s.T())
}

func (s *UserE2ETestSuite) TestUserActivationAndReviews() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "author-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "reviewer-1", Username: "jane", TeamName: "backend", IsActive: true},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create team")
	resp.Body.Close()

	createPRReq := dto.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add new feature",
		AuthorID:        "author-1",
	}

	body, _ = json.Marshal(createPRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to create PR")
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodGet, s.server.BaseURL+"/users/getReview?user_id=reviewer-1", nil)

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to get user reviews")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "get user reviews status code")

	var reviewsResp dto.UserReviewsResponse
	err = json.NewDecoder(resp.Body).Decode(&reviewsResp)
	s.Require().NoError(err, "failed to decode reviews response")

	s.Equal("reviewer-1", reviewsResp.UserID, "reviews user ID")
	s.NotEmpty(reviewsResp.PullRequests, "reviewer should have at least one PR assigned")

	setActiveReq := dto.SetUserActiveRequest{
		UserID:   "reviewer-1",
		IsActive: false,
	}

	body, _ = json.Marshal(setActiveReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/users/setIsActive", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to set user active")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "set user active status code")

	var userResp dto.UserResponse
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	s.Require().NoError(err, "failed to decode user response")

	s.False(userResp.User.IsActive, "user should be inactive after deactivation")
}

func (s *UserE2ETestSuite) TestTeamDeactivation() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "user-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "user-2", Username: "jane", TeamName: "backend", IsActive: true},
			{ID: "user-3", Username: "bob", TeamName: "backend", IsActive: true},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create team")
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/teams/deactivate?team_name=backend", nil)

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to deactivate team")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "deactivate team status code")

	req, _ = http.NewRequest(http.MethodGet, s.server.BaseURL+"/team/get?team_name=backend", nil)

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to get team")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "get team status code")

	var teamResp dto.TeamDTO
	err = json.NewDecoder(resp.Body).Decode(&teamResp)
	s.Require().NoError(err, "failed to decode team response")

	for _, member := range teamResp.Members {
		s.False(member.IsActive, "user %s should be inactive after team deactivation", member.Username)
	}
}

func (s *UserE2ETestSuite) TestStatistics() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "author-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "reviewer-1", Username: "jane", TeamName: "backend", IsActive: true},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create team")
	resp.Body.Close()

	for i := 1; i <= 3; i++ {
		createPRReq := dto.CreatePRRequest{
			PullRequestID:   "pr-" + string(rune(i+'0')),
			PullRequestName: "Feature " + string(rune(i+'0')),
			AuthorID:        "author-1",
		}

		body, _ = json.Marshal(createPRReq)
		req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/create", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ = s.client.Do(req)
		resp.Body.Close()
	}

	mergePRReq := dto.MergePRRequest{
		PullRequestID: "pr-1",
	}
	body, _ = json.Marshal(mergePRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = s.client.Do(req)
	resp.Body.Close()

	req, _ = http.NewRequest(http.MethodGet, s.server.BaseURL+"/stats", nil)

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to get stats")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "get stats status code")

	var statsResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&statsResp)
	s.Require().NoError(err, "failed to decode stats response")

	s.Contains(statsResp, "users", "stats should contain 'users' field")
	s.Contains(statsResp, "prs", "stats should contain 'prs' field")
	s.Contains(statsResp, "teams", "stats should contain 'teams' field")

	prStats, ok := statsResp["prs"].(map[string]interface{})
	s.Require().True(ok, "prs field should be an object")

	totalPRs, ok := prStats["total_prs"].(float64)
	s.True(ok && totalPRs >= 3, "total_prs should be >= 3")

	mergedPRs, ok := prStats["merged_prs"].(float64)
	s.True(ok && mergedPRs >= 1, "merged_prs should be >= 1")
}

func TestUserE2ESuite(t *testing.T) {
	suite.Run(t, new(UserE2ETestSuite))
}

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

type PRE2ETestSuite struct {
	suite.Suite
	server *TestServer
	client *http.Client
}

func (s *PRE2ETestSuite) SetupTest() {
	s.server = SetupTestServer(s.T())
	s.client = &http.Client{}
}

func (s *PRE2ETestSuite) TearDownTest() {
	s.server.Cleanup(s.T())
}

func (s *PRE2ETestSuite) TestPRLifecycle() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "author-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "reviewer-1", Username: "jane", TeamName: "backend", IsActive: true},
			{ID: "reviewer-2", Username: "bob", TeamName: "backend", IsActive: true},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create team")
	resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode, "create team status code")

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
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode, "create PR status code")

	var createPRResp dto.PRResponse
	err = json.NewDecoder(resp.Body).Decode(&createPRResp)
	s.Require().NoError(err, "failed to decode create PR response")

	s.Equal("pr-1", createPRResp.PR.ID, "PR ID")
	s.Equal("OPEN", createPRResp.PR.Status, "PR status")
	s.NotEmpty(createPRResp.PR.Reviewers, "PR should have at least one reviewer assigned")

	mergePRReq := dto.MergePRRequest{
		PullRequestID: "pr-1",
	}

	body, _ = json.Marshal(mergePRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to merge PR")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "merge PR status code")

	var mergePRResp dto.PRResponse
	err = json.NewDecoder(resp.Body).Decode(&mergePRResp)
	s.Require().NoError(err, "failed to decode merge PR response")

	s.Equal("MERGED", mergePRResp.PR.Status, "merged PR status")
}

func (s *PRE2ETestSuite) TestPRReassignment() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "author-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "reviewer-1", Username: "jane", TeamName: "backend", IsActive: true},
			{ID: "reviewer-2", Username: "bob", TeamName: "backend", IsActive: true},
			{ID: "reviewer-3", Username: "alice", TeamName: "backend", IsActive: true},
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

	var createPRResp dto.PRResponse
	json.NewDecoder(resp.Body).Decode(&createPRResp)
	resp.Body.Close()

	s.Require().NotEmpty(createPRResp.PR.Reviewers, "PR should have at least one reviewer")

	oldReviewer := createPRResp.PR.Reviewers[0]

	reassignReq := dto.ReassignPRRequest{
		PullRequestID: "pr-1",
		OldUserID:     oldReviewer,
	}

	body, _ = json.Marshal(reassignReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/reassign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to reassign PR")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "reassign PR status code")

	var reassignResp dto.ReassignPRResponse
	err = json.NewDecoder(resp.Body).Decode(&reassignResp)
	s.Require().NoError(err, "failed to decode reassign PR response")

	s.NotEmpty(reassignResp.PR.Reviewers, "PR should still have a reviewer after reassignment")
	if len(reassignResp.PR.Reviewers) > 0 {
		s.NotEqual(oldReviewer, reassignResp.PR.Reviewers[0], "reviewer should have changed after reassignment")
	}
}

func (s *PRE2ETestSuite) TestPRCreation_DuplicatePR() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "author-1", Username: "john", TeamName: "backend", IsActive: true},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := s.client.Do(req)
	resp.Body.Close()

	createPRReq := dto.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add new feature",
		AuthorID:        "author-1",
	}

	body, _ = json.Marshal(createPRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create PR")
	resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode, "first create PR status code")

	body, _ = json.Marshal(createPRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to send duplicate PR request")
	defer resp.Body.Close()

	s.Equal(http.StatusConflict, resp.StatusCode, "duplicate PR status code")
}

func (s *PRE2ETestSuite) TestPRMerge_AlreadyMerged() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "author-1", Username: "john", TeamName: "backend", IsActive: true},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := s.client.Do(req)
	resp.Body.Close()

	createPRReq := dto.CreatePRRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Add new feature",
		AuthorID:        "author-1",
	}

	body, _ = json.Marshal(createPRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/create", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = s.client.Do(req)
	resp.Body.Close()

	mergePRReq := dto.MergePRRequest{
		PullRequestID: "pr-1",
	}

	body, _ = json.Marshal(mergePRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to merge PR")
	resp.Body.Close()

	body, _ = json.Marshal(mergePRReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/pullRequest/merge", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to send duplicate merge request")
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode, "merge already merged PR status code (idempotent)")
}

func TestPRE2ESuite(t *testing.T) {
	suite.Run(t, new(PRE2ETestSuite))
}

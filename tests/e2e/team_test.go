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

type TeamE2ETestSuite struct {
	suite.Suite
	server *TestServer
	client *http.Client
}

func (s *TeamE2ETestSuite) SetupTest() {
	s.server = SetupTestServer(s.T())
	s.client = &http.Client{}
}

func (s *TeamE2ETestSuite) TearDownTest() {
	s.server.Cleanup(s.T())
}

func (s *TeamE2ETestSuite) TestTeamCreationAndRetrieval() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members: []dto.UserDTO{
			{ID: "user-1", Username: "john", TeamName: "backend", IsActive: true},
			{ID: "user-2", Username: "jane", TeamName: "backend", IsActive: true},
			{ID: "user-3", Username: "bob", TeamName: "backend", IsActive: false},
		},
	}

	body, _ := json.Marshal(createTeamReq)
	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create team")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode, "create team status code")

	var createResp dto.TeamResponse
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	s.Require().NoError(err, "failed to decode create response")

	s.Equal("backend", createResp.Team.TeamName, "team name")
	s.Len(createResp.Team.Members, 3, "team members count")

	req, _ = http.NewRequest(http.MethodGet, s.server.BaseURL+"/team/get?team_name=backend", nil)

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to get team")
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode, "get team status code")

	var getResp dto.TeamDTO
	err = json.NewDecoder(resp.Body).Decode(&getResp)
	s.Require().NoError(err, "failed to decode get response")

	s.Equal("backend", getResp.TeamName, "retrieved team name")
	s.Len(getResp.Members, 3, "retrieved team members count")

	johnFound := false
	for _, member := range getResp.Members {
		if member.Username == "john" && member.IsActive {
			johnFound = true
			break
		}
	}
	s.True(johnFound, "john should be found in team members and be active")
}

func (s *TeamE2ETestSuite) TestTeamCreation_DuplicateTeam() {
	createTeamReq := dto.CreateTeamRequest{
		TeamName: "backend",
		Members:  []dto.UserDTO{},
	}

	body, _ := json.Marshal(createTeamReq)

	req, _ := http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to create team")
	resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode, "first create team status code")

	body, _ = json.Marshal(createTeamReq)
	req, _ = http.NewRequest(http.MethodPost, s.server.BaseURL+"/team/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = s.client.Do(req)
	s.Require().NoError(err, "failed to send duplicate team request")
	defer resp.Body.Close()

	s.Equal(http.StatusBadRequest, resp.StatusCode, "duplicate team status code")
}

func (s *TeamE2ETestSuite) TestTeamRetrieval_NotFound() {
	req, _ := http.NewRequest(http.MethodGet, s.server.BaseURL+"/team/get?team_name=nonexistent", nil)

	resp, err := s.client.Do(req)
	s.Require().NoError(err, "failed to get team")
	defer resp.Body.Close()

	s.Equal(http.StatusNotFound, resp.StatusCode, "get nonexistent team status code")
}

func TestTeamE2ESuite(t *testing.T) {
	suite.Run(t, new(TeamE2ETestSuite))
}

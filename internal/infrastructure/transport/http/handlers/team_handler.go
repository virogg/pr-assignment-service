package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/virogg/pr-assignment-service/internal/application/dto"
	"github.com/virogg/pr-assignment-service/internal/application/mappers"
	"github.com/virogg/pr-assignment-service/internal/domain/entities"
	transportmappers "github.com/virogg/pr-assignment-service/internal/infrastructure/transport/mappers"
	pkg "github.com/virogg/pr-assignment-service/pkg/http"
)

//go:generate mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_team_service.go -package=mocks ./internal/infrastructure/transport/http/handlers teamService
type teamService interface {
	CreateTeam(ctx context.Context, team *entities.Team) (*entities.Team, error)
	GetTeam(ctx context.Context, teamName string) (*entities.Team, error)
}

type TeamHandler struct {
	teamService teamService
	log         *slog.Logger
}

func NewTeamHandler(teamService teamService, log *slog.Logger) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
		log:         log,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request", slog.Any("error", err))
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "invalid request body"))
		return
	}

	if req.TeamName == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "team_name is required"))
		return
	}

	team := mappers.ToEntityTeam(&req)

	createdTeam, err := h.teamService.CreateTeam(r.Context(), team)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusCreated, dto.TeamResponse{
		Team: mappers.ToTeamDTO(createdTeam),
	})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "team_name is required"))
		return
	}

	team, err := h.teamService.GetTeam(r.Context(), teamName)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, mappers.ToTeamDTO(team))
}

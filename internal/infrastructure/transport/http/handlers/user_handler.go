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

//go:generate mockgen -destination=internal/infrastructure/transport/http/handlers/mocks/mock_user_service.go -package=mocks ./internal/infrastructure/transport/http/handlers userService
type userService interface {
	SetUserActive(ctx context.Context, id string, isActive bool) (*entities.User, error)
	GetUserReviews(ctx context.Context, id string) ([]*entities.PullRequest, error)
	DeactivateTeamUsers(ctx context.Context, teamName string) error
}

type UserHandler struct {
	userService userService
	log         *slog.Logger
}

func NewUserHandler(userService userService, log *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log,
	}
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	var req dto.SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request", slog.Any("error", err))
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "invalid request body"))
		return
	}

	if req.UserID == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "user_id is required"))
		return
	}

	user, err := h.userService.SetUserActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, dto.UserResponse{
		User: mappers.ToUserDTO(user),
	})
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "user_id is required"))
		return
	}
	prs, err := h.userService.GetUserReviews(r.Context(), userID)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, dto.UserReviewsResponse{
		UserID:       userID,
		PullRequests: mappers.ToPRShortDTOs(prs),
	})
}

func (h *UserHandler) DeactivateUsers(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "team_name is required"))
		return
	}

	err := h.userService.DeactivateTeamUsers(r.Context(), teamName)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "team users deactivated and PRs reassigned successfully",
	})
}

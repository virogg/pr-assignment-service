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

type prService interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*entities.PullRequest, error)
	GetPR(ctx context.Context, prID string) (*entities.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*entities.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldID string) (*entities.PullRequest, string, error)
}

type PRHandler struct {
	prService prService
	log       *slog.Logger
}

func NewPRHandler(prService prService, log *slog.Logger) *PRHandler {
	return &PRHandler{
		prService: prService,
		log:       log,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req dto.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request", slog.Any("error", err))
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "invalid request body"))
		return
	}

	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "pull_request_id, pull_request_name, and author_id are required"))
		return
	}

	pr, err := h.prService.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusCreated, dto.PRResponse{
		PR: mappers.ToPRDTO(pr),
	})
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req dto.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request", slog.Any("error", err))
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "invalid request body"))
		return
	}

	if req.PullRequestID == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "pull_request_id is required"))
		return
	}

	pr, err := h.prService.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, dto.PRResponse{
		PR: mappers.ToPRDTO(pr),
	})
}

func (h *PRHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	var req dto.ReassignPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("failed to decode request", slog.Any("error", err))
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "invalid request body"))
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		pkg.RespondError(w, http.StatusBadRequest, pkg.NewErrorResponse(transportmappers.ErrCodeInvalidInput, "pull_request_id and old_user_id are required"))
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		statusCode, errResp := transportmappers.MapErrorToHTTP(err)
		pkg.RespondError(w, statusCode, errResp)
		return
	}

	pkg.RespondJSON(w, http.StatusOK, dto.ReassignPRResponse{
		PR:         mappers.ToPRDTO(pr),
		ReplacedBy: newReviewerID,
	})
}

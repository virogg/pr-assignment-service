package mappers

import (
	"errors"
	"net/http"

	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	pkg "github.com/virogg/pr-assignment-service/pkg/http"
)

const (
	ErrCodeTeamExists   = "TEAM_EXISTS"
	ErrCodePRExists     = "PR_EXISTS"
	ErrCodePRMerged     = "PR_MERGED"
	ErrCodeNotAssigned  = "NOT_ASSIGNED"
	ErrCodeNoCandidate  = "NO_CANDIDATE"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeInvalidInput = "INVALID_INPUT"
)

func MapErrorToHTTP(err error) (int, pkg.ErrorResponse) {
	switch {
	case errors.Is(err, domainerr.ErrTeamExists):
		return http.StatusBadRequest, pkg.NewErrorResponse(ErrCodeTeamExists, "team_name already exists")

	case errors.Is(err, domainerr.ErrPRExists):
		return http.StatusConflict, pkg.NewErrorResponse(ErrCodePRExists, "PR id already exists")

	case errors.Is(err, domainerr.ErrPRMerged):
		return http.StatusConflict, pkg.NewErrorResponse(ErrCodePRMerged, "cannot reassign on merged PR")

	case errors.Is(err, domainerr.ErrReviewerNotAssigned):
		return http.StatusConflict, pkg.NewErrorResponse(ErrCodeNotAssigned, "reviewer is not assigned to this PR")

	case errors.Is(err, domainerr.ErrNoActiveUsers):
		return http.StatusConflict, pkg.NewErrorResponse(ErrCodeNoCandidate, "no active replacement candidate in team")

	case errors.Is(err, domainerr.ErrTeamNotFound), errors.Is(err, domainerr.ErrUserNotFound), errors.Is(err, domainerr.ErrPRNotFound), errors.Is(err, domainerr.ErrAuthorNotFound):
		return http.StatusNotFound, pkg.NewErrorResponse(ErrCodeNotFound, "resource not found")

	case errors.Is(err, domainerr.ErrInvalidInput):
		return http.StatusBadRequest, pkg.NewErrorResponse(ErrCodeInvalidInput, "invalid input data")

	default:
		return http.StatusInternalServerError, pkg.NewErrorResponse("INTERNAL_ERROR", "internal server error")
	}
}

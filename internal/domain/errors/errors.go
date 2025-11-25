package errors

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrAlreadyReviewer     = errors.New("user is already a reviewer of this PR")
	ErrNotReviewer         = errors.New("user is not a reviewer of this PR")
	ErrReviewerNotAssigned = errors.New("user is not assigned to this PR")

	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")

	ErrPRExists           = errors.New("PR already exists")
	ErrPRNotFound         = errors.New("PR not found")
	ErrPRMerged           = errors.New("PR already merged")
	ErrNotEnoughReviewers = errors.New("not enough reviewers available")
	ErrAuthorNotFound     = errors.New("author of this PR not found")
	ErrNoActiveUsers      = errors.New("no active replacement users in team")

	ErrInvalidInput = errors.New("invalid input data")
)

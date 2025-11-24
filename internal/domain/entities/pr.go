package entities

import (
	"slices"
	"time"

	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"
)

const (
	MaxReviewers = 2
)

// PullRequest (PR) — сущность с идентификатором, названием, автором, статусом OPEN|MERGED
// и списком назначенных ревьюверов (до 2).
type PullRequest struct {
	ID          string
	Name        string
	AuthorID    string
	Status      vo.PRStatus
	ReviewerIDs []string
	CreatedAt   time.Time
	MergedAt    *time.Time
}

func NewPullRequest(id, name, authorID string, status vo.PRStatus, reviewers []string, createdAt time.Time) *PullRequest {
	return &PullRequest{
		ID:          id,
		Name:        name,
		AuthorID:    authorID,
		Status:      status,
		ReviewerIDs: reviewers,
		CreatedAt:   createdAt,
		MergedAt:    nil,
	}
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == vo.PRStatusMerged
}

func (pr *PullRequest) CanMerge() bool {
	return !pr.IsMerged()
}

func (pr *PullRequest) HasReviewer(userID string) bool {
	reviewerIDs := make([]string, 0, len(pr.ReviewerIDs))
	for _, reviewerID := range pr.ReviewerIDs {
		reviewerIDs = append(reviewerIDs, reviewerID)
	}

	return slices.Contains(reviewerIDs, userID)
}

func (pr *PullRequest) AddReviewer(user *User) error {
	if pr.IsMerged() {
		return domainerr.ErrPRMerged
	}

	if pr.HasReviewer(user.ID) {
		// or nil ?
		return domainerr.ErrAlreadyReviewer
	}

	if len(pr.ReviewerIDs) >= MaxReviewers {
		return domainerr.ErrNotEnoughReviewers
	}

	pr.ReviewerIDs = append(pr.ReviewerIDs, user.ID)
	return nil
}

func (pr *PullRequest) ReplaceReviewer(old, new *User) error {
	if pr.IsMerged() {
		return domainerr.ErrPRMerged
	}

	if !pr.HasReviewer(old.ID) {
		return domainerr.ErrNotReviewer
	}

	for i, reviewerID := range pr.ReviewerIDs {
		if reviewerID == old.ID {
			pr.ReviewerIDs[i] = new.ID
			return nil
		}
	}

	return domainerr.ErrReviewerNotAssigned
}

func (pr *PullRequest) Merge() {
	if !pr.IsMerged() {
		pr.Status = vo.PRStatusMerged
		now := time.Now()
		pr.MergedAt = &now
	}
}

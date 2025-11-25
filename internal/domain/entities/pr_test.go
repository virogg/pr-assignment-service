package entities

import (
	"testing"
	"time"

	domainerr "github.com/virogg/pr-assignment-service/internal/domain/errors"
	vo "github.com/virogg/pr-assignment-service/internal/domain/value_objects"

	"github.com/stretchr/testify/assert"
)

func TestPullRequest_IsMerged(t *testing.T) {
	f := func(status vo.PRStatus, want bool) {
		t.Helper()
		pr := NewPullRequest("pr-1", "Test PR", "author-1", status, time.Now(), "rev-1")
		assert.Equal(t, want, pr.IsMerged())
	}

	f(vo.PRStatusOpen, false)
	f(vo.PRStatusMerged, true)
}

func TestPullRequest_CanMerge(t *testing.T) {
	f := func(status vo.PRStatus, want bool) {
		t.Helper()
		pr := NewPullRequest("pr-1", "Test PR", "author-1", status, time.Now(), "rev-1")
		assert.Equal(t, want, pr.CanMerge())
	}

	f(vo.PRStatusOpen, true)
	f(vo.PRStatusMerged, false)
}

func TestPullRequest_HasReviewer(t *testing.T) {
	reviewers := []string{"rev-1", "rev-2"}
	pr := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), reviewers...)

	f := func(reviewerID string, want bool) {
		t.Helper()
		assert.Equal(t, want, pr.HasReviewer(reviewerID))
	}

	// existing
	f("rev-1", true)
	f("rev-2", true)

	// nonexisting
	f("rev-3", false)
	f("author-1", false)
}

func TestPullRequest_AddReviewer_Success(t *testing.T) {
	pr := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")
	newReviewer := NewUser("rev-2", "reviewer2", "team1", true)

	assert.NoError(t, pr.AddReviewer(newReviewer))
	assert.True(t, pr.HasReviewer("rev-2"))
	assert.Len(t, pr.ReviewerIDs, 2)
}

func TestPullRequest_AddReviewer_Failure(t *testing.T) {
	f := func(pr *PullRequest, user *User, wantErr error) {
		t.Helper()
		err := pr.AddReviewer(user)
		assert.ErrorIs(t, err, wantErr)
	}

	mergedPR := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusMerged, time.Now(), "rev-1")
	newReviewer := NewUser("rev-2", "reviewer2", "team1", true)
	f(mergedPR, newReviewer, domainerr.ErrPRMerged)

	openPR := NewPullRequest("pr-2", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")
	existingReviewer := NewUser("rev-1", "reviewer1", "team1", true)
	f(openPR, existingReviewer, domainerr.ErrAlreadyReviewer)

	fullPR := NewPullRequest("pr-3", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), []string{"rev-1", "rev-2"}...)
	thirdReviewer := NewUser("rev-3", "reviewer3", "team1", true)
	f(fullPR, thirdReviewer, domainerr.ErrNotEnoughReviewers)
}

func TestPullRequest_ReplaceReviewer_Success(t *testing.T) {
	pr := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), []string{"rev-1", "rev-2"}...)
	oldReviewer := NewUser("rev-1", "reviewer1", "team1", true)
	newReviewer := NewUser("rev-3", "reviewer3", "team1", true)

	assert.NoError(t, pr.ReplaceReviewer(oldReviewer, newReviewer))
	assert.False(t, pr.HasReviewer("rev-1"))
	assert.True(t, pr.HasReviewer("rev-3"))
	assert.Len(t, pr.ReviewerIDs, 2)
}

func TestPullRequest_ReplaceReviewer_Failure(t *testing.T) {
	f := func(pr *PullRequest, oldRev, newRev *User, wantErr error) {
		t.Helper()
		err := pr.ReplaceReviewer(oldRev, newRev)
		assert.ErrorIs(t, err, wantErr)
	}

	mergedPR := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusMerged, time.Now(), "rev-1")
	oldRev := NewUser("rev-1", "reviewer1", "team1", true)
	newRev := NewUser("rev-2", "reviewer2", "team1", true)
	f(mergedPR, oldRev, newRev, domainerr.ErrPRMerged)

	openPR := NewPullRequest("pr-2", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")
	notAssigned := NewUser("rev-3", "reviewer3", "team1", true)
	f(openPR, notAssigned, newRev, domainerr.ErrNotReviewer)
}

func TestPullRequest_Merge(t *testing.T) {
	pr := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")

	assert.False(t, pr.IsMerged())
	assert.Nil(t, pr.MergedAt)

	before := time.Now()
	pr.Merge()
	after := time.Now()

	assert.True(t, pr.IsMerged())
	assert.Equal(t, vo.PRStatusMerged, pr.Status)
	assert.NotNil(t, pr.MergedAt)

	if pr.MergedAt.Before(before) || pr.MergedAt.After(after) {
		t.Fatalf("MergedAt timestamp is out of expected range")
	}
}

func TestPullRequest_Merge_Idempotent(t *testing.T) {
	pr := NewPullRequest("pr-1", "Test PR", "author-1", vo.PRStatusOpen, time.Now(), "rev-1")

	pr.Merge()
	firstMergedAt := pr.MergedAt

	// idempotent op
	pr.Merge()

	assert.Equal(t, firstMergedAt, pr.MergedAt)
	assert.True(t, pr.IsMerged())
}

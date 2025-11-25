package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTeam(t *testing.T) {
	users := []*User{
		NewUser("user-1", "john", "backend", true),
		NewUser("user-2", "jane", "backend", false),
	}

	team := NewTeam(1, "backend", users)

	assert.NotNil(t, team)
	assert.Equal(t, int64(1), team.ID)
	assert.Equal(t, "backend", team.Name)
	assert.Len(t, team.Users, 2)
}

func TestTeam_GetActiveUsers(t *testing.T) {
	tests := []struct {
		name      string
		users     []*User
		wantCount int
		wantIDs   []string
	}{
		{
			name: "all active users",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
			},
			wantCount: 2,
			wantIDs:   []string{"user-1", "user-2"},
		},
		{
			name: "mixed active and inactive users",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", false),
				NewUser("user-3", "bob", "backend", true),
			},
			wantCount: 2,
			wantIDs:   []string{"user-1", "user-3"},
		},
		{
			name: "no active users",
			users: []*User{
				NewUser("user-1", "john", "backend", false),
				NewUser("user-2", "jane", "backend", false),
			},
			wantCount: 0,
			wantIDs:   []string{},
		},
		{
			name:      "empty team",
			users:     []*User{},
			wantCount: 0,
			wantIDs:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team := NewTeam(1, "backend", tt.users)
			activeUsers := team.GetActiveUsers()

			assert.Len(t, activeUsers, tt.wantCount)

			actualIDs := make([]string, len(activeUsers))
			for i, u := range activeUsers {
				actualIDs[i] = u.ID
			}

			assert.ElementsMatch(t, tt.wantIDs, actualIDs)
		})
	}
}

func TestTeam_GetActiveUsersExclusive(t *testing.T) {
	tests := []struct {
		name       string
		users      []*User
		excludeIDs []string
		wantCount  int
		wantIDs    []string
	}{
		{
			name: "exclude one user",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
				NewUser("user-3", "bob", "backend", true),
			},
			excludeIDs: []string{"user-2"},
			wantCount:  2,
			wantIDs:    []string{"user-1", "user-3"},
		},
		{
			name: "exclude multiple users",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
				NewUser("user-3", "bob", "backend", true),
				NewUser("user-4", "alice", "backend", true),
			},
			excludeIDs: []string{"user-1", "user-3"},
			wantCount:  2,
			wantIDs:    []string{"user-2", "user-4"},
		},
		{
			name: "exclude inactive user",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", false),
				NewUser("user-3", "bob", "backend", true),
			},
			excludeIDs: []string{"user-2"},
			wantCount:  2,
			wantIDs:    []string{"user-1", "user-3"},
		},
		{
			name: "exclude non-existent user",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
			},
			excludeIDs: []string{"user-999"},
			wantCount:  2,
			wantIDs:    []string{"user-1", "user-2"},
		},
		{
			name: "no exclusions",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
			},
			excludeIDs: []string{},
			wantCount:  2,
			wantIDs:    []string{"user-1", "user-2"},
		},
		{
			name: "exclude all active users",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
			},
			excludeIDs: []string{"user-1", "user-2"},
			wantCount:  0,
			wantIDs:    []string{},
		},
		{
			name: "exclude with inactive users",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", false),
				NewUser("user-3", "bob", "backend", true),
			},
			excludeIDs: []string{"user-1"},
			wantCount:  1,
			wantIDs:    []string{"user-3"},
		},
		{
			name: "duplicate exclusion IDs",
			users: []*User{
				NewUser("user-1", "john", "backend", true),
				NewUser("user-2", "jane", "backend", true),
				NewUser("user-3", "bob", "backend", true),
			},
			excludeIDs: []string{"user-1", "user-1", "user-2"},
			wantCount:  1,
			wantIDs:    []string{"user-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team := NewTeam(1, "backend", tt.users)
			activeUsers := team.GetActiveUsersExclusive(tt.excludeIDs)

			assert.Len(t, activeUsers, tt.wantCount)

			actualIDs := make([]string, len(activeUsers))
			for i, u := range activeUsers {
				actualIDs[i] = u.ID
			}

			assert.ElementsMatch(t, tt.wantIDs, actualIDs)
		})
	}
}

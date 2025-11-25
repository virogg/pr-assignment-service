package value_objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status PRStatus
		want   bool
	}{
		{
			name:   "open status is valid",
			status: PRStatusOpen,
			want:   true,
		},
		{
			name:   "merged status is valid",
			status: PRStatusMerged,
			want:   true,
		},
		{
			name:   "invalid status",
			status: PRStatus("INVALID"),
			want:   false,
		},
		{
			name:   "empty status",
			status: PRStatus(""),
			want:   false,
		},
		{
			name:   "lowercase status",
			status: PRStatus("open"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status PRStatus
		want   string
	}{
		{
			name:   "open status to string",
			status: PRStatusOpen,
			want:   "OPEN",
		},
		{
			name:   "merged status to string",
			status: PRStatusMerged,
			want:   "MERGED",
		},
		{
			name:   "invalid status to string",
			status: PRStatus("INVALID"),
			want:   "INVALID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

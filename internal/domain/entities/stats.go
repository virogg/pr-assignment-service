package entities

type UserStats struct {
	ID                   string `json:"user_id"`
	Username             string `json:"username"`
	TotalAssignments     int    `json:"total_assignments"`
	OpenAssignments      int    `json:"open_assignments"`
	CompletedAssignments int    `json:"completed_assignments"`
}

type PRStats struct {
	TotalPRs          int     `json:"total_prs"`
	OpenPRs           int     `json:"open_prs"`
	MergedPRs         int     `json:"merged_prs"`
	AvgReviewersPerPR float64 `json:"avg_reviewers_per_pr"`
}

type TeamStats struct {
	TeamName      string `json:"team_name"`
	TotalMembers  int    `json:"total_members"`
	ActiveMembers int    `json:"active_members"`
	TotalPRs      int    `json:"total_prs"`
}

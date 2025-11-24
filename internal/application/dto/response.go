package dto

type TeamResponse struct {
	Team TeamDTO `json:"team"`
}

type UserResponse struct {
	User UserDTO `json:"user"`
}

type PRResponse struct {
	PR PullRequestDTO `json:"pr"`
}

type ReassignPRResponse struct {
	PR         PullRequestDTO `json:"pr"`
	ReplacedBy string         `json:"replaced_by"`
}

type UserReviewsResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

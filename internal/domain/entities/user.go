package entities

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

func NewUser(id, username, teamName string, isActive bool) *User {
	return &User{
		ID:       id,
		Username: username,
		TeamName: teamName,
		IsActive: isActive,
	}
}

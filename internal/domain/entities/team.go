package entities

type Team struct {
	ID    int64
	Name  string
	Users []*User
}

func NewTeam(id int64, name string, users []*User) *Team {
	return &Team{
		ID:    id,
		Name:  name,
		Users: users,
	}
}

func (t *Team) GetActiveUsers() []*User {
	active := make([]*User, 0, len(t.Users))
	for _, u := range t.Users {
		if u.IsActive {
			active = append(active, u)
		}
	}
	return active
}

func (t *Team) GetActiveUsersExclusive(ids []string) []*User {
	exclMap := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := exclMap[id]; !ok {
			exclMap[id] = struct{}{}
		}
	}

	active := make([]*User, 0, len(t.Users))
	for _, u := range t.Users {
		if _, ok := exclMap[u.ID]; u.IsActive && !ok {
			active = append(active, u)
		}
	}
	return active
}

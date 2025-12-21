package entities

// User represents a user entity
type User struct {
	ID    string
	Name  string
	Email string
}

// NewUser creates a new user
func NewUser(id, name, email string) *User {
	return &User{
		ID:    id,
		Name:  name,
		Email: email,
	}
}
package commands

// Command represents a user write operation
type Command interface {
	CommandType() string
}

// CreateUserCommand represents a command to create a new user
type CreateUserCommand struct {
	ID    string
	Name  string
	Email string
}

// CommandType returns the command type
func (c CreateUserCommand) CommandType() string {
	return "CreateUser"
}

// UpdateUserCommand represents a command to update an existing user
type UpdateUserCommand struct {
	ID    string
	Name  string
	Email string
}

// CommandType returns the command type
func (c UpdateUserCommand) CommandType() string {
	return "UpdateUser"
}

// DeleteUserCommand represents a command to delete a user
type DeleteUserCommand struct {
	ID string
}

// CommandType returns the command type
func (c DeleteUserCommand) CommandType() string {
	return "DeleteUser"
}
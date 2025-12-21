package commands

// Command represents a project write operation
type Command interface {
	CommandType() string
}

// CreateProjectCommand represents a command to create a new project
type CreateProjectCommand struct {
	ID          string
	Name        string
	Description string
	UserID      string
}

// CommandType returns the command type
func (c CreateProjectCommand) CommandType() string {
	return "CreateProject"
}

// UpdateProjectCommand represents a command to update an existing project
type UpdateProjectCommand struct {
	ID          string
	Name        string
	Description string
}

// CommandType returns the command type
func (c UpdateProjectCommand) CommandType() string {
	return "UpdateProject"
}

// DeleteProjectCommand represents a command to delete a project
type DeleteProjectCommand struct {
	ID string
}

// CommandType returns the command type
func (c DeleteProjectCommand) CommandType() string {
	return "DeleteProject"
}
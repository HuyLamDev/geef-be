package entities

// Project represents a project entity
type Project struct {
	ID          string
	Name        string
	Description string
}

// NewProject creates a new project
func NewProject(id, name, description string) *Project {
	return &Project{
		ID:          id,
		Name:        name,
		Description: description,
	}
}
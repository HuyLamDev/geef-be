package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/project/domain/commands"
	"geef-be/internal/project/domain/entities"
	projectinfrastructure "geef-be/internal/project/infrastructure"
)

// CreateProjectCommandHandler handles CreateProjectCommand
type CreateProjectCommandHandler struct {
	repo projectinfrastructure.ProjectRepository
}

// NewCreateProjectCommandHandler creates a new CreateProjectCommandHandler
func NewCreateProjectCommandHandler(repo projectinfrastructure.ProjectRepository) *CreateProjectCommandHandler {
	return &CreateProjectCommandHandler{repo: repo}
}

// Handle processes the CreateProjectCommand
func (h *CreateProjectCommandHandler) Handle(cmd commands.CreateProjectCommand) error {
	// Validate command
	if cmd.ID == "" {
		return errors.New("project ID cannot be empty")
	}
	if cmd.Name == "" {
		return errors.New("project name cannot be empty")
	}
	if cmd.UserID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Check if project already exists
	existingProject, err := h.repo.FindProjectByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to check existing project: %w", err)
	}
	if existingProject != nil {
		return errors.New("project with this ID already exists")
	}

	// Create domain entity
	project := entities.NewProject(cmd.ID, cmd.Name, cmd.Description)

	// Apply business rules
	if err := h.validateProject(project); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	// Save to repository
	return h.repo.SaveProject(project)
}

// validateProject applies business rules to the project entity
func (h *CreateProjectCommandHandler) validateProject(project *entities.Project) error {
	// Business rules:
	// - Name should be at least 3 characters
	if len(project.Name) < 3 {
		return errors.New("project name must be at least 3 characters")
	}

	// - Description should not exceed 500 characters
	if len(project.Description) > 500 {
		return errors.New("project description must not exceed 500 characters")
	}

	return nil
}
package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/project/domain/commands"
	"geef-be/internal/project/domain/entities"
	projectinfrastructure "geef-be/internal/project/infrastructure"
)

// UpdateProjectCommandHandler handles UpdateProjectCommand
type UpdateProjectCommandHandler struct {
	repo projectinfrastructure.ProjectRepository
}

// NewUpdateProjectCommandHandler creates a new UpdateProjectCommandHandler
func NewUpdateProjectCommandHandler(repo projectinfrastructure.ProjectRepository) *UpdateProjectCommandHandler {
	return &UpdateProjectCommandHandler{repo: repo}
}

// Handle processes the UpdateProjectCommand
func (h *UpdateProjectCommandHandler) Handle(cmd commands.UpdateProjectCommand) error {
	// Validate command
	if cmd.ID == "" {
		return errors.New("project ID cannot be empty")
	}

	// Check if project exists
	existingProject, err := h.repo.FindProjectByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}
	if existingProject == nil {
		return errors.New("project not found")
	}

	// Create updated project entity
	updatedProject := entities.NewProject(cmd.ID, cmd.Name, cmd.Description)

	// Apply business rules
	if err := h.validateProjectUpdate(updatedProject); err != nil {
		return fmt.Errorf("project update validation failed: %w", err)
	}

	// Save updated project
	return h.repo.SaveProject(updatedProject)
}

// validateProjectUpdate applies business rules for project updates
func (h *UpdateProjectCommandHandler) validateProjectUpdate(project *entities.Project) error {
	// Same validation as create, plus any update-specific rules
	if len(project.Name) < 3 {
		return errors.New("project name must be at least 3 characters")
	}

	if len(project.Description) > 500 {
		return errors.New("project description must not exceed 500 characters")
	}

	return nil
}
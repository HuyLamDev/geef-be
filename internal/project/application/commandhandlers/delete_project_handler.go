package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/project/domain/commands"
	projectinfrastructure "geef-be/internal/project/infrastructure"
)

// DeleteProjectCommandHandler handles DeleteProjectCommand
type DeleteProjectCommandHandler struct {
	repo projectinfrastructure.ProjectRepository
}

// NewDeleteProjectCommandHandler creates a new DeleteProjectCommandHandler
func NewDeleteProjectCommandHandler(repo projectinfrastructure.ProjectRepository) *DeleteProjectCommandHandler {
	return &DeleteProjectCommandHandler{repo: repo}
}

// Handle processes the DeleteProjectCommand
func (h *DeleteProjectCommandHandler) Handle(cmd commands.DeleteProjectCommand) error {
	// Validate command
	if cmd.ID == "" {
		return errors.New("project ID cannot be empty")
	}

	// Check if project exists before deletion
	existingProject, err := h.repo.FindProjectByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to find project: %w", err)
	}
	if existingProject == nil {
		return errors.New("project not found")
	}

	// Apply business rules for deletion
	if err := h.validateProjectDeletion(cmd.ID); err != nil {
		return fmt.Errorf("project deletion validation failed: %w", err)
	}

	// Note: In a real implementation, you'd have a DeleteProject method in the repository
	// For now, we'll simulate deletion
	// return h.repo.DeleteProject(cmd.ID)

	// Placeholder: In real implementation, repository would have DeleteProject method
	return nil
}

// validateProjectDeletion applies business rules before project deletion
func (h *DeleteProjectCommandHandler) validateProjectDeletion(projectID string) error {
	// Business rules for deletion:
	// - Check if project has active tasks/issues (would need additional queries)
	// - Check if user has permission to delete
	// - Check retention policies
	// - etc.

	// Example: Prevent deletion of system projects
	if projectID == "default" || projectID == "system" {
		return errors.New("system projects cannot be deleted")
	}

	return nil
}
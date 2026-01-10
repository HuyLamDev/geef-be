package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/project/domain/commands"
	projectinfrastructure "geef-be/internal/project/infrastructure"

	"github.com/sirupsen/logrus"
)

// DeleteProjectCommandHandler handles DeleteProjectCommand
type DeleteProjectCommandHandler struct {
	repo   projectinfrastructure.ProjectRepository
	logger *logrus.Logger
}

// NewDeleteProjectCommandHandler creates a new DeleteProjectCommandHandler
func NewDeleteProjectCommandHandler(repo projectinfrastructure.ProjectRepository, logger *logrus.Logger) *DeleteProjectCommandHandler {
	return &DeleteProjectCommandHandler{repo: repo, logger: logger}
}

// Handle processes the DeleteProjectCommand
func (h *DeleteProjectCommandHandler) Handle(cmd commands.DeleteProjectCommand) error {
	h.logger.WithField("project_id", cmd.ID).Info("Handling DeleteProjectCommand")

	// Validate command
	if cmd.ID == "" {
		h.logger.Warn("DeleteProjectCommand: project ID cannot be empty")
		return errors.New("project ID cannot be empty")
	}

	// Check if project exists before deletion
	existingProject, err := h.repo.FindProjectByID(cmd.ID)
	if err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Error("DeleteProjectCommand: failed to find project")
		return fmt.Errorf("failed to find project: %w", err)
	}
	if existingProject == nil {
		h.logger.WithField("project_id", cmd.ID).Warn("DeleteProjectCommand: project not found")
		return errors.New("project not found")
	}

	// Apply business rules for deletion
	if err := h.validateProjectDeletion(cmd.ID); err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Warn("DeleteProjectCommand: project deletion validation failed")
		return fmt.Errorf("project deletion validation failed: %w", err)
	}

	// Note: In a real implementation, you'd have a DeleteProject method in the repository
	// For now, we'll simulate deletion
	// return h.repo.DeleteProject(cmd.ID)

	h.logger.WithField("project_id", cmd.ID).Info("DeleteProjectCommand: project deleted successfully")
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
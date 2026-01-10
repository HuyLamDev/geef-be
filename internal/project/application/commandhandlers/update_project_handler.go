package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/project/domain/commands"
	"geef-be/internal/project/domain/entities"
	projectinfrastructure "geef-be/internal/project/infrastructure"

	"github.com/sirupsen/logrus"
)

// UpdateProjectCommandHandler handles UpdateProjectCommand
type UpdateProjectCommandHandler struct {
	repo   projectinfrastructure.ProjectRepository
	logger *logrus.Logger
}

// NewUpdateProjectCommandHandler creates a new UpdateProjectCommandHandler
func NewUpdateProjectCommandHandler(repo projectinfrastructure.ProjectRepository, logger *logrus.Logger) *UpdateProjectCommandHandler {
	return &UpdateProjectCommandHandler{repo: repo, logger: logger}
}

// Handle processes the UpdateProjectCommand
func (h *UpdateProjectCommandHandler) Handle(cmd commands.UpdateProjectCommand) error {
	h.logger.WithField("project_id", cmd.ID).Info("Handling UpdateProjectCommand")

	// Validate command
	if cmd.ID == "" {
		h.logger.Warn("UpdateProjectCommand: project ID cannot be empty")
		return errors.New("project ID cannot be empty")
	}

	// Check if project exists
	existingProject, err := h.repo.FindProjectByID(cmd.ID)
	if err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Error("UpdateProjectCommand: failed to find project")
		return fmt.Errorf("failed to find project: %w", err)
	}
	if existingProject == nil {
		h.logger.WithField("project_id", cmd.ID).Warn("UpdateProjectCommand: project not found")
		return errors.New("project not found")
	}

	// Create updated project entity
	updatedProject := entities.NewProject(cmd.ID, cmd.Name, cmd.Description)

	// Apply business rules
	if err := h.validateProjectUpdate(updatedProject); err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Warn("UpdateProjectCommand: project update validation failed")
		return fmt.Errorf("project update validation failed: %w", err)
	}

	// Save updated project
	if err := h.repo.SaveProject(updatedProject); err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Error("UpdateProjectCommand: failed to save updated project")
		return err
	}

	h.logger.WithField("project_id", cmd.ID).Info("UpdateProjectCommand: project updated successfully")
	return nil
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
package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/project/domain/commands"
	"geef-be/internal/project/domain/entities"
	projectinfrastructure "geef-be/internal/project/infrastructure"

	"github.com/sirupsen/logrus"
)

// CreateProjectCommandHandler handles CreateProjectCommand
type CreateProjectCommandHandler struct {
	repo   projectinfrastructure.ProjectRepository
	logger *logrus.Logger
}

// NewCreateProjectCommandHandler creates a new CreateProjectCommandHandler
func NewCreateProjectCommandHandler(repo projectinfrastructure.ProjectRepository, logger *logrus.Logger) *CreateProjectCommandHandler {
	return &CreateProjectCommandHandler{repo: repo, logger: logger}
}

// Handle processes the CreateProjectCommand
func (h *CreateProjectCommandHandler) Handle(cmd commands.CreateProjectCommand) error {
	h.logger.WithField("project_id", cmd.ID).Info("Handling CreateProjectCommand")

	// Validate command
	if cmd.ID == "" {
		h.logger.Warn("CreateProjectCommand: project ID cannot be empty")
		return errors.New("project ID cannot be empty")
	}
	if cmd.Name == "" {
		h.logger.Warn("CreateProjectCommand: project name cannot be empty")
		return errors.New("project name cannot be empty")
	}
	if cmd.UserID == "" {
		h.logger.Warn("CreateProjectCommand: user ID cannot be empty")
		return errors.New("user ID cannot be empty")
	}

	// Check if project already exists
	existingProject, err := h.repo.FindProjectByID(cmd.ID)
	if err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Error("CreateProjectCommand: failed to check existing project")
		return fmt.Errorf("failed to check existing project: %w", err)
	}
	if existingProject != nil {
		h.logger.WithField("project_id", cmd.ID).Warn("CreateProjectCommand: project with this ID already exists")
		return errors.New("project with this ID already exists")
	}

	// Create domain entity
	project := entities.NewProject(cmd.ID, cmd.Name, cmd.Description)

	// Apply business rules
	if err := h.validateProject(project); err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Warn("CreateProjectCommand: project validation failed")
		return fmt.Errorf("project validation failed: %w", err)
	}

	// Save to repository
	if err := h.repo.SaveProject(project); err != nil {
		h.logger.WithError(err).WithField("project_id", cmd.ID).Error("CreateProjectCommand: failed to save project")
		return err
	}

	h.logger.WithField("project_id", cmd.ID).Info("CreateProjectCommand: project created successfully")
	return nil
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
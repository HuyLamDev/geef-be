package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/user/domain/commands"
	"geef-be/internal/user/domain/entities"
	userinfrastructure "geef-be/internal/user/infrastructure"

	"github.com/sirupsen/logrus"
)

// CreateUserCommandHandler handles CreateUserCommand
type CreateUserCommandHandler struct {
	repo   userinfrastructure.UserRepository
	logger *logrus.Logger
}

// NewCreateUserCommandHandler creates a new CreateUserCommandHandler
func NewCreateUserCommandHandler(repo userinfrastructure.UserRepository, logger *logrus.Logger) *CreateUserCommandHandler {
	return &CreateUserCommandHandler{repo: repo, logger: logger}
}

// Handle processes the CreateUserCommand
func (h *CreateUserCommandHandler) Handle(cmd commands.CreateUserCommand) error {
	h.logger.WithField("user_id", cmd.ID).Info("Handling CreateUserCommand")

	// Validate command
	if cmd.ID == "" {
		h.logger.Warn("CreateUserCommand: user ID cannot be empty")
		return errors.New("user ID cannot be empty")
	}
	if cmd.Name == "" {
		h.logger.Warn("CreateUserCommand: user name cannot be empty")
		return errors.New("user name cannot be empty")
	}
	if cmd.Email == "" {
		h.logger.Warn("CreateUserCommand: user email cannot be empty")
		return errors.New("user email cannot be empty")
	}

	// Check if user already exists (business rule)
	existingUser, err := h.repo.FindUserByID(cmd.ID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", cmd.ID).Error("CreateUserCommand: failed to check existing user")
		return fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		h.logger.WithField("user_id", cmd.ID).Warn("CreateUserCommand: user with this ID already exists")
		return errors.New("user with this ID already exists")
	}

	// Create domain entity
	user := entities.NewUser(cmd.ID, cmd.Name, cmd.Email)

	// Apply business rules (e.g., validate email format, etc.)
	if err := h.validateUser(user); err != nil {
		h.logger.WithError(err).WithField("user_id", cmd.ID).Warn("CreateUserCommand: user validation failed")
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Save to repository
	if err := h.repo.SaveUser(user); err != nil {
		h.logger.WithError(err).WithField("user_id", cmd.ID).Error("CreateUserCommand: failed to save user")
		return err
	}

	h.logger.WithField("user_id", cmd.ID).Info("CreateUserCommand: user created successfully")
	return nil
}

// validateUser applies business rules to the user entity
func (h *CreateUserCommandHandler) validateUser(user *entities.User) error {
	// Example business rules:
	// - Email should contain @
	if !contains(user.Email, "@") {
		return errors.New("invalid email format")
	}

	// - Name should be at least 2 characters
	if len(user.Name) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	return nil
}

// contains is a helper function (in real code, use strings.Contains)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/user/domain/commands"
	"geef-be/internal/user/domain/entities"
	userinfrastructure "geef-be/internal/user/infrastructure"

	"github.com/sirupsen/logrus"
)

// UpdateUserCommandHandler handles UpdateUserCommand
type UpdateUserCommandHandler struct {
	repo   userinfrastructure.UserRepository
	logger *logrus.Logger
}

// NewUpdateUserCommandHandler creates a new UpdateUserCommandHandler
func NewUpdateUserCommandHandler(repo userinfrastructure.UserRepository, logger *logrus.Logger) *UpdateUserCommandHandler {
	return &UpdateUserCommandHandler{repo: repo, logger: logger}
}

// Handle processes the UpdateUserCommand
func (h *UpdateUserCommandHandler) Handle(cmd commands.UpdateUserCommand) error {
	// Validate command
	if cmd.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Check if user exists
	existingUser, err := h.repo.FindUserByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// Create updated user entity
	updatedUser := entities.NewUser(cmd.ID, cmd.Name, cmd.Email)

	// Apply business rules
	if err := h.validateUserUpdate(updatedUser); err != nil {
		return fmt.Errorf("user update validation failed: %w", err)
	}

	// Save updated user
	return h.repo.SaveUser(updatedUser)
}

// validateUserUpdate applies business rules for user updates
func (h *UpdateUserCommandHandler) validateUserUpdate(user *entities.User) error {
	// Same validation as create, plus any update-specific rules
	if !contains(user.Email, "@") {
		return errors.New("invalid email format")
	}

	if len(user.Name) < 2 {
		return errors.New("name must be at least 2 characters")
	}

	// Additional update rules could go here
	// e.g., check if email is already taken by another user

	return nil
}
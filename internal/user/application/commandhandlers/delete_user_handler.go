package commandhandlers

import (
	"errors"
	"fmt"

	"geef-be/internal/user/domain/commands"
	userinfrastructure "geef-be/internal/user/infrastructure"
)

// DeleteUserCommandHandler handles DeleteUserCommand
type DeleteUserCommandHandler struct {
	repo userinfrastructure.UserRepository
}

// NewDeleteUserCommandHandler creates a new DeleteUserCommandHandler
func NewDeleteUserCommandHandler(repo userinfrastructure.UserRepository) *DeleteUserCommandHandler {
	return &DeleteUserCommandHandler{repo: repo}
}

// Handle processes the DeleteUserCommand
func (h *DeleteUserCommandHandler) Handle(cmd commands.DeleteUserCommand) error {
	// Validate command
	if cmd.ID == "" {
		return errors.New("user ID cannot be empty")
	}

	// Check if user exists before deletion
	existingUser, err := h.repo.FindUserByID(cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// Apply business rules for deletion
	if err := h.validateUserDeletion(cmd.ID); err != nil {
		return fmt.Errorf("user deletion validation failed: %w", err)
	}

	// Note: In a real implementation, you'd have a DeleteUser method in the repository
	// For now, we'll simulate deletion by "saving" a nil or marked-as-deleted user
	// return h.repo.DeleteUser(cmd.ID)

	// Placeholder: In real implementation, repository would have DeleteUser method
	return nil
}

// validateUserDeletion applies business rules before user deletion
func (h *DeleteUserCommandHandler) validateUserDeletion(userID string) error {
	// Business rules for deletion:
	// - Check if user has active projects/sessions (would need additional queries)
	// - Check if user is admin/superuser
	// - Check retention policies
	// - etc.

	// Example: Prevent deletion of system users
	if userID == "admin" || userID == "system" {
		return errors.New("system users cannot be deleted")
	}

	return nil
}
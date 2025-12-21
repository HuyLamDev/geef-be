package user

import (
	"geef-be/internal/infrastructure"
	"geef-be/internal/user/application/commandhandlers"
	userinfrastructure "geef-be/internal/user/infrastructure"
)

// UserHandlers contains all user command and query handlers
type UserHandlers struct {
	Commands UserCommandHandlers
	Queries  UserQueryHandlers
}

// UserCommandHandlers contains all user command handlers
type UserCommandHandlers struct {
	CreateUserHandler *commandhandlers.CreateUserCommandHandler
	UpdateUserHandler *commandhandlers.UpdateUserCommandHandler
	DeleteUserHandler *commandhandlers.DeleteUserCommandHandler
}

// UserQueryHandlers contains all user query handlers (placeholder for now)
type UserQueryHandlers struct {
	// Add query handlers here when implemented
	// GetUserHandler *queryhandlers.GetUserQueryHandler
}

// NewUserModule initializes the user module with handlers
func NewUserModule(config *infrastructure.Config) *UserHandlers {
	// Initialize user-specific infrastructure
	userRepo := userinfrastructure.NewUserRepository(config.DB, config.Logger)

	// Initialize command handlers
	createHandler := commandhandlers.NewCreateUserCommandHandler(userRepo)
	updateHandler := commandhandlers.NewUpdateUserCommandHandler(userRepo)
	deleteHandler := commandhandlers.NewDeleteUserCommandHandler(userRepo)

	config.Logger.Info("User module initialized")

	return &UserHandlers{
		Commands: UserCommandHandlers{
			CreateUserHandler: createHandler,
			UpdateUserHandler: updateHandler,
			DeleteUserHandler: deleteHandler,
		},
		Queries: UserQueryHandlers{
			// Initialize query handlers here when implemented
		},
	}
}
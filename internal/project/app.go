package project

import (
	"geef-be/internal/infrastructure"
	"geef-be/internal/project/application/commandhandlers"
	projectinfrastructure "geef-be/internal/project/infrastructure"
)

// ProjectHandlers contains all project command and query handlers
type ProjectHandlers struct {
	Commands ProjectCommandHandlers
	Queries  ProjectQueryHandlers
}

// ProjectCommandHandlers contains all project command handlers
type ProjectCommandHandlers struct {
	CreateProjectHandler *commandhandlers.CreateProjectCommandHandler
	UpdateProjectHandler *commandhandlers.UpdateProjectCommandHandler
	DeleteProjectHandler *commandhandlers.DeleteProjectCommandHandler
}

// ProjectQueryHandlers contains all project query handlers (placeholder for now)
type ProjectQueryHandlers struct {
	// Add query handlers here when implemented
	// GetProjectHandler *queryhandlers.GetProjectQueryHandler
}

// NewProjectModule initializes the project module with handlers
func NewProjectModule(config *infrastructure.Config) *ProjectHandlers {
	// Initialize project-specific infrastructure
	projectRepo := projectinfrastructure.NewProjectRepository(config.DB, config.Logger)

	// Initialize command handlers
	createHandler := commandhandlers.NewCreateProjectCommandHandler(projectRepo)
	updateHandler := commandhandlers.NewUpdateProjectCommandHandler(projectRepo)
	deleteHandler := commandhandlers.NewDeleteProjectCommandHandler(projectRepo)

	config.Logger.Info("Project module initialized")

	return &ProjectHandlers{
		Commands: ProjectCommandHandlers{
			CreateProjectHandler: createHandler,
			UpdateProjectHandler: updateHandler,
			DeleteProjectHandler: deleteHandler,
		},
		Queries: ProjectQueryHandlers{
			// Initialize query handlers here when implemented
		},
	}
}

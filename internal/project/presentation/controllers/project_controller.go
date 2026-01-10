package controllers

import (
	"encoding/json"
	"geef-be/internal/project"
	"geef-be/internal/project/application/commandhandlers"
	"geef-be/internal/project/domain/commands"
	"net/http"

	"github.com/sirupsen/logrus"
)

// ProjectController handles project HTTP requests
type ProjectController struct {
	createProjectHandler *commandhandlers.CreateProjectCommandHandler
	updateProjectHandler *commandhandlers.UpdateProjectCommandHandler
	deleteProjectHandler *commandhandlers.DeleteProjectCommandHandler
	logger               *logrus.Logger
}

// NewProjectController creates a new project controller
func NewProjectController(handlers *project.ProjectHandlers, logger *logrus.Logger) *ProjectController {
	return &ProjectController{
		createProjectHandler: handlers.Commands.CreateProjectHandler,
		updateProjectHandler: handlers.Commands.UpdateProjectHandler,
		deleteProjectHandler: handlers.Commands.DeleteProjectHandler,
		logger:               logger,
	}
}

// RegisterRoutes registers project routes on the mux
func (c *ProjectController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/projects/health", c.healthCheck)
	mux.HandleFunc("/projects", c.handleProjects)
	mux.HandleFunc("/projects/", c.handleProjectByID)
}

func (c *ProjectController) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Project service OK"))
}

func (c *ProjectController) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.createProject(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *ProjectController) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.URL.Path[len("/projects/"):]
	if id == "" {
		http.Error(w, "Project ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		c.updateProject(w, r, id)
	case http.MethodDelete:
		c.deleteProject(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *ProjectController) createProject(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		UserID      string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.WithError(err).Warn("createProject: invalid JSON in request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := commands.CreateProjectCommand{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		UserID:      req.UserID,
	}

	if err := c.createProjectHandler.Handle(cmd); err != nil {
		c.logger.WithError(err).WithField("project_id", req.ID).Error("createProject: failed to create project")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.logger.WithField("project_id", req.ID).Info("createProject: project created successfully")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "project created"})
}

func (c *ProjectController) updateProject(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.WithError(err).Warn("updateProject: invalid JSON in request body")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := commands.UpdateProjectCommand{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.updateProjectHandler.Handle(cmd); err != nil {
		c.logger.WithError(err).WithField("project_id", id).Error("updateProject: failed to update project")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.logger.WithField("project_id", id).Info("updateProject: project updated successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "project updated"})
}

func (c *ProjectController) deleteProject(w http.ResponseWriter, r *http.Request, id string) {
	cmd := commands.DeleteProjectCommand{ID: id}

	if err := c.deleteProjectHandler.Handle(cmd); err != nil {
		c.logger.WithError(err).WithField("project_id", id).Error("deleteProject: failed to delete project")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.logger.WithField("project_id", id).Info("deleteProject: project deleted successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "project deleted"})
}
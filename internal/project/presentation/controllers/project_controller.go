package controllers

import (
	"encoding/json"
	"geef-be/internal/project"
	"geef-be/internal/project/application/commandhandlers"
	"geef-be/internal/project/domain/commands"
	"net/http"
)

// ProjectController handles project HTTP requests
type ProjectController struct {
	createProjectHandler *commandhandlers.CreateProjectCommandHandler
	updateProjectHandler *commandhandlers.UpdateProjectCommandHandler
	deleteProjectHandler *commandhandlers.DeleteProjectCommandHandler
}

// NewProjectController creates a new project controller
func NewProjectController(handlers *project.ProjectHandlers) *ProjectController {
	return &ProjectController{
		createProjectHandler: handlers.Commands.CreateProjectHandler,
		updateProjectHandler: handlers.Commands.UpdateProjectHandler,
		deleteProjectHandler: handlers.Commands.DeleteProjectHandler,
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "project created"})
}

func (c *ProjectController) updateProject(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := commands.UpdateProjectCommand{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.updateProjectHandler.Handle(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "project updated"})
}

func (c *ProjectController) deleteProject(w http.ResponseWriter, r *http.Request, id string) {
	cmd := commands.DeleteProjectCommand{ID: id}

	if err := c.deleteProjectHandler.Handle(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "project deleted"})
}
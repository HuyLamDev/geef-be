package controllers

import (
	"encoding/json"
	"geef-be/internal/user"
	"geef-be/internal/user/application/commandhandlers"
	"geef-be/internal/user/domain/commands"
	"net/http"
)

// UserController handles user HTTP requests
type UserController struct {
	createUserHandler *commandhandlers.CreateUserCommandHandler
	updateUserHandler *commandhandlers.UpdateUserCommandHandler
	deleteUserHandler *commandhandlers.DeleteUserCommandHandler
}

// NewUserController creates a new user controller
func NewUserController(handlers *user.UserHandlers) *UserController {
	return &UserController{
		createUserHandler: handlers.Commands.CreateUserHandler,
		updateUserHandler: handlers.Commands.UpdateUserHandler,
		deleteUserHandler: handlers.Commands.DeleteUserHandler,
	}
}

// RegisterRoutes registers user routes on the mux
func (c *UserController) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/users/health", c.healthCheck)
	mux.HandleFunc("/users", c.handleUsers)
	mux.HandleFunc("/users/", c.handleUserByID)
}

func (c *UserController) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User service OK"))
}

func (c *UserController) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		c.createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *UserController) handleUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.URL.Path[len("/users/"):]
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		c.updateUser(w, r, id)
	case http.MethodDelete:
		c.deleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *UserController) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := commands.CreateUserCommand{
		ID:    req.ID,
		Name:  req.Name,
		Email: req.Email,
	}

	if err := c.createUserHandler.Handle(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "user created"})
}

func (c *UserController) updateUser(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cmd := commands.UpdateUserCommand{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}

	if err := c.updateUserHandler.Handle(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "user updated"})
}

func (c *UserController) deleteUser(w http.ResponseWriter, r *http.Request, id string) {
	cmd := commands.DeleteUserCommand{ID: id}

	if err := c.deleteUserHandler.Handle(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "user deleted"})
}
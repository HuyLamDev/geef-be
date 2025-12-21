// Package main Geef Backend API
//
//	@title			Geef Backend API
//	@version		1.0
//	@description	A DDD CQRS backend API for Geef
//	@host			localhost:8080
//	@BasePath		/
package main

import (
	_ "geef-be/docs"
	"geef-be/internal/infrastructure"
	"geef-be/internal/project"
	"geef-be/internal/project/presentation/controllers"
	"geef-be/internal/user"
	userControllers "geef-be/internal/user/presentation/controllers"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize infrastructure config
	config := infrastructure.NewConfig()

	// Wire and initialize user module
	userHandlers := user.NewUserModule(config)

	// Wire and initialize project module
	projectHandlers := project.NewProjectModule(config)

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Initialize controllers for each module
	userController := userControllers.NewUserController(userHandlers)
	projectController := controllers.NewProjectController(projectHandlers)

	// Register routes for each module
	userController.RegisterRoutes(mux)
	projectController.RegisterRoutes(mux)

	// Register common routes
	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Start the server
	config.Logger.Info("Starting Geef Backend API on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		config.Logger.WithError(err).Fatal("Server failed to start")
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API Gateway OK"))
}
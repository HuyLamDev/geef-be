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
	authCfg := infrastructure.NewAuthConfig()

	// Wire and initialize user module
	userHandlers := user.NewUserModule(config)

	// Wire and initialize project module
	projectHandlers := project.NewProjectModule(config)

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Initialize controllers for each module
	userController := userControllers.NewUserController(userHandlers)
	oauthController := userControllers.NewOAuthController(authCfg, config.Logger)
	projectController := controllers.NewProjectController(projectHandlers, config.Logger)

	// Register routes for each module
	userController.RegisterRoutes(mux)
	oauthController.RegisterRoutes(mux)
	projectController.RegisterRoutes(mux)

	// Register common routes
	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Wrap mux with CORS handler
	frontendOrigin := authCfg.FrontendOrigin
	handler := corsMiddleware(mux, frontendOrigin)

	// Start the server
	config.Logger.Info("Starting Geef Backend API on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		config.Logger.WithError(err).Fatal("Server failed to start")
	}
}

// Simple CORS middleware -- in dev it allows the configured FRONTEND_ORIGIN or '*'
func corsMiddleware(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigin == "*" || allowedOrigin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin == allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, traceparent, tracestate, baggage")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API Gateway OK"))
}
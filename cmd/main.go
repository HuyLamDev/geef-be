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
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
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
	oauthController := userControllers.NewOAuthController(config.DB, config.Logger, authCfg)
	projectController := controllers.NewProjectController(projectHandlers, config.Logger)

	// Register routes for each module
	userController.RegisterRoutes(mux)
	oauthController.RegisterRoutes(mux)
	projectController.RegisterRoutes(mux)

	// Register common routes
	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Wrap mux with logging and CORS handlers
	frontendOrigins := authCfg.FrontendOrigins
	fmt.Printf("Main: Frontend origins: %v\n", frontendOrigins)
	handler := loggingMiddleware(mux, config.Logger)
	handler = corsMiddleware(handler, frontendOrigins)

	// Start the server
	config.Logger.Info("Starting Geef Backend API on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		config.Logger.WithError(err).Fatal("Server failed to start")
	}
}

// Simple CORS middleware -- allows configured origins
func corsMiddleware(next http.Handler, allowedOrigins []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimSuffix(r.Header.Get("Origin"), "/")
		fmt.Printf("CORS: Request from origin: %s\n", origin)
		fmt.Printf("CORS: Allowed origins: %v\n", allowedOrigins)

		// Check if origin is allowed and set headers
		originAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			fmt.Printf("CORS: Comparing '%s' == '%s'\n", origin, allowedOrigin)
			if allowedOrigin == "*" || allowedOrigin == "" || origin == allowedOrigin {
				originAllowed = true
				if allowedOrigin != "*" && allowedOrigin != "" {
					w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
					fmt.Printf("CORS: Allowed origin %s, setting header\n", allowedOrigin)
				} else {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					fmt.Printf("CORS: Allowing all origins (*)\n")
				}
				break
			}
		}

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}

		if !originAllowed {
			fmt.Printf("CORS: Origin %s not allowed\n", origin)
		}

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// Logging middleware logs HTTP request details: URL, method, duration
func loggingMiddleware(next http.Handler, logger *logrus.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the request details
		logger.WithFields(logrus.Fields{
			"method":     r.Method,
			"url":        r.URL.Path,
			"query":      r.URL.RawQuery,
			"user_agent": r.Header.Get("User-Agent"),
			"remote_ip":  r.RemoteAddr,
			"status":     wrapped.statusCode,
			"duration":   duration.String(),
			"duration_ms": duration.Milliseconds(),
		}).Info("HTTP Request")
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("API Gateway OK"))
}
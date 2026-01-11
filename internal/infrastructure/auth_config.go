package infrastructure

import (
	"fmt"
	"os"
	"strings"
)

// AuthConfig holds OAuth and JWT related configuration
type AuthConfig struct {
    GoogleClientID     string
    GoogleClientSecret string
    GoogleRedirectURI  string
    FrontendOrigins    []string
    JWTSecret          string
}

// NewAuthConfig reads auth-related environment variables
func NewAuthConfig() *AuthConfig {
    frontendOrigins := parseFrontendOrigins()
    return &AuthConfig{
        GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
        GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
        GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:8080/api/auth/google/callback"),
        FrontendOrigins:    frontendOrigins,
        JWTSecret:          getEnv("JWT_SECRET", os.Getenv("JWT_SECRET")),
    }
}

// parseFrontendOrigins parses frontend origins from environment variables
func parseFrontendOrigins() []string {
    // Check for comma-separated list first
    originsStr := getEnv("FRONTEND_ORIGINS", "")
    if originsStr != "" {
        origins := strings.Split(originsStr, ",")
        for i, origin := range origins {
            origins[i] = strings.TrimSuffix(strings.TrimSpace(origin), "/")
        }
        fmt.Printf("AuthConfig: Parsed FRONTEND_ORIGINS: %v\n", origins)
        return origins
    }

    // Fallback to single origin for backward compatibility
    origin := getEnv("FRONTEND_ORIGIN", "http://localhost:5173")
    origins := []string{strings.TrimSuffix(strings.TrimSpace(origin), "/")}
    fmt.Printf("AuthConfig: Using fallback FRONTEND_ORIGIN: %v\n", origins)
    return origins
}

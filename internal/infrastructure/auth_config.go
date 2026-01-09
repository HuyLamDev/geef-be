package infrastructure

import "os"

// AuthConfig holds OAuth and JWT related configuration
type AuthConfig struct {
    GoogleClientID     string
    GoogleClientSecret string
    GoogleRedirectURI  string
    FrontendOrigin     string
    JWTSecret          string
}

// NewAuthConfig reads auth-related environment variables
func NewAuthConfig() *AuthConfig {
    return &AuthConfig{
        GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
        GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
        GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:8080/api/auth/google/callback"),
        FrontendOrigin:     getEnv("FRONTEND_ORIGIN", "http://localhost:5173"),
        JWTSecret:          getEnv("JWT_SECRET", os.Getenv("JWT_SECRET")),
    }
}

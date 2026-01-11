package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"geef-be/internal/infrastructure"
	userinfra "geef-be/internal/user/infrastructure"

	"github.com/sirupsen/logrus"
)

type OAuthController struct{
    db      *sql.DB
    logger  *logrus.Logger
    authCfg *infrastructure.AuthConfig
}

func NewOAuthController(db *sql.DB, logger *logrus.Logger, authCfg *infrastructure.AuthConfig) *OAuthController {
    return &OAuthController{db: db, logger: logger, authCfg: authCfg}
}

// matchesWildcard checks if origin matches a wildcard pattern like https://*.vercel.app
func matchesWildcard(origin, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return false
	}

	// Handle https://*.domain.com pattern
	if strings.HasPrefix(pattern, "https://*.") {
		suffix := strings.TrimPrefix(pattern, "https://*.") // get domain.com
		expectedPrefix := "https://"
		expectedSuffix := "." + suffix
		return strings.HasPrefix(origin, expectedPrefix) && strings.HasSuffix(origin, expectedSuffix) && len(strings.TrimSuffix(strings.TrimPrefix(origin, expectedPrefix), expectedSuffix)) > 0
	}

	// Handle *.domain.com pattern
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[1:] // remove the *
		return strings.HasSuffix(origin, suffix) && !strings.Contains(strings.TrimSuffix(origin, suffix), ".")
	}

	return false
}

// Register OAuth routes
func (c *OAuthController) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/api/auth/google", c.HandleGoogleStart)
    mux.HandleFunc("/api/auth/google/callback", c.HandleGoogleCallback)
    mux.HandleFunc("/api/auth/me", c.HandleMe)
    mux.HandleFunc("/api/auth/logout", c.HandleLogout)
    mux.HandleFunc("/api/auth/google/url", c.HandleGoogleURL)
}

// Redirects user to Google's OAuth 2.0 consent screen
func (c *OAuthController) HandleGoogleStart(w http.ResponseWriter, r *http.Request) {
    params := url.Values{}
    params.Set("client_id", c.authCfg.GoogleClientID)
    params.Set("redirect_uri", c.authCfg.GoogleRedirectURI)
    params.Set("response_type", "code")
    params.Set("scope", "openid email profile")
    params.Set("access_type", "offline")
    params.Set("prompt", "consent")

    authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
    http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleGoogleURL returns the Google OAuth URL as JSON so the client can open it
func (c *OAuthController) HandleGoogleURL(w http.ResponseWriter, r *http.Request) {
    origin := strings.TrimSuffix(r.Header.Get("Origin"), "/")
    if origin == "" {
        // Fallback for requests without Origin header
        origin = c.authCfg.FrontendOrigins[0]
    }

    params := url.Values{}
    params.Set("client_id", c.authCfg.GoogleClientID)
    params.Set("redirect_uri", c.authCfg.GoogleRedirectURI)
    params.Set("response_type", "code")
    params.Set("scope", "openid email profile")
    params.Set("access_type", "offline")
    params.Set("prompt", "consent")
    params.Set("state", origin) // Include origin in state for callback redirect

    authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"url": authURL})
}

// Callback endpoint where Google will redirect after consent (placeholder)
func (c *OAuthController) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    q := r.URL.Query()
    code := q.Get("code")
    state := q.Get("state") // Get the origin from state
    if code == "" {
        c.logger.Warn("OAuth callback: missing authorization code")
        c.redirectWithError(w, r, "missing_code")
        return
    }

    // Validate state (origin) is allowed
    redirectOrigin := c.authCfg.FrontendOrigins[0] // default fallback
    if state != "" {
        for _, allowed := range c.authCfg.FrontendOrigins {
            if state == allowed || matchesWildcard(state, allowed) {
                redirectOrigin = state
                break
            }
        }
    }

    // Exchange code for tokens
    tokenResp, err := userinfra.ExchangeCode(c.authCfg, ctx, code)
    if err != nil {
        c.logger.WithError(err).Error("OAuth callback: token exchange failed")
        c.redirectWithError(w, r, "token_exchange_failed")
        return
    }

    var claims map[string]interface{}

    // If Google returned an ID token, verify it. Otherwise, fall back to the UserInfo endpoint
    if tokenResp.IdToken != "" {
        claims, err = userinfra.VerifyIDToken(c.authCfg, ctx, tokenResp.IdToken)
        if err != nil {
            c.logger.WithError(err).Error("OAuth callback: ID token verification failed")
            c.redirectWithError(w, r, "id_token_verification_failed")
            return
        }
    } else {
        // Call UserInfo endpoint with access token as fallback
        req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
        req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            c.logger.WithError(err).Error("OAuth callback: failed to fetch userinfo")
            c.redirectWithError(w, r, "userinfo_fetch_failed")
            return
        }
        defer resp.Body.Close()
        if resp.StatusCode != http.StatusOK {
            c.logger.WithField("status", resp.Status).Error("OAuth callback: userinfo endpoint returned non-OK status")
            c.redirectWithError(w, r, "userinfo_status_error")
            return
        }
        if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
            c.logger.WithError(err).Error("OAuth callback: failed to parse userinfo response")
            c.redirectWithError(w, r, "userinfo_parse_error")
            return
        }
    }

    // Map claims to user model (simple map for now)
    user := map[string]interface{}{
        "sub":        claims["sub"],
        "email":      claims["email"],
        "name":       claims["name"], // display name
        "provider":   "google",
        "avatar_url": claims["picture"],
        "full_name":  claims["name"], // Google provides display name as 'name'
    }

    // Persist or update user via repository
    repo := userinfra.NewUserRepository(c.db, c.logger)
    
    // Check if user already exists
    existingUser, err := repo.FindUserByID(user["sub"].(string))
    if err != nil {
        c.logger.WithError(err).WithField("user_id", user["sub"]).Error("OAuth callback: failed to check existing user")
        c.redirectWithError(w, r, "database_error")
        return
    }
    
    if existingUser == nil {
        // First time login - save new user
        c.logger.WithField("user_id", user["sub"]).Info("OAuth callback: first time login, saving new user")
        if err := repo.SaveUser(user); err != nil {
            c.logger.WithError(err).WithField("user_id", user["sub"]).Error("OAuth callback: failed to save new user")
            c.redirectWithError(w, r, "user_save_failed")
            return
        }
    } else {
        // Returning user - just log the login
        c.logger.WithField("user_id", user["sub"]).Info("OAuth callback: returning user login")
    }

    // Create a simple HMAC-signed session token (not a full JWT)
    jwtSecret := []byte(c.authCfg.JWTSecret)
    sessClaims := map[string]interface{}{
        "sub":   claims["sub"],
        "email": claims["email"],
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
    }
    payloadBytes, _ := json.Marshal(sessClaims)
    payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
    mac := hmac.New(sha256.New, jwtSecret)
    mac.Write([]byte(payload))
    sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
    signed := payload + "." + sig

    // Set cookie and redirect back to frontend
    cookie := &http.Cookie{
        Name:     "gfe_session",
        Value:    signed,
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        Expires:  time.Now().Add(24 * time.Hour),
    }
    http.SetCookie(w, cookie)

    // Redirect to frontend callback page (will handle popup closing)
    redirect := redirectOrigin + "/auth/callback"
    http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// HandleMe validates session cookie and returns basic user info
func (c *OAuthController) HandleMe(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("gfe_session")
    if err != nil {
        c.logger.Warn("HandleMe: no session cookie found")
        http.Error(w, "Not authenticated", http.StatusUnauthorized)
        return
    }

    parts := strings.Split(cookie.Value, ".")
    if len(parts) != 2 {
        c.logger.Warn("HandleMe: invalid session format")
        http.Error(w, "Invalid session", http.StatusUnauthorized)
        return
    }

    payloadB, err := base64.RawURLEncoding.DecodeString(parts[0])
    if err != nil {
        c.logger.WithError(err).Warn("HandleMe: invalid session payload")
        http.Error(w, "Invalid session payload", http.StatusUnauthorized)
        return
    }

    sig, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        c.logger.WithError(err).Warn("HandleMe: invalid session signature")
        http.Error(w, "Invalid session signature", http.StatusUnauthorized)
        return
    }

    mac := hmac.New(sha256.New, []byte(c.authCfg.JWTSecret))
    mac.Write([]byte(parts[0]))
    expected := mac.Sum(nil)
    if !hmac.Equal(sig, expected) {
        c.logger.Warn("HandleMe: session signature verification failed")
        http.Error(w, "Invalid session signature", http.StatusUnauthorized)
        return
    }

    var sessClaims map[string]interface{}
    if err := json.Unmarshal(payloadB, &sessClaims); err != nil {
        c.logger.WithError(err).Warn("HandleMe: unable to parse session claims")
        http.Error(w, "Unable to parse session", http.StatusUnauthorized)
        return
    }

    // Check expiration
    if expVal, ok := sessClaims["exp"].(float64); ok {
        if time.Now().Unix() > int64(expVal) {
            c.logger.Warn("HandleMe: session expired")
            http.Error(w, "Session expired", http.StatusUnauthorized)
            return
        }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "user": sessClaims,
    })
}

// HandleLogout clears the session cookie
func (c *OAuthController) HandleLogout(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        c.logger.WithField("method", r.Method).Warn("HandleLogout: invalid HTTP method")
        w.Header().Set("Allow", http.MethodPost)
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    cookie := &http.Cookie{
        Name:     "gfe_session",
        Value:    "",
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        Expires:  time.Unix(0, 0),
        MaxAge:   -1,
    }
    http.SetCookie(w, cookie)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// Helper method to redirect to frontend callback with error
func (c *OAuthController) redirectWithError(w http.ResponseWriter, r *http.Request, errorType string) {
    redirect := c.authCfg.FrontendOrigins[0] + "/auth/callback?error=" + errorType
    http.Redirect(w, r, redirect, http.StatusSeeOther)
}

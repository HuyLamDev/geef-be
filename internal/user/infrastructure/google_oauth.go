package userinfrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"geef-be/internal/infrastructure"
)

type GoogleTokenResponse struct {
    AccessToken  string `json:"access_token"`
    ExpiresIn    int    `json:"expires_in"`
    RefreshToken string `json:"refresh_token"`
    Scope        string `json:"scope"`
    TokenType    string `json:"token_type"`
    IdToken      string `json:"id_token"`
}

// ExchangeCode exchanges authorization code for tokens
func ExchangeCode(cfg *infrastructure.AuthConfig, ctx context.Context, code string) (*GoogleTokenResponse, error) {
    data := url.Values{}
    data.Set("code", code)
    data.Set("client_id", cfg.GoogleClientID)
    data.Set("client_secret", cfg.GoogleClientSecret)
    data.Set("redirect_uri", cfg.GoogleRedirectURI)
    data.Set("grant_type", "authorization_code")

    req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var tr GoogleTokenResponse
    if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
        return nil, err
    }

    return &tr, nil
}

// VerifyIDToken verifies the ID token's audience and signature using oauth2 package's helper
func VerifyIDToken(cfg *infrastructure.AuthConfig, ctx context.Context, idToken string) (map[string]interface{}, error) {
    // Use Google's tokeninfo endpoint for simplicity (note: not recommended for high-volume production)
    resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var claims map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
        return nil, err
    }

    // Simple audience check: accept string or array forms and provide a helpful error message
    audVal, audPresent := claims["aud"]
    if !audPresent {
        return nil, errors.New("id_token missing aud claim")
    }

    match := false
    switch v := audVal.(type) {
    case string:
        if v == cfg.GoogleClientID {
            match = true
        }
    case []interface{}:
        for _, item := range v {
            if s, ok := item.(string); ok && s == cfg.GoogleClientID {
                match = true
                break
            }
        }
    default:
        // unexpected type
    }

    if !match {
        // include the actual audience in the error to help debugging
        return nil, fmt.Errorf("invalid audience: got %v expected %s", audVal, cfg.GoogleClientID)
    }

    return claims, nil
}

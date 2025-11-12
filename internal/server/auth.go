//go:build !js && !wasm
// +build !js,!wasm

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoginRequest represents the login credentials
type LoginRequest struct {
	AccessKey   string `json:"accessKey"`
	SecretKey   string `json:"secretKey"`
	Environment string `json:"environment"` // staging, production, dev
}

// LoginResponse represents the login response
type LoginResponse struct {
	SessionToken string `json:"sessionToken"`
	ExpiresIn    int64  `json:"expiresIn"` // seconds
	Environment  string `json:"environment"`
}

// LogoutRequest represents the logout request
type LogoutRequest struct {
	SessionToken string `json:"sessionToken"`
}

// MegaportAuthResponse represents Megaport API token response
type MegaportAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// getTokenURL returns the appropriate token URL for the environment
func getTokenURL(environment string) string {
	switch environment {
	case "staging":
		return "https://auth-m2m-staging.megaport.com/oauth2/token"
	case "development":
		return "https://auth-m2m-mpone-dev.megaport.com/oauth2/token"
	default:
		return "https://auth-m2m.megaport.com/oauth2/token"
	}
}

// authenticateWithMegaport authenticates with Megaport API and returns a token
func authenticateWithMegaport(accessKey, secretKey, environment string) (*MegaportAuthResponse, error) {
	tokenURL := getTokenURL(environment)

	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(accessKey, secretKey)

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResp MegaportAuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &authResp, nil
}

// HandleLogin handles customer login requests
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var loginReq LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if loginReq.AccessKey == "" || loginReq.SecretKey == "" {
		http.Error(w, "Access key and secret key are required", http.StatusBadRequest)
		return
	}

	// Default to production if not specified
	if loginReq.Environment == "" {
		loginReq.Environment = "production"
	}

	s.logger.Printf("Login attempt for environment: %s", loginReq.Environment)

	// Authenticate with Megaport API
	authResp, err := authenticateWithMegaport(loginReq.AccessKey, loginReq.SecretKey, loginReq.Environment)
	if err != nil {
		s.logger.Printf("Authentication failed: %v", err)
		http.Error(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Create session
	session, err := s.sessionManager.CreateSession(loginReq.AccessKey, loginReq.SecretKey)
	if err != nil {
		s.logger.Printf("Failed to create session: %v", err)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Store Megaport token in session
	tokenExpiry := time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)
	s.sessionManager.UpdateMegaportToken(session.ID, authResp.AccessToken, tokenExpiry)

	s.logger.Printf("Login successful, session created: %s", session.ID)

	// Return response
	response := LoginResponse{
		SessionToken: session.ID,
		ExpiresIn:    int64(s.sessionManager.duration.Seconds()),
		Environment:  loginReq.Environment,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Printf("Error encoding login response: %v", err)
	}
}

// HandleLogout handles customer logout requests
func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session token from header or body
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		var logoutReq LogoutRequest
		if err := json.NewDecoder(r.Body).Decode(&logoutReq); err == nil {
			sessionToken = logoutReq.SessionToken
		}
	}

	if sessionToken == "" {
		http.Error(w, "Session token required", http.StatusBadRequest)
		return
	}

	// Delete session
	s.sessionManager.DeleteSession(sessionToken)
	s.logger.Printf("Logout successful for session: %s", sessionToken)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
		s.logger.Printf("Error writing logout response: %v", err)
	}
}

// HandleSessionCheck checks if a session is valid
func (s *Server) HandleSessionCheck(w http.ResponseWriter, r *http.Request) {
	sessionToken := r.Header.Get("X-Session-Token")
	if sessionToken == "" {
		http.Error(w, "Session token required", http.StatusBadRequest)
		return
	}

	session := s.sessionManager.GetSession(sessionToken)
	if session == nil {
		http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
		return
	}

	// Update activity
	s.sessionManager.UpdateActivity(sessionToken)

	// Return session info
	response := map[string]interface{}{
		"valid":     true,
		"expiresAt": session.ExpiresAt.Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Printf("Error encoding session check response: %v", err)
	}
}

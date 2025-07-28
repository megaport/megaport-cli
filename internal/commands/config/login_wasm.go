//go:build js && wasm
// +build js,wasm

package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall/js"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/wasm/api"
	megaport "github.com/megaport/megaportgo"
)

func Login(ctx context.Context) (*megaport.Client, error) {
	return LoginFunc(ctx)
}

// LoginFunc overrides the standard login for WASM environments
var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	var accessKey, secretKey, env string

	// Add console logging for debugging
	js.Global().Get("console").Call("group", "🔐 Megaport Authentication Debug")

	// Try to get credentials from local storage first
	manager, err := NewConfigManager()
	if err == nil {
		profile, profileName, err := manager.GetCurrentProfile()
		if err == nil {
			accessKey = profile.AccessKey
			secretKey = profile.SecretKey
			env = profile.Environment

			// Log credential source with masked values
			js.Global().Get("console").Call("log", "Using credentials from profile: "+profileName)
			js.Global().Get("console").Call("log", "Access Key: "+maskCredential(accessKey))
			js.Global().Get("console").Call("log", "Secret Key: [HIDDEN]")
			js.Global().Get("console").Call("log", "Environment: "+env)
		} else {
			js.Global().Get("console").Call("log", "No active profile found, error: "+err.Error())
		}
	} else {
		js.Global().Get("console").Call("log", "Failed to load config manager: "+err.Error())
	}

	// Fall back to environment variables
	if accessKey == "" {
		accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
		if accessKey != "" {
			js.Global().Get("console").Call("log", "Using access key from environment variable")
			js.Global().Get("console").Call("log", "Access Key: "+maskCredential(accessKey))
		}
	}
	if secretKey == "" {
		secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
		if secretKey != "" {
			js.Global().Get("console").Call("log", "Using secret key from environment variable")
		}
	}
	if env == "" {
		env = os.Getenv("MEGAPORT_ENVIRONMENT")
		if env != "" {
			js.Global().Get("console").Call("log", "Using environment from environment variable: "+env)
		}
	}

	// Validate credentials
	if accessKey == "" {
		js.Global().Get("console").Call("error", "No access key provided")
		js.Global().Get("console").Call("groupEnd")
		return nil, fmt.Errorf("megaport API access key not provided. Configure a profile or set MEGAPORT_ACCESS_KEY environment variable")
	}
	if secretKey == "" {
		js.Global().Get("console").Call("error", "No secret key provided")
		js.Global().Get("console").Call("groupEnd")
		return nil, fmt.Errorf("megaport API secret key not provided. Configure a profile or set MEGAPORT_SECRET_KEY environment variable")
	}

	// Default to production
	if env == "" {
		env = "production"
		js.Global().Get("console").Call("log", "No environment specified, defaulting to production")
	}

	// Set environment option
	var envOpt megaport.ClientOpt
	var apiEndpoint string

	switch env {
	case "production":
		apiEndpoint = "https://api.megaport.com"
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
	case "staging":
		apiEndpoint = "https://api-staging.megaport.com" // Adjust if needed
		envOpt = megaport.WithEnvironment(megaport.EnvironmentStaging)
	case "development":
		apiEndpoint = "https://api-dev.megaport.com" // Adjust if needed
		envOpt = megaport.WithEnvironment(megaport.EnvironmentDevelopment)
	default:
		apiEndpoint = "https://api.megaport.com"
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
	}

	js.Global().Get("console").Call("log", "Using API endpoint: "+apiEndpoint)

	// Create HTTP client with longer timeout for browser environments
	httpClient := &http.Client{
		Timeout: 45 * time.Second, // Increase from default timeout
	}

	js.Global().Get("console").Call("log", "HTTP client timeout set to: "+httpClient.Timeout.String())

	// Create Megaport client with credentials
	js.Global().Get("console").Call("log", "Creating Megaport client...")
	megaportClient, err := megaport.New(httpClient,
		megaport.WithCredentials(accessKey, secretKey),
		envOpt,
	)
	if err != nil {
		js.Global().Get("console").Call("error", "Failed to create Megaport client: "+err.Error())
		js.Global().Get("console").Call("groupEnd")
		return nil, fmt.Errorf("failed to create Megaport client: %w", err)
	}

	js.Global().Get("console").Call("log", "Megaport client created successfully")

	// Show login indicator
	spinner := output.PrintLoggingIn(false)

	// Create a new context with longer timeout specifically for auth
	authCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()

	js.Global().Get("console").Call("log", "Checking CORS capabilities...")
	js.Global().Get("console").Call("log", "Origin:", js.Global().Get("location").Get("origin").String())
	js.Global().Get("console").Call("log", "Target auth endpoint:", "https://auth-m2m.megaport.com/oauth2/token")

	js.Global().Get("console").Call("log", "Starting authorization with 30 second timeout")

	// Attempt authentication with retry logic
	_, err = RetryWithBackoffAndConsoleLogging(authCtx, 3, megaportClient)

	if err != nil {
		spinner.Stop()
		js.Global().Get("console").Call("error", "Authentication failed: "+err.Error())

		// Check if this might be a CORS error
		if isCORSError(err) {
			js.Global().Get("console").Call("error", "CORS issue detected")
			js.Global().Get("console").Call("groupEnd")
			return nil, fmt.Errorf("authentication failed due to CORS policy: this may be due to browser security restrictions. %w", err)
		}

		js.Global().Get("console").Call("groupEnd")
		return nil, fmt.Errorf("authentication failed after multiple attempts: %w", err)
	} else {
		js.Global().Get("console").Call("log", "✅ Authentication successful!")
		js.Global().Get("console").Call("groupEnd")
		spinner.StopWithSuccess("Successfully logged in to Megaport")
	}

	return megaportClient, nil
}

// Helper function to mask credentials for logging
func maskCredential(cred string) string {
	if len(cred) <= 8 {
		return cred[:2] + "..." + cred[len(cred)-2:]
	}
	return cred[:4] + "..." + cred[len(cred)-4:]
}

// Enhanced retryWithBackoff that includes console logging and token caching
func RetryWithBackoffAndConsoleLogging(ctx context.Context, attempts int, client *megaport.Client) (*megaport.AuthInfo, error) {
	// Get environment from client
	environment := "production"
	if client != nil && client.BaseURL != nil {
		switch client.BaseURL.Host {
		case "api-staging.megaport.com":
			environment = "staging"
		case "api-dev.megaport.com":
			environment = "development"
		}
	} else {
		return nil, errors.New("megaport client is nil or has no valid BaseURL")
	}

	// First try to get a cached token
	if cachedAuth := CheckCachedToken(environment); cachedAuth != nil {
		js.Global().Get("console").Call("log", "Using cached authentication token")
		return cachedAuth, nil
	}

	// If no valid cached token, proceed with authentication attempts
	var err error
	for i := 0; i < attempts; i++ {
		// Log attempt information
		if i == 0 {
			js.Global().Get("console").Call("log", "Initial authentication attempt using proxy...")
		} else {
			js.Global().Get("console").Call("log", fmt.Sprintf("Retry attempt %d of %d...", i+1, attempts))
		}

		// Create a longer timeout for each attempt
		authCtx, cancel := context.WithTimeout(ctx, 45*time.Second)

		// Attempt the authorization
		startTime := time.Now()

		// Use the new proxy-based authentication method
		authInfo, err := authorizeWithProxy(authCtx, client)

		// Must cancel the context after use
		cancel()

		elapsed := time.Since(startTime)

		if err == nil {
			js.Global().Get("console").Call("log", fmt.Sprintf("Proxy authorization successful (took %v)", elapsed))

			// Verify we actually got a token
			if authInfo != nil && authInfo.AccessToken != "" {
				js.Global().Get("console").Call("log", "Valid access token received")
				return authInfo, nil
			} else {
				js.Global().Get("console").Call("error", "No valid token in auth info")
				err = errors.New("no valid token received")
			}
		}

		// Log error details
		js.Global().Get("console").Call("error", fmt.Sprintf("Authorization failed (took %v): %s", elapsed, err.Error()))

		// Check if we should abort early
		if strings.Contains(err.Error(), "invalid_client") ||
			strings.Contains(err.Error(), "unauthorized") {
			js.Global().Get("console").Call("warn", "Credential error detected - aborting retry attempts")
			return nil, err
		}

		// Don't wait after the last attempt
		if i < attempts-1 {
			waitTime := time.Duration(500*(1<<i)) * time.Millisecond
			js.Global().Get("console").Call("log", fmt.Sprintf("Waiting %v before next attempt", waitTime))
			time.Sleep(waitTime)
		}
	}
	return nil, err
}

// isCORSError detects common CORS-related error patterns
func isCORSError(err error) bool {
	// Network errors in browsers often don't have detailed information
	// but we can look for common patterns
	errStr := err.Error()
	return strings.Contains(errStr, "CORS") ||
		strings.Contains(errStr, "cross-origin") ||
		strings.Contains(errStr, "access control") ||
		strings.Contains(errStr, "blocked by") ||
		strings.Contains(errStr, "not allowed") ||
		strings.Contains(errStr, "No 'Access-Control-Allow-Origin'")
}

// checkCachedToken tries to get a cached token from the browser
func CheckCachedToken(environment string) *megaport.AuthInfo {
	js.Global().Get("console").Call("log", "Checking for cached token...")

	if js.Global().Get("tokenManager").IsUndefined() {
		js.Global().Get("console").Call("warn", "Token manager not available")
		return nil
	}

	// Get token from token manager
	token := js.Global().Get("tokenManager").Call("getToken", environment)

	if token.IsNull() || token.IsUndefined() {
		js.Global().Get("console").Call("log", "No valid cached token found")
		return nil
	}

	// If we have a valid token, create auth info
	tokenStr := token.String()
	if tokenStr != "" {
		js.Global().Get("console").Call("log", "Found cached token")

		// For cached tokens, set expiry to 24 hours from now
		// The token manager already checked if it's valid
		expiry := time.Now().Add(24 * time.Hour)

		return &megaport.AuthInfo{
			AccessToken: tokenStr,
			Expiration:  expiry,
		}
	}

	return nil
}

// clearCachedToken removes any stored token
func ClearCachedToken() {
	if !js.Global().Get("tokenManager").IsUndefined() {
		js.Global().Get("tokenManager").Call("clearToken")
	}
}

func Logout() {
	ClearCachedToken()
	js.Global().Get("console").Call("log", "User logged out and tokens cleared")
}

// authorizeWithProxy handles authentication through the server-side proxy
func authorizeWithProxy(_ context.Context, client *megaport.Client) (*megaport.AuthInfo, error) {
	// Determine the token URL based on environment
	var tokenHostname string
	var tokenPath string

	switch client.BaseURL.Host {
	case "api.megaport.com":
		tokenHostname = "auth-m2m.megaport.com"
	case "api-staging.megaport.com":
		tokenHostname = "auth-m2m-staging.megaport.com"
	default:
		tokenHostname = "auth-m2m-mpone-dev.megaport.com"
	}

	tokenPath = "oauth2/token"

	js.Global().Get("console").Call("log", fmt.Sprintf("Using proxied auth request to %s/%s", tokenHostname, tokenPath))

	// Create form data for the token request
	formData := map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     client.AccessKey,
		"client_secret": client.SecretKey,
	}

	// Convert this to JSON for the request body
	formJSON, err := json.Marshal(formData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling auth request: %w", err)
	}

	// Use our API.MakeProxiedRequest function but with special token endpoint handling
	state, err := api.MakeProxiedRequest(
		js.ValueOf(nil),
		fmt.Sprintf("https://%s/%s", tokenHostname, tokenPath),
		"", // No auth token needed for authentication request
		js.ValueOf(map[string]interface{}{
			"method": "POST",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": string(formJSON),
		}),
	)

	if err != nil {
		js.Global().Get("console").Call("error", "Proxied auth request failed:", err.Error())
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}

	if state.Error != nil {
		js.Global().Get("console").Call("error", "Auth response error:", state.Error.Error())
		return nil, state.Error
	}

	// Parse the response
	if len(state.Result) == 0 {
		return nil, fmt.Errorf("empty response from authentication endpoint")
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("Auth response received (%d bytes)", len(state.Result)))

	// Parse the JSON to extract token data
	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.Unmarshal(state.Result, &response); err != nil {
		js.Global().Get("console").Call("error", "Failed to parse auth response:", err.Error())
		return nil, fmt.Errorf("failed to parse authentication response: %w", err)
	}

	// Check for errors in the response
	if response.Error != "" {
		errMsg := response.Error
		if response.ErrorDesc != "" {
			errMsg += ": " + response.ErrorDesc
		}
		js.Global().Get("console").Call("error", "Auth error from server:", errMsg)
		return nil, fmt.Errorf("authentication error: %s", errMsg)
	}

	// Validate access token
	if response.AccessToken == "" {
		js.Global().Get("console").Call("error", "No access token in response")
		return nil, errors.New("no access token received")
	}

	// Calculate token expiration
	expiry := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)

	js.Global().Get("console").Call("log", fmt.Sprintf(
		"Token acquired successfully: type=%s, expires_in=%d",
		response.TokenType, response.ExpiresIn))

	// If we have a token manager, store the token
	if !js.Global().Get("tokenManager").IsUndefined() {
		environment := "production"
		if client != nil && client.BaseURL != nil {
			switch client.BaseURL.Host {
			case "api-staging.megaport.com":
				environment = "staging"
			case "api-dev.megaport.com":
				environment = "development"
			}
		}
		js.Global().Get("tokenManager").Call("setToken", response.AccessToken, environment)
		js.Global().Get("console").Call("log", "Token cached in browser storage")
	}

	// Create the auth info with the token
	return &megaport.AuthInfo{
		AccessToken: response.AccessToken,
		Expiration:  expiry,
	}, nil
}

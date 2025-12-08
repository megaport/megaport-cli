//go:build js && wasm
// +build js,wasm

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall/js"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/wasm/wasmhttp"
	megaport "github.com/megaport/megaportgo"
)

func Login(ctx context.Context) (*megaport.Client, error) {
	return LoginFunc(ctx)
}

// LoginFunc overrides the standard login for WASM environments
// Note: WASM version uses session-based authentication managed by the browser UI.
// Config profiles are not supported in the WASM version.
var LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
	var accessKey, secretKey, env string

	// Add console logging for debugging
	js.Global().Get("console").Call("group", "🔐 Megaport Authentication Debug")
	js.Global().Get("console").Call("info", "ℹ️  WASM version uses session-based authentication")
	js.Global().Get("console").Call("info", "ℹ️  Config profiles are not supported - please use the login form in the UI")

	// PRIORITY 1: Check for external token from portal (bypasses OAuth flow)
	megaportTokenGlobal := js.Global().Get("megaportToken")
	if !megaportTokenGlobal.IsUndefined() && !megaportTokenGlobal.IsNull() {
		token := megaportTokenGlobal.Get("token").String()
		tokenEnv := megaportTokenGlobal.Get("environment").String()

		if token != "" {
			js.Global().Get("console").Call("log", "✅ Using external token from portal (bypassing OAuth flow)")
			js.Global().Get("console").Call("log", "Environment: "+tokenEnv)

			// Default to production
			if tokenEnv == "" {
				tokenEnv = "production"
			}

			// Set environment option
			var envOpt megaport.ClientOpt
			switch tokenEnv {
			case "production":
				envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
			case "staging":
				envOpt = megaport.WithEnvironment(megaport.EnvironmentStaging)
			case "development":
				envOpt = megaport.WithEnvironment(megaport.EnvironmentDevelopment)
			default:
				envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
			}

			// Create WASM HTTP client
			httpClient := wasmhttp.NewWasmHTTPClient()
			httpClient.Timeout = 45 * time.Second

			// Create Megaport client with the external token (no OAuth flow needed!)
			megaportClient, err := megaport.New(httpClient,
				megaport.WithAccessToken(token, time.Time{}), // Token managed externally by portal
				envOpt,
			)
			if err != nil {
				js.Global().Get("console").Call("error", "Failed to create Megaport client: "+err.Error())
				js.Global().Get("console").Call("groupEnd")
				return nil, fmt.Errorf("failed to create Megaport client: %w", err)
			}

			js.Global().Get("console").Call("log", "✅ Client created with external token - no OAuth needed!")
			js.Global().Get("console").Call("groupEnd")
			return megaportClient, nil
		}
	}

	// PRIORITY 2: Check for API Key/Secret credentials (uses OAuth flow)
	// First, try to get credentials from JavaScript global (set by browser login)
	megaportCredsGlobal := js.Global().Get("megaportCredentials")
	if !megaportCredsGlobal.IsUndefined() && !megaportCredsGlobal.IsNull() {
		js.Global().Get("console").Call("log", "✅ Found credentials from browser login")
		accessKey = megaportCredsGlobal.Get("accessKey").String()
		secretKey = megaportCredsGlobal.Get("secretKey").String()
		env = megaportCredsGlobal.Get("environment").String()
		js.Global().Get("console").Call("log", "Access Key: "+maskCredential(accessKey))
		js.Global().Get("console").Call("log", "Environment: "+env)
	} else {
		// Fallback to environment variables
		accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
		if accessKey != "" {
			js.Global().Get("console").Call("log", "Using access key from environment variable")
			js.Global().Get("console").Call("log", "Access Key: "+maskCredential(accessKey))
		}

		secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
		if secretKey != "" {
			js.Global().Get("console").Call("log", "Using secret key from environment variable")
		}

		env = os.Getenv("MEGAPORT_ENVIRONMENT")
		if env != "" {
			js.Global().Get("console").Call("log", "Using environment from environment variable: "+env)
		}
	}

	// Validate credentials
	if accessKey == "" {
		js.Global().Get("console").Call("error", "No access key or token provided")
		js.Global().Get("console").Call("error", "💡 WASM Tip: Use setAuthToken() for portal tokens or setAuthCredentials() for API keys")
		js.Global().Get("console").Call("groupEnd")
		return nil, fmt.Errorf("megaport API access key not provided. Please use setAuthToken() for portal tokens or setAuthCredentials() for API keys")
	}
	if secretKey == "" {
		js.Global().Get("console").Call("error", "No secret key provided")
		js.Global().Get("console").Call("error", "💡 WASM Tip: Use the login form in the browser UI or set MEGAPORT_SECRET_KEY environment variable")
		js.Global().Get("console").Call("groupEnd")
		return nil, fmt.Errorf("megaport API secret key not provided. Please use the login form in the browser UI or set MEGAPORT_SECRET_KEY environment variable")
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

	// Create WASM HTTP client that uses browser fetch API
	httpClient := wasmhttp.NewWasmHTTPClient()
	httpClient.Timeout = 45 * time.Second // Increase from default timeout

	js.Global().Get("console").Call("log", "HTTP client timeout set to: "+httpClient.Timeout.String())
	js.Global().Get("console").Call("log", "✨ Using WASM HTTP transport with browser fetch")

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

		// Use the SDK's built-in Authorize() method with our WASM HTTP transport
		authInfo, err := client.Authorize(authCtx)

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

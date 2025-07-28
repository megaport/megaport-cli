//go:build js && wasm
// +build js,wasm

package locations

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/wasm/api"
	megaport "github.com/megaport/megaportgo"
)

// Override the standard function with a browser-compatible version
func init() {
	listLocationsFunc = listLocationsWasmImpl
}
func listLocationsWasmImpl(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
	js.Global().Get("console").Call("log", "Using proxied implementation for locations")

	// Setup authenticated client
	authClient, authInfo, err := setupAuthenticatedClient(ctx, client)
	if err != nil {
		return nil, err
	}

	// Prepare the request URL
	url := buildAPIRequestURL(authClient, "/v2/locations")
	js.Global().Get("console").Call("log", fmt.Sprintf("Fetching locations from: %s", url))

	// Use our proxied request function - modified to handle request polling
	state, err := api.MakeProxiedRequest(
		js.ValueOf(nil), // No JS context needed
		url,
		authInfo.AccessToken,
		js.ValueOf(map[string]interface{}{}), // Default options
	)

	// Handle request errors
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("API request failed: %v", err))
		return nil, fmt.Errorf("error making API request: %v", err)
	}

	// Validate response data
	if len(state.Result) == 0 {
		js.Global().Get("console").Call("error", "Empty response received from API")
		return nil, fmt.Errorf("empty response received")
	}

	// Log successful response
	js.Global().Get("console").Call("log", fmt.Sprintf("Received %d bytes of location data", len(state.Result)))

	// Parse the response
	js.Global().Get("console").Call("log", "Parsing locations response...")
	var apiResponse struct {
		Data []*megaport.Location `json:"data"`
	}

	if err := json.Unmarshal(state.Result, &apiResponse); err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("JSON parse error: %v", err))
		return nil, fmt.Errorf("error parsing location data: %v", err)
	}

	// Success
	js.Global().Get("console").Call("log", fmt.Sprintf("Successfully parsed %d locations", len(apiResponse.Data)))
	return apiResponse.Data, nil
}

// setupAuthenticatedClient ensures we have a valid client and auth token
func setupAuthenticatedClient(ctx context.Context, client *megaport.Client) (*megaport.Client, *megaport.AuthInfo, error) {
	// Create a context with sufficient timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	var err error
	// Check if client is nil or needs authentication
	if client == nil {
		client, err = config.Login(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	// Ensure we have a valid auth token
	authInfo, err := config.RetryWithBackoffAndConsoleLogging(ctx, 4, client)
	if err != nil {
		return nil, nil, fmt.Errorf("error retrieving auth token: %v", err)
	}

	// Double check we have a valid client and auth token
	if client == nil || authInfo == nil || authInfo.AccessToken == "" {
		return nil, nil, fmt.Errorf("no valid authentication token available")
	}

	return client, authInfo, nil
}

// buildAPIRequestURL safely builds the API endpoint URL
func buildAPIRequestURL(client *megaport.Client, path string) string {
	if client.BaseURL != nil {
		return client.BaseURL.JoinPath(path).String()
	}
	// Fall back to default URL if BaseURL is nil
	return "https://api.megaport.com" + path
}

//go:build js && wasm
// +build js,wasm

package locations

import (
	"context"
	"fmt"
	"strings"
	"syscall/js"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
)

// Override the standard implementations with WASM-compatible versions
// These now use the SDK directly via the WASM HTTP transport!
func init() {
	listLocationsFunc = listLocationsWasmImpl
}

// isAuthError checks if an error is a 401 authentication error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "Bad session token") ||
		strings.Contains(errStr, "Unauthorized") ||
		strings.Contains(errStr, "authentication failed")
}

// listLocationsWasmImpl uses the SDK's LocationService.ListLocations() method
// The WASM HTTP transport automatically handles the fetch API calls
func listLocationsWasmImpl(ctx context.Context, client *megaport.Client) ([]*megaport.Location, error) {
	js.Global().Get("console").Call("log", "🚀 Using SDK LocationService.ListLocations() with WASM HTTP transport")

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK LocationService.ListLocations()...")
	locations, err := client.LocationService.ListLocations(ctx)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK ListLocations failed: %v", err))
		
		// Check if this is a 401 error and clear the cached token
		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}
		
		return nil, fmt.Errorf("error listing locations: %v", err)
	}
	
	js.Global().Get("console").Call("log", fmt.Sprintf("✅ SDK returned %d locations successfully", len(locations)))
	return locations, nil
}

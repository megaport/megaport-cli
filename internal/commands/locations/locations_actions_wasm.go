//go:build js && wasm
// +build js,wasm

package locations

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
)

// Override the standard implementations with WASM-compatible versions
// These now use the SDK directly via the WASM HTTP transport!
func init() {
	listLocationsFunc = listLocationsWasmImpl
}

// listLocationsWasmImpl uses the SDK's LocationService.ListLocations() method
// The WASM HTTP transport automatically handles the fetch API calls
func listLocationsWasmImpl(ctx context.Context, client *megaport.Client) ([]*megaport.LocationV3, error) {
	js.Global().Get("console").Call("log", "🚀 Using SDK LocationService.ListLocations() with WASM HTTP transport")

	if client == nil {
		var err error
		client, err = config.NewUnauthenticatedClient()
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Failed to create API client: %v", err))
			return nil, fmt.Errorf("failed to create API client: %w", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK LocationService.ListLocations()...")
	locations, err := client.LocationService.ListLocationsV3(ctx)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK ListLocations failed: %v", err))
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("✅ SDK returned %d locations successfully", len(locations)))
	return locations, nil
}

//go:build js && wasm
// +build js,wasm

package mve

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
)

// Override the standard implementations with WASM-compatible versions
func init() {
	listMVEResourceTagsFunc = listMVEResourceTagsWasmImpl
}

// isAuthError checks if the error is an authentication/authorization error
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return (errStr == "unauthorized" || errStr == "forbidden" ||
		errStr == "authentication required" || errStr == "token expired")
}

// listMVEResourceTagsWasmImpl uses the SDK's MVEService.ListMVEResourceTags() method
func listMVEResourceTagsWasmImpl(ctx context.Context, client *megaport.Client, mveID string) (map[string]string, error) {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK MVEService.ListMVEResourceTags() for MVE %s", mveID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK MVEService.ListMVEResourceTags()...")
	tags, err := client.MVEService.ListMVEResourceTags(ctx, mveID)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK ListMVEResourceTags failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return nil, fmt.Errorf("error listing MVE resource tags: %v", err)
	}

	js.Global().Get("console").Call("log", fmt.Sprintf("✅ SDK returned %d resource tags successfully", len(tags)))
	return tags, nil
}

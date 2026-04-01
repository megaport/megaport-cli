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
	lockMVEFunc = lockMVEWasmImpl
	unlockMVEFunc = unlockMVEWasmImpl
	restoreMVEFunc = restoreMVEWasmImpl
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

// lockMVEWasmImpl uses the SDK's ProductService.ManageProductLock() method
func lockMVEWasmImpl(ctx context.Context, client *megaport.Client, mveUID string) (*megaport.ManageProductLockResponse, error) {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK ProductService.ManageProductLock() to lock MVE %s", mveUID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK ProductService.ManageProductLock()...")
	response, err := client.ProductService.ManageProductLock(ctx, &megaport.ManageProductLockRequest{ProductID: mveUID, ShouldLock: true})
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK ManageProductLock (lock) failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return nil, fmt.Errorf("error locking MVE: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK ManageProductLock (lock) successful")
	return response, nil
}

// unlockMVEWasmImpl uses the SDK's ProductService.ManageProductLock() method
func unlockMVEWasmImpl(ctx context.Context, client *megaport.Client, mveUID string) (*megaport.ManageProductLockResponse, error) {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK ProductService.ManageProductLock() to unlock MVE %s", mveUID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK ProductService.ManageProductLock()...")
	response, err := client.ProductService.ManageProductLock(ctx, &megaport.ManageProductLockRequest{ProductID: mveUID, ShouldLock: false})
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK ManageProductLock (unlock) failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return nil, fmt.Errorf("error unlocking MVE: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK ManageProductLock (unlock) successful")
	return response, nil
}

// restoreMVEWasmImpl uses the SDK's ProductService.RestoreProduct() method
func restoreMVEWasmImpl(ctx context.Context, client *megaport.Client, mveUID string) (*megaport.RestoreProductResponse, error) {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK ProductService.RestoreProduct() for MVE %s", mveUID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK ProductService.RestoreProduct()...")
	response, err := client.ProductService.RestoreProduct(ctx, mveUID)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK RestoreProduct failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return nil, fmt.Errorf("error restoring MVE: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK RestoreProduct successful")
	return response, nil
}

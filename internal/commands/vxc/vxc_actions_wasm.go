//go:build js && wasm
// +build js,wasm

package vxc

import (
	"context"
	"fmt"
	"syscall/js"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
)

// Override the standard implementations with WASM-compatible versions
func init() {
	deleteVXCFunc = deleteVXCWasmImpl
	buyVXCFunc = buyVXCWasmImpl
	updateVXCFunc = updateVXCWasmImpl
	getVXCFunc = getVXCWasmImpl
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

// deleteVXCWasmImpl uses the SDK's VXCService.DeleteVXC() method
func deleteVXCWasmImpl(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK VXCService.DeleteVXC() for VXC %s", vxcUID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK VXCService.DeleteVXC()...")
	err := client.VXCService.DeleteVXC(ctx, vxcUID, req)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK DeleteVXC failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return fmt.Errorf("error deleting VXC: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK DeleteVXC successful")
	return nil
}

// buyVXCWasmImpl uses the SDK's VXCService.BuyVXC() method
func buyVXCWasmImpl(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	js.Global().Get("console").Call("log", "🚀 Using SDK VXCService.BuyVXC()")

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK VXCService.BuyVXC()...")
	response, err := client.VXCService.BuyVXC(ctx, req)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK BuyVXC failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return nil, fmt.Errorf("error buying VXC: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK BuyVXC successful")
	return response, nil
}

// updateVXCWasmImpl uses the SDK's VXCService.UpdateVXC() method
func updateVXCWasmImpl(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK VXCService.UpdateVXC() for VXC %s", vxcUID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK VXCService.UpdateVXC()...")
	_, err := client.VXCService.UpdateVXC(ctx, vxcUID, req)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK UpdateVXC failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return fmt.Errorf("error updating VXC: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK UpdateVXC successful")
	return nil
}

// getVXCWasmImpl uses the SDK's VXCService.GetVXC() method
func getVXCWasmImpl(ctx context.Context, client *megaport.Client, vxcUID string) (*megaport.VXC, error) {
	js.Global().Get("console").Call("log", fmt.Sprintf("🚀 Using SDK VXCService.GetVXC() for VXC %s", vxcUID))

	if client == nil {
		var err error
		client, err = config.Login(ctx)
		if err != nil {
			js.Global().Get("console").Call("error", fmt.Sprintf("❌ Login failed: %v", err))
			return nil, fmt.Errorf("error logging in: %v", err)
		}
	}

	js.Global().Get("console").Call("log", "📡 Calling SDK VXCService.GetVXC()...")
	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)
	if err != nil {
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ SDK GetVXC failed: %v", err))

		if isAuthError(err) {
			js.Global().Get("console").Call("warn", "🔓 Authentication token expired or invalid, clearing cache")
			config.ClearCachedToken()
			return nil, fmt.Errorf("authentication token expired. Please run the command again to re-authenticate")
		}

		return nil, fmt.Errorf("error getting VXC: %v", err)
	}

	js.Global().Get("console").Call("log", "✅ SDK GetVXC successful")
	return vxc, nil
}

//go:build js && wasm
// +build js,wasm

package config

import (
	"fmt"
	"syscall/js"
)

const (
	// Storage keys
	configStorageKey = "megaport_cli_config"
)

// GetConfigDir overrides the standard version for WASM environments
func GetConfigDir() (string, error) {
	// In WASM context, we don't need an actual directory
	return "/virtual/.megaport", nil
}

// GetConfigFilePath overrides the standard version for WASM environments
func GetConfigFilePath() (string, error) {
	// Return a virtual file path for reference
	return "/virtual/.megaport/config.json", nil
}

// Checks if the browser supports localStorage
func hasLocalStorage() bool {
	return !js.Global().Get("localStorage").IsUndefined() &&
		!js.Global().Get("localStorage").IsNull()
}

// Additional WASM-specific functions
func SaveToLocalStorage(data []byte) error {
	if !hasLocalStorage() {
		return fmt.Errorf("localStorage is not available in this browser")
	}

	js.Global().Get("localStorage").Call(
		"setItem",
		configStorageKey,
		string(data),
	)

	return nil
}

func LoadFromLocalStorage() ([]byte, error) {
	if !hasLocalStorage() {
		return nil, fmt.Errorf("localStorage is not available in this browser")
	}

	value := js.Global().Get("localStorage").Call("getItem", configStorageKey)

	if value.IsNull() || value.IsUndefined() {
		return nil, nil // No data stored yet
	}

	return []byte(value.String()), nil
}

func ClearLocalStorage() error {
	if !hasLocalStorage() {
		return fmt.Errorf("localStorage is not available in this browser")
	}

	js.Global().Get("localStorage").Call("removeItem", configStorageKey)
	return nil
}

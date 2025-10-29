//go:build !js && !wasm

package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetConfigDir() (string, error) {
	if envDir := os.Getenv("MEGAPORT_CONFIG_DIR"); envDir != "" {
		if err := os.MkdirAll(envDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create config directory: %w", err)
		}
		return envDir, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to find home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".megaport")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

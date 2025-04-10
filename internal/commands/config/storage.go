package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetConfigDir returns the directory where the config file is stored
func GetConfigDir() (string, error) {
	// If MEGAPORT_CONFIG_DIR is set, use that
	if envDir := os.Getenv("MEGAPORT_CONFIG_DIR"); envDir != "" {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(envDir, 0700); err != nil {
			return "", fmt.Errorf("failed to create config directory: %w", err)
		}
		return envDir, nil
	}

	// Otherwise, use ~/.megaport
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

// GetConfigFilePath returns the path to the config file
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

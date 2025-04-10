package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	// ErrProfileNotFound is returned when a profile doesn't exist
	ErrProfileNotFound = errors.New("profile not found")
)

// ConfigManager provides methods for managing the configuration
type ConfigManager struct {
	config     *ConfigFile
	configPath string // Store the actual path used
}

// NewConfigManager creates a new config manager
func NewConfigManager() (*ConfigManager, error) {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %w", err)
	}

	manager := &ConfigManager{
		configPath: configPath, // Store the actual path
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		config := NewConfigFile()
		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		// Make sure directory exists
		configDir, err := GetConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		if err := os.WriteFile(configPath, configData, 0600); err != nil {
			return nil, fmt.Errorf("failed to create default config: failed to write config file: %w", err)
		}

		manager.config = config

		return manager, nil
	}

	// Read existing config
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// When reading the config file:
	var config ConfigFile
	err = json.Unmarshal(configData, &config)
	if err != nil {
		// Instead of returning an error, log a warning and create a new config
		fmt.Fprintf(os.Stderr, "Warning: Config file is corrupted, creating a new default config\n")
		config = ConfigFile{
			Version:       ConfigVersion,
			ActiveProfile: "",
			Profiles:      make(map[string]*Profile),
			Defaults:      make(map[string]interface{}),
		}

		// Write the new config
		configData, err = json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

		err = os.WriteFile(configPath, configData, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to write default config: %w", err)
		}
	}

	if config.Version < ConfigVersion {
		// Upgrade the version
		config.Version = ConfigVersion

		// Save the changes
		manager.config = &config
		err = manager.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save upgraded config: %w", err)
		}
	}

	if config.Profiles == nil {
		config.Profiles = make(map[string]*Profile)
	}
	if config.Defaults == nil {
		config.Defaults = make(map[string]interface{})
	}
	manager.config = &config
	err = manager.Save()
	if err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return manager, nil
}

// GetCurrentProfile returns the currently active profile
func (m *ConfigManager) GetCurrentProfile() (*Profile, string, error) {
	profileName := m.config.ActiveProfile
	profile, exists := m.config.Profiles[profileName]
	if !exists {
		return nil, "", ErrProfileNotFound
	}
	return profile, profileName, nil
}

// CreateProfile creates a new profile with the given credentials and settings
func (m *ConfigManager) CreateProfile(name, accessKey, secretKey, environment, description string) error {
	// Validate profile name
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	// Check for whitespace-only name
	if len(strings.TrimSpace(name)) == 0 {
		return fmt.Errorf("profile name cannot be just whitespace")
	}

	// Ensure config and maps are initialized
	if m.config == nil {
		m.config = NewConfigFile()
	}
	if m.config.Profiles == nil {
		m.config.Profiles = make(map[string]*Profile)
	}

	// Test if we can write to the config file before attempting
	configPath, err := GetConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to get config file path: %w", err)
	}
	file, err := os.OpenFile(configPath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot write to config file: %w", err)
	}
	file.Close()
	m.config.Profiles[name] = &Profile{
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Environment: environment,
		Description: description,
	}
	return m.Save()
}

// UpdateProfile updates an existing profile
func (m *ConfigManager) UpdateProfile(name, accessKey, secretKey, environment string, updateDescription bool, description string) error {
	profile, exists := m.config.Profiles[name]
	if !exists {
		return ErrProfileNotFound
	}

	// Only update fields that have values
	if accessKey != "" {
		profile.AccessKey = accessKey
	}
	if secretKey != "" {
		profile.SecretKey = secretKey
	}
	if environment != "" {
		profile.Environment = environment
	}

	// Only update description if explicitly requested
	if updateDescription {
		profile.Description = description
	}

	return m.Save()
}

// DeleteProfile deletes a profile
func (m *ConfigManager) DeleteProfile(name string) error {
	// Check if profile exists
	if _, exists := m.config.Profiles[name]; !exists {
		return ErrProfileNotFound
	}

	// Check if profile is the active profile
	if m.config.ActiveProfile == name {
		return fmt.Errorf("cannot delete active profile; use 'config use-profile' to switch profiles first")
	}

	// Delete the profile
	delete(m.config.Profiles, name)
	return m.Save()
}

// ListProfiles returns all profiles
func (m *ConfigManager) ListProfiles() (map[string]*Profile, error) {
	if m == nil || m.config == nil || m.config.Profiles == nil {
		return make(map[string]*Profile), nil
	}

	return m.config.Profiles, nil
}

// UseProfile sets the active profile
func (m *ConfigManager) UseProfile(name string) error {
	if _, exists := m.config.Profiles[name]; !exists {
		return ErrProfileNotFound
	}
	m.config.ActiveProfile = name
	return m.Save()
}

// GetDefault gets a default value from config
func (m *ConfigManager) GetDefault(key string) (interface{}, bool) {
	value, exists := m.config.Defaults[key]
	return value, exists
}

// SetDefault sets a default value in config
func (m *ConfigManager) SetDefault(key string, value interface{}) error {
	m.config.Defaults[key] = value
	return m.Save()
}

// Save saves the configuration
func (m *ConfigManager) Save() error {
	configPath := m.configPath

	configData, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Export exports the configuration (excluding sensitive data)
func (m *ConfigManager) Export() (*ConfigFile, error) {
	export := &ConfigFile{
		Version:       m.config.Version,
		ActiveProfile: m.config.ActiveProfile,
		Profiles:      make(map[string]*Profile),
		Defaults:      m.config.Defaults,
	}

	// Redact sensitive information
	for name, profile := range m.config.Profiles {
		export.Profiles[name] = &Profile{
			AccessKey:   "[REDACTED]",
			SecretKey:   "[REDACTED]",
			Environment: profile.Environment,
			Description: profile.Description,
		}
	}

	return export, nil
}

func (m *ConfigManager) RemoveDefault(key string) error {
	if _, exists := m.config.Defaults[key]; exists {
		delete(m.config.Defaults, key)
		return m.Save()
	}
	return nil
}

func (m *ConfigManager) ClearDefaults() error {
	m.config.Defaults = make(map[string]interface{})
	return m.Save()
}

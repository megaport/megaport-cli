//go:build js && wasm
// +build js,wasm

package config

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
)

// ConfigManagerInterface defines the methods that must be implemented by both
// standard and WASM versions of the ConfigManager
type ConfigManagerInterface interface {
	CreateProfile(name, accessKey, secretKey, environment, description string) error
	UpdateProfile(name, accessKey, secretKey, environment string, updateDescription bool, description string) error
	DeleteProfile(name string) error
	ListProfiles() (map[string]*Profile, error)
	UseProfile(name string) error
	GetDefault(key string) (interface{}, bool)
	SetDefault(key string, value interface{}) error
	RemoveDefault(key string) error
	ClearDefaults() error
	Save() error
	Export() (*ConfigFile, error)
	GetCurrentProfile() (*Profile, string, error)
}

// Ensure ConfigManager implements the interface
var _ ConfigManagerInterface = (*ConfigManager)(nil)

// NewConfigManager creates a WASM-specific configuration manager
func NewConfigManager() (*ConfigManager, error) {
	configPath, _ := GetConfigFilePath() // Just for reference, not actually used

	manager := &ConfigManager{
		configPath: configPath,
	}

	// Try to load existing config from localStorage
	configData, err := LoadFromLocalStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to access local storage: %w", err)
	}

	// If no data found, create a new config
	if len(configData) == 0 {
		config := NewConfigFile()
		manager.config = config

		// Save the new config to localStorage
		if err := saveToLocalStorage(manager); err != nil {
			return nil, fmt.Errorf("failed to save new config: %w", err)
		}

		return manager, nil
	}

	// Parse existing config
	var config ConfigFile
	err = json.Unmarshal(configData, &config)
	if err != nil {
		// Config is corrupted, create new one
		config = *NewConfigFile()
		manager.config = &config
		if err := saveToLocalStorage(manager); err != nil {
			return nil, fmt.Errorf("failed to save new config: %w", err)
		}
	} else {
		// Upgrade config version if needed
		if config.Version < ConfigVersion {
			config.Version = ConfigVersion
		}

		// Ensure map fields are initialized
		if config.Profiles == nil {
			config.Profiles = make(map[string]*Profile)
		}
		if config.Defaults == nil {
			config.Defaults = make(map[string]interface{})
		}

		manager.config = &config
		if err := saveToLocalStorage(manager); err != nil {
			return nil, fmt.Errorf("failed to save loaded config: %w", err)
		}
	}

	return manager, nil
}

// Helper function to save to localStorage
func saveToLocalStorage(m *ConfigManager) error {
	configData, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return SaveToLocalStorage(configData)
}

// Override the Save method for WASM environment
func (m *ConfigManager) Save() error {
	return saveToLocalStorage(m)
}

func (m *ConfigManager) GetCurrentProfile() (*Profile, string, error) {
	profileName := m.config.ActiveProfile
	profile, exists := m.config.Profiles[profileName]
	if !exists {
		return nil, "", ErrProfileNotFound
	}
	return profile, profileName, nil
}

// ClearAllCredentials clears all stored credentials
func ClearAllCredentials() error {
	return ClearLocalStorage()
}

func (m *ConfigManager) CreateProfile(name, accessKey, secretKey, environment, description string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}

	if m.config == nil {
		return fmt.Errorf("config not initialized")
	}

	if m.config.Profiles == nil {
		m.config.Profiles = make(map[string]*Profile)
	}

	// Check if profile already exists
	if _, exists := m.config.Profiles[name]; exists {
		return fmt.Errorf("profile '%s' already exists", name)
	}

	m.config.Profiles[name] = &Profile{
		AccessKey:   accessKey,
		SecretKey:   secretKey,
		Environment: environment,
		Description: description,
	}

	return m.Save()
}

func (m *ConfigManager) UpdateProfile(name, accessKey, secretKey, environment string, updateDescription bool, description string) error {
	if m.config == nil || m.config.Profiles == nil {
		return fmt.Errorf("config not initialized")
	}

	profile, exists := m.config.Profiles[name]
	if !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	if accessKey != "" {
		profile.AccessKey = accessKey
	}

	if secretKey != "" {
		profile.SecretKey = secretKey
	}

	if environment != "" {
		profile.Environment = environment
	}

	if updateDescription {
		profile.Description = description
	}

	return m.Save()
}

func (m *ConfigManager) DeleteProfile(name string) error {
	if m.config == nil || m.config.Profiles == nil {
		return fmt.Errorf("config not initialized")
	}

	if _, exists := m.config.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	delete(m.config.Profiles, name)

	// If we deleted the active profile, clear it
	if m.config.ActiveProfile == name {
		m.config.ActiveProfile = ""
	}

	return m.Save()
}

func (m *ConfigManager) ListProfiles() (map[string]*Profile, error) {
	if m.config == nil {
		return nil, fmt.Errorf("config not initialized")
	}

	if m.config.Profiles == nil {
		m.config.Profiles = make(map[string]*Profile)
	}

	return m.config.Profiles, nil
}

func (m *ConfigManager) UseProfile(name string) error {
	if m.config == nil || m.config.Profiles == nil {
		return fmt.Errorf("config not initialized")
	}

	if _, exists := m.config.Profiles[name]; !exists {
		return fmt.Errorf("profile '%s' not found", name)
	}

	m.config.ActiveProfile = name
	return m.Save()
}

func (m *ConfigManager) GetDefault(key string) (interface{}, bool) {
	if m.config == nil || m.config.Defaults == nil {
		return nil, false
	}

	value, exists := m.config.Defaults[key]
	return value, exists
}

func (m *ConfigManager) SetDefault(key string, value interface{}) error {
	if m.config == nil {
		return fmt.Errorf("config not initialized")
	}

	if m.config.Defaults == nil {
		m.config.Defaults = make(map[string]interface{})
	}

	m.config.Defaults[key] = value
	return m.Save()
}

func (m *ConfigManager) RemoveDefault(key string) error {
	if m.config == nil || m.config.Defaults == nil {
		return fmt.Errorf("config not initialized")
	}

	if _, exists := m.config.Defaults[key]; !exists {
		return fmt.Errorf("default setting '%s' not found", key)
	}

	delete(m.config.Defaults, key)
	return m.Save()
}

func (m *ConfigManager) ClearDefaults() error {
	if m.config == nil {
		return fmt.Errorf("config not initialized")
	}

	m.config.Defaults = make(map[string]interface{})
	return m.Save()
}

func (m *ConfigManager) Export() (*ConfigFile, error) {
	if m.config == nil {
		return nil, fmt.Errorf("config not initialized")
	}

	// Create a copy to avoid modifying the original
	exportConfig := &ConfigFile{
		Version:       m.config.Version,
		ActiveProfile: m.config.ActiveProfile,
		Profiles:      make(map[string]*Profile),
		Defaults:      make(map[string]interface{}),
	}

	// Copy profiles
	for name, profile := range m.config.Profiles {
		exportConfig.Profiles[name] = &Profile{
			AccessKey:   profile.AccessKey,
			SecretKey:   profile.SecretKey,
			Environment: profile.Environment,
			Description: profile.Description,
		}
	}

	// Copy defaults
	for key, value := range m.config.Defaults {
		exportConfig.Defaults[key] = value
	}

	return exportConfig, nil
}

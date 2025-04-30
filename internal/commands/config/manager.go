package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
)

type ConfigManager struct {
	config     *ConfigFile
	configPath string
}

func NewConfigManager() (*ConfigManager, error) {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %w", err)
	}

	manager := &ConfigManager{
		configPath: configPath,
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := NewConfigFile()
		configData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}

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

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config ConfigFile
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Config file is corrupted, creating a new default config\n")
		config = ConfigFile{
			Version:       ConfigVersion,
			ActiveProfile: "",
			Profiles:      make(map[string]*Profile),
			Defaults:      make(map[string]interface{}),
		}

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
		config.Version = ConfigVersion
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

func (m *ConfigManager) GetCurrentProfile() (*Profile, string, error) {
	profileName := m.config.ActiveProfile
	profile, exists := m.config.Profiles[profileName]
	if !exists {
		return nil, "", ErrProfileNotFound
	}
	return profile, profileName, nil
}

func (m *ConfigManager) CreateProfile(name, accessKey, secretKey, environment, description string) error {
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if len(strings.TrimSpace(name)) == 0 {
		return fmt.Errorf("profile name cannot be just whitespace")
	}
	if m.config == nil {
		m.config = NewConfigFile()
	}
	if m.config.Profiles == nil {
		m.config.Profiles = make(map[string]*Profile)
	}
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

func (m *ConfigManager) UpdateProfile(name, accessKey, secretKey, environment string, updateDescription bool, description string) error {
	profile, exists := m.config.Profiles[name]
	if !exists {
		return ErrProfileNotFound
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
	if _, exists := m.config.Profiles[name]; !exists {
		return ErrProfileNotFound
	}
	if m.config.ActiveProfile == name {
		return fmt.Errorf("cannot delete active profile; use 'config use-profile' to switch profiles first")
	}
	delete(m.config.Profiles, name)
	return m.Save()
}

func (m *ConfigManager) ListProfiles() (map[string]*Profile, error) {
	if m == nil || m.config == nil || m.config.Profiles == nil {
		return make(map[string]*Profile), nil
	}
	return m.config.Profiles, nil
}

func (m *ConfigManager) UseProfile(name string) error {
	if _, exists := m.config.Profiles[name]; !exists {
		return ErrProfileNotFound
	}
	m.config.ActiveProfile = name
	return m.Save()
}

func (m *ConfigManager) GetDefault(key string) (interface{}, bool) {
	value, exists := m.config.Defaults[key]
	return value, exists
}

func (m *ConfigManager) SetDefault(key string, value interface{}) error {
	m.config.Defaults[key] = value
	return m.Save()
}

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

func (m *ConfigManager) Export() (*ConfigFile, error) {
	export := &ConfigFile{
		Version:       m.config.Version,
		ActiveProfile: m.config.ActiveProfile,
		Profiles:      make(map[string]*Profile),
		Defaults:      m.config.Defaults,
	}
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

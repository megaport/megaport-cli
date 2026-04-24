package config

import (
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// cliHeaders identifies CLI traffic to the Megaport API.
var cliHeaders = map[string]string{"x-app": "cli"}

// ConfigManager handles configuration operations
type ConfigManager struct {
	config     *ConfigFile
	configPath string
}

// ConfigFile represents the configuration file structure
type ConfigFile struct {
	Version       int                    `json:"version"`
	ActiveProfile string                 `json:"activeProfile,omitempty"`
	Profiles      map[string]*Profile    `json:"profiles,omitempty"`
	Defaults      map[string]interface{} `json:"defaults"`
}

// Profile represents a credential profile
type Profile struct {
	AccessKey   string `json:"accessKey"`
	SecretKey   string `json:"secretKey"`
	Environment string `json:"environment"`
	Description string `json:"description,omitempty"`
}

// ConfigVersion is the current version of the config file format
const ConfigVersion = 1

// NewConfigFile creates a new empty configuration file
func NewConfigFile() *ConfigFile {
	return &ConfigFile{
		Version:  ConfigVersion,
		Profiles: make(map[string]*Profile),
		Defaults: make(map[string]interface{}),
	}
}

// normalizeEnvironment maps short aliases and normalizes the environment string
// to a canonical name. Accepts "prod"/"production", "dev"/"development",
// "staging". Unknown values default to "production".
func normalizeEnvironment(env string) string {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production", "prod":
		return "production"
	case "staging":
		return "staging"
	case "development", "dev":
		return "development"
	default:
		return "production"
	}
}

// environmentOption returns the megaport.ClientOpt for the given environment string.
func environmentOption(env string) megaport.ClientOpt {
	switch env {
	case "production":
		return megaport.WithEnvironment(megaport.EnvironmentProduction)
	case "staging":
		return megaport.WithEnvironment(megaport.EnvironmentStaging)
	case "development":
		return megaport.WithEnvironment(megaport.EnvironmentDevelopment)
	default:
		return megaport.WithEnvironment(megaport.EnvironmentProduction)
	}
}

package config

const (
	// ConfigVersion is the current version of the config file format
	ConfigVersion = 1
)

// Profile stores authentication and environment settings
type Profile struct {
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Environment string `json:"environment"`           // "production", "staging", or "development"
	Description string `json:"description,omitempty"` // Optional description
}

// ConfigFile represents the structure of the configuration file
type ConfigFile struct {
	Version       int                    `json:"version"`
	ActiveProfile string                 `json:"active_profile"`
	Profiles      map[string]*Profile    `json:"profiles"`
	Defaults      map[string]interface{} `json:"defaults"`
}

// NewConfigFile creates a new configuration file with default values
func NewConfigFile() *ConfigFile {
	return &ConfigFile{
		Version:       ConfigVersion,
		ActiveProfile: "",
		Profiles:      make(map[string]*Profile),
		Defaults:      make(map[string]interface{}),
	}
}

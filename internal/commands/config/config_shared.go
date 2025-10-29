package config

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

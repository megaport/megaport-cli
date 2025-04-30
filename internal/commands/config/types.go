package config

const (
	ConfigVersion = 1
)

type Profile struct {
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Environment string `json:"environment"`
	Description string `json:"description,omitempty"`
}

type ConfigFile struct {
	Version       int                    `json:"version"`
	ActiveProfile string                 `json:"active_profile"`
	Profiles      map[string]*Profile    `json:"profiles"`
	Defaults      map[string]interface{} `json:"defaults"`
}

func NewConfigFile() *ConfigFile {
	return &ConfigFile{
		Version:       ConfigVersion,
		ActiveProfile: "",
		Profiles:      make(map[string]*Profile),
		Defaults:      make(map[string]interface{}),
	}
}

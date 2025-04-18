package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// Command implementations

// CreateProfile creates a new profile
func CreateProfile(cmd *cobra.Command, args []string, noColor bool) error {
	profileName := args[0]
	accessKey, _ := cmd.Flags().GetString("access-key")
	secretKey, _ := cmd.Flags().GetString("secret-key")
	environment, _ := cmd.Flags().GetString("environment")
	description, _ := cmd.Flags().GetString("description")

	// Validate environment
	if environment != "production" && environment != "staging" && environment != "development" {
		return fmt.Errorf("environment must be 'production', 'staging', or 'development'")
	}

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	if err := manager.CreateProfile(profileName, accessKey, secretKey, environment, description); err != nil {
		return err
	}

	output.PrintSuccess("Profile '%s' created successfully", noColor, profileName)
	return nil
}

// UpdateProfile updates an existing profile
func UpdateProfile(cmd *cobra.Command, args []string, noColor bool) error {
	profileName := args[0]

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	// Check which fields to update
	accessKeyChanged := cmd.Flags().Changed("access-key")
	secretKeyChanged := cmd.Flags().Changed("secret-key")
	environmentChanged := cmd.Flags().Changed("environment")
	descriptionChanged := cmd.Flags().Changed("description")

	// Get values
	accessKey := ""
	if accessKeyChanged {
		accessKey, _ = cmd.Flags().GetString("access-key")
	}

	secretKey := ""
	if secretKeyChanged {
		secretKey, _ = cmd.Flags().GetString("secret-key")
	}

	environment := ""
	if environmentChanged {
		environment, _ = cmd.Flags().GetString("environment")
	}

	description := ""
	if descriptionChanged {
		description, _ = cmd.Flags().GetString("description")
	}

	if err := manager.UpdateProfile(profileName, accessKey, secretKey, environment, descriptionChanged, description); err != nil {
		return err
	}

	output.PrintSuccess("Profile '%s' updated successfully", noColor, profileName)
	return nil
}

// DeleteProfile deletes a profile
func DeleteProfile(cmd *cobra.Command, args []string, noColor bool) error {
	profileName := args[0]

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	profiles, err := manager.ListProfiles()
	if err != nil {
		return err
	}
	if _, exists := profiles[profileName]; !exists {
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	// Confirm deletion
	confirmed := utils.ConfirmPrompt(fmt.Sprintf("Are you sure you want to delete profile '%s'? (y/n): ", profileName), noColor)
	if !confirmed {
		output.PrintInfo("Profile deletion cancelled", noColor)
		return nil
	}

	if err := manager.DeleteProfile(profileName); err != nil {
		return err
	}

	output.PrintSuccess("Profile '%s' deleted successfully", noColor, profileName)
	return nil
}

// ProfileOutput represents the output format for profiles
type ProfileOutput struct {
	output.Output `json:"-" header:"-"`
	Name          string `json:"name" header:"Name"`
	AccessKey     string `json:"access_key" header:"Access Key"`
	Environment   string `json:"environment" header:"Environment"`
	Description   string `json:"description" header:"Description"`
	IsActive      bool   `json:"is_active" header:"Active"`
}

// ListProfiles lists all profiles
func ListProfiles(cmd *cobra.Command, args []string, noColor bool, outputFormat string) error {
	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	profiles, err := manager.ListProfiles()
	if err != nil {
		output.PrintError("Failed to list profiles: %s", noColor, err)
		return err
	}
	activeProfile := manager.config.ActiveProfile

	// Prepare output
	var profileOutputs []ProfileOutput
	for name, profile := range profiles {
		profileOutputs = append(profileOutputs, ProfileOutput{
			Name:        name,
			AccessKey:   profile.AccessKey,
			Environment: profile.Environment,
			Description: profile.Description,
			IsActive:    name == activeProfile,
		})
	}

	// Print output
	if len(profileOutputs) == 0 {
		output.PrintInfo("No profiles found", noColor)
		return nil
	}

	return output.PrintOutput(profileOutputs, outputFormat, noColor)
}

// UseProfile sets the active profile
func UseProfile(cmd *cobra.Command, args []string, noColor bool) error {
	profileName := args[0]

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	profiles, err := manager.ListProfiles()
	if err != nil {
		output.PrintError("Failed to list profiles: %s", noColor, err)
		return err
	}
	if _, exists := profiles[profileName]; !exists {
		output.PrintError("Profile '%s' not found", noColor, profileName)
		return fmt.Errorf("profile '%s' not found", profileName)
	}

	if err := manager.UseProfile(profileName); err != nil {
		output.PrintError("Failed to switch to profile '%s': %s", noColor, profileName, err)
		return err
	}

	output.PrintSuccess("Switched to profile '%s'", noColor, profileName)
	return nil
}

// SetDefault sets a default value
func SetDefault(cmd *cobra.Command, args []string, noColor bool) error {
	key := args[0]
	valueStr := args[1]

	// Define allowed configuration keys and their validators
	allowedSettings := map[string]func(string) (interface{}, error){
		"output": func(v string) (interface{}, error) {
			validFormats := map[string]bool{"json": true, "yaml": true, "table": true}
			if !validFormats[v] {
				return nil, fmt.Errorf("output format must be one of: json, yaml, table")
			}
			return v, nil
		},
		"no-color": func(v string) (interface{}, error) {
			if strings.ToLower(v) == "true" {
				return true, nil
			} else if strings.ToLower(v) == "false" {
				return false, nil
			}
			return nil, fmt.Errorf("no-color must be true or false")
		},
		// Add other valid settings here
	}

	// Check if the key is allowed
	validator, exists := allowedSettings[key]
	if !exists {
		return fmt.Errorf("unknown configuration key: %s. Valid keys are: %s",
			key, strings.Join(mapKeys(allowedSettings), ", "))
	}

	// Validate and convert the value
	value, err := validator(valueStr)
	if err != nil {
		return err
	}

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	if err := manager.SetDefault(key, value); err != nil {
		return err
	}

	output.PrintSuccess("Default '%s' set to '%v'", noColor, key, value)
	return nil
}

// GetDefault gets a default value
func GetDefault(cmd *cobra.Command, args []string, noColor bool) error {
	key := args[0]

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	value, exists := manager.GetDefault(key)
	if !exists {
		return fmt.Errorf("default '%s' not found", key)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%v\n", value)
	return nil
}

// ExportConfig exports the configuration
func ExportConfig(cmd *cobra.Command, args []string, noColor bool) error {
	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	exportConfig, err := manager.Export()
	if err != nil {
		return err
	}

	// Convert config to JSON
	data, err := json.MarshalIndent(exportConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file or stdout
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config to file: %w", err)
		}
		output.PrintSuccess("Configuration exported to '%s'", noColor, filePath)
	} else {
		fmt.Println(string(data))
	}

	return nil
}

// ImportConfig imports configuration from a file
func ImportConfig(cmd *cobra.Command, args []string, noColor bool) error {
	filePath, _ := cmd.Flags().GetString("file")

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON
	var importConfig ConfigFile
	if err := json.Unmarshal(data, &importConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}
	if importConfig.ActiveProfile != "" {
		// Check if the profile actually exists
		if _, exists := importConfig.Profiles[importConfig.ActiveProfile]; !exists {
			// Either skip setting active profile or create an empty one
			return fmt.Errorf("import specifies active profile '%s' but the profile was not found", importConfig.ActiveProfile)
		}

		// Set default values for missing fields
		for profileName, profile := range importConfig.Profiles {
			// Set environment to production if missing
			if profile.Environment == "" {
				profile.Environment = "production"
			}

			// Ensure other required fields have values
			if profile.AccessKey == "" || profile.SecretKey == "" {
				return fmt.Errorf("profile '%s' is missing required credential fields", profileName)
			}
		}
	}

	// Create a new config manager
	manager, err := NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Import profiles
	for name, profile := range importConfig.Profiles {
		err = manager.CreateProfile(
			name,
			profile.AccessKey,
			profile.SecretKey,
			profile.Environment,
			profile.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to import profile '%s': %w", name, err)
		}
	}

	// Import default settings
	for key, value := range importConfig.Defaults {
		err = manager.SetDefault(key, value)
		if err != nil {
			return fmt.Errorf("failed to import default setting '%s': %w", key, err)
		}
	}

	// Set active profile if specified and exists
	if importConfig.ActiveProfile != "" {
		err = manager.UseProfile(importConfig.ActiveProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not set active profile to '%s': %v\n",
				importConfig.ActiveProfile, err)
			// Continue without error - this is a non-critical issue
		}
	}

	// Confirm import
	confirmed := utils.ConfirmPrompt("This will overwrite any existing profiles with the same names. Continue? (y/n): ", noColor)
	if !confirmed {
		output.PrintInfo("Import cancelled", noColor)
		return nil
	}

	// Update current config with imported values
	// Merge profiles (skip ones with [REDACTED] credentials)
	for name, profile := range importConfig.Profiles {
		if profile.AccessKey != "[REDACTED]" && profile.SecretKey != "[REDACTED]" {
			manager.config.Profiles[name] = profile
		}
	}

	// Merge defaults
	for key, value := range importConfig.Defaults {
		manager.config.Defaults[key] = value
	}

	// Save config
	if err := manager.Save(); err != nil {
		return err
	}

	output.PrintSuccess("Configuration imported successfully", noColor)
	return nil
}

// ViewConfig displays the current configuration
func ViewConfig(cmd *cobra.Command, args []string, noColor bool) error {
	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	// Get current settings
	activeProfile, profileName, err := manager.GetCurrentProfile()
	if err != nil {
		// No active profile
		fmt.Fprintf(cmd.OutOrStdout(), "Current Configuration:\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  No active profile set.\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Use 'megaport-cli config use-profile <name>' to set an active profile.\n\n")
	} else {
		// Print current settings
		fmt.Fprintf(cmd.OutOrStdout(), "Current Configuration:\n\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Active Profile: %s\n", profileName)
		fmt.Fprintf(cmd.OutOrStdout(), "  Access Key: %s\n", activeProfile.AccessKey)
		fmt.Fprintf(cmd.OutOrStdout(), "  Environment: %s\n", activeProfile.Environment)

		if activeProfile.Description != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "  Description: %s\n", activeProfile.Description)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\n")
	}

	// Print defaults
	fmt.Fprintf(cmd.OutOrStdout(), "  Default Settings:\n")
	if len(manager.config.Defaults) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "    No default settings configured.\n")
	} else {
		for key, value := range manager.config.Defaults {
			fmt.Fprintf(cmd.OutOrStdout(), "    %s: %v\n", key, value)
		}
	}

	return nil
}

// RemoveDefault removes a default setting
func RemoveDefault(cmd *cobra.Command, args []string, noColor bool) error {
	key := args[0]

	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	if err := manager.RemoveDefault(key); err != nil {
		return err
	}

	output.PrintSuccess("Default setting '%s' removed", noColor, key)
	return nil
}

// ClearDefaults removes all default settings
func ClearDefaults(cmd *cobra.Command, args []string, noColor bool) error {
	manager, err := NewConfigManager()
	if err != nil {
		return err
	}

	// Confirm deletion
	confirmed := utils.ConfirmPrompt("Are you sure you want to clear all default settings? (y/n): ", noColor)
	if !confirmed {
		output.PrintInfo("Operation cancelled", noColor)
		return nil
	}

	if err := manager.ClearDefaults(); err != nil {
		return err
	}

	output.PrintSuccess("All default settings cleared", noColor)
	return nil
}

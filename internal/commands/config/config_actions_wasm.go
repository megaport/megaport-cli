//go:build js && wasm

package config

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// AddWASMSpecificCommands adds browser-specific commands for managing credentials
func AddWASMSpecificCommands(configCmd *cobra.Command) {
	browserCmd := &cobra.Command{
		Use:   "browser",
		Short: "Browser-specific configuration commands",
		Long:  "Commands for managing configuration in browser environments",
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear credentials from browser storage",
		Long:  "Removes all stored credentials from browser localStorage",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ClearLocalStorage()
			if err != nil {
				return err
			}
			fmt.Println("Browser credentials cleared successfully")
			return nil
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show browser storage status",
		Long:  "Displays information about the credentials stored in browser localStorage",
		RunE: func(cmd *cobra.Command, args []string) error {
			configData, err := LoadFromLocalStorage()
			if err != nil {
				return err
			}

			if len(configData) == 0 {
				fmt.Println("No credentials found in browser storage")
			} else {
				fmt.Println("Browser storage contains credentials")
				fmt.Printf("Storage size: %d bytes\n", len(configData))

				// Parse and show profile names without showing the actual credentials
				var config ConfigFile
				if err := json.Unmarshal(configData, &config); err == nil {
					fmt.Println("Profiles stored:")
					for name := range config.Profiles {
						fmt.Printf("- %s\n", name)
					}
					if config.ActiveProfile != "" {
						fmt.Printf("Active profile: %s\n", config.ActiveProfile)
					}
				}
			}
			return nil
		},
	}

	browserCmd.AddCommand(clearCmd, statusCmd)
	configCmd.AddCommand(browserCmd)
}

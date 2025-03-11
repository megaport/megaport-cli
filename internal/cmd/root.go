package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "megaport-cli",
	Short: "A CLI tool to interact with the Megaport API",
	Long: `A CLI tool to interact with the Megaport API.

This CLI supports the following features:
  - Locations: List and manage locations.
  - Ports: List all ports and get details for a specific port.
  - MCRs: Get details for Megaport Cloud Routers.
  - MVEs: Get details for Megaport Virtual Edge devices.
  - VXCs: Get details for Virtual Cross Connects.
  - Partner Ports: List and filter partner ports based on product name, connect type, company name, location ID, and diversity zone.
`,
}

const (
	formatTable = "table"
	formatJSON  = "json"
	formatCSV   = "csv"
	formatXML   = "xml"
)

var (
	env          string
	outputFormat string
	validFormats = []string{formatTable, formatJSON, formatCSV, formatXML}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", formatTable,
		fmt.Sprintf("Output format (%s)", strings.Join(validFormats, ", ")))

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		format := strings.ToLower(outputFormat)
		for _, validFormat := range validFormats {
			if format == validFormat {
				outputFormat = format
				return nil
			}
		}
		return fmt.Errorf("invalid output format: %s. Must be one of: %s",
			outputFormat, strings.Join(validFormats, ", "))
	}

	// Check if the env flag is already defined before adding it
	if rootCmd.PersistentFlags().Lookup("env") == nil {
		rootCmd.PersistentFlags().StringVarP(&env, "env", "e", "production", "Environment to use (production, staging, development)")
	}

	err := rootCmd.PersistentFlags().SetAnnotation("output", cobra.BashCompCustom, validFormats)
	if err != nil {
		fmt.Println(err)
	}
	rootCmd.AddCommand(completionCmd)
}

func initConfig() {
	// Any additional initialization can be done here
}

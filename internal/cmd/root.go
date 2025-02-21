package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "megaport",
	Short: "A CLI tool to interact with the Megaport API",
	Long: `A CLI tool to interact with the Megaport API.

This CLI supports the following features:
  - Configure credentials: Use "megaport configure" to set your access and secret keys.
  - Locations: List and manage locations.
  - Ports: List all ports and get details for a specific port.
  - MCRs: Get details for Megaport Cloud Routers.
  - MVEs: Get details for Megaport Virtual Edge devices.
  - VXCs: Get details for Virtual Cross Connects.
  - Partner Ports: List and filter partner ports based on product name, connect type, company name, location ID, and diversity zone.
`,
}

var (
	outputFormat string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json, table, or csv)")
}

func initConfig() {
	// Any additional initialization can be done here
}

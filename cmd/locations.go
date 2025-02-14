/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// locationsCmd represents the locations command
var locationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List all available locations",
	Long: `The locations command provides a list of all available locations 
where services can be provisioned. This command can be used to get 
detailed information about each location, including its name, 
region, and availability. For example:

megaport-cli locations`,
}

func init() {
	rootCmd.AddCommand(locationsCmd)
}

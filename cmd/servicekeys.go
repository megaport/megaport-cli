package cmd

import (
	"github.com/spf13/cobra"
)

// servicekeysCmd represents the servicekeys command
var servicekeysCmd = &cobra.Command{
	Use:   "servicekeys",
	Short: "Manage service keys in the Megaport API",
	Long:  `Manage service keys in the Megaport API. This command allows you to create, list, update, and get service keys.`,
}

func init() {
	rootCmd.AddCommand(servicekeysCmd)
}

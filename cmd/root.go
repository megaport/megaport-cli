package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mp1",
	Short: "A CLI tool to interact with the Megaport API",
	Long:  `A CLI tool to interact with the Megaport API, allowing you to manage locations, service keys, and more.`,
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

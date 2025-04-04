package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// WrapRunE wraps a RunE function to set SilenceUsage to true if an error occurs and formats the error message.
func WrapRunE(runE func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runE(cmd, args)
		if err != nil {
			// Prevent usage output if an error occurs
			cmd.SilenceUsage = true
			// Silence duplicate error message
			cmd.SilenceErrors = true

			// Return a formatted error message with additional context
			return fmt.Errorf("error running %s command\n\nError: %v\nCommand: %s\nArguments: %v\n\nFor more information, use the --help flag", cmd.Name(), err, cmd.Name(), args)
		}
		return nil
	}
}

var prompt = func(msg string) (string, error) {
	if !noColor {
		fmt.Print(color.BlueString(msg))
	} else {
		fmt.Print(msg)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

var confirmPrompt = func(question string) bool {
	var response string

	if !noColor {
		fmt.Print(color.YellowString("%s [y/N]: ", question))
	} else {
		fmt.Printf("%s [y/N]: ", question)
	}

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println("Error reading input:", err)
		return false // Or handle the error as appropriate for your use case
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

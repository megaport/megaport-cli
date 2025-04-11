package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var Prompt = func(msg string, noColor bool) (string, error) {
	if !noColor {
		// Add contextual icon and use Megaport's red
		fmt.Print(color.New(color.FgHiRed, color.Bold).Sprint("❯ " + msg + " "))
	} else {
		fmt.Print("❯ " + msg + " ")
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

var ConfirmPrompt = func(question string, noColor bool) bool {
	var response string

	if !noColor {
		// Add warning icon for confirmation prompts
		fmt.Print(color.New(color.FgHiRed).Sprint("⚠️  " + question + " "))
		fmt.Print(color.New(color.FgHiWhite, color.Bold).Sprint("[y/N]") + " ")
	} else {
		fmt.Printf("⚠️  %s [y/N] ", question)
	}

	_, err := fmt.Scanln(&response)
	if err != nil {
		// Handle empty response (just pressing Enter)
		if err.Error() == "unexpected newline" {
			return false // Default to "No"
		}
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

var ResourcePrompt = func(resourceType string, msg string, noColor bool) (string, error) {
	// Choose icon based on resource type
	icon := "❯"
	switch strings.ToLower(resourceType) {
	case "port":
		icon = "🔌"
	case "mve":
		icon = "🌐"
	case "mcr":
		icon = "🛰️"
	case "vxc":
		icon = "🔗"
	case "location":
		icon = "📍"
	}

	if !noColor {
		fmt.Print(color.New(color.FgHiRed, color.Bold).Sprint(icon + " " + msg + " "))
	} else {
		fmt.Print(icon + " " + msg + " ")
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

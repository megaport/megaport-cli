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

var ResourceTagsPrompt = func(noColor bool) (map[string]string, error) {
	addTags := ConfirmPrompt("Would you like to add resource tags?", noColor)
	if !addTags {
		return nil, nil
	}

	tags := make(map[string]string)
	fmt.Println("Enter tags (key and value). Enter empty key to finish.")

	for {
		// Prompt for tag key
		key, err := Prompt("Tag key:", noColor)
		if err != nil {
			return nil, err
		}

		// Empty key means we're done
		if key == "" {
			break
		}

		// Prompt for tag value
		value, err := Prompt("Tag value for '"+key+"':", noColor)
		if err != nil {
			return nil, err
		}

		tags[key] = value
	}

	if len(tags) > 0 {
		fmt.Println("Tags added:")
		for k, v := range tags {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return tags, nil
}

var UpdateResourceTagsPrompt = func(existingTags map[string]string, noColor bool) (map[string]string, error) {
	// Display warning about replacing tags
	if !noColor {
		fmt.Println(color.New(color.FgHiYellow).Sprint("⚠️  Warning: This operation will replace all existing tags with the new set of tags you define."))
	} else {
		fmt.Println("⚠️  Warning: This operation will replace all existing tags with the new set of tags you define.")
	}

	// Show existing tags if any
	if len(existingTags) > 0 {
		fmt.Println("Current tags:")
		for k, v := range existingTags {
			fmt.Printf("  %s: %s\n", k, v)
		}
	} else {
		fmt.Println("No existing tags found.")
	}

	// Ask if they want to proceed
	proceed := ConfirmPrompt("Do you want to continue with updating tags?", noColor)
	if !proceed {
		return nil, fmt.Errorf("tag update cancelled by user")
	}

	// Options for common tag operations
	if len(existingTags) > 0 {
		fmt.Println("\nChoose how you want to update tags:")
		fmt.Println("1. Start with a clean slate (remove all existing tags)")
		fmt.Println("2. Start with existing tags and modify them")

		choice, err := Prompt("Enter choice (1 or 2):", noColor)
		if err != nil {
			return nil, err
		}

		// Start with existing tags if requested
		tags := make(map[string]string)
		if choice == "2" {
			// Copy existing tags
			for k, v := range existingTags {
				tags[k] = v
			}

			// Allow modifying existing tags
			fmt.Println("\nYou can now modify, add, or remove tags.")
			fmt.Println("To remove a tag, enter its key and an empty value.")
		} else if choice != "1" {
			return nil, fmt.Errorf("invalid choice: %s", choice)
		}

		// Now prompt for new/updated tags
		fmt.Println("\nEnter tags (key and value). Enter empty key to finish.")
		for {
			// Prompt for tag key
			key, err := Prompt("Tag key:", noColor)
			if err != nil {
				return nil, err
			}

			// Empty key means we're done
			if key == "" {
				break
			}

			// Prompt for tag value
			value, err := Prompt(fmt.Sprintf("Tag value for '%s':", key), noColor)
			if err != nil {
				return nil, err
			}

			// Empty value means remove the tag
			if value == "" && tags[key] != "" {
				delete(tags, key)
				fmt.Printf("  Removed tag: %s\n", key)
			} else if value != "" {
				tags[key] = value
				fmt.Printf("  Updated tag: %s: %s\n", key, value)
			}
		}

		fmt.Println("\nFinal tags that will be applied:")
		if len(tags) > 0 {
			for k, v := range tags {
				fmt.Printf("  %s: %s\n", k, v)
			}
		} else {
			fmt.Println("  No tags - all existing tags will be removed")
		}

		// Final confirmation
		confirmApply := ConfirmPrompt("Apply these changes?", noColor)
		if !confirmApply {
			return nil, fmt.Errorf("tag update cancelled by user")
		}

		return tags, nil
	} else {
		// No existing tags, just use the normal tag prompt
		return ResourceTagsPrompt(noColor)
	}
}

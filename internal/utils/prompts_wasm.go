//go:build js && wasm
// +build js,wasm

package utils

import (
	"fmt"
	"strings"

	"github.com/megaport/megaport-cli/internal/wasm"
)

// WASM versions of the prompt functions that use JavaScript callbacks
// instead of stdin/stdout

func init() {
	// Override the default prompt functions with WASM-specific ones
	Prompt = wasmPrompt
	ConfirmPrompt = wasmConfirmPrompt
	ResourcePrompt = wasmResourcePrompt
	ResourceTagsPrompt = wasmResourceTagsPrompt
	UpdateResourceTagsPrompt = wasmUpdateResourceTagsPrompt
}

func wasmPrompt(msg string, noColor bool) (string, error) {
	// In WASM, we use the JavaScript prompt callback
	return wasm.PromptForInput(msg, "text", "")
}

func wasmConfirmPrompt(question string, noColor bool) bool {
	// Add [y/N] to the question for clarity
	fullQuestion := question + " [y/N]"
	
	response, err := wasm.PromptForInput(fullQuestion, "confirm", "")
	if err != nil {
		return false
	}
	
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func wasmResourcePrompt(resourceType string, msg string, noColor bool) (string, error) {
	// Use the resource-specific prompt type
	return wasm.PromptForInput(msg, "resource", resourceType)
}

func wasmResourceTagsPrompt(noColor bool) (map[string]string, error) {
	addTags := wasmConfirmPrompt("Would you like to add resource tags?", noColor)
	if !addTags {
		return nil, nil
	}

	tags := make(map[string]string)
	
	// Inform user how to finish entering tags
	fmt.Println("Enter tags (key and value). Enter empty key to finish.")

	for {
		// Prompt for tag key
		key, err := wasmPrompt("Tag key (empty to finish):", noColor)
		if err != nil {
			return nil, err
		}

		// Empty key means we're done
		if key == "" {
			break
		}

		// Prompt for tag value
		value, err := wasmPrompt(fmt.Sprintf("Tag value for '%s':", key), noColor)
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

func wasmUpdateResourceTagsPrompt(existingTags map[string]string, noColor bool) (map[string]string, error) {
	// Display warning about replacing tags
	fmt.Println("⚠️  Warning: This operation will replace all existing tags with the new set of tags you define.")

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
	proceed := wasmConfirmPrompt("Do you want to continue with updating tags?", noColor)
	if !proceed {
		return nil, fmt.Errorf("tag update cancelled by user")
	}

	// Options for common tag operations
	tags := make(map[string]string)
	
	if len(existingTags) > 0 {
		fmt.Println("\nChoose how you want to update tags:")
		fmt.Println("1. Start with a clean slate (remove all existing tags)")
		fmt.Println("2. Start with existing tags and modify them")

		choice, err := wasmPrompt("Enter choice (1 or 2):", noColor)
		if err != nil {
			return nil, err
		}

		// Start with existing tags if requested
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
	}

	// Now prompt for new/updated tags
	fmt.Println("\nEnter tags (key and value). Enter empty key to finish.")
	for {
		// Prompt for tag key
		key, err := wasmPrompt("Tag key (empty to finish):", noColor)
		if err != nil {
			return nil, err
		}

		// Empty key means we're done
		if key == "" {
			break
		}

		// Prompt for tag value
		value, err := wasmPrompt(fmt.Sprintf("Tag value for '%s' (empty to remove):", key), noColor)
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
	confirmApply := wasmConfirmPrompt("Apply these changes?", noColor)
	if !confirmApply {
		return nil, fmt.Errorf("tag update cancelled by user")
	}

	return tags, nil
}

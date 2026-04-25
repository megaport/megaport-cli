package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/spf13/cobra"
)

const defaultTagsTimeout = 90 * time.Second

// TagListerFunc is a function that lists resource tags for a given UID.
type TagListerFunc func(ctx context.Context, uid string) (map[string]string, error)

// TagUpdaterFunc is a function that updates resource tags for a given UID.
type TagUpdaterFunc func(ctx context.Context, uid string, tags map[string]string) error

// ListResourceTags handles the common pattern of listing resource tags:
// setting output format, calling the list function, converting to ResourceTag slice,
// sorting, and printing output.
func ListResourceTags(resourceType, uid string, noColor bool, outputFormat string, listFunc TagListerFunc) error {
	output.SetOutputFormat(outputFormat)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTagsTimeout)
	defer cancel()

	tagsMap, err := listFunc(ctx, uid)
	if err != nil {
		output.PrintError("Failed to get resource tags for %s %s: %v", noColor, resourceType, uid, err)
		return fmt.Errorf("failed to get resource tags for %s %s: %w", resourceType, uid, err)
	}

	tags := make([]output.ResourceTag, 0, len(tagsMap))
	for k, v := range tagsMap {
		tags = append(tags, output.ResourceTag{Key: k, Value: v})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	return output.PrintOutput(tags, outputFormat, noColor)
}

// UpdateTagsOptions contains the options for UpdateResourceTags.
type UpdateTagsOptions struct {
	ResourceType  string
	UID           string
	NoColor       bool
	Cmd           *cobra.Command
	ListFunc      TagListerFunc
	UpdateFunc    TagUpdaterFunc
	ExtraTagFlags bool // When true, also check --tags, --tags-file, --resource-tags flags
}

// UpdateResourceTags handles the common pattern of updating resource tags:
// fetching existing tags, parsing input (interactive/JSON/JSON file), calling
// the update function, and printing results.
func UpdateResourceTags(opts UpdateTagsOptions) error {
	// Use a dedicated context for the initial list call so interactive prompts
	// don't consume the timeout budget. Cancel immediately after the call to
	// release timer resources before any potentially long interactive prompt.
	listCtx, listCancel := context.WithTimeout(context.Background(), defaultTagsTimeout)
	existingTags, err := opts.ListFunc(listCtx, opts.UID)
	listCancel()
	if err != nil {
		output.PrintError("Failed to log in or list existing resource tags: %v", opts.NoColor, err)
		return fmt.Errorf("failed to log in or list existing resource tags: %w", err)
	}

	interactive, _ := opts.Cmd.Flags().GetBool("interactive")

	var resourceTags map[string]string

	if interactive {
		// No timeout around interactive prompts — user input can take any amount of time.
		resourceTags, err = UpdateResourceTagsPrompt(existingTags, opts.NoColor)
		if err != nil {
			output.PrintError("Failed to update resource tags: %v", opts.NoColor, err)
			return err
		}
	} else if opts.ExtraTagFlags {
		resourceTags, err = parseResourceTagsInputExtended(opts.Cmd)
		if err != nil {
			output.PrintError("Failed to parse resource tags input: %v", opts.NoColor, err)
			return err
		}
	} else {
		resourceTags, err = ParseResourceTagsInput(opts.Cmd)
		if err != nil {
			output.PrintError("Failed to parse resource tags input: %v", opts.NoColor, err)
			return err
		}
	}

	// Normalize nil to empty map so the API receives {} rather than null.
	if resourceTags == nil {
		resourceTags = map[string]string{}
	}

	if len(resourceTags) == 0 {
		output.PrintWarning("No tags provided. The %s will have all existing tags removed.", opts.NoColor, opts.ResourceType)
	}

	force, _ := opts.Cmd.Flags().GetBool("force")
	if !force && !interactive && len(existingTags) > 0 {
		msg := fmt.Sprintf("This will replace %d existing tag(s). Continue? [y/N]: ", len(existingTags))
		if !ConfirmPrompt(msg, opts.NoColor) {
			output.PrintInfo("Update cancelled", opts.NoColor)
			return exitcodes.NewCancelledError(fmt.Errorf("cancelled by user"))
		}
	}

	// Use a separate context for the update call.
	updateCtx, updateCancel := context.WithTimeout(context.Background(), defaultTagsTimeout)
	defer updateCancel()

	spinner := output.PrintResourceUpdating(opts.ResourceType+"-Resource-Tags", opts.UID, opts.NoColor)

	err = opts.UpdateFunc(updateCtx, opts.UID, resourceTags)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update resource tags: %v", opts.NoColor, err)
		return fmt.Errorf("failed to update resource tags: %w", err)
	}

	output.PrintSuccess("Resource tags updated for %s %s", opts.NoColor, opts.ResourceType, opts.UID)
	return nil
}

// ParseResourceTagsInput reads resource tags from --json or --json-file flags.
func ParseResourceTagsInput(cmd *cobra.Command) (map[string]string, error) {
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")

	var resourceTags map[string]string

	if jsonStr != "" {
		if err := json.Unmarshal([]byte(jsonStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	} else if jsonFile != "" {
		jsonData, err := os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read JSON file: %w", err)
		}
		if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse JSON file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no input provided, use --interactive, --json, or --json-file to specify resource tags")
	}

	return resourceTags, nil
}

// parseResourceTagsInputExtended reads resource tags from --json, --json-file,
// --tags, --tags-file, or --resource-tags flags (ports-specific extended flags).
func parseResourceTagsInputExtended(cmd *cobra.Command) (map[string]string, error) {
	jsonStr, _ := cmd.Flags().GetString("json")
	jsonFile, _ := cmd.Flags().GetString("json-file")
	tagsStr, _ := cmd.Flags().GetString("tags")
	tagsFile, _ := cmd.Flags().GetString("tags-file")
	resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")

	var resourceTags map[string]string

	switch {
	case jsonStr != "":
		if err := json.Unmarshal([]byte(jsonStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case jsonFile != "":
		jsonData, err := os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read JSON file: %w", err)
		}
		if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse JSON file: %w", err)
		}
	case tagsStr != "":
		if err := json.Unmarshal([]byte(tagsStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse tags JSON: %w", err)
		}
	case resourceTagsStr != "":
		if err := json.Unmarshal([]byte(resourceTagsStr), &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse resource-tags JSON: %w", err)
		}
	case tagsFile != "":
		tagData, err := os.ReadFile(tagsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read tags file: %w", err)
		}
		if err := json.Unmarshal(tagData, &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse tags file JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("no input provided, use --interactive, --json, --json-file, --tags, --resource-tags, or --tags-file to specify resource tags")
	}

	return resourceTags, nil
}

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/spf13/cobra"
)

// TagListerFunc is a function that lists resource tags for a given UID.
type TagListerFunc func(ctx context.Context, uid string) (map[string]string, error)

// TagUpdaterFunc is a function that updates resource tags for a given UID.
type TagUpdaterFunc func(ctx context.Context, uid string, tags map[string]string) error

// ListResourceTags handles the common pattern of listing resource tags:
// setting output format, calling the list function, converting to ResourceTag slice,
// sorting, and printing output.
func ListResourceTags(resourceType, uid string, noColor bool, outputFormat string, listFunc TagListerFunc) error {
	output.SetOutputFormat(outputFormat)

	ctx := context.Background()

	tagsMap, err := listFunc(ctx, uid)
	if err != nil {
		output.PrintError("Error getting resource tags for %s %s: %v", noColor, resourceType, uid, err)
		return fmt.Errorf("error getting resource tags for %s %s: %v", resourceType, uid, err)
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
	ResourceType string
	UID          string
	NoColor      bool
	Cmd          *cobra.Command
	ListFunc     TagListerFunc
	UpdateFunc   TagUpdaterFunc
}

// UpdateResourceTags handles the common pattern of updating resource tags:
// fetching existing tags, parsing input (interactive/JSON/JSON file), calling
// the update function, and printing results.
func UpdateResourceTags(opts UpdateTagsOptions) error {
	ctx := context.Background()

	existingTags, err := opts.ListFunc(ctx, opts.UID)
	if err != nil {
		output.PrintError("Failed to get existing resource tags: %v", opts.NoColor, err)
		return fmt.Errorf("failed to get existing resource tags: %v", err)
	}

	interactive, _ := opts.Cmd.Flags().GetBool("interactive")

	var resourceTags map[string]string

	if interactive {
		resourceTags, err = UpdateResourceTagsPrompt(existingTags, opts.NoColor)
		if err != nil {
			output.PrintError("Failed to update resource tags: %v", opts.NoColor, err)
			return err
		}
	} else {
		resourceTags, err = ParseResourceTagsInput(opts.Cmd)
		if err != nil {
			return err
		}
	}

	if len(resourceTags) == 0 {
		output.PrintWarning("No tags provided. The %s will have all existing tags removed.", opts.NoColor, opts.ResourceType)
	}

	spinner := output.PrintResourceUpdating(opts.ResourceType+"-Resource-Tags", opts.UID, opts.NoColor)

	err = opts.UpdateFunc(ctx, opts.UID, resourceTags)

	spinner.Stop()

	if err != nil {
		output.PrintError("Failed to update resource tags: %v", opts.NoColor, err)
		return fmt.Errorf("failed to update resource tags: %v", err)
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
			return nil, fmt.Errorf("error parsing JSON: %v", err)
		}
	} else if jsonFile != "" {
		jsonData, err := os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
		if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
			return nil, fmt.Errorf("error parsing JSON file: %v", err)
		}
	} else {
		return nil, fmt.Errorf("no input provided, use --interactive, --json, or --json-file to specify resource tags")
	}

	return resourceTags, nil
}

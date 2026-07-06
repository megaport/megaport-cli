package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/spf13/cobra"
)

const maxTagsFileSize = 1 << 20 // 1 MiB

// readTagsFile reads a user-supplied JSON file path safely: it cleans the path,
// rejects upward traversal, rejects non-regular files, and enforces a 1 MiB
// size limit while reading (open + stat + LimitReader avoids a TOCTOU race).
func readTagsFile(path string) ([]byte, error) {
	clean := filepath.Clean(path)
	// Reject paths that still navigate above the current directory after cleaning.
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid file path %q: path traversal not allowed", path)
	}
	f, err := os.Open(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("file %q is not a regular file", path)
	}
	if info.Size() > maxTagsFileSize {
		return nil, fmt.Errorf("file %q exceeds maximum allowed size of 1 MiB (%d bytes)", path, info.Size())
	}
	data, err := io.ReadAll(io.LimitReader(f, maxTagsFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if len(data) > maxTagsFileSize {
		return nil, fmt.Errorf("file %q exceeds maximum allowed size of 1 MiB (%d bytes)", path, len(data))
	}
	return data, nil
}

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
	interactive, _ := opts.Cmd.Flags().GetBool("interactive")
	if err := CheckInteractiveConflict(interactive, HasConflictingInputFlags(opts.Cmd)); err != nil {
		output.PrintError("%v", opts.NoColor, err)
		return err
	}

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
		var msg string
		if len(resourceTags) == 0 {
			msg = fmt.Sprintf("This will remove all %d existing tag(s). Continue?", len(existingTags))
		} else {
			msg = fmt.Sprintf("This will replace %d existing tag(s). Continue?", len(existingTags))
		}
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

// RejectEmptyTagKeys rejects a tag map with an empty key, matching the
// interactive tag entry which treats an empty key as "stop" rather than a tag.
func RejectEmptyTagKeys(tags map[string]string) error {
	if _, ok := tags[""]; ok {
		return fmt.Errorf("tag key must not be empty")
	}
	return nil
}

// TagMapFromObject converts a decoded JSON object into a string tag map,
// rejecting non-string values (including null) and empty keys. It is the shared
// validation behind both the --resource-tags flag and the JSON-body
// resourceTags field so the two paths accept and reject identically.
func TagMapFromObject(raw map[string]interface{}) (map[string]string, error) {
	if raw == nil {
		return nil, nil
	}
	tags := make(map[string]string, len(raw))
	for k, v := range raw {
		strValue, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("resourceTags value for key %q must be a string", k)
		}
		tags[k] = strValue
	}
	if err := RejectEmptyTagKeys(tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// ParseResourceTagsFlag parses the --resource-tags JSON string into a tag map,
// applying the same value and empty-key validation as the JSON path. An empty
// string (flag unset) yields a nil map and no error.
func ParseResourceTagsFlag(resourceTagsStr string) (map[string]string, error) {
	return ParseResourceTagsFlagOrFile(resourceTagsStr, "")
}

// ParseResourceTagsFlagOrFile parses resource tags from the --resource-tags JSON
// string or, when that is empty, the --resource-tags-file path. The string takes
// precedence over the file (matching utils.ReadJSONInput). It applies the same
// value and empty-key validation as the JSON path via TagMapFromObject. When
// neither is set it returns a nil map and no error.
func ParseResourceTagsFlagOrFile(resourceTagsStr, resourceTagsFile string) (map[string]string, error) {
	var data []byte
	switch {
	case resourceTagsStr != "":
		data = []byte(resourceTagsStr)
	case resourceTagsFile != "":
		fileData, err := readTagsFile(resourceTagsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read resource tags file: %w", err)
		}
		data = fileData
	default:
		return nil, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse resource tags JSON: %w", err)
	}
	return TagMapFromObject(raw)
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
		jsonData, err := readTagsFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read JSON file: %w", err)
		}
		if err := json.Unmarshal(jsonData, &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse JSON file: %w", err)
		}
	} else {
		return nil, fmt.Errorf("no input provided, use --interactive, --json, or --json-file to specify resource tags")
	}

	if err := RejectEmptyTagKeys(resourceTags); err != nil {
		return nil, err
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
		jsonData, err := readTagsFile(jsonFile)
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
		tagData, err := readTagsFile(tagsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read tags file: %w", err)
		}
		if err := json.Unmarshal(tagData, &resourceTags); err != nil {
			return nil, fmt.Errorf("failed to parse tags file JSON: %w", err)
		}
	default:
		return nil, fmt.Errorf("no input provided, use --interactive, --json, --json-file, --tags, --resource-tags, or --tags-file to specify resource tags")
	}

	if err := RejectEmptyTagKeys(resourceTags); err != nil {
		return nil, err
	}

	return resourceTags, nil
}

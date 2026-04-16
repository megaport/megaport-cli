package utils

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
)

// ConfirmDelete prompts the user to confirm deletion of a resource.
// If force is true, confirmation is skipped and (true, nil) is returned.
// If the user declines, it returns (false, err) with a Cancelled exit code.
func ConfirmDelete(resourceType, resourceID string, force, noColor bool) (bool, error) {
	if force {
		return true, nil
	}
	message := fmt.Sprintf("Are you sure you want to delete %s %s?", resourceType, resourceID)
	if !ConfirmPrompt(message, noColor) {
		output.PrintInfo("Deletion cancelled", noColor)
		return false, exitcodes.New(exitcodes.Cancelled, fmt.Errorf("cancelled by user"))
	}
	return true, nil
}

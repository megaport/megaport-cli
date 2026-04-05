package utils

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
)

// ApplyLimitAndPrint handles the common post-filter pipeline for all List
// commands: validate and apply --limit, check for empty results, and print.
//
// Returns nil (no error) when items is empty and outputFormat is table,
// after printing an informational message.
func ApplyLimitAndPrint[T any](
	items []T,
	limit int,
	outputFormat string,
	noColor bool,
	emptyMessage string,
	printFunc func([]T, string, bool) error,
) error {
	if limit < 0 {
		return fmt.Errorf("--limit must be a non-negative integer")
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	if len(items) == 0 {
		if outputFormat == FormatTable {
			output.PrintInfo(emptyMessage, noColor)
		}
		return nil
	}

	if err := printFunc(items, outputFormat, noColor); err != nil {
		output.PrintError("Failed to print output: %v", noColor, err)
		return fmt.Errorf("error printing output: %w", err)
	}
	return nil
}

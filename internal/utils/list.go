package utils

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
)

// ApplyLimitAndPrint handles the common post-filter pipeline for all List
// commands: validate and apply --limit, check for empty results, and print.
//
// For table output an empty result prints a human-readable message instead of
// an empty table. For machine formats (json/csv/xml/go-template) it always
// calls printFunc so an empty result still emits a valid document ([] for json,
// header-only or empty for csv, <items></items> for xml) rather than zero bytes.
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

	if len(items) == 0 && outputFormat == FormatTable {
		output.PrintInfo(emptyMessage, noColor)
		return nil
	}

	return printFunc(items, outputFormat, noColor)
}

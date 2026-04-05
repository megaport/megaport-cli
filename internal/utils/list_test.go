package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyLimitAndPrint(t *testing.T) {
	noop := func(items []string, format string, noColor bool) error { return nil }

	t.Run("negative limit returns error", func(t *testing.T) {
		err := ApplyLimitAndPrint([]string{"a"}, -1, FormatTable, true, "none", noop)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--limit must be a non-negative integer")
	})

	t.Run("applies limit", func(t *testing.T) {
		var printed []string
		err := ApplyLimitAndPrint([]string{"a", "b", "c"}, 2, FormatTable, true, "none",
			func(items []string, format string, noColor bool) error {
				printed = items
				return nil
			})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, printed)
	})

	t.Run("zero limit means no limit", func(t *testing.T) {
		var printed []string
		err := ApplyLimitAndPrint([]string{"a", "b", "c"}, 0, FormatTable, true, "none",
			func(items []string, format string, noColor bool) error {
				printed = items
				return nil
			})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, printed)
	})

	t.Run("empty items prints info for table format", func(t *testing.T) {
		err := ApplyLimitAndPrint([]string{}, 0, FormatTable, true, "No items found.", noop)
		assert.NoError(t, err)
	})

	t.Run("empty items returns nil for non-table format", func(t *testing.T) {
		err := ApplyLimitAndPrint([]string{}, 0, "json", true, "No items found.", noop)
		assert.NoError(t, err)
	})

	t.Run("propagates print error with wrapping", func(t *testing.T) {
		err := ApplyLimitAndPrint([]string{"a"}, 0, FormatTable, true, "none",
			func(items []string, format string, noColor bool) error {
				return fmt.Errorf("print failed")
			})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error printing output")
		assert.Contains(t, err.Error(), "print failed")
	})

	t.Run("limit greater than items returns all", func(t *testing.T) {
		var printed []string
		err := ApplyLimitAndPrint([]string{"a", "b"}, 10, FormatTable, true, "none",
			func(items []string, format string, noColor bool) error {
				printed = items
				return nil
			})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, printed)
	})

	t.Run("nil slice behaves like empty", func(t *testing.T) {
		err := ApplyLimitAndPrint[string](nil, 0, FormatTable, true, "No items.", noop)
		assert.NoError(t, err)
	})
}

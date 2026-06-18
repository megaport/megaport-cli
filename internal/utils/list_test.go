package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
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

	t.Run("empty items prints info and skips printFunc for table format", func(t *testing.T) {
		called := false
		err := ApplyLimitAndPrint([]string{}, 0, FormatTable, true, "No items found.",
			func(items []string, format string, noColor bool) error {
				called = true
				return nil
			})
		assert.NoError(t, err)
		assert.False(t, called, "printFunc must not be called for empty table output")
	})

	t.Run("empty items still call printFunc for non-table formats", func(t *testing.T) {
		for _, format := range []string{FormatJSON, FormatCSV, FormatXML} {
			called := false
			err := ApplyLimitAndPrint([]string{}, 0, format, true, "No items found.",
				func(items []string, f string, noColor bool) error {
					called = true
					assert.Empty(t, items)
					return nil
				})
			assert.NoError(t, err)
			assert.True(t, called, "printFunc must be called for empty %s output", format)
		}
	})

	t.Run("propagates print error", func(t *testing.T) {
		err := ApplyLimitAndPrint([]string{"a"}, 0, FormatTable, true, "none",
			func(items []string, format string, noColor bool) error {
				return fmt.Errorf("print failed")
			})
		assert.Error(t, err)
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

type listItem struct {
	UID  string `json:"uid" header:"UID" csv:"uid"`
	Name string `json:"name" header:"Name" csv:"name"`
}

// TestApplyLimitAndPrint_EmptyOutputPerFormat verifies an empty result still
// emits a valid document for the machine formats instead of zero bytes.
func TestApplyLimitAndPrint_EmptyOutputPerFormat(t *testing.T) {
	printItems := func(items []listItem, format string, noColor bool) error {
		return output.PrintOutput(items, format, noColor)
	}

	t.Run("json emits empty array", func(t *testing.T) {
		out := output.CaptureOutput(func() {
			assert.NoError(t, ApplyLimitAndPrint([]listItem{}, 0, FormatJSON, true, "none", printItems))
		})
		assert.Equal(t, "[]", strings.TrimSpace(out))
	})

	t.Run("csv emits header only", func(t *testing.T) {
		out := output.CaptureOutput(func() {
			assert.NoError(t, ApplyLimitAndPrint([]listItem{}, 0, FormatCSV, true, "none", printItems))
		})
		assert.Equal(t, "uid,name", strings.TrimSpace(out))
	})

	t.Run("xml emits empty document", func(t *testing.T) {
		out := output.CaptureOutput(func() {
			assert.NoError(t, ApplyLimitAndPrint([]listItem{}, 0, FormatXML, true, "none", printItems))
		})
		assert.NotEmpty(t, strings.TrimSpace(out))
		assert.Contains(t, out, "<items>")
	})

	t.Run("table prints info message and no document", func(t *testing.T) {
		out := output.CaptureOutput(func() {
			assert.NoError(t, ApplyLimitAndPrint([]listItem{}, 0, FormatTable, true, "No items found.", printItems))
		})
		assert.Contains(t, out, "No items found.")
	})
}

package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayChanges(t *testing.T) {
	t.Run("prints changed fields", func(t *testing.T) {
		out := captureOutput(func() {
			DisplayChanges([]FieldChange{
				{Label: "Name", OldValue: "old", NewValue: "new"},
				{Label: "Term", OldValue: "12 months", NewValue: "12 months"}, // unchanged
				{Label: "Cost", OldValue: "(none)", NewValue: "IT"},
			}, true)
		})
		assert.Contains(t, out, "Name:")
		assert.Contains(t, out, "old")
		assert.Contains(t, out, "new")
		assert.Contains(t, out, "Cost:")
		assert.NotContains(t, out, "Term:")
	})

	t.Run("prints no changes detected when all equal", func(t *testing.T) {
		out := captureOutput(func() {
			DisplayChanges([]FieldChange{
				{Label: "Name", OldValue: "same", NewValue: "same"},
			}, true)
		})
		assert.Contains(t, out, "No changes detected")
	})

	t.Run("prints no changes detected for empty slice", func(t *testing.T) {
		out := captureOutput(func() {
			DisplayChanges(nil, true)
		})
		assert.Contains(t, out, "No changes detected")
	})
}

func TestFormatBool(t *testing.T) {
	assert.Equal(t, "Yes", FormatBool(true))
	assert.Equal(t, "No", FormatBool(false))
}

func TestFormatOptionalString(t *testing.T) {
	assert.Equal(t, "(none)", FormatOptionalString(""))
	assert.Equal(t, "IT", FormatOptionalString("IT"))
}

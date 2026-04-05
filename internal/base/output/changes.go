package output

import "fmt"

// FieldChange represents a single field that changed between two versions of a resource.
type FieldChange struct {
	Label    string
	OldValue string
	NewValue string
}

// DisplayChanges prints a formatted list of field changes for a resource update.
// Only changes where OldValue != NewValue are displayed. If no changes are found,
// a "No changes detected" message is printed.
func DisplayChanges(changes []FieldChange, noColor bool) {
	fmt.Println()
	PrintInfo("Changes applied:", noColor)

	found := false
	for _, c := range changes {
		if c.OldValue == c.NewValue {
			continue
		}
		found = true
		fmt.Printf("  • %s: %s → %s\n",
			c.Label,
			FormatOldValue(c.OldValue, noColor),
			FormatNewValue(c.NewValue, noColor))
	}

	if !found {
		fmt.Println("  No changes detected")
	}
}

// FormatBool returns "Yes" or "No" for a boolean value.
func FormatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// FormatOptionalString returns the string value or "(none)" if empty.
func FormatOptionalString(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

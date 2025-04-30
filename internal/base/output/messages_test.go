package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	if err != nil {
		panic(fmt.Sprintf("Failed to copy from pipe: %v", err))
	}
	return buf.String()
}

func TestPrintSuccess(t *testing.T) {
	tests := []struct {
		name             string
		format           string
		args             []interface{}
		noColor          bool
		expectedContains string
	}{
		{
			name:             "with color",
			format:           "Test %s message",
			args:             []interface{}{"success"},
			noColor:          false,
			expectedContains: "Test success message",
		},
		{
			name:             "without color",
			format:           "Test %s message",
			args:             []interface{}{"success"},
			noColor:          true,
			expectedContains: "✓ Test success message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintSuccess(tt.format, tt.noColor, tt.args...)
			})
			assert.Contains(t, output, tt.expectedContains)
		})
	}
}

func TestFormatSuccess(t *testing.T) {
	origNoColor := color.NoColor
	defer func() { color.NoColor = origNoColor }()

	color.NoColor = false
	result := FormatSuccess("test", false)
	assert.NotEqual(t, "successfully", result)

	result = FormatSuccess("test", true)
	assert.Equal(t, "successfully", result)
}

func TestPrintResourceSuccess(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		action       string
		uid          string
		noColor      bool
		expected     string
	}{
		{
			name:         "action with -ed suffix",
			resourceType: "Port",
			action:       "created",
			uid:          "port-123",
			noColor:      true,
			expected:     "✓ Port created port-123\n",
		},
		{
			name:         "action without -ed suffix",
			resourceType: "Port",
			action:       "create",
			uid:          "port-123",
			noColor:      true,
			expected:     "✓ Port created port-123\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintResourceSuccess(tt.resourceType, tt.action, tt.uid, tt.noColor)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestPrintResourceCreated(t *testing.T) {
	output := captureOutput(func() {
		PrintResourceCreated("Port", "port-123", true)
	})
	assert.Equal(t, "✓ Port created port-123\n", output)
}

func TestPrintResourceUpdated(t *testing.T) {
	output := captureOutput(func() {
		PrintResourceUpdated("Port", "port-123", true)
	})
	assert.Equal(t, "✓ Port updated port-123\n", output)
}

func TestPrintResourceDeleted(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		uid          string
		immediate    bool
		noColor      bool
		expected     string
	}{
		{
			name:         "delete immediate",
			resourceType: "Port",
			uid:          "port-123",
			immediate:    true,
			noColor:      true,
			expected:     "✓ Port deleted port-123\nThe resource will be deleted immediately\n",
		},
		{
			name:         "delete at end of billing period",
			resourceType: "Port",
			uid:          "port-123",
			immediate:    false,
			noColor:      true,
			expected:     "✓ Port deleted port-123\nThe resource will be deleted at the end of the current billing period\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintResourceDeleted(tt.resourceType, tt.uid, tt.immediate, tt.noColor)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestSpinner(t *testing.T) {
	spinner := NewSpinner(true)
	assert.NotNil(t, spinner)
	assert.Equal(t, 100*time.Millisecond, spinner.frameRate)
	assert.True(t, spinner.noColor)

	output := captureOutput(func() {
		spinner.Start("Testing spinner")
		time.Sleep(500 * time.Millisecond)
		spinner.Stop()
	})
	assert.NotEmpty(t, output)
}

func TestPrintResourceSpinners(t *testing.T) {
	tests := []struct {
		name         string
		function     func(string, string, bool) *Spinner
		resourceType string
		uid          string
		noColor      bool
		expected     string
	}{
		{
			name:         "creating spinner",
			function:     PrintResourceCreating,
			resourceType: "Port",
			uid:          "port-123",
			noColor:      true,
			expected:     "Creating Port port-123...",
		},
		{
			name:         "updating spinner",
			function:     PrintResourceUpdating,
			resourceType: "Port",
			uid:          "port-123",
			noColor:      true,
			expected:     "Updating Port port-123...",
		},
		{
			name:         "deleting spinner",
			function:     PrintResourceDeleting,
			resourceType: "Port",
			uid:          "port-123",
			noColor:      true,
			expected:     "Deleting Port port-123...",
		},
		{
			name:         "getting spinner",
			function:     PrintResourceGetting,
			resourceType: "Port",
			uid:          "port-123",
			noColor:      true,
			expected:     "Getting Port port-123 details...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				spinner := tt.function(tt.resourceType, tt.uid, tt.noColor)
				time.Sleep(200 * time.Millisecond)
				spinner.Stop()
			})
			assert.Contains(t, output, tt.expected)
		})
	}
}

func TestPrintResourceListing(t *testing.T) {
	output := captureOutput(func() {
		spinner := PrintResourceListing("Port", true)
		time.Sleep(200 * time.Millisecond)
		spinner.Stop()
	})
	assert.Contains(t, output, "Listing Ports...")
}

func TestPrintError(t *testing.T) {
	output := captureOutput(func() {
		PrintError("Test %s message", true, "error")
	})
	assert.Equal(t, "✗ Test error message\n", output)

	output = captureOutput(func() {
		PrintError("Test %s message", false, "error")
	})
	assert.Contains(t, output, "Test error message")
}

func TestPrintWarning(t *testing.T) {
	output := captureOutput(func() {
		PrintWarning("Test %s message", true, "warning")
	})
	assert.Equal(t, "⚠ Test warning message\n", output)

	output = captureOutput(func() {
		PrintWarning("Test %s message", false, "warning")
	})
	assert.Contains(t, output, "Test warning message")
}

func TestPrintInfo(t *testing.T) {
	output := captureOutput(func() {
		PrintInfo("Test %s message", true, "info")
	})
	assert.Equal(t, "ℹ Test info message\n", output)

	output = captureOutput(func() {
		PrintInfo("Test %s message", false, "info")
	})
	assert.Contains(t, output, "Test info message")
}

func TestFormatConfirmation(t *testing.T) {
	result := FormatConfirmation("Delete this resource?", true)
	assert.Equal(t, "Delete this resource? [y/N]", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatConfirmation("Delete this resource?", false)
	assert.Contains(t, result, "Delete this resource?")
	assert.Contains(t, result, "[y/N]")
	assert.NotEqual(t, "Delete this resource? [y/N]", result)
}

func TestFormatPrompt(t *testing.T) {
	result := FormatPrompt("Enter name:", true)
	assert.Equal(t, "Enter name:", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatPrompt("Enter name:", false)
	assert.NotEqual(t, "Enter name:", result)
	assert.Contains(t, StripANSIColors(result), "Enter name:")
}

func TestFormatExample(t *testing.T) {
	result := FormatExample("megaport-cli port list", true)
	assert.Equal(t, "megaport-cli port list", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatExample("megaport-cli port list", false)
	assert.NotEqual(t, "megaport-cli port list", result)
	assert.Contains(t, StripANSIColors(result), "megaport-cli port list")
}

func TestFormatCommandName(t *testing.T) {
	result := FormatCommandName("port list", true)
	assert.Equal(t, "port list", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatCommandName("port list", false)
	assert.NotEqual(t, "port list", result)
	assert.Contains(t, StripANSIColors(result), "port list")
}

func TestFormatRequiredFlag(t *testing.T) {
	result := FormatRequiredFlag("--name", "Name of the resource", true)
	assert.Equal(t, "--name (REQUIRED): Name of the resource", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatRequiredFlag("--name", "Name of the resource", false)
	assert.NotEqual(t, "--name (REQUIRED): Name of the resource", result)
	assert.Contains(t, StripANSIColors(result), "--name (REQUIRED): Name of the resource")
}

func TestFormatOptionalFlag(t *testing.T) {
	result := FormatOptionalFlag("--location-id", "ID of the location", true)
	assert.Equal(t, "--location-id: ID of the location", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatOptionalFlag("--location-id", "ID of the location", false)
	assert.NotEqual(t, "--location-id: ID of the location", result)
	assert.Contains(t, StripANSIColors(result), "--location-id: ID of the location")
}

func TestFormatJSONExample(t *testing.T) {
	json := `{"name": "test"}`
	result := FormatJSONExample(json, true)
	assert.Equal(t, json, result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatJSONExample(json, false)
	assert.NotEqual(t, json, result)
	assert.Contains(t, StripANSIColors(result), json)
}

func TestFormatUID(t *testing.T) {
	result := FormatUID("port-123", true)
	assert.Equal(t, "port-123", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatUID("port-123", false)
	assert.NotEqual(t, "port-123", result)
	assert.Contains(t, StripANSIColors(result), "port-123")
}

func TestStripANSIColors(t *testing.T) {
	coloredString := "\x1b[31mRed\x1b[0m \x1b[32mGreen\x1b[0m"
	result := StripANSIColors(coloredString)
	assert.Equal(t, "Red Green", result)
}

func TestFormatOldValue(t *testing.T) {
	result := FormatOldValue("old-value", true)
	assert.Equal(t, "old-value", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatOldValue("old-value", false)
	assert.NotEqual(t, "old-value", result)
	assert.Contains(t, StripANSIColors(result), "old-value")
}

func TestFormatNewValue(t *testing.T) {
	result := FormatNewValue("new-value", true)
	assert.Equal(t, "new-value", result)

	oldNoColor := color.NoColor
	color.NoColor = false
	defer func() { color.NoColor = oldNoColor }()

	result = FormatNewValue("new-value", false)
	assert.NotEqual(t, "new-value", result)
	assert.Contains(t, StripANSIColors(result), "new-value")
}

func TestSpinnerStopWithSuccess(t *testing.T) {
	output := captureOutput(func() {
		spinner := NewSpinner(true)
		spinner.Start("Testing")
		time.Sleep(200 * time.Millisecond)
		spinner.StopWithSuccess("Operation completed")
	})
	assert.Contains(t, output, "✓ Operation completed")
}

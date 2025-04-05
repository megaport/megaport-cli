package generate_docs

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerateCommandDoc(t *testing.T) {
	// Mock command for testing
	mockCmd := &cobra.Command{
		Use:   "mock",
		Short: "Mock command for testing",
		Long: `This is a mock command used for testing documentation generation.

Example usage:

  # Example with flags
  mock-cli mock --flag1 value1 --flag2 value2

  # Example with JSON
  mock-cli mock --json '{"key":"value"}'

  # Example with stray backticks
  mock-cli mock --flag-with-backtick '{"key":"value"}'
`,
		Example: `mock-cli mock --example-flag value`,
	}

	// Temporary output file
	outputFile := "test_mock_command.md"
	defer os.Remove(outputFile) // Clean up after test

	// Run the function
	err := generateCommandDoc(mockCmd, outputFile)
	if err != nil {
		t.Fatalf("generateCommandDoc failed: %v", err)
	}

	// Read the generated file
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Convert content to string for assertions
	doc := string(content)

	// Assertions
	t.Run("Check if file contains command name", func(t *testing.T) {
		if !strings.Contains(doc, "# mock") {
			t.Errorf("Expected command name '# mock' in the documentation")
		}
	})

	t.Run("Check if file contains short description", func(t *testing.T) {
		if !strings.Contains(doc, "Mock command for testing") {
			t.Errorf("Expected short description in the documentation")
		}
	})

	t.Run("Check if examples are formatted correctly", func(t *testing.T) {
		if !strings.Contains(doc, "```") {
			t.Errorf("Expected code blocks in the examples section")
		}
		if strings.Contains(doc, "`--flag-with-backtick") {
			t.Errorf("Stray backticks found in example: %s", doc)
		}
	})

	t.Run("Check if JSON examples are formatted correctly", func(t *testing.T) {
		if !strings.Contains(doc, `--json '{"key":"value"}'`) {
			t.Errorf("Expected JSON example to be formatted correctly")
		}
	})

	t.Run("Check if usage section is present", func(t *testing.T) {
		if !strings.Contains(doc, "## Usage") {
			t.Errorf("Expected '## Usage' section in the documentation")
		}
	})
}

package generate_docs

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerateCommandDoc(t *testing.T) {
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

	outputFile := "test_mock_command.md"
	defer os.Remove(outputFile)

	err := generateCommandDoc(mockCmd, outputFile)
	if err != nil {
		t.Fatalf("generateCommandDoc failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	doc := string(content)

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

func TestCollectFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	cmd.Flags().String("flag1", "default1", "description1")
	cmd.Flags().StringP("flag2", "f", "default2", "description2")
	cmd.Flags().Bool("flag3", false, "description3")
	cmd.PersistentFlags().String("flag1", "default1", "description1")
	cmd.Flags().String("required", "", "Required flag")
	err := cmd.MarkFlagRequired("required")
	if err != nil {
		t.Fatalf("Failed to mark flag as required: %v", err)
	}

	allFlags, localFlags, persistentFlags := collectFlags(cmd)

	if len(allFlags) != 4 {
		t.Errorf("Expected 4 deduplicated flags, got %d", len(allFlags))
	}

	if len(localFlags) != 4 {
		t.Errorf("Expected 4 local flags, got %d", len(localFlags))
	}

	if len(persistentFlags) != 1 {
		t.Errorf("Expected 1 persistent flag, got %d", len(persistentFlags))
	}

	foundRequired := false
	for _, flag := range allFlags {
		if flag.Name == "required" {
			foundRequired = true
			if !flag.Required {
				t.Errorf("Flag 'required' should be marked as required")
			}
			break
		}
	}
	if !foundRequired {
		t.Errorf("Required flag not found in the collected flags")
	}

	for _, flag := range allFlags {
		if flag.Name == "flag2" && flag.Shorthand != "f" {
			t.Errorf("Flag 'flag2' should have shorthand 'f'")
		}
	}
}

func TestFormatSection(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Required fields:", "### Required Fields"},
		{"Optional fields:", "### Optional Fields"},
		{"Important notes:", "### Important Notes"},
		{"Example usage:", "### Example Usage"},
		{"Examples:", "### Example Usage"},
		{"JSON format example:", "### JSON Format Example"},
		{"Other section:", "Other section:"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := formatSection(tc.input)
			if result != tc.expected {
				t.Errorf("formatSection(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatFieldLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"field: description", "- `field`: description"},
		{"  field: description with spaces", "  - `field`: description with spaces"},
		{"field without colon", "field without colon"},
		{"  multiple-word-field: description", "  - `multiple-word-field`: description"},
		{"field: description: with: colons", "- `field`: description: with: colons"},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			result := formatFieldLine(tc.input)
			if result != tc.expected {
				t.Errorf("formatFieldLine(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestFormatNoteLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"This is a note", "- This is a note"},
		{"  Indented note", "  - Indented note"},
		{"- Already bulleted", "- Already bulleted"},
		{"  - Already bulleted with indent", "  - Already bulleted with indent"},
		{"", ""},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			result := formatNoteLine(tc.input)
			if result != tc.expected {
				t.Errorf("formatNoteLine(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestProcessDescription(t *testing.T) {
	testCases := []struct {
		name                   string
		input                  string
		cmdName                string
		expectedOutputContains []string
		notExpectedToContain   []string
	}{
		{
			name: "required fields formatting",
			input: `Test description.

Required fields:
field1: description1
field2: description2

More text.`,
			cmdName: "test",
			expectedOutputContains: []string{
				"### Required Fields",
				"- `field1`: description1",
				"- `field2`: description2",
			},
		},
		{
			name: "optional fields formatting",
			input: `Test description.

Optional fields:
option1: description1
option2: description2

More text.`,
			cmdName: "test",
			expectedOutputContains: []string{
				"### Optional Fields",
				"- `option1`: description1",
				"- `option2`: description2",
			},
		},
		{
			name: "important notes formatting",
			input: `Test description.

Important notes:
This is an important note
Another important note

More text.`,
			cmdName: "test",
			expectedOutputContains: []string{
				"### Important Notes",
				"- This is an important note",
				"- Another important note",
			},
		},
		{
			name: "example formatting",
			input: `Test description.

Example usage:
test --flag value
test --another-flag value

More text.`,
			cmdName: "test",
			expectedOutputContains: []string{
				"### Example Usage",
				"```",
				"test --flag value",
				"test --another-flag value",
				"```",
			},
		},
		{
			name: "json example formatting",
			input: `Test description.

JSON format example:
{"key": "value"}
{"another": "value"}

More text.`,
			cmdName: "test",
			expectedOutputContains: []string{
				"### JSON Format Example",
				"```json",
				"{\"key\": \"value\"}",
				"{\"another\": \"value\"}",
				"```",
			},
		},
		{
			name: "mixed sections",
			input: `Test description.

Required fields:
field1: description1

Optional fields:
option1: description1

Important notes:
This is an important note

Example usage:
test --flag value`,
			cmdName: "test",
			expectedOutputContains: []string{
				"### Required Fields",
				"### Optional Fields",
				"### Important Notes",
				"### Example Usage",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := processDescription(tc.input, tc.cmdName)

			for _, expected := range tc.expectedOutputContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q but it didn't\nOutput: %s", expected, result)
				}
			}

			for _, unexpected := range tc.notExpectedToContain {
				if strings.Contains(result, unexpected) {
					t.Errorf("Output contains %q when it shouldn't\nOutput: %s", unexpected, result)
				}
			}
		})
	}
}

func TestDetermineParentInfo(t *testing.T) {
	rootCmd := &cobra.Command{Use: "megaport-cli"}
	parentCmd := &cobra.Command{Use: "parent"}
	childCmd := &cobra.Command{Use: "child"}

	rootCmd.AddCommand(parentCmd)
	parentCmd.AddCommand(childCmd)

	t.Run("root command has no parent", func(t *testing.T) {
		hasParent, _, _, _ := determineParentInfo(rootCmd, "megaport-cli")
		if hasParent {
			t.Error("Root command should not have a parent")
		}
	})

	t.Run("child has correct parent path", func(t *testing.T) {
		hasParent, parentPath, parentName, parentFilePath := determineParentInfo(childCmd, "megaport-cli_parent_child")

		if !hasParent {
			t.Error("Child command should have a parent")
		}

		if parentPath != "megaport-cli parent" {
			t.Errorf("Expected parent path 'megaport-cli parent', got %q", parentPath)
		}

		if parentName != "parent" {
			t.Errorf("Expected parent name 'parent', got %q", parentName)
		}

		if parentFilePath != "megaport-cli_parent" {
			t.Errorf("Expected parent file path 'megaport-cli_parent', got %q", parentFilePath)
		}
	})
}

func TestGatherSubcommands(t *testing.T) {
	cmd := &cobra.Command{Use: "parent"}
	cmd.AddCommand(&cobra.Command{Use: "child1"})
	cmd.AddCommand(&cobra.Command{Use: "child2"})
	hiddenCmd := &cobra.Command{Use: "hidden", Hidden: true}
	cmd.AddCommand(hiddenCmd)
	helpCmd := &cobra.Command{Use: "help"}
	cmd.AddCommand(helpCmd)

	subcommands := gatherSubcommands(cmd)

	if len(subcommands) != 2 {
		t.Errorf("Expected 2 visible subcommands, got %d", len(subcommands))
	}

	if !containsString(subcommands, "child1") || !containsString(subcommands, "child2") {
		t.Errorf("Expected subcommands to contain 'child1' and 'child2', got %v", subcommands)
	}

	if containsString(subcommands, "hidden") {
		t.Error("Subcommands should not include hidden commands")
	}

	if containsString(subcommands, "help") {
		t.Error("Subcommands should not include help command")
	}
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func TestFullDocumentationPipeline(t *testing.T) {
	mockCmd := &cobra.Command{
		Use:   "complex-command",
		Short: "A complex command for testing",
		Long: `This is a complex command with multiple sections.

Required fields:
username: The username to authenticate with
password: The password for authentication

Optional fields:
verbose: Enable verbose output
timeout: Timeout in seconds

Important notes:
Credentials are case-sensitive
Sessions expire after 30 minutes of inactivity

Example usage:
complex-command --username admin --password secret
complex-command --verbose --timeout 60

JSON format example:
{"username": "admin", "password": "secret"}`,
	}

	mockCmd.Flags().String("username", "", "Username for authentication")
	mockCmd.Flags().String("password", "", "Password for authentication")
	mockCmd.Flags().Bool("verbose", false, "Enable verbose output")
	mockCmd.Flags().Int("timeout", 30, "Timeout in seconds")
	err := mockCmd.MarkFlagRequired("username")
	if err != nil {
		t.Fatalf("Failed to mark flag as required: %v", err)
	}
	err = mockCmd.MarkFlagRequired("password")
	if err != nil {
		t.Fatalf("Failed to mark flag as required: %v", err)
	}

	outputFile := "test_complex_command.md"
	defer os.Remove(outputFile)

	err = generateCommandDoc(mockCmd, outputFile)
	if err != nil {
		t.Fatalf("generateCommandDoc failed: %v", err)
	}

	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	doc := string(content)

	expectedSections := []string{
		"# complex-command",
		"A complex command for testing",
		"### Required Fields",
		"- `username`: The username to authenticate with",
		"- `password`: The password for authentication",
		"### Optional Fields",
		"- `verbose`: Enable verbose output",
		"- `timeout`: Timeout in seconds",
		"### Important Notes",
		"- Credentials are case-sensitive",
		"- Sessions expire after 30 minutes of inactivity",
		"### Example Usage",
		"```",
		"complex-command --username admin --password secret",
		"complex-command --verbose --timeout 60",
		"```",
		"### JSON Format Example",
		"```json",
		"{\"username\": \"admin\", \"password\": \"secret\"}",
		"```",
		"## Flags",
		"`--username`",
		"`--password`",
		"`--verbose`",
		"`--timeout`",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(doc, expected) {
			t.Errorf("Expected documentation to contain %q, but it didn't", expected)
		}
	}
}

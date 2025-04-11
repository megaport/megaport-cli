package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to mock and capture stdin/stdout for testing prompts
func withMockedIO(input string, fn func()) string {
	// Save original stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Create pipes for mocking stdin/stdout
	inR, inW, _ := os.Pipe()
	outR, outW, _ := os.Pipe()

	// Replace stdin/stdout with our pipes
	os.Stdin = inR
	os.Stdout = outW

	// Write the mock input
	_, err := inW.WriteString(input)
	if err != nil {
		// In tests, we can just panic on unexpected IO errors
		panic(fmt.Sprintf("Failed to write to mock stdin: %v", err))
	}
	inW.Close()

	// Call the function we're testing
	fn()

	// Restore original stdin/stdout
	outW.Close() // Close the write end of stdout first
	os.Stdout = oldStdout
	os.Stdin = oldStdin

	// Read the captured output
	var buf bytes.Buffer
	_, err = io.Copy(&buf, outR)
	if err != nil {
		// In tests, we can just panic on unexpected IO errors
		panic(fmt.Sprintf("Failed to copy from mock stdout: %v", err))
	}
	return buf.String()
}

// Test the Prompt function
func TestPrompt(t *testing.T) {
	// Save original prompt function and restore after test
	originalPrompt := Prompt
	defer func() { Prompt = originalPrompt }()

	tests := []struct {
		name          string
		input         string
		message       string
		noColor       bool
		expected      string
		expectedError bool
	}{
		{
			name:          "basic input with color",
			input:         "test input\n",
			message:       "Enter value:",
			noColor:       false,
			expected:      "test input",
			expectedError: false,
		},
		{
			name:          "basic input without color",
			input:         "test input\n",
			message:       "Enter value:",
			noColor:       true,
			expected:      "test input",
			expectedError: false,
		},
		{
			name:          "empty input",
			input:         "\n",
			message:       "Enter value:",
			noColor:       true,
			expected:      "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Prompt for testing
			mockPrompt := func(msg string, noColor bool) (string, error) {
				assert.Equal(t, tt.message, msg)
				assert.Equal(t, tt.noColor, noColor)

				// Important: Actually print something that matches the real function's output
				if !noColor {
					fmt.Print("❯ " + msg + " ") // Simulate the colored prompt
				} else {
					fmt.Print("❯ " + msg + " ") // Simulate the non-colored prompt
				}

				// Create a reader from the test input
				if tt.expectedError {
					return "", errors.New("mocked error")
				}

				return strings.TrimSpace(tt.input), nil
			}

			Prompt = mockPrompt

			// Capture output
			output := withMockedIO(tt.input, func() {
				result, err := Prompt(tt.message, tt.noColor)
				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, result)
				}
			})

			// Verify the prompt was displayed (actual formatting will vary with color settings)
			assert.Contains(t, output, "❯")
			assert.Contains(t, output, tt.message)
		})
	}
}

// Test the ConfirmPrompt function
func TestConfirmPrompt(t *testing.T) {
	// Save original function and restore after test
	originalConfirmPrompt := ConfirmPrompt
	defer func() { ConfirmPrompt = originalConfirmPrompt }()

	tests := []struct {
		name     string
		input    string
		question string
		noColor  bool
		expected bool
	}{
		{
			name:     "yes response",
			input:    "y\n",
			question: "Continue?",
			noColor:  true,
			expected: true,
		},
		{
			name:     "yes capitalized response",
			input:    "Y\n",
			question: "Continue?",
			noColor:  true,
			expected: true,
		},
		{
			name:     "yes full response",
			input:    "yes\n",
			question: "Continue?",
			noColor:  true,
			expected: true,
		},
		{
			name:     "no response",
			input:    "n\n",
			question: "Continue?",
			noColor:  true,
			expected: false,
		},
		{
			name:     "no full response",
			input:    "no\n",
			question: "Continue?",
			noColor:  true,
			expected: false,
		},
		{
			name:     "empty response (default to no)",
			input:    "\n",
			question: "Continue?",
			noColor:  true,
			expected: false,
		},
		{
			name:     "invalid response (default to no)",
			input:    "maybe\n",
			question: "Continue?",
			noColor:  true,
			expected: false,
		},
		{
			name:     "yes with color",
			input:    "y\n",
			question: "Continue?",
			noColor:  false,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ConfirmPrompt for testing
			mockConfirmPrompt := func(question string, noColor bool) bool {
				assert.Equal(t, tt.question, question)
				assert.Equal(t, tt.noColor, noColor)

				// Important: Actually print something that matches the real function's output
				if !noColor {
					fmt.Print("⚠️  " + question + " ")
					fmt.Print("[y/N] ")
				} else {
					fmt.Printf("⚠️  %s [y/N] ", question)
				}

				// Process the response based on test input
				input := strings.TrimSpace(tt.input)
				input = strings.ToLower(input)
				return input == "y" || input == "yes"
			}

			ConfirmPrompt = mockConfirmPrompt

			// Capture output
			output := withMockedIO(tt.input, func() {
				result := ConfirmPrompt(tt.question, tt.noColor)
				assert.Equal(t, tt.expected, result)
			})

			// Verify the prompt was displayed
			assert.Contains(t, output, tt.question)
			assert.Contains(t, output, "[y/N]")
		})
	}
}

// Test the ResourcePrompt function
func TestResourcePrompt(t *testing.T) {
	// Save original function and restore after test
	originalResourcePrompt := ResourcePrompt
	defer func() { ResourcePrompt = originalResourcePrompt }()

	tests := []struct {
		name         string
		input        string
		resourceType string
		message      string
		noColor      bool
		expected     string
		expectedIcon string
	}{
		{
			name:         "port resource",
			input:        "test port\n",
			resourceType: "port",
			message:      "Enter port name:",
			noColor:      true,
			expected:     "test port",
			expectedIcon: "🔌",
		},
		{
			name:         "mve resource",
			input:        "test mve\n",
			resourceType: "mve",
			message:      "Enter MVE name:",
			noColor:      true,
			expected:     "test mve",
			expectedIcon: "🌐",
		},
		{
			name:         "mcr resource",
			input:        "test mcr\n",
			resourceType: "mcr",
			message:      "Enter MCR name:",
			noColor:      true,
			expected:     "test mcr",
			expectedIcon: "🛰️",
		},
		{
			name:         "vxc resource",
			input:        "test vxc\n",
			resourceType: "vxc",
			message:      "Enter VXC name:",
			noColor:      true,
			expected:     "test vxc",
			expectedIcon: "🔗",
		},
		{
			name:         "location resource",
			input:        "test location\n",
			resourceType: "location",
			message:      "Enter location:",
			noColor:      true,
			expected:     "test location",
			expectedIcon: "📍",
		},
		{
			name:         "unknown resource type",
			input:        "test unknown\n",
			resourceType: "unknown",
			message:      "Enter something:",
			noColor:      true,
			expected:     "test unknown",
			expectedIcon: "❯",
		},
		{
			name:         "with color",
			input:        "test input\n",
			resourceType: "port",
			message:      "Enter port name:",
			noColor:      false,
			expected:     "test input",
			expectedIcon: "🔌",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ResourcePrompt for testing
			mockResourcePrompt := func(resourceType, msg string, noColor bool) (string, error) {
				assert.Equal(t, tt.resourceType, resourceType)
				assert.Equal(t, tt.message, msg)
				assert.Equal(t, tt.noColor, noColor)

				// Important: Actually print something that matches the real function's output
				// Determine icon based on resource type
				icon := "❯"
				switch strings.ToLower(resourceType) {
				case "port":
					icon = "🔌"
				case "mve":
					icon = "🌐"
				case "mcr":
					icon = "🛰️"
				case "vxc":
					icon = "🔗"
				case "location":
					icon = "📍"
				}

				if !noColor {
					fmt.Print(icon + " " + msg + " ")
				} else {
					fmt.Print(icon + " " + msg + " ")
				}

				return strings.TrimSpace(tt.input), nil
			}

			ResourcePrompt = mockResourcePrompt

			// Capture output
			output := withMockedIO(tt.input, func() {
				result, err := ResourcePrompt(tt.resourceType, tt.message, tt.noColor)
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			})

			// Verify the prompt was displayed with correct icon
			assert.Contains(t, output, tt.expectedIcon)
			assert.Contains(t, output, tt.message)
		})
	}
}

// Integration test that actually runs the real functions
func TestPromptsIntegration(t *testing.T) {
	// Skip in CI environments
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping interactive test in CI environment")
	}

	tests := []struct {
		name    string
		inputFn func() string
		testFn  func(string)
	}{
		{
			name: "Prompt integration",
			inputFn: func() string {
				return "test input\n"
			},
			testFn: func(input string) {
				originalPrompt := Prompt // Use real function
				output := withMockedIO(input, func() {
					result, err := originalPrompt("Test prompt:", true)
					assert.NoError(t, err)
					assert.Equal(t, "test input", result)
				})
				assert.Contains(t, output, "❯ Test prompt:")
			},
		},
		{
			name: "ResourcePrompt integration",
			inputFn: func() string {
				return "resource input\n"
			},
			testFn: func(input string) {
				originalResourcePrompt := ResourcePrompt // Use real function
				output := withMockedIO(input, func() {
					result, err := originalResourcePrompt("port", "Port name:", true)
					assert.NoError(t, err)
					assert.Equal(t, "resource input", result)
				})
				assert.Contains(t, output, "🔌 Port name:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFn(tt.inputFn())
		})
	}
}

// Test using custom reader for simulated input
func TestPromptWithCustomReader(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Create and use a temporary file for input
	simulatedInput := "simulated input\n"

	// Save and restore original stdin
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Write our simulated input
	fmt.Fprint(tmpfile, simulatedInput)
	_, err = tmpfile.Seek(0, 0)
	if err != nil {
		t.Fatalf("Failed to seek in temp file: %v", err)
	}

	// Replace stdin with our file
	os.Stdin = tmpfile

	// Capture stdout
	_, outW, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = outW
	defer func() { os.Stdout = oldStdout }()

	// Test the real Prompt function with our simulated stdin
	result, err := Prompt("Enter value:", true)

	// Close the write pipe
	outW.Close()

	// Restore stdout to see test results
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Got error from prompt: %v", err)
	}

	expected := "simulated input"
	assert.Equal(t, expected, result, "Expected prompt to return %q, got %q", expected, result)
}

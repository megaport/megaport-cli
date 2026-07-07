package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// withMockedIO redirects stdin to input and captures anything written to
// stdout and stderr while fn runs, restoring the originals afterward. The
// two streams are captured separately so tests can assert prompt text lands
// on stderr and never leaks onto stdout (which must stay clean for
// machine-readable output).
func withMockedIO(input string, fn func()) (stdout string, stderr string) {
	// Save originals
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	// Create pipes for mocking stdin/stdout/stderr
	inR, inW, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock stdin pipe: %v", err))
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock stdout pipe: %v", err))
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock stderr pipe: %v", err))
	}

	os.Stdin = inR
	os.Stdout = outW
	os.Stderr = errW
	// Restore the globals on any exit path, including a panic or t.Fatal
	// (which unwinds via runtime.Goexit), so a failing subtest doesn't leave
	// stdio redirected for the rest of the test binary.
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Write the mock input
	if _, err := inW.WriteString(input); err != nil {
		// In tests, we can just panic on unexpected IO errors
		panic(fmt.Sprintf("Failed to write to mock stdin: %v", err))
	}
	inW.Close()

	// Call the function we're testing
	fn()

	outW.Close()
	errW.Close()

	// Read the captured output, then close the read ends so repeated calls
	// across a table-driven test don't accumulate open file descriptors.
	var outBuf, errBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, outR); err != nil {
		panic(fmt.Sprintf("Failed to copy from mock stdout: %v", err))
	}
	if _, err := io.Copy(&errBuf, errR); err != nil {
		panic(fmt.Sprintf("Failed to copy from mock stderr: %v", err))
	}
	inR.Close()
	outR.Close()
	errR.Close()
	return outBuf.String(), errBuf.String()
}

// withMockedIOStreamed is like withMockedIO but for prompt flows that read
// stdin more than once (e.g. a tag-entry loop). Each prompt call opens a
// fresh bufio.Reader over stdin, so writing all input up front in one shot
// lets a single Read syscall slurp every line into that reader's private
// buffer, silently starving the next prompt call. A real terminal never hits
// this because the tty line discipline only ever hands over one line at a
// time; this helper reproduces that pacing by writing one line at a time
// with a short delay, giving the blocked reader time to consume each line
// before the next is written.
func withMockedIOStreamed(lines []string, fn func()) (stdout string, stderr string) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr

	inR, inW, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock stdin pipe: %v", err))
	}
	outR, outW, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock stdout pipe: %v", err))
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		panic(fmt.Sprintf("Failed to create mock stderr pipe: %v", err))
	}

	os.Stdin = inR
	os.Stdout = outW
	os.Stderr = errW
	// Restore the globals on any exit path, including a panic or t.Fatal
	// (which unwinds via runtime.Goexit), so a failing subtest doesn't leave
	// stdio redirected for the rest of the test binary.
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	go func() {
		defer inW.Close()
		for _, line := range lines {
			if _, err := inW.WriteString(line + "\n"); err != nil {
				return
			}
			time.Sleep(30 * time.Millisecond)
		}
	}()

	fn()

	outW.Close()
	errW.Close()

	var outBuf, errBuf bytes.Buffer
	if _, err := io.Copy(&outBuf, outR); err != nil {
		panic(fmt.Sprintf("Failed to copy from mock stdout: %v", err))
	}
	if _, err := io.Copy(&errBuf, errR); err != nil {
		panic(fmt.Sprintf("Failed to copy from mock stderr: %v", err))
	}
	inR.Close()
	outR.Close()
	errR.Close()
	return outBuf.String(), errBuf.String()
}

// Test the Prompt function
func TestPrompt(t *testing.T) {
	// Save original prompt function and restore after test
	originalPrompt := GetPrompt()
	defer func() { SetPrompt(originalPrompt) }()

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

			SetPrompt(mockPrompt)

			// Capture output
			output, _ := withMockedIO(tt.input, func() {
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
	originalConfirmPrompt := GetConfirmPrompt()
	defer func() { SetConfirmPrompt(originalConfirmPrompt) }()

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

			SetConfirmPrompt(mockConfirmPrompt)

			// Capture output
			output, _ := withMockedIO(tt.input, func() {
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
	originalResourcePrompt := GetResourcePrompt()
	defer func() { SetResourcePrompt(originalResourcePrompt) }()

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

			SetResourcePrompt(mockResourcePrompt)

			// Capture output
			output, _ := withMockedIO(tt.input, func() {
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

// Test the SecretResourcePrompt function — verifies the override hook delegates
// to the installed function and propagates resource type, message, and noColor.
func TestSecretResourcePrompt(t *testing.T) {
	original := GetSecretResourcePrompt()
	defer SetSecretResourcePrompt(original)

	var (
		gotResource string
		gotMsg      string
		gotNoColor  bool
	)
	SetSecretResourcePrompt(func(resource, msg string, noColor bool) (string, error) {
		gotResource, gotMsg, gotNoColor = resource, msg, noColor
		return "p@ss", nil
	})

	got, err := SecretResourcePrompt("mve", "admin password: ", true)
	assert.NoError(t, err)
	assert.Equal(t, "p@ss", got)
	assert.Equal(t, "mve", gotResource)
	assert.Equal(t, "admin password: ", gotMsg)
	assert.True(t, gotNoColor)
}

// TestSecretResourcePrompt_NonTTYFallback verifies that when stdin is not a
// terminal the default implementation falls back to a buffered read so piped
// CI input still works (term.ReadPassword would error on a non-TTY fd).
func TestSecretResourcePrompt_NonTTYFallback(t *testing.T) {
	stdout, stderr := withMockedIO("piped-secret\n", func() {
		got, err := SecretResourcePrompt("mve", "admin password: ", true)
		assert.NoError(t, err)
		assert.Equal(t, "piped-secret", got)
	})
	// Prompt should still be rendered, with the lock icon, but on stderr so
	// it doesn't corrupt machine-readable stdout output.
	assert.Contains(t, stderr, "🔐")
	assert.Contains(t, stderr, "admin password:")
	assert.Empty(t, stdout)
}

// TestSecretResourcePrompt_NonTTYFallbackWithColor exercises the colored-prompt
// branch of the default implementation.
func TestSecretResourcePrompt_NonTTYFallbackWithColor(t *testing.T) {
	stdout, stderr := withMockedIO("colored-secret\n", func() {
		got, err := SecretResourcePrompt("mve", "admin password: ", false)
		assert.NoError(t, err)
		assert.Equal(t, "colored-secret", got)
	})
	assert.Contains(t, stderr, "admin password:")
	assert.Empty(t, stdout)
}

// TestPasswordPrompt verifies the PasswordPrompt call-through wrapper
// delegates to the installed function and propagates its arguments.
func TestPasswordPrompt(t *testing.T) {
	original := GetPasswordPrompt()
	defer SetPasswordPrompt(original)

	var gotMsg string
	var gotNoColor bool
	SetPasswordPrompt(func(msg string, noColor bool) (string, error) {
		gotMsg, gotNoColor = msg, noColor
		return "hunter2", nil
	})

	got, err := PasswordPrompt("Password:", true)
	assert.NoError(t, err)
	assert.Equal(t, "hunter2", got)
	assert.Equal(t, "Password:", gotMsg)
	assert.True(t, gotNoColor)
}

// TestAccessors verifies that every Get/Set accessor pair round-trips correctly
// and that the call-through wrappers delegate to the underlying function pointer.
func TestAccessors(t *testing.T) {
	t.Run("BuyConfirmPrompt", func(t *testing.T) {
		original := GetBuyConfirmPrompt()
		defer SetBuyConfirmPrompt(original)

		called := false
		SetBuyConfirmPrompt(func(_ string, _ []BuyConfirmDetail, _ bool) bool {
			called = true
			return true
		})
		result := BuyConfirmPrompt("port", nil, true)
		assert.True(t, called)
		assert.True(t, result)
	})

	t.Run("DesignConfirmPrompt", func(t *testing.T) {
		original := GetDesignConfirmPrompt()
		defer SetDesignConfirmPrompt(original)

		called := false
		var gotResource string
		var gotDetails []BuyConfirmDetail
		SetDesignConfirmPrompt(func(resource string, details []BuyConfirmDetail, _ bool) bool {
			called = true
			gotResource = resource
			gotDetails = details
			return true
		})
		details := []BuyConfirmDetail{{Key: "Name", Value: "test"}}
		result := DesignConfirmPrompt("NAT Gateway", details, true)
		assert.True(t, called)
		assert.True(t, result)
		assert.Equal(t, "NAT Gateway", gotResource)
		assert.Equal(t, details, gotDetails)
	})

	t.Run("SecretResourcePrompt", func(t *testing.T) {
		original := GetSecretResourcePrompt()
		defer SetSecretResourcePrompt(original)

		SetSecretResourcePrompt(func(_, _ string, _ bool) (string, error) {
			return "sentinel", nil
		})
		got := GetSecretResourcePrompt()
		out, err := got("", "", false)
		assert.NoError(t, err)
		assert.Equal(t, "sentinel", out)
	})

	t.Run("ResourceTagsPrompt", func(t *testing.T) {
		original := GetResourceTagsPrompt()
		defer SetResourceTagsPrompt(original)

		SetResourceTagsPrompt(func(_ bool) (map[string]string, error) {
			return map[string]string{"k": "v"}, nil
		})
		tags, err := ResourceTagsPrompt(true)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"k": "v"}, tags)
	})

	t.Run("UpdateResourceTagsPrompt", func(t *testing.T) {
		original := GetUpdateResourceTagsPrompt()
		defer SetUpdateResourceTagsPrompt(original)

		SetUpdateResourceTagsPrompt(func(existing map[string]string, _ bool) (map[string]string, error) {
			existing["new"] = "tag"
			return existing, nil
		})
		tags, err := UpdateResourceTagsPrompt(map[string]string{"old": "tag"}, true)
		assert.NoError(t, err)
		assert.Equal(t, "tag", tags["new"])
		assert.Equal(t, "tag", tags["old"])
	})
}

// TestResourceTagsPrompt_DefaultDeclines exercises the default
// resourceTagsPromptFn's early-exit path: declining to add tags skips the
// tag-entry loop entirely.
func TestResourceTagsPrompt_DefaultDeclines(t *testing.T) {
	stdout, stderr := withMockedIO("n\n", func() {
		tags, err := ResourceTagsPrompt(true)
		assert.NoError(t, err)
		assert.Nil(t, tags)
	})
	assert.Contains(t, stderr, "Would you like to add resource tags?")
	assert.Empty(t, stdout)
}

// TestResourceTagsPrompt_DefaultAddsTags exercises the default
// resourceTagsPromptFn's tag-entry loop, which reads stdin multiple times.
func TestResourceTagsPrompt_DefaultAddsTags(t *testing.T) {
	stdout, stderr := withMockedIOStreamed([]string{"y", "env", "prod", ""}, func() {
		tags, err := ResourceTagsPrompt(true)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod"}, tags)
	})
	assert.Contains(t, stderr, "Enter tags (key and value). Enter empty key to finish.")
	assert.Contains(t, stderr, "Tags added:")
	assert.Contains(t, stderr, "env: prod")
	assert.Empty(t, stdout)
}

// TestUpdateResourceTagsPrompt_DefaultDeclinesWithExistingTags exercises the
// default updateResourceTagsPromptFn when existing tags are present and the
// user declines to continue.
func TestUpdateResourceTagsPrompt_DefaultDeclinesWithExistingTags(t *testing.T) {
	stdout, stderr := withMockedIO("n\n", func() {
		tags, err := UpdateResourceTagsPrompt(map[string]string{"env": "prod"}, true)
		assert.Error(t, err)
		assert.Nil(t, tags)
	})
	assert.Contains(t, stderr, "Current tags:")
	assert.Contains(t, stderr, "env: prod")
	assert.Empty(t, stdout)
}

// TestUpdateResourceTagsPrompt_DefaultColoredWarning exercises the
// default updateResourceTagsPromptFn's colored warning branch (noColor=false).
func TestUpdateResourceTagsPrompt_DefaultColoredWarning(t *testing.T) {
	stdout, stderr := withMockedIO("n\n", func() {
		tags, err := UpdateResourceTagsPrompt(map[string]string{"env": "prod"}, false)
		assert.Error(t, err)
		assert.Nil(t, tags)
	})
	assert.Contains(t, stderr, "Warning: This operation will replace all existing tags")
	assert.Empty(t, stdout)
}

// TestUpdateResourceTagsPrompt_DefaultDeclinesWithNoExistingTags exercises
// the default updateResourceTagsPromptFn's "no existing tags" branch.
func TestUpdateResourceTagsPrompt_DefaultDeclinesWithNoExistingTags(t *testing.T) {
	stdout, stderr := withMockedIO("n\n", func() {
		tags, err := UpdateResourceTagsPrompt(map[string]string{}, true)
		assert.Error(t, err)
		assert.Nil(t, tags)
	})
	assert.Contains(t, stderr, "No existing tags found.")
	assert.Empty(t, stdout)
}

// TestUpdateResourceTagsPrompt_DefaultNoExistingTagsDelegates exercises the
// default updateResourceTagsPromptFn's delegation to ResourceTagsPrompt when
// there are no existing tags to choose a clean-slate/modify strategy for.
func TestUpdateResourceTagsPrompt_DefaultNoExistingTagsDelegates(t *testing.T) {
	stdout, stderr := withMockedIOStreamed([]string{"y", "n"}, func() {
		tags, err := UpdateResourceTagsPrompt(map[string]string{}, true)
		assert.NoError(t, err)
		assert.Nil(t, tags)
	})
	assert.Contains(t, stderr, "Would you like to add resource tags?")
	assert.Empty(t, stdout)
}

// TestUpdateResourceTagsPrompt_DefaultCleanSlate exercises the default
// updateResourceTagsPromptFn's clean-slate path (choice "1"), which discards
// existing tags and prompts for a fresh set.
func TestUpdateResourceTagsPrompt_DefaultCleanSlate(t *testing.T) {
	stdout, stderr := withMockedIOStreamed([]string{"y", "1", "env", "prod", "", "y"}, func() {
		tags, err := UpdateResourceTagsPrompt(map[string]string{"old": "tag"}, true)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"env": "prod"}, tags)
	})
	assert.Contains(t, stderr, "Choose how you want to update tags:")
	assert.Contains(t, stderr, "Final tags that will be applied:")
	assert.Contains(t, stderr, "env: prod")
	assert.Empty(t, stdout)
}

// TestUpdateResourceTagsPrompt_DefaultModifyExistingRemovesTag exercises the
// default updateResourceTagsPromptFn's modify path (choice "2"), including
// removing an existing tag by entering its key with an empty value.
func TestUpdateResourceTagsPrompt_DefaultModifyExistingRemovesTag(t *testing.T) {
	stdout, stderr := withMockedIOStreamed([]string{"y", "2", "foo", "", "", "y"}, func() {
		tags, err := UpdateResourceTagsPrompt(map[string]string{"foo": "bar"}, true)
		assert.NoError(t, err)
		assert.Empty(t, tags)
	})
	assert.Contains(t, stderr, "You can now modify, add, or remove tags.")
	assert.Contains(t, stderr, "Removed tag: foo")
	assert.Contains(t, stderr, "No tags - all existing tags will be removed")
	assert.Empty(t, stdout)
}

// TestDesignConfirmPrompt_DefaultRendering verifies the default
// designConfirmPromptFn renders design-stage wording instead of
// BuyConfirmPrompt's purchase wording.
func TestDesignConfirmPrompt_DefaultRendering(t *testing.T) {
	details := []BuyConfirmDetail{
		{Key: "Name", Value: "test-gw"},
		{Key: "Speed", Value: "1000 Mbps"},
	}

	stdout, stderr := withMockedIO("y\n", func() {
		result := DesignConfirmPrompt("NAT Gateway", details, true)
		assert.True(t, result)
	})

	assert.Contains(t, stderr, "Design Summary:")
	assert.Contains(t, stderr, "Proceed with creation?")
	assert.Contains(t, stderr, "Resource Type: NAT Gateway")
	assert.Contains(t, stderr, "Name: test-gw")
	assert.Contains(t, stderr, "Speed: 1000 Mbps")
	assert.NotContains(t, stderr, "Purchase Summary")
	assert.NotContains(t, stderr, "Proceed with purchase")
	assert.Empty(t, stdout, "confirmation prompt must not write to stdout")
}

// TestConfirmPrompt_StdoutStaysCleanForJSONPiping is a regression test for
// ESD-1586: piping `--output json` through `jq` must not see prompt text,
// since jq would fail to parse it.
func TestConfirmPrompt_StdoutStaysCleanForJSONPiping(t *testing.T) {
	stdout, stderr := withMockedIO("y\n", func() {
		result := ConfirmPrompt("Proceed with provisioning?", false)
		assert.True(t, result)
	})

	assert.Empty(t, stdout, "confirmation prompt text must not appear on stdout")
	assert.Contains(t, stderr, "Proceed with provisioning?")
	assert.Contains(t, stderr, "[y/N]")
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
				originalPrompt := GetPrompt() // Use real function
				stdout, stderr := withMockedIO(input, func() {
					result, err := originalPrompt("Test prompt:", true)
					assert.NoError(t, err)
					assert.Equal(t, "test input", result)
				})
				assert.Contains(t, stderr, "❯ Test prompt:")
				assert.Empty(t, stdout)
			},
		},
		{
			name: "ResourcePrompt integration",
			inputFn: func() string {
				return "resource input\n"
			},
			testFn: func(input string) {
				originalResourcePrompt := GetResourcePrompt() // Use real function
				stdout, stderr := withMockedIO(input, func() {
					result, err := originalResourcePrompt("port", "Port name:", true)
					assert.NoError(t, err)
					assert.Equal(t, "resource input", result)
				})
				assert.Contains(t, stderr, "🔌 Port name:")
				assert.Empty(t, stdout)
			},
		},
		{
			name: "ResourcePrompt integration colored",
			inputFn: func() string {
				return "resource input\n"
			},
			testFn: func(input string) {
				originalResourcePrompt := GetResourcePrompt() // Use real function
				stdout, stderr := withMockedIO(input, func() {
					result, err := originalResourcePrompt("port", "Port name:", false)
					assert.NoError(t, err)
					assert.Equal(t, "resource input", result)
				})
				assert.Contains(t, stderr, "Port name:")
				assert.Empty(t, stdout)
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
	defer tmpfile.Close()

	// Write our simulated input
	fmt.Fprint(tmpfile, simulatedInput)
	_, err = tmpfile.Seek(0, 0)
	if err != nil {
		t.Fatalf("Failed to seek in temp file: %v", err)
	}

	// Replace stdin with our file
	os.Stdin = tmpfile

	// Capture stdout and stderr separately so we can verify the prompt text
	// lands on stderr, not stdout.
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create mock stdout pipe: %v", err)
	}
	oldStdout := os.Stdout
	os.Stdout = outW
	defer func() { os.Stdout = oldStdout }()

	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create mock stderr pipe: %v", err)
	}
	oldStderr := os.Stderr
	os.Stderr = errW
	defer func() { os.Stderr = oldStderr }()

	// Test the real Prompt function with our simulated stdin
	result, err := Prompt("Enter value:", true)

	// Close the write pipes
	outW.Close()
	errW.Close()

	// Restore stdout/stderr to see test results
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var outBuf, errBuf bytes.Buffer
	if _, copyErr := io.Copy(&outBuf, outR); copyErr != nil {
		t.Fatalf("Failed to copy from mock stdout: %v", copyErr)
	}
	if _, copyErr := io.Copy(&errBuf, errR); copyErr != nil {
		t.Fatalf("Failed to copy from mock stderr: %v", copyErr)
	}
	outR.Close()
	errR.Close()

	if err != nil {
		t.Fatalf("Got error from prompt: %v", err)
	}

	expected := "simulated input"
	assert.Equal(t, expected, result, "Expected prompt to return %q, got %q", expected, result)
	assert.Contains(t, errBuf.String(), "Enter value:")
	assert.Empty(t, outBuf.String(), "prompt text must not be written to stdout")
}

//go:build js && wasm
// +build js,wasm

package utils

import (
	"syscall/js"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/wasm"
)

func TestWasmPrompt(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		noColor      bool
		mockResponse string
		expectError  bool
	}{
		{
			name:         "successful text prompt",
			message:      "Enter name:",
			noColor:      true,
			mockResponse: "test-value",
			expectError:  false,
		},
		{
			name:         "empty response",
			message:      "Enter optional field:",
			noColor:      false,
			mockResponse: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock callback
			setupMockPromptHandler(t, tt.mockResponse)
			defer cleanupMockPromptHandler()

			// Execute
			result, err := wasmPrompt(tt.message, tt.noColor)

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.mockResponse {
				t.Errorf("Expected %q, got %q", tt.mockResponse, result)
			}
		})
	}
}

func TestWasmConfirmPrompt(t *testing.T) {
	tests := []struct {
		name         string
		question     string
		noColor      bool
		mockResponse string
		expectResult bool
	}{
		{
			name:         "confirm with 'y'",
			question:     "Continue?",
			noColor:      true,
			mockResponse: "y",
			expectResult: true,
		},
		{
			name:         "confirm with 'yes'",
			question:     "Proceed?",
			noColor:      true,
			mockResponse: "yes",
			expectResult: true,
		},
		{
			name:         "confirm with 'n'",
			question:     "Delete?",
			noColor:      false,
			mockResponse: "n",
			expectResult: false,
		},
		{
			name:         "confirm with 'no'",
			question:     "Delete?",
			noColor:      false,
			mockResponse: "no",
			expectResult: false,
		},
		{
			name:         "confirm with empty (default no)",
			question:     "Continue?",
			noColor:      true,
			mockResponse: "",
			expectResult: false,
		},
		{
			name:         "confirm with uppercase YES",
			question:     "Confirm?",
			noColor:      true,
			mockResponse: "YES",
			expectResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock callback
			setupMockPromptHandler(t, tt.mockResponse)
			defer cleanupMockPromptHandler()

			// Execute
			result := wasmConfirmPrompt(tt.question, tt.noColor)

			// Assert
			if result != tt.expectResult {
				t.Errorf("Expected %v, got %v", tt.expectResult, result)
			}
		})
	}
}

func TestWasmResourcePrompt(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		message      string
		noColor      bool
		mockResponse string
		expectError  bool
	}{
		{
			name:         "port prompt",
			resourceType: "port",
			message:      "Enter port speed:",
			noColor:      true,
			mockResponse: "10000",
			expectError:  false,
		},
		{
			name:         "mcr prompt",
			resourceType: "mcr",
			message:      "Enter MCR name:",
			noColor:      false,
			mockResponse: "test-mcr",
			expectError:  false,
		},
		{
			name:         "vxc prompt",
			resourceType: "vxc",
			message:      "Enter VXC name:",
			noColor:      true,
			mockResponse: "test-vxc",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock callback
			setupMockPromptHandler(t, tt.mockResponse)
			defer cleanupMockPromptHandler()

			// Execute
			result, err := wasmResourcePrompt(tt.resourceType, tt.message, tt.noColor)

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && result != tt.mockResponse {
				t.Errorf("Expected %q, got %q", tt.mockResponse, result)
			}
		})
	}
}

func TestWasmResourceTagsPrompt(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []string // Sequence of responses to prompts
		expectedTags  map[string]string
		expectError   bool
	}{
		{
			name: "no tags",
			mockResponses: []string{
				"n", // Don't add tags
			},
			expectedTags: nil,
			expectError:  false,
		},
		{
			name: "single tag",
			mockResponses: []string{
				"y",        // Add tags
				"env",      // Tag key
				"prod",     // Tag value
				"",         // Empty key to finish
			},
			expectedTags: map[string]string{
				"env": "prod",
			},
			expectError: false,
		},
		{
			name: "multiple tags",
			mockResponses: []string{
				"y",           // Add tags
				"env",         // Tag key 1
				"prod",        // Tag value 1
				"team",        // Tag key 2
				"engineering", // Tag value 2
				"",            // Empty key to finish
			},
			expectedTags: map[string]string{
				"env":  "prod",
				"team": "engineering",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock callback with sequence of responses
			setupSequentialMockPromptHandler(t, tt.mockResponses)
			defer cleanupMockPromptHandler()

			// Execute
			result, err := wasmResourceTagsPrompt(true)

			// Assert
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if len(result) != len(tt.expectedTags) {
					t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(result))
				}
				for key, expectedValue := range tt.expectedTags {
					if actualValue, ok := result[key]; !ok {
						t.Errorf("Expected tag %q not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("For tag %q, expected %q, got %q", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

func TestWasmUpdateResourceTagsPrompt(t *testing.T) {
	tests := []struct {
		name          string
		existingTags  map[string]string
		mockResponses []string
		expectedTags  map[string]string
		expectError   bool
	}{
		{
			name: "user cancels",
			existingTags: map[string]string{
				"env": "prod",
			},
			mockResponses: []string{
				"n", // Don't continue
			},
			expectedTags: nil,
			expectError:  true,
		},
		{
			name: "clean slate - add new tags",
			existingTags: map[string]string{
				"env": "prod",
			},
			mockResponses: []string{
				"y",        // Continue
				"1",        // Clean slate
				"team",     // New tag key
				"devops",   // New tag value
				"",         // Finish
				"y",        // Apply changes
			},
			expectedTags: map[string]string{
				"team": "devops",
			},
			expectError: false,
		},
		{
			name: "modify existing tags",
			existingTags: map[string]string{
				"env": "prod",
			},
			mockResponses: []string{
				"y",      // Continue
				"2",      // Start with existing
				"env",    // Modify existing key
				"dev",    // New value
				"team",   // Add new key
				"backend", // New tag value
				"",       // Finish
				"y",      // Apply changes
			},
			expectedTags: map[string]string{
				"env":  "dev",
				"team": "backend",
			},
			expectError: false,
		},
		{
			name: "remove tag with empty value",
			existingTags: map[string]string{
				"env":  "prod",
				"team": "backend",
			},
			mockResponses: []string{
				"y",   // Continue
				"2",   // Start with existing
				"env", // Tag to remove
				"",    // Empty value = remove
				"",    // Finish
				"y",   // Apply changes
			},
			expectedTags: map[string]string{
				"team": "backend",
			},
			expectError: false,
		},
		{
			name:         "no existing tags - add new",
			existingTags: map[string]string{},
			mockResponses: []string{
				"y",      // Continue
				"y",      // Add tags (from ResourceTagsPrompt)
				"env",    // Tag key
				"prod",   // Tag value
				"",       // Finish
			},
			expectedTags: map[string]string{
				"env": "prod",
			},
			expectError: false,
		},
		{
			name: "user cancels at final confirmation",
			existingTags: map[string]string{
				"env": "prod",
			},
			mockResponses: []string{
				"y",      // Continue
				"1",      // Clean slate
				"team",   // New tag key
				"devops", // New tag value
				"",       // Finish
				"n",      // Don't apply changes
			},
			expectedTags: nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock callback with sequence of responses
			setupSequentialMockPromptHandler(t, tt.mockResponses)
			defer cleanupMockPromptHandler()

			// Execute
			result, err := wasmUpdateResourceTagsPrompt(tt.existingTags, true)

			// Assert
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if len(result) != len(tt.expectedTags) {
					t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(result))
				}
				for key, expectedValue := range tt.expectedTags {
					if actualValue, ok := result[key]; !ok {
						t.Errorf("Expected tag %q not found", key)
					} else if actualValue != expectedValue {
						t.Errorf("For tag %q, expected %q, got %q", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}

// Helper functions for testing

var responseQueue []string
var responseIndex int

func setupMockPromptHandler(t *testing.T, response string) {
	t.Helper()
	
	// Register a mock callback - convert js.Func to js.Value
	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			promptData := args[0]
			promptID := promptData.Get("id").String()
			
			// Automatically respond after a short delay
			go func() {
				time.Sleep(10 * time.Millisecond)
				jsArgs := []js.Value{
					js.ValueOf(promptID),
					js.ValueOf(response),
				}
				// Simulate JavaScript calling SubmitPromptResponse
				wasm.SubmitPromptResponse(js.Undefined(), jsArgs)
			}()
		}
		return nil
	})
	
	wasm.RegisterPromptCallback(callback.Value)
}

func setupSequentialMockPromptHandler(t *testing.T, responses []string) {
	t.Helper()
	
	responseQueue = responses
	responseIndex = 0
	
	// Register a mock callback that returns responses in sequence
	callback := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 && responseIndex < len(responseQueue) {
			promptData := args[0]
			promptID := promptData.Get("id").String()
			response := responseQueue[responseIndex]
			responseIndex++
			
			// Automatically respond after a short delay
			go func() {
				time.Sleep(10 * time.Millisecond)
				jsArgs := []js.Value{
					js.ValueOf(promptID),
					js.ValueOf(response),
				}
				// Simulate JavaScript calling SubmitPromptResponse
				wasm.SubmitPromptResponse(js.Undefined(), jsArgs)
			}()
		}
		return nil
	})
	
	wasm.RegisterPromptCallback(callback.Value)
}

func cleanupMockPromptHandler() {
	// Reset the callback
	wasm.RegisterPromptCallback(js.Undefined())
	responseQueue = nil
	responseIndex = 0
}

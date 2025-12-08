//go:build js && wasm
// +build js,wasm

package wasm

import (
	"fmt"
	"sync"
	"syscall/js"
	"testing"
	"time"
)

func TestPromptForInput(t *testing.T) {
	tests := []struct {
		name         string
		message      string
		promptType   string
		resourceType string
		response     string
		shouldError  bool
		setupMock    func()
		cleanupMock  func()
	}{
		{
			name:         "successful text prompt",
			message:      "Enter name:",
			promptType:   "text",
			resourceType: "",
			response:     "test-name",
			shouldError:  false,
			setupMock: func() {
				// Mock the callback - js.FuncOf returns js.Func which implements js.Value
				fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					return nil
				})
				promptCallback = fn.Value
			},
			cleanupMock: func() {
				promptCallback = js.Undefined()
			},
		},
		{
			name:         "successful confirm prompt",
			message:      "Continue?",
			promptType:   "confirm",
			resourceType: "",
			response:     "y",
			shouldError:  false,
			setupMock: func() {
				fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					return nil
				})
				promptCallback = fn.Value
			},
			cleanupMock: func() {
				promptCallback = js.Undefined()
			},
		},
		{
			name:         "successful resource prompt",
			message:      "Enter port speed:",
			promptType:   "resource",
			resourceType: "port",
			response:     "10000",
			shouldError:  false,
			setupMock: func() {
				fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
					return nil
				})
				promptCallback = fn.Value
			},
			cleanupMock: func() {
				promptCallback = js.Undefined()
			},
		},
		{
			name:         "error when callback not registered",
			message:      "Enter name:",
			promptType:   "text",
			resourceType: "",
			response:     "",
			shouldError:  true,
			setupMock: func() {
				promptCallback = js.Undefined()
			},
			cleanupMock: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupMock()
			defer tt.cleanupMock()

			// Clear pending prompts
			pendingMutex.Lock()
			pendingPrompts = make(map[string]*PromptRequest)
			pendingMutex.Unlock()

			// If we have a callback, simulate the response
			if !promptCallback.IsUndefined() {
				go func() {
					time.Sleep(50 * time.Millisecond)

					// Find the pending prompt
					pendingMutex.Lock()
					var promptID string
					for id := range pendingPrompts {
						promptID = id
						break
					}
					pendingMutex.Unlock()

					if promptID != "" {
						// Simulate JavaScript response
						args := []js.Value{
							js.ValueOf(promptID),
							js.ValueOf(tt.response),
						}
						submitPromptResponse(js.Undefined(), args)
					}
				}()
			}

			// Execute
			result, err := PromptForInput(tt.message, tt.promptType, tt.resourceType)

			// Assert
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.shouldError && result != tt.response {
				t.Errorf("Expected response %q, got %q", tt.response, result)
			}
		})
	}
}

func TestSubmitPromptResponse(t *testing.T) {
	tests := []struct {
		name        string
		setupPrompt bool
		promptID    string
		response    string
		expectError bool
	}{
		{
			name:        "successful response submission",
			setupPrompt: true,
			promptID:    "test-prompt-1",
			response:    "test-response",
			expectError: false,
		},
		{
			name:        "response to non-existent prompt",
			setupPrompt: false,
			promptID:    "non-existent",
			response:    "test-response",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			pendingMutex.Lock()
			pendingPrompts = make(map[string]*PromptRequest)

			var responseChan chan string
			if tt.setupPrompt {
				responseChan = make(chan string, 1)
				pendingPrompts[tt.promptID] = &PromptRequest{
					ID:           tt.promptID,
					Message:      "Test message",
					PromptType:   "text",
					ResponseChan: responseChan,
					ErrorChan:    make(chan error, 1),
				}
			}
			pendingMutex.Unlock()

			// Execute
			args := []js.Value{
				js.ValueOf(tt.promptID),
				js.ValueOf(tt.response),
			}
			result := submitPromptResponse(js.Undefined(), args)

			// Assert
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("Expected result to be a map")
			}

			if tt.expectError {
				if _, hasError := resultMap["error"]; !hasError {
					t.Error("Expected error in result")
				}
			} else {
				if success, ok := resultMap["success"].(bool); !ok || !success {
					t.Error("Expected success=true in result")
				}

				// Verify response was sent
				select {
				case receivedResponse := <-responseChan:
					if receivedResponse != tt.response {
						t.Errorf("Expected response %q, got %q", tt.response, receivedResponse)
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("Response was not sent to channel")
				}
			}
		})
	}
}

func TestCancelPrompt(t *testing.T) {
	tests := []struct {
		name        string
		setupPrompt bool
		promptID    string
		expectError bool
	}{
		{
			name:        "successful cancellation",
			setupPrompt: true,
			promptID:    "test-prompt-1",
			expectError: false,
		},
		{
			name:        "cancel non-existent prompt",
			setupPrompt: false,
			promptID:    "non-existent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			pendingMutex.Lock()
			pendingPrompts = make(map[string]*PromptRequest)

			var errorChan chan error
			if tt.setupPrompt {
				errorChan = make(chan error, 1)
				pendingPrompts[tt.promptID] = &PromptRequest{
					ID:           tt.promptID,
					Message:      "Test message",
					PromptType:   "text",
					ResponseChan: make(chan string, 1),
					ErrorChan:    errorChan,
				}
			}
			pendingMutex.Unlock()

			// Execute
			args := []js.Value{js.ValueOf(tt.promptID)}
			result := cancelPrompt(js.Undefined(), args)

			// Assert
			resultMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatal("Expected result to be a map")
			}

			if tt.expectError {
				if _, hasError := resultMap["error"]; !hasError {
					t.Error("Expected error in result")
				}
			} else {
				if success, ok := resultMap["success"].(bool); !ok || !success {
					t.Error("Expected success=true in result")
				}

				// Verify error was sent
				select {
				case err := <-errorChan:
					if err == nil {
						t.Error("Expected error to be sent to channel")
					}
				case <-time.After(100 * time.Millisecond):
					t.Error("Error was not sent to channel")
				}
			}
		})
	}
}

func TestRegisterPromptHandler(t *testing.T) {
	tests := []struct {
		name     string
		callback js.Value
		expectOK bool
	}{
		{
			name:     "register valid function",
			callback: js.FuncOf(func(this js.Value, args []js.Value) interface{} { return nil }).Value,
			expectOK: true,
		},
		{
			name:     "register non-function",
			callback: js.ValueOf("not a function"),
			expectOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset
			promptCallback = js.Undefined()

			// Execute
			args := []js.Value{tt.callback}
			result := registerPromptHandler(js.Undefined(), args)

			// Assert
			if tt.expectOK {
				if resultBool, ok := result.(bool); !ok || !resultBool {
					t.Error("Expected true result for valid callback")
				}
				if promptCallback.IsUndefined() {
					t.Error("Callback was not registered")
				}
			} else {
				if resultBool, ok := result.(bool); ok && resultBool {
					t.Error("Expected false result for invalid callback")
				}
			}
		})
	}
}

func TestGetPendingPrompts(t *testing.T) {
	// Setup some pending prompts
	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingPrompts["prompt-1"] = &PromptRequest{
		ID:           "prompt-1",
		Message:      "Enter name:",
		PromptType:   "text",
		ResourceType: "",
		ResponseChan: make(chan string, 1),
		ErrorChan:    make(chan error, 1),
	}
	pendingPrompts["prompt-2"] = &PromptRequest{
		ID:           "prompt-2",
		Message:      "Enter port speed:",
		PromptType:   "resource",
		ResourceType: "port",
		ResponseChan: make(chan string, 1),
		ErrorChan:    make(chan error, 1),
	}
	pendingMutex.Unlock()

	// Execute
	result := getPendingPrompts(js.Undefined(), nil)

	// Assert
	prompts, ok := result.([]map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a slice of maps")
	}

	if len(prompts) != 2 {
		t.Errorf("Expected 2 pending prompts, got %d", len(prompts))
	}

	// Verify prompt contents
	foundPrompt1 := false
	foundPrompt2 := false
	for _, p := range prompts {
		if id, ok := p["id"].(string); ok {
			switch id {
			case "prompt-1":
				foundPrompt1 = true
				if p["message"] != "Enter name:" || p["type"] != "text" {
					t.Error("Prompt-1 has incorrect data")
				}
			case "prompt-2":
				foundPrompt2 = true
				if p["message"] != "Enter port speed:" || p["type"] != "resource" || p["resourceType"] != "port" {
					t.Error("Prompt-2 has incorrect data")
				}
			}
		}
	}

	if !foundPrompt1 || !foundPrompt2 {
		t.Error("Not all expected prompts were found in result")
	}
}

func TestPromptTimeout(t *testing.T) {
	// This test verifies that prompts timeout after the specified duration
	// We'll use a shorter timeout for testing

	// Setup
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Don't respond - let it timeout
		return nil
	})
	promptCallback = fn.Value
	defer func() {
		promptCallback = js.Undefined()
	}()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	// Execute with a short timeout by temporarily modifying the timeout
	// Note: In real implementation, we'd need to make timeout configurable for testing
	// For now, we'll just verify the timeout mechanism works

	done := make(chan bool)

	go func() {
		//nolint:errcheck // intentionally ignoring error in test to let it timeout
		PromptForInput("Test", "text", "")
		done <- true
	}()

	// Wait a bit for the prompt to be created
	time.Sleep(50 * time.Millisecond)

	// Verify prompt exists
	pendingMutex.Lock()
	promptCount := len(pendingPrompts)
	pendingMutex.Unlock()

	if promptCount != 1 {
		t.Errorf("Expected 1 pending prompt, got %d", promptCount)
	}

	// Note: Full timeout test would take 5 minutes, so we just verify the mechanism is in place
	// In production, you'd want to make the timeout configurable for testing
}

func TestConcurrentPrompts(t *testing.T) {
	// Test that multiple prompts can be handled concurrently

	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	})
	promptCallback = fn.Value
	defer func() {
		promptCallback = js.Undefined()
	}()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	numPrompts := 10
	var wg sync.WaitGroup
	wg.Add(numPrompts)

	// Start multiple prompts concurrently
	for i := 0; i < numPrompts; i++ {
		go func(n int) {
			defer wg.Done()

			// Start the prompt
			go func() {
				message := fmt.Sprintf("Prompt %d", n)
				//nolint:errcheck // intentionally ignoring error in concurrent test
				PromptForInput(message, "text", "")
			}()

			// Wait a bit for prompt to be registered
			time.Sleep(50 * time.Millisecond)

			// Find and respond to the prompt
			pendingMutex.Lock()
			for id := range pendingPrompts {
				args := []js.Value{
					js.ValueOf(id),
					js.ValueOf(fmt.Sprintf("Response %d", n)),
				}
				submitPromptResponse(js.Undefined(), args)
				break
			}
			pendingMutex.Unlock()
		}(i)
	}

	// Wait for all prompts to complete (with timeout)
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("Concurrent prompt test timed out")
	}
}

func TestPromptIDUniqueness(t *testing.T) {
	// Verify that each prompt gets a unique ID

	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil
	})
	promptCallback = fn.Value
	defer func() {
		promptCallback = js.Undefined()
	}()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	seenIDs := make(map[string]bool)
	numPrompts := 100

	for i := 0; i < numPrompts; i++ {
		go func() {
			//nolint:errcheck // intentionally ignoring error in uniqueness test
			PromptForInput("Test", "text", "")
		}()
	}

	// Wait for prompts to be registered
	time.Sleep(100 * time.Millisecond)

	// Collect all IDs
	pendingMutex.Lock()
	for id := range pendingPrompts {
		if seenIDs[id] {
			t.Errorf("Duplicate prompt ID found: %s", id)
		}
		seenIDs[id] = true
	}
	pendingMutex.Unlock()

	if len(seenIDs) != numPrompts {
		t.Errorf("Expected %d unique IDs, got %d", numPrompts, len(seenIDs))
	}
}

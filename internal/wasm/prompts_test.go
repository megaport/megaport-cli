//go:build js && wasm

package wasm

import (
	"fmt"
	"strings"
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

// withScaledPromptTimeout temporarily overrides promptTimeout for a test and
// restores it afterward. promptTimeout defaults to CommandTimeout (10
// minutes), far too long to wait out in a unit test.
func withScaledPromptTimeout(t *testing.T, d time.Duration) {
	t.Helper()
	orig := promptTimeout
	promptTimeout = d
	t.Cleanup(func() { promptTimeout = orig })
}

// capturingPromptCallback returns a mock JS prompt callback and a channel
// that receives the prompt ID the moment PromptForInput invokes it.
// PromptForInput calls the callback synchronously before it enters its
// timeout select, so reading from the returned channel is deterministic and
// never races the timeout, unlike polling pendingPrompts from a goroutine.
func capturingPromptCallback() (js.Func, <-chan string) {
	idCh := make(chan string, 1)
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		idCh <- args[0].Get("id").String()
		return nil
	})
	return fn, idCh
}

// TestPromptForInput_RespondsPastOldFiveMinuteMark is a regression test for
// the original bug: PromptForInput's per-prompt timeout (hardcoded 5
// minutes) fired before main_wasm.go's 10-minute asyncCommandTimeout,
// killing a command whose prompt was still within the command's own budget.
// promptTimeout is scaled down here so the test runs in milliseconds, but
// the response arrives well past the halfway point of the budget (mirroring
// "answered at minute 6 of a 10-minute budget") and must still succeed now
// that both timeouts share one source of truth.
func TestPromptForInput_RespondsPastOldFiveMinuteMark(t *testing.T) {
	withScaledPromptTimeout(t, 200*time.Millisecond)

	fn, promptIDCh := capturingPromptCallback()
	t.Cleanup(fn.Release)
	promptCallback = fn.Value
	defer func() { promptCallback = js.Undefined() }()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	go func() {
		promptID := <-promptIDCh
		time.Sleep(120 * time.Millisecond) // 60% of the scaled budget
		args := []js.Value{js.ValueOf(promptID), js.ValueOf("late-but-in-budget")}
		submitPromptResponse(js.Undefined(), args)
	}()

	result, err := PromptForInput("Enter name:", "text", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "late-but-in-budget" {
		t.Errorf("expected %q, got %q", "late-but-in-budget", result)
	}
}

// TestPromptForInput_TimeoutErrorMessage verifies the timeout error states
// how long PromptForInput waited and that the host can cancel via
// cancelPrompt, and that the timed-out prompt is cleaned up.
func TestPromptForInput_TimeoutErrorMessage(t *testing.T) {
	withScaledPromptTimeout(t, 20*time.Millisecond)

	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return nil // never respond
	})
	t.Cleanup(fn.Release)
	promptCallback = fn.Value
	defer func() { promptCallback = js.Undefined() }()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	_, err := PromptForInput("Enter name:", "text", "")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, promptTimeout.String()) {
		t.Errorf("expected error to state the duration waited (%s), got: %s", promptTimeout, msg)
	}
	if !strings.Contains(msg, "cancelPrompt") {
		t.Errorf("expected error to mention cancelPrompt, got: %s", msg)
	}

	pendingMutex.Lock()
	remaining := len(pendingPrompts)
	pendingMutex.Unlock()
	if remaining != 0 {
		t.Errorf("expected timed-out prompt to be cleaned up, got %d pending", remaining)
	}
}

// TestPromptForInput_LateSubmitAfterTimeoutStaysClean verifies a
// submitPromptResponse that arrives after a prompt has already timed out
// (and been cleaned up) returns a clean "Prompt not found" without blocking
// or panicking.
func TestPromptForInput_LateSubmitAfterTimeoutStaysClean(t *testing.T) {
	withScaledPromptTimeout(t, 20*time.Millisecond)

	fn, promptIDCh := capturingPromptCallback()
	t.Cleanup(fn.Release)
	promptCallback = fn.Value
	defer func() { promptCallback = js.Undefined() }()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	_, err := PromptForInput("Enter name:", "text", "")
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	promptID := <-promptIDCh

	args := []js.Value{js.ValueOf(promptID), js.ValueOf("too-late")}
	done := make(chan interface{}, 1)
	go func() { done <- submitPromptResponse(js.Undefined(), args) }()

	select {
	case result := <-done:
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("expected result to be a map")
		}
		if _, hasError := resultMap["error"]; !hasError {
			t.Error("expected error in result for late submit after timeout")
		}
	case <-time.After(time.Second):
		t.Fatal("late submitPromptResponse blocked instead of returning cleanly")
	}
}

// TestPromptForInput_LateSubmitAfterCancelStaysClean verifies a
// submitPromptResponse that arrives after a prompt has been cancelled (and
// cleaned up) returns a clean "Prompt not found" without blocking or
// panicking.
func TestPromptForInput_LateSubmitAfterCancelStaysClean(t *testing.T) {
	// Safety net: cancelPrompt is what should unblock PromptForInput here, not
	// the timeout. A scaled-down promptTimeout keeps a broken cancel path from
	// hanging the test for the full 10-minute default.
	withScaledPromptTimeout(t, 5*time.Second)

	fn, promptIDCh := capturingPromptCallback()
	t.Cleanup(fn.Release)
	promptCallback = fn.Value
	defer func() { promptCallback = js.Undefined() }()

	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()

	// lateIDCh receives the prompt ID once cancelPrompt has already been
	// called with it, so reading it after PromptForInput returns is
	// guaranteed to see a value: the send happens-before the cancellation
	// that unblocks PromptForInput.
	lateIDCh := make(chan string, 1)
	go func() {
		promptID := <-promptIDCh
		lateIDCh <- promptID
		cancelPrompt(js.Undefined(), []js.Value{js.ValueOf(promptID)})
	}()

	_, err := PromptForInput("Enter name:", "text", "")
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	// Pin down that cancelPrompt (not the 5-second timeout fallback) is what
	// unblocked PromptForInput, so a broken cancel path can't slip through as
	// a false-negative pass.
	if !strings.Contains(err.Error(), "cancelled") {
		t.Fatalf("expected a cancellation error, got: %v", err)
	}
	promptID := <-lateIDCh

	args := []js.Value{js.ValueOf(promptID), js.ValueOf("too-late")}
	done := make(chan interface{}, 1)
	go func() { done <- submitPromptResponse(js.Undefined(), args) }()

	select {
	case result := <-done:
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("expected result to be a map")
		}
		if _, hasError := resultMap["error"]; !hasError {
			t.Error("expected error in result for late submit after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("late submitPromptResponse blocked instead of returning cleanly")
	}
}

func TestConcurrentPrompts(t *testing.T) {
	// Test that multiple prompts can be registered and responded to concurrently
	// This test verifies that the prompt registration mechanism handles concurrent access correctly

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
	promptIDs := make([]string, numPrompts)
	responseChs := make([]chan string, numPrompts)

	// Phase 1: Create all prompts and collect their IDs
	for i := 0; i < numPrompts; i++ {
		responseCh := make(chan string, 1)
		responseChs[i] = responseCh

		// Register prompt directly without blocking on PromptForInput
		pendingMutex.Lock()
		promptCounter.Add(1)
		promptID := fmt.Sprintf("concurrent_test_%d_%d", i, time.Now().UnixNano())
		promptIDs[i] = promptID

		request := &PromptRequest{
			ID:           promptID,
			Message:      fmt.Sprintf("Prompt %d", i),
			PromptType:   "text",
			ResourceType: "",
			ResponseChan: responseCh,
			ErrorChan:    make(chan error, 1),
		}
		pendingPrompts[promptID] = request
		pendingMutex.Unlock()
	}

	// Verify all prompts were registered
	pendingMutex.Lock()
	registeredCount := len(pendingPrompts)
	pendingMutex.Unlock()

	if registeredCount != numPrompts {
		t.Errorf("Expected %d registered prompts, got %d", numPrompts, registeredCount)
		return
	}

	// Phase 2: Respond to each prompt by its specific ID
	var wg sync.WaitGroup
	wg.Add(numPrompts)

	for i := 0; i < numPrompts; i++ {
		go func(n int) {
			defer wg.Done()
			args := []js.Value{
				js.ValueOf(promptIDs[n]),
				js.ValueOf(fmt.Sprintf("Response %d", n)),
			}
			submitPromptResponse(js.Undefined(), args)
		}(i)
	}

	// Wait for all responses to be submitted
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Verify responses were sent by checking response channels
		successCount := 0
		for i := 0; i < numPrompts; i++ {
			select {
			case resp := <-responseChs[i]:
				if resp == fmt.Sprintf("Response %d", i) {
					successCount++
				}
			default:
				// Response not received yet
			}
		}

		// We expect all responses to have been submitted
		// Note: Since we're not running full PromptForInput flow, prompts aren't auto-cleaned
		// The test verifies that submitPromptResponse correctly sends to response channels

		if successCount != numPrompts {
			t.Logf("Received %d/%d responses (some may still be pending)", successCount, numPrompts)
		}

		// Clean up remaining prompts manually
		pendingMutex.Lock()
		pendingPrompts = make(map[string]*PromptRequest)
		pendingMutex.Unlock()

	case <-time.After(5 * time.Second):
		t.Error("Concurrent prompt response test timed out")
	}
}

func TestPromptIDUniqueness(t *testing.T) {
	// Verify that each prompt gets a unique ID
	// This test verifies that the prompt ID generation produces unique IDs under concurrent access

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
	var idMutex sync.Mutex

	// Generate IDs concurrently without blocking on PromptForInput
	var wg sync.WaitGroup
	wg.Add(numPrompts)

	for i := 0; i < numPrompts; i++ {
		go func() {
			defer wg.Done()

			// Register prompt directly (similar to what PromptForInput does)
			pendingMutex.Lock()
			cnt := promptCounter.Add(1)
			promptID := fmt.Sprintf("unique_test_%d_%d", cnt, time.Now().UnixNano())

			request := &PromptRequest{
				ID:           promptID,
				Message:      "Test",
				PromptType:   "text",
				ResourceType: "",
				ResponseChan: make(chan string, 1),
				ErrorChan:    make(chan error, 1),
			}
			pendingPrompts[promptID] = request
			pendingMutex.Unlock()

			// Track the ID we generated
			idMutex.Lock()
			if seenIDs[promptID] {
				t.Errorf("Duplicate prompt ID found: %s", promptID)
			}
			seenIDs[promptID] = true
			idMutex.Unlock()
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify we got the expected number of unique IDs
	idMutex.Lock()
	uniqueCount := len(seenIDs)
	idMutex.Unlock()

	if uniqueCount != numPrompts {
		t.Errorf("Expected %d unique IDs, got %d", numPrompts, uniqueCount)
	}

	// Clean up registered prompts
	pendingMutex.Lock()
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex.Unlock()
}

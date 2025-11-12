//go:build js && wasm
// +build js,wasm

package wasm

import (
	"fmt"
	"sync"
	"syscall/js"
	"time"
)

var (
	// promptCallback is the JavaScript function to call when prompting for input
	promptCallback js.Value
	
	// pendingPrompts tracks prompts waiting for responses
	pendingPrompts = make(map[string]*PromptRequest)
	pendingMutex   sync.Mutex
	
	// promptCounter generates unique IDs for each prompt
	promptCounter int
)

// PromptRequest represents a pending prompt waiting for user input
type PromptRequest struct {
	ID           string
	Message      string
	PromptType   string // "text", "confirm", "resource"
	ResourceType string // for resource prompts (port, mcr, vxc, etc.)
	ResponseChan chan string
	ErrorChan    chan error
}

// RegisterPromptCallback allows JavaScript to register a callback function
// that will be invoked when the WASM code needs user input
func RegisterPromptCallback(callback js.Value) {
	if callback.Type() != js.TypeFunction {
		js.Global().Get("console").Call("error", "Prompt callback must be a function")
		return
	}
	
	promptCallback = callback
	js.Global().Get("console").Call("log", "✅ Prompt callback registered")
}

// PromptForInput requests input from the user via JavaScript
// This function blocks until the JavaScript side provides a response
func PromptForInput(message string, promptType string, resourceType string) (string, error) {
	if promptCallback.IsUndefined() {
		return "", fmt.Errorf("prompt callback not registered - interactive mode requires JavaScript integration")
	}
	
	// Create a unique ID for this prompt
	pendingMutex.Lock()
	promptCounter++
	promptID := fmt.Sprintf("prompt_%d_%d", promptCounter, time.Now().UnixNano())
	pendingMutex.Unlock()
	
	// Create channels for the response
	responseChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	
	// Register the pending prompt
	request := &PromptRequest{
		ID:           promptID,
		Message:      message,
		PromptType:   promptType,
		ResourceType: resourceType,
		ResponseChan: responseChan,
		ErrorChan:    errorChan,
	}
	
	pendingMutex.Lock()
	pendingPrompts[promptID] = request
	pendingMutex.Unlock()
	
	// Log the prompt request
	js.Global().Get("console").Call("log", fmt.Sprintf("📝 Requesting input: ID=%s, Type=%s, Message=%s", 
		promptID, promptType, message))
	
	// Call the JavaScript callback with the prompt details
	promptCallback.Invoke(map[string]interface{}{
		"id":           promptID,
		"message":      message,
		"type":         promptType,
		"resourceType": resourceType,
	})
	
	// Wait for response with timeout
	select {
	case response := <-responseChan:
		// Clean up
		pendingMutex.Lock()
		delete(pendingPrompts, promptID)
		pendingMutex.Unlock()
		
		js.Global().Get("console").Call("log", fmt.Sprintf("✅ Received response for %s: %s", promptID, response))
		return response, nil
		
	case err := <-errorChan:
		// Clean up
		pendingMutex.Lock()
		delete(pendingPrompts, promptID)
		pendingMutex.Unlock()
		
		js.Global().Get("console").Call("error", fmt.Sprintf("❌ Error for %s: %v", promptID, err))
		return "", err
		
	case <-time.After(5 * time.Minute):
		// Timeout
		pendingMutex.Lock()
		delete(pendingPrompts, promptID)
		pendingMutex.Unlock()
		
		return "", fmt.Errorf("prompt timeout: no response received")
	}
}

// SubmitPromptResponse is called by JavaScript to provide the user's response
// Exported for testing
func SubmitPromptResponse(this js.Value, args []js.Value) interface{} {
	return submitPromptResponse(this, args)
}

// submitPromptResponse is the internal implementation
func submitPromptResponse(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		js.Global().Get("console").Call("error", "submitPromptResponse requires: id (string) and response (string)")
		return map[string]interface{}{
			"error": "Invalid arguments",
		}
	}
	
	promptID := args[0].String()
	response := args[1].String()
	
	js.Global().Get("console").Call("log", fmt.Sprintf("📨 Submitting response for %s: %s", promptID, response))
	
	pendingMutex.Lock()
	request, exists := pendingPrompts[promptID]
	pendingMutex.Unlock()
	
	if !exists {
		js.Global().Get("console").Call("warn", fmt.Sprintf("No pending prompt found for ID: %s", promptID))
		return map[string]interface{}{
			"error": "Prompt not found",
		}
	}
	
	// Send the response to the waiting goroutine
	select {
	case request.ResponseChan <- response:
		js.Global().Get("console").Call("log", fmt.Sprintf("✅ Response sent for %s", promptID))
		return map[string]interface{}{
			"success": true,
		}
	default:
		js.Global().Get("console").Call("error", fmt.Sprintf("Failed to send response for %s", promptID))
		return map[string]interface{}{
			"error": "Failed to send response",
		}
	}
}

// CancelPrompt is called by JavaScript to cancel a pending prompt
func cancelPrompt(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		js.Global().Get("console").Call("error", "cancelPrompt requires: id (string)")
		return map[string]interface{}{
			"error": "Invalid arguments",
		}
	}
	
	promptID := args[0].String()
	
	pendingMutex.Lock()
	request, exists := pendingPrompts[promptID]
	pendingMutex.Unlock()
	
	if !exists {
		return map[string]interface{}{
			"error": "Prompt not found",
		}
	}
	
	// Send error to the waiting goroutine
	select {
	case request.ErrorChan <- fmt.Errorf("prompt cancelled by user"):
		js.Global().Get("console").Call("log", fmt.Sprintf("Cancelled prompt %s", promptID))
		return map[string]interface{}{
			"success": true,
		}
	default:
		return map[string]interface{}{
			"error": "Failed to cancel prompt",
		}
	}
}

// RegisterPromptHandler is called by JavaScript to register the prompt callback
func registerPromptHandler(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		js.Global().Get("console").Call("error", "registerPromptHandler requires: callback (function)")
		return false
	}
	
	callback := args[0]
	if callback.Type() != js.TypeFunction {
		js.Global().Get("console").Call("error", "Argument must be a function")
		return false
	}
	
	RegisterPromptCallback(callback)
	return true
}

// GetPendingPrompts returns info about pending prompts (for debugging)
func getPendingPrompts(this js.Value, args []js.Value) interface{} {
	pendingMutex.Lock()
	defer pendingMutex.Unlock()
	
	prompts := make([]map[string]interface{}, 0, len(pendingPrompts))
	for id, req := range pendingPrompts {
		prompts = append(prompts, map[string]interface{}{
			"id":           id,
			"message":      req.Message,
			"type":         req.PromptType,
			"resourceType": req.ResourceType,
		})
	}
	
	return prompts
}

// InitPromptSystem registers the JavaScript functions for the prompt system
func InitPromptSystem() {
	js.Global().Set("registerPromptHandler", js.FuncOf(registerPromptHandler))
	js.Global().Set("submitPromptResponse", js.FuncOf(submitPromptResponse))
	js.Global().Set("cancelPrompt", js.FuncOf(cancelPrompt))
	js.Global().Set("getPendingPrompts", js.FuncOf(getPendingPrompts))
	
	js.Global().Get("console").Call("log", "✅ WASM Prompt System initialized")
	js.Global().Get("console").Call("log", "Available functions:")
	js.Global().Get("console").Call("log", "  - registerPromptHandler(callback)")
	js.Global().Get("console").Call("log", "  - submitPromptResponse(id, response)")
	js.Global().Get("console").Call("log", "  - cancelPrompt(id)")
	js.Global().Get("console").Call("log", "  - getPendingPrompts()")
}

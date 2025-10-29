//go:build js && wasm
// +build js,wasm

package api

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"
)

// FetchResponse represents a response from a direct fetch call
type FetchResponse struct {
	StatusCode int
	Body       []byte
	Error      error
}

// MakeDirectFetch makes a direct fetch call to the API without going through a proxy
func MakeDirectFetch(url string, token string) (*FetchResponse, error) {
	console := js.Global().Get("console")
	console.Call("log", fmt.Sprintf("🚀 Making direct fetch to: %s", url))

	startTime := time.Now()

	// Create fetch options with compression support
	headers := map[string]interface{}{
		"Authorization":   "Bearer " + token,
		"Accept":          "application/json",
		"Accept-Encoding": "gzip, deflate, br", // Request compressed responses
	}

	options := map[string]interface{}{
		"method":  "GET",
		"headers": headers,
	}

	// Make the fetch call
	promise := js.Global().Call("fetch", url, options)

	// Create channels for async handling
	resultChan := make(chan *FetchResponse, 1)
	errorChan := make(chan error, 1)

	// Handle the promise
	var thenFunc, catchFunc js.Func
	var responseStatus int

	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer thenFunc.Release()
		defer catchFunc.Release()

		if len(args) == 0 {
			errorChan <- fmt.Errorf("no response from fetch")
			return nil
		}

		response := args[0]
		responseStatus = response.Get("status").Int()

		// Check if response was compressed
		contentEncoding := ""
		if !response.Get("headers").IsUndefined() {
			headers := response.Get("headers")
			if !headers.Get("get").IsUndefined() {
				encoding := headers.Call("get", "content-encoding")
				if !encoding.IsNull() && !encoding.IsUndefined() {
					contentEncoding = encoding.String()
				}
			}
		}

		if contentEncoding != "" {
			console.Call("log", fmt.Sprintf("📥 Response status: %d (compressed: %s)", responseStatus, contentEncoding))
		} else {
			console.Call("log", fmt.Sprintf("📥 Response status: %d", responseStatus))
		}

		// Create handlers for text promise
		var textThen, textCatch js.Func

		textThen = js.FuncOf(func(this js.Value, textArgs []js.Value) interface{} {
			defer textThen.Release()
			defer textCatch.Release()

			if len(textArgs) == 0 {
				errorChan <- fmt.Errorf("no body in response")
				return nil
			}

			bodyText := textArgs[0].String()
			elapsed := time.Since(startTime).Milliseconds()

			console.Call("log", fmt.Sprintf("✅ Received %d bytes in %dms", len(bodyText), elapsed))

			resultChan <- &FetchResponse{
				StatusCode: responseStatus,
				Body:       []byte(bodyText),
				Error:      nil,
			}

			return nil
		})

		textCatch = js.FuncOf(func(this js.Value, textArgs []js.Value) interface{} {
			defer textThen.Release()
			defer textCatch.Release()

			errMsg := "unknown error reading response body"
			if len(textArgs) > 0 {
				errMsg = textArgs[0].Get("message").String()
			}

			console.Call("error", fmt.Sprintf("❌ Error reading body: %s", errMsg))
			errorChan <- fmt.Errorf("error reading response body: %s", errMsg)
			return nil
		})

		// Read the response body as text
		textPromise := response.Call("text")
		textPromise.Call("then", textThen).Call("catch", textCatch)
		return nil
	})

	catchFunc = js.FuncOf(func(this js.Value, catchArgs []js.Value) interface{} {
		defer thenFunc.Release()
		defer catchFunc.Release()

		errMsg := "unknown fetch error"
		if len(catchArgs) > 0 {
			errMsg = catchArgs[0].Get("message").String()
		}

		console.Call("error", fmt.Sprintf("❌ Fetch failed: %s", errMsg))
		errorChan <- fmt.Errorf("fetch failed: %s", errMsg)
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc) // Wait for result or error (with timeout)
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("fetch timeout after 60 seconds")
	}
}

// MakeDirectFetchJSON makes a direct fetch and unmarshals JSON response
func MakeDirectFetchJSON(url string, token string, target interface{}) error {
	response, err := MakeDirectFetch(url, token)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d: %s", response.StatusCode, string(response.Body))
	}

	if err := json.Unmarshal(response.Body, target); err != nil {
		return fmt.Errorf("failed to parse JSON response: %v", err)
	}

	return nil
}

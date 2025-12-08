//go:build js && wasm
// +build js,wasm

package wasmhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"syscall/js"
	"time"
)

// WasmHTTPTransport implements http.RoundTripper using browser fetch API
// This allows the standard megaportgo SDK to work in WASM without modification
type WasmHTTPTransport struct {
	Timeout time.Duration // Request timeout (default: 60s)
}

// NewWasmHTTPClient returns an http.Client configured for WASM using browser fetch
func NewWasmHTTPClient() *http.Client {
	return &http.Client{
		Transport: &WasmHTTPTransport{
			Timeout: 60 * time.Second,
		},
	}
}

// RoundTrip executes a single HTTP transaction using browser fetch API
// This is the core method that bridges Go's http.Client with JavaScript fetch
func (t *WasmHTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	console := js.Global().Get("console")

	// Determine timeout
	timeout := t.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	// Log the request
	console.Call("log", fmt.Sprintf("🌐 HTTP %s %s", req.Method, req.URL.String()))

	// Convert http.Request to fetch options
	fetchOpts, err := t.buildFetchOptions(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build fetch options: %w", err)
	}

	// Make the fetch call
	response, err := t.doFetch(req.URL.String(), fetchOpts, timeout)
	if err != nil {
		return nil, err
	}

	// Convert fetch response to http.Response
	httpResponse := t.buildHTTPResponse(response, req)

	console.Call("log", fmt.Sprintf("✅ HTTP %d %s (%d bytes)",
		httpResponse.StatusCode,
		req.URL.String(),
		len(response.Body)))

	return httpResponse, nil
}

// buildFetchOptions converts an http.Request to fetch API options
func (t *WasmHTTPTransport) buildFetchOptions(req *http.Request) (map[string]interface{}, error) {
	// Build headers map
	headers := make(map[string]interface{})
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0] // fetch expects single string values
		}
	}

	// Add compression support
	if _, exists := headers["Accept-Encoding"]; !exists {
		headers["Accept-Encoding"] = "gzip, deflate, br"
	}

	fetchOpts := map[string]interface{}{
		"method":  req.Method,
		"headers": headers,
	}

	// Add body for POST, PUT, PATCH, DELETE requests
	if req.Body != nil && req.Method != http.MethodGet && req.Method != http.MethodHead {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}

		// Restore body for potential retries
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		if len(bodyBytes) > 0 {
			fetchOpts["body"] = string(bodyBytes)
		}
	}

	return fetchOpts, nil
}

// doFetch performs the actual fetch call and handles the promise
func (t *WasmHTTPTransport) doFetch(url string, options map[string]interface{}, timeout time.Duration) (*fetchResponse, error) {
	console := js.Global().Get("console")
	startTime := time.Now()

	// Make the fetch call
	promise := js.Global().Call("fetch", url, options)

	// Channels for async response handling
	resultChan := make(chan *fetchResponse, 1)
	errorChan := make(chan error, 1)

	var thenFunc, catchFunc js.Func
	var responseStatus int
	var responseHeaders map[string]string

	// Success handler
	thenFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer thenFunc.Release()
		defer catchFunc.Release()

		if len(args) == 0 {
			errorChan <- fmt.Errorf("no response from fetch")
			return nil
		}

		response := args[0]
		responseStatus = response.Get("status").Int()
		responseHeaders = t.extractHeaders(response)

		// Log compression if present
		if encoding, ok := responseHeaders["content-encoding"]; ok && encoding != "" {
			console.Call("log", fmt.Sprintf("📥 Response status: %d (compressed: %s)", responseStatus, encoding))
		} else {
			console.Call("log", fmt.Sprintf("📥 Response status: %d", responseStatus))
		}

		// Create handlers for reading response body
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

			console.Call("log", fmt.Sprintf("📦 Received %d bytes in %dms", len(bodyText), elapsed))

			resultChan <- &fetchResponse{
				StatusCode: responseStatus,
				Headers:    responseHeaders,
				Body:       []byte(bodyText),
			}

			return nil
		})

		textCatch = js.FuncOf(func(this js.Value, textArgs []js.Value) interface{} {
			defer textThen.Release()
			defer textCatch.Release()

			errMsg := "unknown error reading response body"
			if len(textArgs) > 0 && !textArgs[0].IsUndefined() {
				if msg := textArgs[0].Get("message"); !msg.IsUndefined() {
					errMsg = msg.String()
				}
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

	// Error handler
	catchFunc = js.FuncOf(func(this js.Value, catchArgs []js.Value) interface{} {
		defer thenFunc.Release()
		defer catchFunc.Release()

		errMsg := "unknown fetch error"
		if len(catchArgs) > 0 && !catchArgs[0].IsUndefined() {
			if msg := catchArgs[0].Get("message"); !msg.IsUndefined() {
				errMsg = msg.String()
			}
		}

		console.Call("error", fmt.Sprintf("❌ Fetch failed: %s", errMsg))
		errorChan <- fmt.Errorf("fetch failed: %s", errMsg)
		return nil
	})

	promise.Call("then", thenFunc).Call("catch", catchFunc)

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("fetch timeout after %v", timeout)
	}
}

// extractHeaders extracts headers from the fetch Response object
func (t *WasmHTTPTransport) extractHeaders(response js.Value) map[string]string {
	headers := make(map[string]string)

	if response.Get("headers").IsUndefined() {
		return headers
	}

	jsHeaders := response.Get("headers")

	// Try to use forEach if available
	if !jsHeaders.Get("forEach").IsUndefined() {
		forEachFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) >= 2 {
				value := args[0].String()
				key := args[1].String()
				headers[key] = value
			}
			return nil
		})
		defer forEachFunc.Release()

		jsHeaders.Call("forEach", forEachFunc)
	}

	return headers
}

// buildHTTPResponse converts a fetchResponse to http.Response
func (t *WasmHTTPTransport) buildHTTPResponse(fetchResp *fetchResponse, req *http.Request) *http.Response {
	// Build http.Header from map
	header := make(http.Header)
	for key, value := range fetchResp.Headers {
		header.Set(key, value)
	}

	return &http.Response{
		Status:        http.StatusText(fetchResp.StatusCode),
		StatusCode:    fetchResp.StatusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(fetchResp.Body)),
		ContentLength: int64(len(fetchResp.Body)),
		Request:       req,
	}
}

// fetchResponse represents the result of a fetch call
type fetchResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

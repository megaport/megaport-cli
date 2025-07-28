//go:build js && wasm
// +build js,wasm

package api

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"syscall/js"
	"time"
)

// RequestState tracks everything about an API request
type RequestState struct {
	RequestID       string // Changed from int to string to match usage
	URL             string
	StartTime       time.Time
	MaxWaitTime     time.Duration
	PollInterval    time.Duration
	HeaderReceived  bool
	NotifyChan      chan bool
	PromiseResolved chan bool
	StatusObj       js.Value
	LastStatus      string
	Completed       bool   // Added missing field
	Error           error  // Added missing field
	Result          []byte // Added missing field
}

// Global request tracking
var (
	requestNotifications = make(map[string]chan bool) // Changed from int to string
	requestMutex         sync.Mutex
)

// init registers the JS callback into api.requestNotifications
func init() {
	js.Global().Set("notifyRequestComplete",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				js.Global().Get("console").Call("warn", "notifyRequestComplete called with no args")
				return nil
			}

			// Updated to handle string IDs
			var id string
			if args[0].Type() == js.TypeString {
				id = args[0].String()
			} else {
				id = fmt.Sprintf("%d", args[0].Int())
			}

			js.Global().Get("console").Call("log", fmt.Sprintf("Go: notifyRequestComplete called from JS for ID %s", id))

			requestMutex.Lock()
			ch, ok := requestNotifications[id]
			requestMutex.Unlock() // Unlock sooner

			if ok {
				js.Global().Get("console").Call("log", fmt.Sprintf("Go: Found channel for ID %s. Attempting to send notification.", id))
				select {
				case ch <- true:
					js.Global().Get("console").Call("log", fmt.Sprintf("Go: Notification sent to channel for ID %s.", id))
				default:
					js.Global().Get("console").Call("log", fmt.Sprintf("Go: Channel for ID %s not ready (full or closed).", id))
				}
			} else {
				js.Global().Get("console").Call("warn", fmt.Sprintf("Go: No active notification channel found for ID %s.", id))
			}
			return nil
		}),
	)
}

// Add this new function to convert API URLs to proxy URLs
func convertToProxyURL(originalURL string) (string, error) {
	// Parse the original URL
	parsedURL := js.Global().Get("URL").New(originalURL)

	// Extract the hostname and path
	hostname := parsedURL.Get("hostname").String()
	path := parsedURL.Get("pathname").String()

	// Remove leading slash from path if present
	path = strings.TrimPrefix(path, "/")

	// Construct proxy URL
	proxyURL := fmt.Sprintf("/proxy/%s?base=%s", path, hostname)

	js.Global().Get("console").Call("log",
		fmt.Sprintf("Converting API URL %s to proxy URL %s", originalURL, proxyURL))

	return proxyURL, nil
}

// Then modify MakeProxiedRequest to use this function
func MakeProxiedRequest(jsThis js.Value, url string, token string, options js.Value) (RequestState, error) {
	// Create unique request ID - fixed to handle the float value correctly
	randomValue := js.Global().Get("Math").Call("random").Float()
	randomInt := int64(randomValue * 1000000) // Convert to integer for ID generation
	requestID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), randomInt)

	// Debug output
	js.Global().Get("console").Call("log", fmt.Sprintf("Starting proxied request to %s with ID %s", url, requestID))

	// Convert API URL to proxy URL
	proxyURL, err := convertToProxyURL(url)
	if err != nil {
		return RequestState{}, fmt.Errorf("invalid API URL: %v", err)
	}

	// Debug output
	js.Global().Get("console").Call("log", fmt.Sprintf("Making fetch request to proxy: %s", proxyURL))

	// Create promise-based request with the converted proxy URL
	js.Global().Call("directBrowserFetch", proxyURL, token,
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return nil // Success handler - we don't use this directly anymore
		}),
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return nil // Error handler - we don't use this directly anymore
		}))

	// Rest of the function remains the same
	// ...

	// Poll for result with timeout (30 seconds)
	startTime := time.Now()
	timeout := 30 * time.Second

	for time.Since(startTime) < timeout {
		// Check if result is available
		status := js.Global().Call("checkFetchRequestStatus", requestID)

		if status.Get("completed").Bool() {
			js.Global().Get("console").Call("log", fmt.Sprintf("Received response for request %s", requestID))

			if status.Get("success").Bool() {
				// Get the result data
				result := []byte(status.Get("data").String())
				return RequestState{
					Result: result,
					Error:  nil,
				}, nil
			} else {
				errMsg := status.Get("error").String()
				return RequestState{}, fmt.Errorf(errMsg)
			}
		}

		// Wait a bit before checking again
		time.Sleep(100 * time.Millisecond)
	}

	// Timeout occurred
	return RequestState{}, errors.New("request timed out after 30 seconds")
}

// CleanupRequest performs cleanup operations when done with a request
func CleanupRequest(state *RequestState) {
	requestMutex.Lock()
	defer requestMutex.Unlock()

	delete(requestNotifications, state.RequestID)
	if state.NotifyChan != nil {
		close(state.NotifyChan)
	}

	// Close promise channel if it exists
	if state.PromiseResolved != nil {
		close(state.PromiseResolved)
	}
}

// YieldToJS yields to the JavaScript event loop and waits for callback
func YieldToJS() {
	yieldComplete := make(chan struct{}, 1)
	js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		yieldComplete <- struct{}{}
		return nil
	}), 0)

	// Wait for yield to complete before continuing
	select {
	case <-yieldComplete:
		// Continue processing
	case <-time.After(100 * time.Millisecond):
		// Timeout on yield - continue anyway
	}
}

// DirectNotificationCheck does a quick JS lookup to see if the URL has already returned
func DirectNotificationCheck(state *RequestState) bool {
	if !js.Global().Get("checkResponseLookup").IsUndefined() {
		status := js.Global().Call("checkResponseLookup", state.URL)
		if !status.IsNull() && status.Get("complete").Bool() {
			js.Global().Get("console").Call("log",
				fmt.Sprintf("Direct lookup: complete response found for %s", state.URL))
			return true
		}
	}
	return false
}

// InitiateRequest starts a WebAssembly API request and returns a RequestState.
func InitiateRequest(url, token string, options map[string]interface{}) (*RequestState, error) {
	// build JS options object
	jsOpts := js.Global().Get("Object").New()
	for k, v := range options {
		jsOpts.Set(k, v)
	}

	// ensure the JS bridge exists
	if js.Global().Get("startApiRequest").IsUndefined() {
		return nil, fmt.Errorf("startApiRequest is not defined in JS")
	}

	// fire off the request and get ID (convert to string)
	requestIdVal := js.Global().Call("startApiRequest", url, token, jsOpts)
	var requestId string
	if requestIdVal.Type() == js.TypeString {
		requestId = requestIdVal.String()
	} else {
		requestId = fmt.Sprintf("%d", requestIdVal.Int())
	}

	js.Global().Get("console").Call("log", fmt.Sprintf(
		"InitiateRequest → URL=%s, id=%s", url, requestId,
	))

	// register the optional JS status listener
	if !js.Global().Get("registerWasmStatusListener").IsUndefined() {
		js.Global().Call("registerWasmStatusListener", requestId)
	}

	// create and record the Go notification channel
	notifyChan := make(chan bool, 1)
	requestMutex.Lock()
	requestNotifications[requestId] = notifyChan
	requestMutex.Unlock()

	// create promise channel in case JS uses waitUntilRequestComplete
	promiseResolved := make(chan bool, 1)
	if !js.Global().Get("waitUntilRequestComplete").IsUndefined() {
		cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			select {
			case promiseResolved <- true:
			default:
			}
			return nil
		})
		js.Global().
			Call("waitUntilRequestComplete", requestId, 60000).
			Call("then", cb).
			Call("catch", cb)
	}

	return &RequestState{
		RequestID:       requestId, // Changed from ID to RequestID
		URL:             url,
		StartTime:       time.Now(),
		MaxWaitTime:     45 * time.Second,
		PollInterval:    500 * time.Millisecond,
		NotifyChan:      notifyChan,
		PromiseResolved: promiseResolved,
	}, nil
}

// CheckForNotifications tries multiple channels in order:
//  1. Direct JS lookup via checkResponseLookup()
//  2. The promiseResolved channel
//  3. The notifyChan channel (driven by window.notifyRequestComplete)
//  4. A brief 5ms blocking wait
//
// If any notify is seen, it also calls forceRequestStatusCheck(id).
func CheckForNotifications(state *RequestState) bool {
	// 1) direct JS lookup
	if !js.Global().Get("checkResponseLookup").IsUndefined() {
		stat := js.Global().Call("checkResponseLookup", state.URL)
		if !stat.IsNull() && stat.Get("complete").Bool() {
			return true
		}
	}

	got := false

	// 2) promiseResolved
	select {
	case <-state.PromiseResolved:
		got = true
	default:
	}

	// 3) notifyChan
	select {
	case <-state.NotifyChan:
		got = true
	default:
	}

	// 4) brief blocking fallback
	if !got {
		select {
		case <-state.PromiseResolved:
			got = true
		case <-state.NotifyChan:
			got = true
		case <-time.After(5 * time.Millisecond):
		}
	}

	// if we did get a notification, force JS to re-check status immediately
	if got && !js.Global().Get("forceRequestStatusCheck").IsUndefined() {
		js.Global().Call("forceRequestStatusCheck", state.RequestID)
	}
	return got
}

// UpdateRequestStatus fetches the current status from JavaScript
func UpdateRequestStatus(state *RequestState) string {
	state.StatusObj = js.Global().Call("checkRequestStatus", state.RequestID)
	state.LastStatus = state.StatusObj.Get("status").String()
	return state.LastStatus
}

// HandleHeadersReceived processes the "headers_received" status
func HandleHeadersReceived(state *RequestState) {
	if !state.HeaderReceived {
		state.HeaderReceived = true
		headerInfo := state.StatusObj.Get("headers")
		if !headerInfo.IsUndefined() && !headerInfo.Get("contentLength").IsUndefined() {
			contentLength := headerInfo.Get("contentLength").String()
			js.Global().Get("console").Call("log", fmt.Sprintf("Headers received for #%s, content length: %s bytes",
				state.RequestID, contentLength))

			// Check if status is 200 and add more time immediately
			if headerInfo.Get("status").Int() == 200 {
				contentLengthInt := 0
				fmt.Sscanf(contentLength, "%d", &contentLengthInt)

				// Calculate additional time based on content size
				extraTime := 60 * time.Second
				if contentLengthInt > 100000 {
					// Add 3 minutes for very large responses
					extraTime = 180 * time.Second
				}

				// Reset the MaxWaitTime completely
				state.MaxWaitTime = time.Now().Add(extraTime).Sub(state.StartTime)
				js.Global().Get("console").Call("log",
					fmt.Sprintf("⏰ Reset timeout! New total wait time: %v for %d bytes",
						state.MaxWaitTime, contentLengthInt))

				// Also reduce polling interval for more responsiveness
				state.PollInterval = 250 * time.Millisecond
			}
		}
	}
}

// HandleProcessingStatus handles the "processing" status
func HandleProcessingStatus(state *RequestState) {
	// If we get a processing notification, reset the poll interval
	state.PollInterval = 250 * time.Millisecond
	js.Global().Get("console").Call("log",
		fmt.Sprintf("Worker is processing large response, reset poll interval to %v",
			state.PollInterval))

	// Add more time if processing large response
	extraTime := 60 * time.Second
	elapsed := time.Since(state.StartTime)
	if elapsed+extraTime > state.MaxWaitTime {
		state.MaxWaitTime = elapsed + extraTime
		js.Global().Get("console").Call("log",
			fmt.Sprintf("Extended timeout to %v for processing", state.MaxWaitTime))
	}
}

// HandlePartialResult handles the "partial_result" status
func HandlePartialResult(state *RequestState) {
	js.Global().Get("console").Call("log", "Received partial result, waiting for complete data")

	// Add more time to wait for full result
	extraTime := 60 * time.Second
	elapsed := time.Since(state.StartTime)
	if elapsed+extraTime > state.MaxWaitTime {
		state.MaxWaitTime = elapsed + extraTime
		js.Global().Get("console").Call("log",
			fmt.Sprintf("Extended timeout to %v to wait for full result", state.MaxWaitTime))
	}

	// Decrease polling interval for faster updates
	state.PollInterval = 250 * time.Millisecond
}

// LogRequestStatus logs status at appropriate intervals
func LogRequestStatus(state *RequestState) {
	elapsed := time.Since(state.StartTime).Seconds()
	if elapsed < 5 || int(elapsed)%5 == 0 {
		elapsedMs := state.StatusObj.Get("elapsedMs").Float()
		js.Global().Get("console").Call("log", fmt.Sprintf(
			"Request #%s status: %s (%.1f seconds elapsed, JS reports %.1f ms)",
			state.RequestID, state.LastStatus, elapsed, elapsedMs)) // Changed %d to %s
	}
}

// AdjustPollingInterval adjusts polling based on elapsed time
func AdjustPollingInterval(state *RequestState) {
	elapsed := time.Since(state.StartTime)
	if elapsed > 10*time.Second && state.PollInterval < 1*time.Second {
		state.PollInterval = 1 * time.Second
	} else if elapsed > 30*time.Second && state.PollInterval < 2*time.Second {
		state.PollInterval = 2 * time.Second
	}
}

// ForceStatusCheck performs a force notification check
func ForceStatusCheck(state *RequestState) {
	elapsed := time.Since(state.StartTime).Seconds()
	if int(elapsed)%10 == 0 {
		js.Global().Get("console").Call("log", "Performing force notification check")
		if !js.Global().Get("forceRequestStatusCheck").IsUndefined() {
			js.Global().Call("forceRequestStatusCheck", state.RequestID)
		}
	}
}

// HasPartialResult checks if there's a partial result in the response
func HasPartialResult(state *RequestState) bool {
	return state.StatusObj.Get("partialResult").Type() == js.TypeString
}

// ExtendTimeout adds more time to the request timeout
func ExtendTimeout(state *RequestState, extraTime time.Duration, reason string) {
	elapsed := time.Since(state.StartTime)
	if elapsed+extraTime > state.MaxWaitTime {
		state.MaxWaitTime = elapsed + extraTime
		js.Global().Get("console").Call("log",
			fmt.Sprintf("Extended timeout to %v: %s", state.MaxWaitTime, reason))
	}
}

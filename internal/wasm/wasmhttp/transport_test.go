//go:build js && wasm

package wasmhttp

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"syscall/js"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// installMockFetch replaces the global fetch for the duration of a test and
// restores the original on cleanup. behavior receives the request URL and the
// fetch options (including the AbortController's signal) and must return a
// JS Promise, mirroring what the real browser fetch returns.
func installMockFetch(t *testing.T, behavior func(url string, opts js.Value) js.Value) {
	t.Helper()

	original := js.Global().Get("fetch")
	fetchFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return behavior(args[0].String(), args[1])
	})
	js.Global().Set("fetch", fetchFunc)

	t.Cleanup(func() {
		js.Global().Set("fetch", original)
		fetchFunc.Release()
	})
}

// newHangingFetch returns fetch behavior that never resolves on its own, only
// rejecting once the request's AbortSignal fires. aborted is set to 1 when
// that happens, so tests can assert the browser fetch was actually cancelled.
func newHangingFetch(aborted *int32) func(url string, opts js.Value) js.Value {
	return func(url string, opts js.Value) js.Value {
		signal := opts.Get("signal")

		var executor js.Func
		executor = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer executor.Release()
			reject := args[1]

			var onAbort js.Func
			onAbort = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				defer onAbort.Release()
				atomic.StoreInt32(aborted, 1)
				abortErr := js.Global().Get("Error").New("The operation was aborted")
				abortErr.Set("name", "AbortError")
				reject.Invoke(abortErr)
				return nil
			})

			if !signal.IsUndefined() {
				signal.Call("addEventListener", "abort", onAbort)
			}
			return nil
		})

		return js.Global().Get("Promise").New(executor)
	}
}

// newResolvingFetch returns fetch behavior that resolves immediately with a
// minimal Response-like object carrying status and a JSON body.
func newResolvingFetch(status int, body string) func(url string, opts js.Value) js.Value {
	return func(url string, opts js.Value) js.Value {
		var executor js.Func
		executor = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			defer executor.Release()
			resolve := args[0]
			resolve.Invoke(newFakeResponse(status, body))
			return nil
		})
		return js.Global().Get("Promise").New(executor)
	}
}

func newFakeResponse(status int, body string) js.Value {
	var forEachFunc js.Func
	forEachFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer forEachFunc.Release()
		return nil
	})

	headers := js.ValueOf(map[string]interface{}{})
	headers.Set("forEach", forEachFunc)

	var textFunc js.Func
	textFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer textFunc.Release()
		return js.Global().Get("Promise").Call("resolve", body)
	})

	resp := js.ValueOf(map[string]interface{}{
		"status": status,
	})
	resp.Set("headers", headers)
	resp.Set("text", textFunc)
	return resp
}

func newTestRequest(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.invalid/test", nil)
	require.NoError(t, err)
	return req
}

func TestRoundTrip_ContextCancelledReturnsPromptly(t *testing.T) {
	var aborted int32
	installMockFetch(t, newHangingFetch(&aborted))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	transport := &WasmHTTPTransport{Timeout: 5 * time.Second}
	req := newTestRequest(t, ctx)

	start := time.Now()
	resp, err := transport.RoundTrip(req)
	elapsed := time.Since(start)

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled), "expected context.Canceled, got %v", err)
	assert.Less(t, elapsed, 2*time.Second, "cancellation should not wait for the full transport timeout")
	assert.Equal(t, int32(1), atomic.LoadInt32(&aborted), "expected the browser fetch to be aborted")
}

func TestRoundTrip_TimeoutAbortsAndReturnsError(t *testing.T) {
	var aborted int32
	installMockFetch(t, newHangingFetch(&aborted))

	transport := &WasmHTTPTransport{Timeout: 50 * time.Millisecond}
	req := newTestRequest(t, context.Background())

	start := time.Now()
	resp, err := transport.RoundTrip(req)
	elapsed := time.Since(start)

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.Less(t, elapsed, 2*time.Second, "timeout error should arrive close to the configured timeout")
	assert.Equal(t, int32(1), atomic.LoadInt32(&aborted), "expected the browser fetch to be aborted on timeout")
}

func TestRoundTrip_SuccessNormalResponseUnaffected(t *testing.T) {
	installMockFetch(t, newResolvingFetch(http.StatusOK, `{"ok":true}`))

	transport := &WasmHTTPTransport{Timeout: 5 * time.Second}
	req := newTestRequest(t, context.Background())

	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := make([]byte, 32)
	n, _ := resp.Body.Read(body)
	assert.Equal(t, `{"ok":true}`, string(body[:n]))
}

func TestBuildFetchOptions_NoForbiddenAcceptEncodingHeader(t *testing.T) {
	transport := &WasmHTTPTransport{}
	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/test", nil)
	require.NoError(t, err)

	opts, err := transport.buildFetchOptions(req)
	require.NoError(t, err)

	headers, ok := opts["headers"].(map[string]interface{})
	require.True(t, ok)

	_, exists := headers["Accept-Encoding"]
	assert.False(t, exists, "Accept-Encoding is a forbidden fetch header and must not be set")
}

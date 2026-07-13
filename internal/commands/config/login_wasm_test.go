//go:build js && wasm

package config

import (
	"context"
	"net/http"
	"syscall/js"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setTokenManager installs a mock window.tokenManager whose getToken always
// returns result, restoring the previous global (if any) after the test.
func setTokenManager(t *testing.T, result interface{}) {
	t.Helper()
	prev := js.Global().Get("tokenManager")
	tm := js.Global().Get("Object").New()
	tm.Set("getToken", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		return result
	}))
	js.Global().Set("tokenManager", tm)
	t.Cleanup(func() {
		if prev.IsUndefined() {
			js.Global().Delete("tokenManager")
		} else {
			js.Global().Set("tokenManager", prev)
		}
	})
}

func TestCheckCachedToken_NoTokenManager(t *testing.T) {
	prev := js.Global().Get("tokenManager")
	js.Global().Delete("tokenManager")
	defer func() {
		if !prev.IsUndefined() {
			js.Global().Set("tokenManager", prev)
		}
	}()

	assert.Nil(t, CheckCachedToken("production"))
}

func TestCheckCachedToken_NoCachedToken(t *testing.T) {
	setTokenManager(t, js.Null())
	assert.Nil(t, CheckCachedToken("production"))
}

func TestCheckCachedToken_LegacyStringFallsBackTo24h(t *testing.T) {
	setTokenManager(t, "legacy-bare-token")

	before := time.Now()
	auth := CheckCachedToken("production")
	require.NotNil(t, auth)
	assert.Equal(t, "legacy-bare-token", auth.AccessToken)

	// Historical fallback: ~24h from now, not zero and not the token's own TTL
	// (there isn't one on the legacy contract).
	assert.WithinDuration(t, before.Add(24*time.Hour), auth.Expiration, 5*time.Second)
}

func TestCheckCachedToken_ObjectWithExpiryHonorsRealTTL(t *testing.T) {
	want := time.Date(2031, 3, 4, 5, 6, 7, 0, time.UTC)

	tokenObj := js.Global().Get("Object").New()
	tokenObj.Set("token", "object-form-token")
	tokenObj.Set("expiry", float64(want.UnixMilli()))
	setTokenManager(t, tokenObj)

	auth := CheckCachedToken("production")
	require.NotNil(t, auth)
	assert.Equal(t, "object-form-token", auth.AccessToken)
	assert.True(t, want.Equal(auth.Expiration), "expiration = %v, want %v", auth.Expiration, want)
}

func TestCheckCachedToken_ObjectWithoutExpiryFallsBackTo24h(t *testing.T) {
	tokenObj := js.Global().Get("Object").New()
	tokenObj.Set("token", "object-form-token-no-expiry")
	setTokenManager(t, tokenObj)

	before := time.Now()
	auth := CheckCachedToken("production")
	require.NotNil(t, auth)
	assert.WithinDuration(t, before.Add(24*time.Hour), auth.Expiration, 5*time.Second)
}

func TestCheckCachedToken_ObjectWithEmptyTokenIsTreatedAsNoToken(t *testing.T) {
	tokenObj := js.Global().Get("Object").New()
	setTokenManager(t, tokenObj)

	assert.Nil(t, CheckCachedToken("production"))
}

func TestRetryWithBackoffAndConsoleLogging_NilClient(t *testing.T) {
	_, err := RetryWithBackoffAndConsoleLogging(context.Background(), 1, nil)
	require.Error(t, err)
}

// TestRetryWithBackoffAndConsoleLogging_UnrecognisedHostUsesCache verifies that
// a non-standard API host still buckets to production for the token cache
// (rather than erroring) and, critically, that the cached-token short-circuit
// still fires before any network Authorize() call would happen.
func TestRetryWithBackoffAndConsoleLogging_UnrecognisedHostUsesCache(t *testing.T) {
	setTokenManager(t, "cached-for-unrecognised-host")

	client, err := megaport.New(http.DefaultClient, megaport.WithBaseURL("https://api-custom-env.example.com"))
	require.NoError(t, err)

	auth, err := RetryWithBackoffAndConsoleLogging(context.Background(), 1, client)
	require.NoError(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, "cached-for-unrecognised-host", auth.AccessToken)
}

//go:build js && wasm

package config

import (
	"context"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setTamperedTokenGlobal publishes a window.megaportToken global whose apiURL
// field has been overwritten to an attacker-controlled host, mimicking a script
// tampering with the page-writable global after a legitimate setAuthToken call.
func setTamperedTokenGlobal(env, apiURL string) {
	obj := js.Global().Get("Object").New()
	obj.Set("environment", env)
	obj.Set("apiURL", apiURL)
	js.Global().Set("megaportToken", obj)
}

// TestLoginReadsAPIURLFromEnvNotTamperedGlobal asserts the authenticated token
// login path routes to the host stored in MEGAPORT_API_URL by setAuthToken, even
// when window.megaportToken.apiURL has been overwritten with a hostile value.
func TestLoginReadsAPIURLFromEnvNotTamperedGlobal(t *testing.T) {
	const validHost = "api-staging.megaport.com"
	const tamperedURL = "https://evil.attacker.com"

	t.Setenv("MEGAPORT_ACCESS_TOKEN", "test-token-12345")
	t.Setenv("MEGAPORT_API_URL", "https://"+validHost)

	setTamperedTokenGlobal("staging", tamperedURL)
	defer js.Global().Delete("megaportToken")

	client, err := loginFunc(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.BaseURL)

	assert.Equal(t, validHost, client.BaseURL.Host,
		"login must route to the validated env-var host, not the tampered global")
}

// TestLoginFallbackEnvIgnoresTamperedGlobal asserts that when MEGAPORT_API_URL
// is unset and login falls back to environment-based host selection, it uses the
// bucket stored in MEGAPORT_ENVIRONMENT rather than the page-writable global.
func TestLoginFallbackEnvIgnoresTamperedGlobal(t *testing.T) {
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "test-token-12345")
	t.Setenv("MEGAPORT_API_URL", "")
	t.Setenv("MEGAPORT_ENVIRONMENT", "staging")

	// Global claims production; the env-var bucket (staging) must win.
	setTamperedTokenGlobal("production", "https://evil.attacker.com")
	defer js.Global().Delete("megaportToken")

	client, err := loginFunc(context.Background())
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.BaseURL)

	assert.Equal(t, "api-staging.megaport.com", client.BaseURL.Host,
		"fallback host selection must use the env-var bucket, not the tampered global")
}

// TestUnauthenticatedClientReadsAPIURLFromEnvNotTamperedGlobal asserts the same
// for the unauthenticated client factory used by public endpoints.
func TestUnauthenticatedClientReadsAPIURLFromEnvNotTamperedGlobal(t *testing.T) {
	const validHost = "api-staging.megaport.com"
	const tamperedURL = "https://evil.attacker.com"

	t.Setenv("MEGAPORT_API_URL", "https://"+validHost)

	setTamperedTokenGlobal("staging", tamperedURL)
	defer js.Global().Delete("megaportToken")

	client, err := newUnauthenticatedClientFunc()
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.BaseURL)

	assert.Equal(t, validHost, client.BaseURL.Host,
		"unauthenticated client must route to the validated env-var host, not the tampered global")
}

// These tests call undefined JS globals (fetch, config-file/localStorage
// bridges) that only exist in a real browser host, so they panic under the
// headless node runtime. The `browser` tag keeps them out of the default CI
// run; a browser harness opts in with -tags js,wasm,browser.
//go:build js && wasm && browser

package wasm

import (
	"encoding/json"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResetOutputBuffers verifies that buffer reset clears all content
func TestResetOutputBuffers(t *testing.T) {
	// Write some data to buffers
	stdoutBuffer.WriteString("test stdout")
	stderrBuffer.WriteString("test stderr")
	_, _ = WasmOutputBuffer.Write([]byte("test direct"))

	// Set globals
	js.Global().Set("wasmJSONOutput", "test json")
	js.Global().Set("wasmCSVOutput", "test csv")
	js.Global().Set("wasmTableOutput", "test table")

	// Reset
	ResetOutputBuffers()

	// Verify all buffers are empty
	assert.Equal(t, "", stdoutBuffer.String(), "stdout buffer should be empty")
	assert.Equal(t, "", stderrBuffer.String(), "stderr buffer should be empty")
	assert.Equal(t, "", WasmOutputBuffer.String(), "direct buffer should be empty")

	// Verify globals are cleared
	assert.True(t, js.Global().Get("wasmJSONOutput").IsUndefined(), "wasmJSONOutput should be undefined")
	assert.True(t, js.Global().Get("wasmCSVOutput").IsUndefined(), "wasmCSVOutput should be undefined")
	assert.True(t, js.Global().Get("wasmTableOutput").IsUndefined(), "wasmTableOutput should be undefined")
}

// TestEnableDebugMode verifies debug mode activation
func TestEnableDebugMode(t *testing.T) {
	// Enable debug mode
	EnableDebugMode()

	// Verify debug mode is enabled
	assert.True(t, debugMode.Load())

	// Verify JS function is registered
	wasmDebugFunc := js.Global().Get("wasmDebug")
	assert.False(t, wasmDebugFunc.IsUndefined())

	// Call the JS function and verify it returns true
	result := wasmDebugFunc.Invoke()
	assert.Equal(t, true, result.Bool())
}

// TestRegisterJSFunctions verifies all JS functions are registered
func TestRegisterJSFunctions(t *testing.T) {
	RegisterJSFunctions()

	// List of expected functions
	expectedFunctions := []string{
		"readConfigFile",
		"writeConfigFile",
		"debugAuthInfo",
		"saveToLocalStorage",
		"loadFromLocalStorage",
		"resetWasmOutput",
		"getWasmOutput",
		"toggleWasmDebug",
		"dumpBuffers",
		"logLocationCommand",
	}

	// Verify each function is registered
	for _, funcName := range expectedFunctions {
		jsFunc := js.Global().Get(funcName)
		assert.False(t, jsFunc.IsUndefined(), "Function %s should be registered", funcName)
		assert.Equal(t, js.TypeFunction, jsFunc.Type(), "Function %s should be a function", funcName)
	}
}

// TestReadConfigFile_NotFound verifies handling of missing config
func TestReadConfigFile_NotFound(t *testing.T) {
	// Clear localStorage
	js.Global().Get("localStorage").Call("removeItem", "megaport_fs_test.json")

	// Call readConfigFile
	result := readConfigFile(js.Null(), []js.Value{js.ValueOf("test.json")})

	// Should return error map
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Contains(t, resultMap, "error")
}

// TestReadConfigFile_ConfigJSON verifies default config creation
func TestReadConfigFile_ConfigJSON(t *testing.T) {
	// Clear localStorage
	js.Global().Get("localStorage").Call("removeItem", "megaport_fs_config.json")

	// Call readConfigFile for config.json
	result := readConfigFile(js.Null(), []js.Value{js.ValueOf("config.json")})

	// Should return default config
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Contains(t, resultMap, "content")

	// Parse the content
	var configData map[string]interface{}
	contentStr, ok := resultMap["content"].(string)
	assert.True(t, ok, "content should be a string")
	err := json.Unmarshal([]byte(contentStr), &configData)
	assert.NoError(t, err)

	// Verify default structure
	assert.Contains(t, configData, "profiles")
	assert.Contains(t, configData, "activeProfile")
	assert.Contains(t, configData, "defaults")
}

// TestWriteConfigFile verifies file writing to localStorage
func TestWriteConfigFile(t *testing.T) {
	testContent := `{"test": "data"}`

	// Write file
	result := writeConfigFile(js.Null(), []js.Value{
		js.ValueOf("test.json"),
		js.ValueOf(testContent),
	})

	// Should return success
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	assert.Contains(t, resultMap, "success")
	assert.Equal(t, true, resultMap["success"])

	// Verify it was written to localStorage
	stored := js.Global().Get("localStorage").Call("getItem", "megaport_fs_test.json")
	assert.False(t, stored.IsNull())
	assert.Equal(t, testContent, stored.String())
}

// TestSaveToLocalStorage verifies localStorage save
func TestSaveToLocalStorage(t *testing.T) {
	result := saveToLocalStorage(js.Null(), []js.Value{
		js.ValueOf("test_key"),
		js.ValueOf("test_value"),
	})

	resultBool, ok := result.(bool)
	assert.True(t, ok, "result should be a bool")
	assert.Equal(t, true, resultBool)

	// Verify in localStorage
	stored := js.Global().Get("localStorage").Call("getItem", "test_key")
	assert.Equal(t, "test_value", stored.String())
}

// TestLoadFromLocalStorage verifies localStorage load
func TestLoadFromLocalStorage(t *testing.T) {
	// Set a value
	js.Global().Get("localStorage").Call("setItem", "test_key", "test_value")

	// Load it
	result := loadFromLocalStorage(js.Null(), []js.Value{
		js.ValueOf("test_key"),
	})

	resultVal, ok := result.(js.Value)
	assert.True(t, ok, "result should be a js.Value")
	assert.Equal(t, js.TypeString, resultVal.Type())
	assert.Equal(t, "test_value", resultVal.String())
}

// TestResetWasmOutput_JSFunction verifies JS function for resetting output
func TestResetWasmOutput_JSFunction(t *testing.T) {
	RegisterJSFunctions()

	// Add some content to buffers
	stdoutBuffer.WriteString("test")
	_, _ = WasmOutputBuffer.Write([]byte("test"))

	// Call the JS function
	resetFunc := js.Global().Get("resetWasmOutput")
	result := resetFunc.Invoke()

	// Should return true
	assert.Equal(t, true, result.Bool())

	// Buffers should be empty
	assert.Equal(t, "", stdoutBuffer.String())
	assert.Equal(t, "", WasmOutputBuffer.String())
}

// TestGetWasmOutput_JSFunction verifies JS function for getting output
func TestGetWasmOutput_JSFunction(t *testing.T) {
	RegisterJSFunctions()
	ResetOutputBuffers()

	// Add some content
	testOutput := "test output content"
	stdoutBuffer.WriteString(testOutput)

	// Call the JS function
	getFunc := js.Global().Get("getWasmOutput")
	result := getFunc.Invoke()

	// Should return the output
	assert.Equal(t, testOutput, result.String())
}

// TestDumpBuffers_JSFunction verifies JS function for dumping all buffers
func TestDumpBuffers_JSFunction(t *testing.T) {
	RegisterJSFunctions()
	ResetOutputBuffers()

	// Add content to different buffers
	stdoutBuffer.WriteString("stdout content")
	stderrBuffer.WriteString("stderr content")
	_, _ = WasmOutputBuffer.Write([]byte("direct content"))

	// Call the JS function
	dumpFunc := js.Global().Get("dumpBuffers")
	result := dumpFunc.Invoke()

	// Should return an object with all buffers
	assert.Equal(t, "stdout content", result.Get("stdout").String())
	assert.Equal(t, "stderr content", result.Get("stderr").String())
	assert.Equal(t, "direct content", result.Get("direct").String())
}

// TestToggleWasmDebug_JSFunction verifies JS function for toggling debug mode
func TestToggleWasmDebug_JSFunction(t *testing.T) {
	RegisterJSFunctions()

	// Get initial state
	toggleFunc := js.Global().Get("toggleWasmDebug")
	initialState := debugMode.Load()

	// Toggle
	result := toggleFunc.Invoke()

	// Should return opposite state
	assert.Equal(t, !initialState, result.Bool())
	assert.Equal(t, !initialState, debugMode.Load())

	// Toggle again
	result = toggleFunc.Invoke()
	assert.Equal(t, initialState, result.Bool())
	assert.Equal(t, initialState, debugMode.Load())
}

// TestSetAuthToken verifies token-based authentication with hostname mapping
func TestSetAuthToken(t *testing.T) {
	RegisterJSFunctions()

	tests := []struct {
		name           string
		token          string
		hostname       string
		explicitEnv    string
		expectError    bool
		expectedEnv    string // real env name returned in the JS result
		expectedURL    string
		expectedBucket string // MEGAPORT_ENVIRONMENT — always one of production/staging/development
	}{
		// --- hostname-derived cases (no override) ---
		{
			name:           "production portal host",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			hostname:       "portal.megaport.com",
			expectedEnv:    "production",
			expectedURL:    "https://api.megaport.com/",
			expectedBucket: "production",
		},
		{
			name:           "staging portal host",
			token:          "valid-staging-token-12345",
			hostname:       "portal-staging.megaport.com",
			expectedEnv:    "staging",
			expectedURL:    "https://api-staging.megaport.com/",
			expectedBucket: "staging",
		},
		{
			name:           "portal-qa returns real qa env, bucketed to development",
			token:          "qa-token-12345",
			hostname:       "portal-qa.megaport.com",
			expectedEnv:    "qa",
			expectedURL:    "https://api-qa.megaport.com/",
			expectedBucket: "development",
		},
		{
			name:           "portal-uat returns real uat env, bucketed to development",
			token:          "uat-token-12345",
			hostname:       "portal-uat.megaport.com",
			expectedEnv:    "uat",
			expectedURL:    "https://api-uat.megaport.com/",
			expectedBucket: "development",
		},
		{
			name:           "another app hostname (dashboard) derives env via app-env convention",
			token:          "dashboard-token-12345",
			hostname:       "dashboard-qa.megaport.com",
			expectedEnv:    "qa",
			expectedURL:    "https://api-qa.megaport.com/",
			expectedBucket: "development",
		},
		{
			name:           "another app at the apex (dashboard.megaport.com) is production",
			token:          "dashboard-prod-token-12345",
			hostname:       "dashboard.megaport.com",
			expectedEnv:    "production",
			expectedURL:    "https://api.megaport.com/",
			expectedBucket: "production",
		},
		{
			name:           "multi-segment env (api-mpone-dev) preserved",
			token:          "mpone-dev-token-12345",
			hostname:       "api-mpone-dev.megaport.com",
			expectedEnv:    "mpone-dev",
			expectedURL:    "https://api-mpone-dev.megaport.com/",
			expectedBucket: "development",
		},

		// --- hostname-derived failure cases (no override) ---
		{
			name:        "localhost without override now errors (no silent staging fallback)",
			token:       "dev-token-12345",
			hostname:    "localhost",
			expectError: true,
		},
		{
			name:        "non-megaport hostname without override errors",
			token:       "random-token-12345",
			hostname:    "example.com",
			expectError: true,
		},
		{
			name:        "look-alike hostname (megaport.com.attacker.com) fails closed",
			token:       "lookalike-token-12345",
			hostname:    "megaport.com.attacker.com",
			expectError: true,
		},

		// --- explicit override cases ---
		{
			name:           "explicit qa override from production hostname (mismatch)",
			token:          "explicit-env-token-12345",
			hostname:       "portal.megaport.com",
			explicitEnv:    "qa",
			expectedEnv:    "qa",
			expectedURL:    "https://api-qa.megaport.com/",
			expectedBucket: "development",
		},
		{
			name:           "explicit production override from staging hostname",
			token:          "explicit-prod-token-12345",
			hostname:       "portal-staging.megaport.com",
			explicitEnv:    "production",
			expectedEnv:    "production",
			expectedURL:    "https://api.megaport.com/",
			expectedBucket: "production",
		},
		{
			name:           "explicit production override unblocks localhost",
			token:          "local-override-token-12345",
			hostname:       "localhost",
			explicitEnv:    "production",
			expectedEnv:    "production",
			expectedURL:    "https://api.megaport.com/",
			expectedBucket: "production",
		},
		{
			name:           "explicit qa override unblocks localhost",
			token:          "local-qa-token-12345",
			hostname:       "localhost",
			explicitEnv:    "qa",
			expectedEnv:    "qa",
			expectedURL:    "https://api-qa.megaport.com/",
			expectedBucket: "development",
		},
		{
			name:           "explicit override unblocks unrecognised host (non-megaport)",
			token:          "unknown-override-token-12345",
			hostname:       "example.com",
			explicitEnv:    "qa",
			expectedEnv:    "qa",
			expectedURL:    "https://api-qa.megaport.com/",
			expectedBucket: "development",
		},
		{
			name:           "explicit override matching derived (no mismatch)",
			token:          "match-token-12345",
			hostname:       "portal-staging.megaport.com",
			explicitEnv:    "staging",
			expectedEnv:    "staging",
			expectedURL:    "https://api-staging.megaport.com/",
			expectedBucket: "staging",
		},
		{
			name:           "explicit development pin",
			token:          "dev-pin-token-12345",
			hostname:       "localhost",
			explicitEnv:    "development",
			expectedEnv:    "development",
			expectedURL:    "https://api-development.megaport.com/",
			expectedBucket: "development",
		},
		{
			// "prod" passes the env-name regex and gets used as-is for the API URL
			// (yielding api-prod.megaport.com/), but restrictEnvironmentName recognises
			// it as an alias for production so MEGAPORT_ENVIRONMENT lands on
			// "production". This mirrors normalizeEnvironment in config_shared.go.
			name:           "'prod' alias buckets to production (matches normalizeEnvironment)",
			token:          "prod-alias-token-12345",
			hostname:       "portal.megaport.com",
			explicitEnv:    "prod",
			expectedEnv:    "prod",
			expectedURL:    "https://api-prod.megaport.com/",
			expectedBucket: "production",
		},

		// --- normalisation of override input ---
		{
			name:           "whitespace-only override is treated as no override",
			token:          "ws-token-12345",
			hostname:       "portal-staging.megaport.com",
			explicitEnv:    "   ",
			expectedEnv:    "staging",
			expectedURL:    "https://api-staging.megaport.com/",
			expectedBucket: "staging",
		},
		{
			name:           "uppercase override is lowercased",
			token:          "case-token-12345",
			hostname:       "portal.megaport.com",
			explicitEnv:    "PRODUCTION",
			expectedEnv:    "production",
			expectedURL:    "https://api.megaport.com/",
			expectedBucket: "production",
		},

		// --- injection rejection ---
		{
			name:        "override with host-injection chars is rejected",
			token:       "inject-token-12345",
			hostname:    "portal.megaport.com",
			explicitEnv: "foo.attacker.com/",
			expectError: true,
		},
		{
			name:        "override with slash is rejected",
			token:       "slash-token-12345",
			hostname:    "portal.megaport.com",
			explicitEnv: "qa/evil",
			expectError: true,
		},
		{
			name:        "override with @ is rejected",
			token:       "at-token-12345",
			hostname:    "portal.megaport.com",
			explicitEnv: "qa@evil.com",
			expectError: true,
		},

		// --- empty-input errors ---
		{
			name:        "empty token errors",
			token:       "",
			hostname:    "portal.megaport.com",
			expectError: true,
		},
		{
			name:        "empty hostname errors",
			token:       "valid-token-12345",
			hostname:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any previous auth
			js.Global().Get("clearAuthCredentials").Invoke()

			// Call setAuthToken
			setAuthFunc := js.Global().Get("setAuthToken")
			assert.False(t, setAuthFunc.IsUndefined(), "setAuthToken should be registered")

			var result js.Value
			if tt.explicitEnv != "" {
				result = setAuthFunc.Invoke(tt.token, tt.hostname, tt.explicitEnv)
			} else {
				result = setAuthFunc.Invoke(tt.token, tt.hostname)
			}

			success := result.Get("success").Bool()

			if tt.expectError {
				assert.False(t, success, "should fail for invalid input")
				errMsg := result.Get("error").String()
				assert.NotEmpty(t, errMsg, "error message should be set on failure to guide the caller")
			} else {
				assert.True(t, success, "should succeed for valid input")

				// The JS return surface carries the real resolved env name (e.g. "qa"),
				// not the bucket — that's what the portal UI displays back to the user.
				returnedEnv := result.Get("environment").String()
				assert.Equal(t, tt.expectedEnv, returnedEnv, "returned environment should match expected")

				returnedURL := result.Get("apiURL").String()
				assert.Equal(t, tt.expectedURL, returnedURL, "returned API URL should match expected")

				// Verify auth info shows token is set
				authInfo := js.Global().Get("debugAuthInfo").Invoke()
				tokenSet := authInfo.Get("accessTokenSet").Bool()
				assert.True(t, tokenSet, "token should be marked as set")

				// MEGAPORT_ENVIRONMENT (surfaced via debugAuthInfo) holds the *bucket*
				// — one of production/staging/development — so downstream
				// normalizeEnvironment consumers don't silently coerce non-canonical
				// values to production.
				envVar := authInfo.Get("environment").String()
				assert.Equal(t, tt.expectedBucket, envVar, "MEGAPORT_ENVIRONMENT should be bucketed")

				authMethod := authInfo.Get("authMethod").String()
				assert.Equal(t, "token", authMethod, "authMethod should be 'token'")
			}
		})
	}
}

// TestSetAuthToken_NonStringThirdArg verifies that JS callers passing
// `undefined` or `null` (and other non-string values) as the third argument
// are treated as if no override was provided — rather than having
// Value.String() return "<undefined>"/"<null>" and triggering the regex
// rejection with a misleading error.
func TestSetAuthToken_NonStringThirdArg(t *testing.T) {
	RegisterJSFunctions()
	setAuthFunc := js.Global().Get("setAuthToken")

	tests := []struct {
		name     string
		thirdArg js.Value
	}{
		{"undefined third arg", js.Undefined()},
		{"null third arg", js.Null()},
		{"number third arg", js.ValueOf(0)},
		{"bool third arg", js.ValueOf(false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js.Global().Get("clearAuthCredentials").Invoke()

			// Recognised host so the call should succeed when the third arg is ignored.
			result := setAuthFunc.Invoke("tok", "portal-staging.megaport.com", tt.thirdArg)

			assert.True(t, result.Get("success").Bool(), "non-string third arg should be ignored, not trigger validation")
			assert.Equal(t, "staging", result.Get("environment").String(), "should fall through to hostname-derived env")
		})
	}

	t.Run("undefined third arg + unrecognised host fails closed (not regex error)", func(t *testing.T) {
		js.Global().Get("clearAuthCredentials").Invoke()
		result := setAuthFunc.Invoke("tok", "localhost", js.Undefined())
		assert.False(t, result.Get("success").Bool())
		errMsg := result.Get("error").String()
		assert.Contains(t, errMsg, "could not determine environment", "should hit fail-closed path, not regex validation")
		assert.NotContains(t, errMsg, "lowercase letters", "must not surface the regex rejection message")
	})
}

// TestSetAuthToken_TwoArgsSkipsValidation is a regression test for a bug
// where a flipped comparison in setAuthToken caused the env-name regex
// validation to fire on every no-override call, surfacing the misleading
// "environment must contain only lowercase letters..." error instead of the
// real fail-closed message. Two-argument calls must never reach validation.
func TestSetAuthToken_TwoArgsSkipsValidation(t *testing.T) {
	RegisterJSFunctions()
	js.Global().Get("clearAuthCredentials").Invoke()

	// Hostname that the derivation step rejects, so the call errors and we can
	// inspect which error path produced the message.
	result := js.Global().Get("setAuthToken").Invoke("tok", "localhost")
	assert.False(t, result.Get("success").Bool())
	errMsg := result.Get("error").String()
	assert.NotContains(t, errMsg, "lowercase letters", "two-arg call must not reach env-name validation")
	assert.Contains(t, errMsg, "could not determine environment", "two-arg call should hit the fail-closed path")
}

// TestAuthMethodPriority verifies that token auth takes precedence over API key auth
func TestAuthMethodPriority(t *testing.T) {
	RegisterJSFunctions()

	// Clear any existing auth state first
	js.Global().Get("clearAuthCredentials").Invoke()

	// First set API key auth
	js.Global().Get("setAuthCredentials").Invoke("api-key", "api-secret", "staging")
	authInfo := js.Global().Get("debugAuthInfo").Invoke()
	assert.Equal(t, "apikey", authInfo.Get("authMethod").String())

	// Now set token auth - should override
	js.Global().Get("setAuthToken").Invoke("test-token-12345", "portal.megaport.com")
	authInfo = js.Global().Get("debugAuthInfo").Invoke()
	assert.Equal(t, "token", authInfo.Get("authMethod").String())
	assert.Equal(t, "production", authInfo.Get("environment").String())

	// Clear and verify
	js.Global().Get("clearAuthCredentials").Invoke()
	authInfo = js.Global().Get("debugAuthInfo").Invoke()
	assert.Equal(t, "none", authInfo.Get("authMethod").String())
}

// TestSetAuthTokenMasking verifies token preview masking
func TestSetAuthTokenMasking(t *testing.T) {
	RegisterJSFunctions()

	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.test"
	js.Global().Get("setAuthToken").Invoke(testToken, "portal.megaport.com")

	authInfo := js.Global().Get("debugAuthInfo").Invoke()
	preview := authInfo.Get("accessTokenPreview").String()

	// Verify preview is masked
	assert.Contains(t, preview, "...")
	assert.NotEqual(t, testToken, preview, "full token should not be in preview")
	assert.True(t, len(preview) < len(testToken), "preview should be shorter than full token")
}

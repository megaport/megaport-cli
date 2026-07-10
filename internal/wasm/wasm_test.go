// These tests touch only buffers and JS globals that node's headless
// wasm_exec.js runtime already provides (no fetch, localStorage, or other
// browser-only bridges), so they run under the default `js,wasm` CI step.
// Tests that need a real browser host live in wasm_browser_test.go, gated
// behind the `browser` build tag.
//go:build js && wasm

package wasm

import (
	"bytes"
	"math"
	"strings"
	"syscall/js"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetCapturedOutput_Priority verifies output priority order
func TestGetCapturedOutput_Priority(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func()
		expectedSource string
		expectedOutput string
	}{
		{
			name: "JSON has highest priority",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmJSONOutput", "json output")
				js.Global().Set("wasmCSVOutput", "csv output")
				js.Global().Set("wasmTableOutput", "table output")
				stdoutBuffer.WriteString("stdout output")
				_, _ = WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "JSON buffer",
			expectedOutput: "json output",
		},
		{
			name: "CSV has second priority",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmCSVOutput", "csv output")
				js.Global().Set("wasmTableOutput", "table output")
				stdoutBuffer.WriteString("stdout output")
				_, _ = WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "CSV buffer",
			expectedOutput: "csv output",
		},
		{
			name: "Table output is prefixed with direct status buffer",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmTableOutput", "table output")
				stdoutBuffer.WriteString("stdout output")
				_, _ = WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "table buffer",
			expectedOutput: "direct outputtable output",
		},
		{
			name: "Table without status returns table only",
			setupFn: func() {
				ResetOutputBuffers()
				js.Global().Set("wasmTableOutput", "table output")
			},
			expectedSource: "table buffer",
			expectedOutput: "table output",
		},
		{
			name: "Direct buffer has fourth priority",
			setupFn: func() {
				ResetOutputBuffers()
				stdoutBuffer.WriteString("stdout output")
				_, _ = WasmOutputBuffer.Write([]byte("direct output"))
			},
			expectedSource: "direct buffer",
			expectedOutput: "direct output",
		},
		{
			name: "Stdout/stderr is lowest priority",
			setupFn: func() {
				ResetOutputBuffers()
				stdoutBuffer.WriteString("stdout output")
				stderrBuffer.WriteString("stderr output")
			},
			expectedSource: "combined stdout/stderr",
			expectedOutput: "stdout outputstderr output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFn()
			output := GetCapturedOutput()
			assert.Equal(t, tt.expectedOutput, output)
		})
	}
}

// TestDirectOutputBuffer_Write verifies thread-safe writes
func TestDirectOutputBuffer_Write(t *testing.T) {
	buffer := &DirectOutputBuffer{
		buffer: &bytes.Buffer{},
	}

	testData := []byte("test data")
	n, err := buffer.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, "test data", buffer.String())
}

// TestDirectOutputBuffer_Concurrent verifies thread safety
func TestDirectOutputBuffer_Concurrent(t *testing.T) {
	buffer := &DirectOutputBuffer{
		buffer: &bytes.Buffer{},
	}

	done := make(chan bool)
	iterations := 100

	// Multiple goroutines writing concurrently
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				_, _ = buffer.Write([]byte("x"))
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have exactly 1000 'x' characters (10 goroutines * 100 iterations)
	result := buffer.String()
	assert.Equal(t, 1000, len(result))
	assert.Equal(t, strings.Repeat("x", 1000), result)
}

// TestDirectOutputBuffer_Reset verifies buffer reset
func TestDirectOutputBuffer_Reset(t *testing.T) {
	buffer := &DirectOutputBuffer{
		buffer: &bytes.Buffer{},
	}

	_, _ = buffer.Write([]byte("test data"))
	assert.NotEqual(t, "", buffer.String())

	buffer.Reset()
	assert.Equal(t, "", buffer.String())
}

// TestCustomWriter verifies custom writer implementation
func TestCustomWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := &customWriter{writer: &buf}

	testData := []byte("test data")
	n, err := writer.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, "test data", buf.String())
}

// TestCaptureOutput verifies output capture functionality
func TestCaptureOutput(t *testing.T) {
	testOutput := "test output from function"

	captured := CaptureOutput(func() {
		_, _ = WasmOutputBuffer.Write([]byte(testOutput))
	})

	// Should contain the output (may have additional content from pipes)
	assert.Contains(t, captured, testOutput)
}

// TestSplitArgs verifies command argument parsing
func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple command",
			input:    "locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "command with flags",
			input:    "ports list --format json",
			expected: []string{"ports", "list", "--format", "json"},
		},
		{
			name:     "command with double quotes",
			input:    `ports create --name "My Port"`,
			expected: []string{"ports", "create", "--name", "My Port"},
		},
		{
			name:     "command with single quotes",
			input:    `ports create --name 'My Port'`,
			expected: []string{"ports", "create", "--name", "My Port"},
		},
		{
			name:     "removes program name",
			input:    "megaport-cli locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "removes program name with path",
			input:    "./megaport-cli locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "removes leading program name literal megaport",
			input:    "megaport locations list",
			expected: []string{"locations", "list"},
		},
		{
			name:     "preserves megaport as a flag value",
			input:    "ports update abc123 --name megaport",
			expected: []string{"ports", "update", "abc123", "--name", "megaport"},
		},
		{
			name:     "preserves megaport as a positional argument",
			input:    "tag megaport ports",
			expected: []string{"tag", "megaport", "ports"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "complex with mixed quotes",
			input:    `ports create --name "Port 1" --description 'Test port' --speed 1000`,
			expected: []string{"ports", "create", "--name", "Port 1", "--description", "Test port", "--speed", "1000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitArgs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSetTerminalWidth verifies the exported width setter/getter round-trip,
// that negative values reset to "unset" (0), and that absurdly large values
// are clamped rather than stored as-is.
func TestSetTerminalWidth(t *testing.T) {
	defer SetTerminalWidth(0)

	assert.Equal(t, 0, TerminalWidth(), "width should be unset by default")

	SetTerminalWidth(120)
	assert.Equal(t, 120, TerminalWidth())

	SetTerminalWidth(-5)
	assert.Equal(t, 0, TerminalWidth(), "negative width should reset to unset")

	SetTerminalWidth(1_000_000_000)
	assert.Equal(t, maxTerminalWidthCols, TerminalWidth(), "absurd width should clamp to the max")
}

// TestSetTerminalWidth_JSFunction verifies the JS-exposed setTerminalWidth
// function for present, absent (invalid), and absurd (tiny/huge) widths.
func TestSetTerminalWidth_JSFunction(t *testing.T) {
	RegisterJSFunctions()
	defer SetTerminalWidth(0)

	setFunc := js.Global().Get("setTerminalWidth")

	tests := []struct {
		name    string
		args    []interface{}
		wantOK  bool
		wantCol int
	}{
		{name: "normal width", args: []interface{}{80}, wantOK: true, wantCol: 80},
		{name: "tiny width", args: []interface{}{1}, wantOK: true, wantCol: 1},
		{name: "huge width clamps to max", args: []interface{}{100000}, wantOK: true, wantCol: maxTerminalWidthCols},
		{name: "out-of-int-range width clamps to max", args: []interface{}{1e300}, wantOK: true, wantCol: maxTerminalWidthCols},
		{name: "large negative width resets to unset", args: []interface{}{-1e300}, wantOK: true, wantCol: 0},
		{name: "zero width", args: []interface{}{0}, wantOK: true, wantCol: 0},
		{name: "missing argument", args: []interface{}{}, wantOK: false},
		{name: "non-numeric argument", args: []interface{}{"wide"}, wantOK: false},
		{name: "NaN argument", args: []interface{}{math.NaN()}, wantOK: false},
		{name: "positive infinity argument", args: []interface{}{math.Inf(1)}, wantOK: false},
		{name: "negative infinity argument", args: []interface{}{math.Inf(-1)}, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTerminalWidth(0)
			result := setFunc.Invoke(tt.args...)

			assert.Equal(t, tt.wantOK, result.Get("success").Bool())
			if tt.wantOK {
				assert.Equal(t, tt.wantCol, TerminalWidth())
			} else {
				assert.Equal(t, 0, TerminalWidth())
				assert.False(t, result.Get("error").IsUndefined())
			}
		})
	}
}

// TestSplitArgs_EdgeCases verifies edge cases in argument parsing
func TestSplitArgs_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "unclosed double quote",
			input:    `command --flag "unclosed`,
			expected: []string{"command", "--flag", "unclosed"},
		},
		{
			name:     "unclosed single quote",
			input:    `command --flag 'unclosed`,
			expected: []string{"command", "--flag", "unclosed"},
		},
		{
			name:     "empty quotes",
			input:    `command --flag "" --other ''`,
			expected: []string{"command", "--flag", "", "--other", ""},
		},
		{
			name:     "multiple spaces",
			input:    "command    with    spaces",
			expected: []string{"command", "with", "spaces"},
		},
		{
			name:     "quotes with spaces inside",
			input:    `command "arg with   spaces"`,
			expected: []string{"command", "arg with   spaces"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitArgs(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEnvironmentToAPIURL verifies the URL-builder helper. The helper assumes
// its input has already been validated upstream (see envNamePattern), so it
// is exercised only with values the upstream gates accept. Injection-vector
// rejection is covered at the boundary in TestSetAuthToken.
func TestEnvironmentToAPIURL(t *testing.T) {
	tests := []struct {
		env         string
		expectedURL string
	}{
		{"production", "https://api.megaport.com/"},
		{"staging", "https://api-staging.megaport.com/"},
		{"development", "https://api-development.megaport.com/"},
		{"qa", "https://api-qa.megaport.com/"},
		{"uat", "https://api-uat.megaport.com/"},
		{"mpone-dev", "https://api-mpone-dev.megaport.com/"},
		{"prod", "https://api-prod.megaport.com/"}, // typo "prod" still passes envNamePattern; reviewer wanted coverage
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			assert.Equal(t, tt.expectedURL, environmentToAPIURL(tt.env))
		})
	}
}

// TestEnvironmentFromHostname verifies the hostname extractor against the two
// conventions Megaport uses: <app>.megaport.com (production) and
// <app>-<env>.megaport.com.
func TestEnvironmentFromHostname(t *testing.T) {
	tests := []struct {
		hostname    string
		expectedEnv string
		expectedOK  bool
	}{
		// Production: apex, www, and any <app>.megaport.com.
		{"megaport.com", "production", true},
		{"www.megaport.com", "production", true},
		{"portal.megaport.com", "production", true},
		{"api.megaport.com", "production", true},
		{"dashboard.megaport.com", "production", true},
		{"tools.megaport.com", "production", true},

		// <app>-<env>.megaport.com — any app, env after the first hyphen.
		{"portal-staging.megaport.com", "staging", true},
		{"portal-qa.megaport.com", "qa", true},
		{"portal-uat.megaport.com", "uat", true},
		{"api-staging.megaport.com", "staging", true},
		{"api-qa.megaport.com", "qa", true},
		{"dashboard-staging.megaport.com", "staging", true},
		{"tools-uat.megaport.com", "uat", true},

		// Multi-segment env names (split on the FIRST hyphen).
		{"api-mpone-dev.megaport.com", "mpone-dev", true},
		{"portal-mpone-dev.megaport.com", "mpone-dev", true},

		// Case- and whitespace-insensitive.
		{"PORTAL-QA.MEGAPORT.COM", "qa", true},
		{"  portal-staging.megaport.com  ", "staging", true},

		// Fail closed for localhost and private IPs.
		{"localhost", "", false},
		{"127.0.0.1", "", false},
		{"192.168.1.100", "", false},
		{"10.0.0.1", "", false},

		// Fail closed for non-.megaport.com hostnames.
		{"example.com", "", false},
		{"attacker.com", "", false},
		// Sanity check the suffix guard: looks similar but isn't .megaport.com.
		{"megaport.com.attacker.com", "", false},
		{"xmegaport.com", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			env, ok := environmentFromHostname(tt.hostname)
			assert.Equal(t, tt.expectedOK, ok)
			assert.Equal(t, tt.expectedEnv, env)
		})
	}
}

// TestRestrictEnvironmentName verifies the shim that collapses any
// environment name into the three canonical values that downstream
// MEGAPORT_ENVIRONMENT consumers understand.
func TestRestrictEnvironmentName(t *testing.T) {
	tests := []struct {
		env      string
		expected string
	}{
		{"production", "production"},
		{"staging", "staging"},
		{"development", "development"},
		{"qa", "development"},
		{"uat", "development"},
		{"mpone-dev", "development"},
		{"prod", "production"}, // alias accepted, matches normalizeEnvironment
		{"", "development"},    // shouldn't be called with empty, but defined behaviour
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			assert.Equal(t, tt.expected, restrictEnvironmentName(tt.env))
		})
	}
}

// TestBufferThreadSafety verifies all buffers are thread-safe
func TestBufferThreadSafety(t *testing.T) {
	ResetOutputBuffers()

	done := make(chan bool)
	iterations := 100

	// Test stdout buffer
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				bufferMutex.Lock()
				stdoutBuffer.WriteString("x")
				bufferMutex.Unlock()
			}
			done <- true
		}()
	}

	// Wait for completion
	for i := 0; i < 5; i++ {
		<-done
	}

	// Should have exactly 500 'x' characters
	result := stdoutBuffer.String()
	assert.Equal(t, 500, len(result))
}

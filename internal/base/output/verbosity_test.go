//go:build !wasm
// +build !wasm

package output

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetVerbosity(t *testing.T) {
	t.Cleanup(func() {
		SetVerbosity("normal")
	})
}

// captureStdout captures stdout output from a function call.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// captureStderr captures stderr output from a function call.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestIsQuiet(t *testing.T) {
	resetVerbosity(t)

	assert.False(t, IsQuiet(), "should not be quiet by default")

	SetVerbosity("quiet")
	assert.True(t, IsQuiet(), "should be quiet after SetVerbosity(quiet)")

	SetVerbosity("normal")
	assert.False(t, IsQuiet(), "should not be quiet after SetVerbosity(normal)")
}

func TestIsVerbose(t *testing.T) {
	resetVerbosity(t)

	assert.False(t, IsVerbose(), "should not be verbose by default")

	SetVerbosity("verbose")
	assert.True(t, IsVerbose(), "should be verbose after SetVerbosity(verbose)")

	SetVerbosity("normal")
	assert.False(t, IsVerbose(), "should not be verbose after SetVerbosity(normal)")
}

func TestQuietSuppressesPrintInfo(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	out := captureStdout(t, func() {
		PrintInfo("test message %s", true, "arg")
	})
	assert.Empty(t, out, "PrintInfo should produce no output in quiet mode")
}

func TestQuietSuppressesPrintSuccess(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	out := captureStdout(t, func() {
		PrintSuccess("test message", true)
	})
	assert.Empty(t, out, "PrintSuccess should produce no output in quiet mode")
}

func TestQuietSuppressesPrintWarning(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	out := captureStdout(t, func() {
		PrintWarning("test message", true)
	})
	assert.Empty(t, out, "PrintWarning should produce no output in quiet mode")
}

func TestQuietDoesNotSuppressPrintError(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	out := captureStdout(t, func() {
		PrintError("error message", true)
	})
	assert.Contains(t, out, "error message", "PrintError should still produce output in quiet mode")
}

func TestQuietReturnsNoOpSpinner(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	spinner := PrintResourceListing("Port", true)
	assert.NotNil(t, spinner, "should return a non-nil spinner")
	assert.True(t, spinner.stopped, "spinner should be already stopped in quiet mode")

	// Verify Stop() and StopWithSuccess() don't panic
	spinner.Stop()
	spinner.StopWithSuccess("done")
}

func TestQuietSuppressesAllSpinnerCreators(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	tests := []struct {
		name    string
		spinner *Spinner
	}{
		{"PrintResourceCreating", PrintResourceCreating("Port", "uid-123", true)},
		{"PrintResourceUpdating", PrintResourceUpdating("Port", "uid-123", true)},
		{"PrintResourceDeleting", PrintResourceDeleting("Port", "uid-123", true)},
		{"PrintResourceListing", PrintResourceListing("Port", true)},
		{"PrintResourceGetting", PrintResourceGetting("Port", "uid-123", true)},
		{"PrintResourceGettingWithOutput", PrintResourceGettingWithOutput("Port", "uid-123", true, "table")},
		{"PrintListingResourceTags", PrintListingResourceTags("Port", "uid-123", true)},
		{"PrintResourceValidating", PrintResourceValidating("Port", true)},
		{"PrintLoggingIn", PrintLoggingIn(true)},
		{"PrintLoggingInWithOutput", PrintLoggingInWithOutput(true, "table")},
		{"PrintCustomSpinner", PrintCustomSpinner("Restoring", "uid-123", true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.spinner)
			assert.True(t, tt.spinner.stopped, "%s should return stopped spinner in quiet mode", tt.name)
		})
	}
}

func TestQuietSuppressesStopWithSuccess(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	spinner := newNoOpSpinner(true)
	out := captureStdout(t, func() {
		spinner.StopWithSuccess("done message")
	})
	assert.Empty(t, out, "StopWithSuccess should produce no output in quiet mode")
}

func TestVerboseEnablesPrintVerbose(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("verbose")

	out := captureStdout(t, func() {
		PrintVerbose("debug info %s", true, "details")
	})
	assert.Contains(t, out, "[DEBUG]", "PrintVerbose should print in verbose mode")
	assert.Contains(t, out, "debug info details", "PrintVerbose should contain the message")
}

func TestNormalSuppressesPrintVerbose(t *testing.T) {
	resetVerbosity(t)

	out := captureStdout(t, func() {
		PrintVerbose("debug info", true)
	})
	assert.Empty(t, out, "PrintVerbose should produce no output in normal mode")
}

func TestQuietSuppressesPrintVerbose(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("quiet")

	out := captureStdout(t, func() {
		PrintVerbose("debug info", true)
	})
	assert.Empty(t, out, "PrintVerbose should produce no output in quiet mode")
}

func TestNormalModeShowsAllMessages(t *testing.T) {
	resetVerbosity(t)

	infoOut := captureStdout(t, func() {
		PrintInfo("info message", true)
	})
	assert.Contains(t, infoOut, "info message", "PrintInfo should produce output in normal mode")

	successOut := captureStdout(t, func() {
		PrintSuccess("success message", true)
	})
	assert.Contains(t, successOut, "success message", "PrintSuccess should produce output in normal mode")

	warningOut := captureStdout(t, func() {
		PrintWarning("warning message", true)
	})
	assert.Contains(t, warningOut, "warning message", "PrintWarning should produce output in normal mode")

	errorOut := captureStdout(t, func() {
		PrintError("error message", true)
	})
	assert.Contains(t, errorOut, "error message", "PrintError should produce output in normal mode")
}

func TestVerbosePrintVerboseWithJSONFormat(t *testing.T) {
	resetVerbosity(t)
	SetVerbosity("verbose")

	oldFormat := getOutputFormat()
	SetOutputFormat("json")
	defer SetOutputFormat(oldFormat)

	out := captureStderr(t, func() {
		PrintVerbose("debug info", true)
	})
	assert.Contains(t, out, "[DEBUG]", "PrintVerbose should write to stderr in JSON format")
	assert.Contains(t, out, "debug info", "PrintVerbose should contain the message")
}

//go:build js && wasm

package megaport

import (
	"net/http"
	"strings"
	"syscall/js"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/wasm"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecuteWithArgs_UnknownCommand_RoutesErrorNotOutput is the ESD-1666
// regression test: a Cobra parser-layer failure (the ticket's `location list`
// repro) must be returned as an error so the host can route it to result.error,
// not buried in the captured output. Only the usage hint stays in output.
func TestExecuteWithArgs_UnknownCommand_RoutesErrorNotOutput(t *testing.T) {
	wasm.ResetOutputBuffers()

	err := ExecuteWithArgs([]string{"megaport-cli", "location", "list"})

	require.Error(t, err, "an unknown command must return an error")
	assert.Contains(t, err.Error(), "unknown command", "the returned error should name the failure")

	out := wasm.GetCapturedOutput()
	assert.NotContains(t, out, "unknown command", "the error text must not stay in output")
	assert.Contains(t, out, "Run 'megaport-cli --help'", "the usage hint may stay in output")
}

// TestExecuteWithArgs_MissingRequiredArg_RoutesError covers the other parser-layer
// failure named in the ticket: a command that requires an argument (cobra.ExactArgs)
// invoked without one. Cobra's arg validation fails before RunE and is not a
// *CLIError, so it must route to the returned error, not sit in output.
func TestExecuteWithArgs_MissingRequiredArg_RoutesError(t *testing.T) {
	wasm.ResetOutputBuffers()

	// `vxc get` is built with WithArgs(cobra.ExactArgs(1)); omit the UID.
	err := ExecuteWithArgs([]string{"megaport-cli", "vxc", "get"})

	require.Error(t, err, "a missing required arg must return an error")
	assert.Contains(t, err.Error(), "arg(s)", "the returned error should name the arg-count failure")

	out := wasm.GetCapturedOutput()
	assert.NotContains(t, out, "arg(s)", "the error text must not stay in output")
}

// TestExecuteWithArgs_Success_ReturnsNil guards the other half of the contract:
// a command that succeeds returns no error, so the host never sets result.error
// on success. --help returns before any network call is made.
func TestExecuteWithArgs_Success_ReturnsNil(t *testing.T) {
	wasm.ResetOutputBuffers()

	err := ExecuteWithArgs([]string{"megaport-cli", "--help"})

	assert.NoError(t, err, "a successful command must not return an error")
	assert.NotEmpty(t, wasm.GetCapturedOutput(), "help output should still be captured")
}

// TestExecuteWithArgs_UnknownCommand_WithHandler_RoutesError is the Portal
// scenario: a live-output handler is registered, and the ticket's `location list`
// repro fails at the parser layer. Because cobra's error print is silenced, the
// error never streams as a normal chunk, so ExecuteWithArgs returns it for
// result.error (red Error: line + failure telemetry) while only the usage hint
// streams to the terminal.
func TestExecuteWithArgs_UnknownCommand_WithHandler_RoutesError(t *testing.T) {
	wasm.ResetOutputBuffers()

	var streamed []string
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			streamed = append(streamed, args[0].String())
		}
		return nil
	})
	defer fn.Release()
	wasm.RegisterOutputCallback(fn.Value)
	defer wasm.UnregisterOutputCallback()

	err := ExecuteWithArgs([]string{"megaport-cli", "location", "list"})

	require.Error(t, err, "an unknown command must return an error even with a handler")
	assert.Contains(t, err.Error(), "unknown command", "the returned error should name the failure")
	assert.NotContains(t, strings.Join(streamed, ""), "unknown command",
		"the error text must reach result.error, not stream as an uncolored chunk")
}

// TestExecuteWithArgs_StreamedError_ReturnsNil locks in the no-double-render
// rule: when an action streams its own error to a live-output handler before
// returning, ExecuteWithArgs returns nil so the host does not render a second
// red Error: line on top of the streamed text.
func TestExecuteWithArgs_StreamedError_ReturnsNil(t *testing.T) {
	t.Setenv("MEGAPORT_ACCESS_TOKEN", "some-token")

	wasm.ResetOutputBuffers()

	var streamed []string
	fn := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			streamed = append(streamed, args[0].String())
		}
		return nil
	})
	defer fn.Release()
	wasm.RegisterOutputCallback(fn.Value)
	defer wasm.UnregisterOutputCallback()

	cmd := &cobra.Command{
		Use: "streamederrortest",
		RunE: utils.WrapRunE(func(cmd *cobra.Command, args []string) error {
			apiErr := &megaport.ErrorResponse{
				Response: &http.Response{StatusCode: 500, Header: http.Header{}, Request: &http.Request{}},
				Message:  "boom",
			}
			output.PrintError("Failed: %v", false, apiErr)
			return apiErr
		}),
	}
	rootCmd.AddCommand(cmd)
	defer rootCmd.RemoveCommand(cmd)

	err := ExecuteWithArgs([]string{"megaport-cli", "streamederrortest"})

	require.True(t, wasm.DidStreamOutput(), "the action should have streamed its error")
	assert.NoError(t, err, "a streamed error must not also be returned for result.error")
	assert.NotEmpty(t, strings.Join(streamed, ""), "the streamed error should have reached the handler")
}

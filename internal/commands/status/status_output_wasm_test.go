//go:build js && wasm

package status

import (
	"testing"

	op "github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/wasm"
	"github.com/stretchr/testify/assert"
)

// TestPrintDashboard_TableWASMCapture guards the ESD-1283 fix: the multi-section
// table dashboard must render in full through the WASM capture path. Routing each
// section through PrintOutput("table") previously left only the last section in
// the per-call wasmTableOutput global, so the browser saw a truncated dashboard.
// PrintTableToWriter renders every section into WasmOutputBuffer instead, so
// GetCapturedOutput returns headers, all five tables, and the summary in order.
func TestPrintDashboard_TableWASMCapture(t *testing.T) {
	wasm.ResetOutputBuffers()
	op.SetOutputFormat("table")

	dashboard := statusTestDashboard(t)
	err := printDashboard(wasm.WasmOutputBuffer, dashboard, "table", true)
	assert.NoError(t, err)

	out := wasm.GetCapturedOutput()

	// Every section header and its row must survive, not just the last one.
	for _, want := range []string{
		"PORTS (1)", "port-1",
		"MCRS (1)", "mcr-1",
		"MVES (1)", "mve-1",
		"VXCS (1)", "vxc-1",
		"IXS (1)", "ix-1",
		"Total: 1 port(s), 1 MCR(s), 1 MVE(s), 1 VXC(s), 1 IX(s)",
	} {
		assert.Contains(t, out, want, "WASM dashboard capture should include %q", want)
	}
}

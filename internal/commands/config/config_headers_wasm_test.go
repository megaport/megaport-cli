//go:build js && wasm

package config

import "testing"

func TestCLIHeadersWasm(t *testing.T) {
	if got := cliHeaders["x-app"]; got != "cli-wasm" {
		t.Errorf("cliHeaders[\"x-app\"] = %q, want %q", got, "cli-wasm")
	}
}

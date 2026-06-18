//go:build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestE2E_Smoke proves the harness end to end: build, exec, capture, exit code.
func TestE2E_Smoke(t *testing.T) {
	res := Run(t, "version")

	assert.Equal(t, 0, res.Exit, "version should exit 0; stderr: %s", res.Stderr)
	assert.NotEmpty(t, res.Stdout, "version should print to stdout")
}

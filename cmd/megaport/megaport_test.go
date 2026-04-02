//go:build !js && !wasm
// +build !js,!wasm

package megaport

import (
	"errors"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/stretchr/testify/assert"
)

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		// CLIError with each code
		{"CLIError general", exitcodes.New(exitcodes.General, errors.New("unknown")), exitcodes.General},
		{"CLIError usage", exitcodes.NewUsageError(errors.New("bad flag")), exitcodes.Usage},
		{"CLIError auth", exitcodes.NewAuthError(errors.New("no creds")), exitcodes.Authentication},
		{"CLIError api", exitcodes.NewAPIError(errors.New("500")), exitcodes.API},
		{"CLIError cancelled", exitcodes.NewCancelledError(errors.New("cancelled by user")), exitcodes.Cancelled},

		// Cobra-style errors
		{"cobra unknown command", errors.New(`unknown command "foo" for "megaport-cli"`), exitcodes.Usage},
		{"cobra unknown flag", errors.New(`unknown flag: --bogus`), exitcodes.Usage},
		{"cobra unknown shorthand flag", errors.New(`unknown shorthand flag: 'x' in -x`), exitcodes.Usage},
		{"cobra accepts at most", errors.New(`accepts at most 1 arg(s), received 2`), exitcodes.Usage},
		{"cobra required flag(s)", errors.New(`required flag(s) "name" not set`), exitcodes.Usage},

		// PersistentPreRunE format validation
		{"invalid output format", errors.New("invalid output format: yaml"), exitcodes.Usage},

		// Unknown errors
		{"unknown error", errors.New("something unexpected"), exitcodes.General},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, exitCodeFromError(tt.err))
		})
	}
}

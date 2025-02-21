package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureOutput captures and returns any output written to stdout during execution of f.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

// createTempConfigPath creates a temporary config file, overrides the configFile path for testing,
// and returns a cleanup function to restore the original configuration.
func createTempConfigPath(t *testing.T) func() {
	tmpFile, err := os.CreateTemp("", "megaport-config-*.json")
	assert.NoError(t, err)

	// Save the original configFile path.
	originalConfigFile := configFile
	// Override the configFile with the temporary file's name.
	configFile = tmpFile.Name()

	return func() {
		os.Remove(tmpFile.Name())
		configFile = originalConfigFile
	}
}

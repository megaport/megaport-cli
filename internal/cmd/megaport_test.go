package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigureCmd(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("MEGAPORT_ACCESS_KEY", "test-access-key")
	t.Setenv("MEGAPORT_SECRET_KEY", "test-secret-key")
	t.Setenv("MEGAPORT_ENVIRONMENT", "staging")

	defer func() {
		os.Unsetenv("MEGAPORT_ACCESS_KEY")
		os.Unsetenv("MEGAPORT_SECRET_KEY")
		os.Unsetenv("MEGAPORT_ENVIRONMENT")
	}()

	cmd := configureCmd
	err := cmd.RunE(cmd, []string{})
	assert.NoError(t, err)
}

func TestConfigureCmdMissingEnvVars(t *testing.T) {
	// Unset environment variables for testing
	os.Unsetenv("MEGAPORT_ACCESS_KEY")
	os.Unsetenv("MEGAPORT_SECRET_KEY")
	os.Unsetenv("MEGAPORT_ENVIRONMENT")

	cmd := configureCmd
	err := cmd.RunE(cmd, []string{})
	assert.Error(t, err)
}

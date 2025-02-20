package cmd

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSaveAndLoadConfig(t *testing.T) {
	// Use a temp file instead of the real config file
	tmpFile, err := os.CreateTemp("", "megaport-cli-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	testConfig := Config{
		AccessKey: "testAccessKey",
		SecretKey: "testSecretKey",
	}

	// Temporarily override configFile path
	originalConfigFile := configFile
	configFile = tmpFile.Name()
	defer func() { configFile = originalConfigFile }()

	// Test saving
	err = saveConfig(testConfig)
	assert.NoError(t, err)

	// Test loading
	loaded, err := loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, testConfig.AccessKey, loaded.AccessKey)
	assert.Equal(t, testConfig.SecretKey, loaded.SecretKey)
}

func TestConfigureCmd(t *testing.T) {
	cmd := configureCmd
	// Create a temp file for simulating config
	tmpFile, err := os.CreateTemp("", "megaport-cli-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	originalConfigFile := configFile
	configFile = tmpFile.Name()
	defer func() { configFile = originalConfigFile }()

	// Capture output
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set flags
	cmd.Flags().Set("access-key", "myAccessKey")
	cmd.Flags().Set("secret-key", "mySecretKey")

	// Run command
	cmd.Run(cmd, []string{})

	// Restore stdout
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldOut

	// Check output
	output := string(out)
	assert.Contains(t, output, "Configuration saved")

	// Verify config was actually saved
	data, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "myAccessKey", cfg.AccessKey)
	assert.Equal(t, "mySecretKey", cfg.SecretKey)
}

// Helper function to capture stdout
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

func TestConfigureCmd_NoFlags(t *testing.T) {
	// Create temp config file
	tmpFile, err := os.CreateTemp("", "megaport-cli-config-*.json")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Save original config path and restore after test
	origConfig := configFile
	configFile = tmpFile.Name()
	defer func() { configFile = origConfig }()

	// Create clean command with empty flags
	cmd := &cobra.Command{
		Use: "configure",
		Run: configureCmd.Run,
	}
	cmd.Flags().String("access-key", "", "")
	cmd.Flags().String("secret-key", "", "")

	// Capture output and run command
	output := captureOutput(func() {
		cmd.Run(cmd, []string{})
	})

	// Verify error message
	assert.Contains(t, output, "Error saving configuration: both access key and secret key are required")
}

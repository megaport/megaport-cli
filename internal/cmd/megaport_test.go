package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// createTestRootCmd returns a new root command with the configure subcommand.
func createTestRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "megaport",
		Short: "A CLI tool to interact with the Megaport API (test mode).",
	}
	// Note: configureCmd is global so its flag values may persist.
	root.AddCommand(configureCmd)
	return root
}

func TestConfigureCmd_ValidFlags(t *testing.T) {
	cleanup := createTempConfigPath(t)
	defer cleanup()

	// Ensure env vars do not interfere.
	os.Unsetenv("MEGAPORT_ACCESS_KEY")
	os.Unsetenv("MEGAPORT_SECRET_KEY")

	rootCmd := createTestRootCmd()
	rootCmd.SetArgs([]string{
		"configure",
		"--access-key", "flag-access",
		"--secret-key", "flag-secret",
	})

	output := captureOutput(func() {
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})

	assert.Contains(t, output, "Credentials from flags saved successfully.")

	cfg, err := loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "flag-access", cfg.AccessKey)
	assert.Equal(t, "flag-secret", cfg.SecretKey)
}

func TestConfigureCmd_MissingFlags(t *testing.T) {
	// Setup temp directory for config file.
	tmpDir, err := os.MkdirTemp("", "megaport-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	originalConfigFile := configFile
	configFile = filepath.Join(tmpDir, "config.json")
	defer func() { configFile = originalConfigFile }()

	// Clear environment variables.
	t.Log("Clearing environment variables")
	os.Unsetenv("MEGAPORT_ACCESS_KEY")
	os.Unsetenv("MEGAPORT_SECRET_KEY")

	// Create and execute command without providing flags.
	t.Log("Executing configure command without flags")
	rootCmd := createTestRootCmd()
	rootCmd.SetArgs([]string{"configure"})
	// Explicitly clear flag values (in case they persisted from prior tests)
	rootCmd.Commands()[0].Flags().Set("access-key", "")
	rootCmd.Commands()[0].Flags().Set("secret-key", "")

	var cmdErr error
	output := captureOutput(func() {
		cmdErr = rootCmd.Execute()
	})
	t.Logf("Command output: %s", output)
	t.Logf("Command error: %v", cmdErr)

	// Assertions: Expect an error since no credentials provided.
	assert.Error(t, cmdErr, "Expected an error when no credentials provided")
	if cmdErr != nil {
		assert.Equal(t, "no valid credentials provided", cmdErr.Error())
	}

	// Verify output message.
	expectedMsg := "Please provide credentials either through environment variables MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY\nor through flags --access-key and --secret-key"
	assert.Contains(t, output, expectedMsg)

	// Verify that no config file was written.
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("Config file should not exist")
	}
}

func TestConfigureCmd_ValidEnvVars(t *testing.T) {
	cleanup := createTempConfigPath(t)
	defer cleanup()

	// Set environment variables.
	os.Setenv("MEGAPORT_ACCESS_KEY", "env-access")
	os.Setenv("MEGAPORT_SECRET_KEY", "env-secret")
	defer func() {
		os.Unsetenv("MEGAPORT_ACCESS_KEY")
		os.Unsetenv("MEGAPORT_SECRET_KEY")
	}()

	rootCmd := createTestRootCmd()
	rootCmd.SetArgs([]string{"configure"})

	output := captureOutput(func() {
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})
	assert.Contains(t, output, "Credentials from environment saved successfully.")

	cfg, err := loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "env-access", cfg.AccessKey)
	assert.Equal(t, "env-secret", cfg.SecretKey)
}

func TestConfigureCmd_EnvVarsPrecedence(t *testing.T) {
	cleanup := createTempConfigPath(t)
	defer cleanup()

	// Set environment variables.
	os.Setenv("MEGAPORT_ACCESS_KEY", "env-access")
	os.Setenv("MEGAPORT_SECRET_KEY", "env-secret")
	defer func() {
		os.Unsetenv("MEGAPORT_ACCESS_KEY")
		os.Unsetenv("MEGAPORT_SECRET_KEY")
	}()

	rootCmd := createTestRootCmd()
	rootCmd.SetArgs([]string{
		"configure",
		"--access-key", "flag-access",
		"--secret-key", "flag-secret",
	})

	output := captureOutput(func() {
		err := rootCmd.Execute()
		assert.NoError(t, err)
	})
	// Env vars take precedence.
	assert.Contains(t, output, "Credentials from environment saved successfully.")

	cfg, err := loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "env-access", cfg.AccessKey)
	assert.Equal(t, "env-secret", cfg.SecretKey)
}

func TestWriteConfigFile(t *testing.T) {
	cleanup := createTempConfigPath(t)
	defer cleanup()

	cfg := Config{
		AccessKey: "writeTest",
		SecretKey: "writeSecret",
	}
	err := writeConfigFile(cfg)
	assert.NoError(t, err)

	data, err := os.ReadFile(configFile)
	assert.NoError(t, err)

	var loaded Config
	err = json.Unmarshal(data, &loaded)
	assert.NoError(t, err)
	assert.Equal(t, "writeTest", loaded.AccessKey)
	assert.Equal(t, "writeSecret", loaded.SecretKey)
}

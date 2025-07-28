//go:build !js && !wasm
// +build !js,!wasm

package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutputFromAction(action func() error) (string, error) {
	var outputText string
	var err error

	outputText = output.CaptureOutput(func() {
		err = action()
	})

	return outputText, err
}

func setupTestCmd() (*cobra.Command, *bytes.Buffer) {
	outBuf := new(bytes.Buffer)
	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(outBuf)
	cmd.SetErr(outBuf)
	return cmd, outBuf
}

func setupTestConfigEnv(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "megaport-config-actions-test")
	require.NoError(t, err)

	oldEnv := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

	return tempDir, func() {
		os.Setenv("MEGAPORT_CONFIG_DIR", oldEnv)
		os.RemoveAll(tempDir)
	}
}

func TestUpdateProfile_CMD(t *testing.T) {
	_, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Create a profile first
	manager, err := NewConfigManager()
	require.NoError(t, err)
	err = manager.CreateProfile("test-profile", "old-access", "old-secret", "production", "Old description")
	require.NoError(t, err)

	// Update only some fields - this time let the command run directly
	cmd, _ := setupTestCmd()
	cmd.Flags().String("access-key", "new-access", "")
	cmd.Flags().String("environment", "staging", "")

	// Mark flags as changed by parsing args
	err = cmd.ParseFlags([]string{"--access-key=new-access", "--environment=staging"})
	require.NoError(t, err)

	outputText, err := captureOutputFromAction(func() error {
		return UpdateProfile(cmd, []string{"test-profile"}, false)
	})
	require.NoError(t, err)
	assert.Contains(t, outputText, "Profile 'test-profile' updated successfully")

	// Very important: Create a new manager instance to read from disk
	manager, err = NewConfigManager()
	require.NoError(t, err)

	// Verify the profile was updated
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)
	assert.Equal(t, "new-access", profiles["test-profile"].AccessKey)
	assert.Equal(t, "old-secret", profiles["test-profile"].SecretKey) // Should be unchanged
	assert.Equal(t, "staging", profiles["test-profile"].Environment)
	assert.Equal(t, "Old description", profiles["test-profile"].Description) // Should be unchanged

	// Test description update
	cmd, _ = setupTestCmd()
	cmd.Flags().String("description", "New description", "")

	// Mark flag as changed by parsing args
	err = cmd.ParseFlags([]string{"--description=New description"})
	require.NoError(t, err)

	outputText, err = captureOutputFromAction(func() error {
		return UpdateProfile(cmd, []string{"test-profile"}, false)
	})
	require.NoError(t, err)
	assert.Contains(t, outputText, "Profile 'test-profile' updated successfully")

	// Very important: Create a new manager instance to read from disk
	manager, err = NewConfigManager()
	require.NoError(t, err)

	// Verify the description was updated
	profiles, err = manager.ListProfiles()
	require.NoError(t, err)
	assert.Equal(t, "New description", profiles["test-profile"].Description)
}

func TestUseProfile_CMD(t *testing.T) {
	_, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Create some profiles first
	manager, err := NewConfigManager()
	require.NoError(t, err)
	err = manager.CreateProfile("profile1", "access1", "secret1", "production", "")
	require.NoError(t, err)
	err = manager.CreateProfile("profile2", "access2", "secret2", "staging", "")
	require.NoError(t, err)

	// Test switching to profile1
	cmd, _ := setupTestCmd()
	outputText, err := captureOutputFromAction(func() error {
		return UseProfile(cmd, []string{"profile1"}, false)
	})
	require.NoError(t, err)
	assert.Contains(t, outputText, "Switched to profile 'profile1'")

	// Create a new manager to read from disk
	manager, err = NewConfigManager()
	require.NoError(t, err)

	// Verify active profile
	_, name, err := manager.GetCurrentProfile()
	require.NoError(t, err)
	assert.Equal(t, "profile1", name)

	// Test switching to profile2
	cmd, _ = setupTestCmd()
	outputText, err = captureOutputFromAction(func() error {
		return UseProfile(cmd, []string{"profile2"}, false)
	})
	require.NoError(t, err)
	assert.Contains(t, outputText, "Switched to profile 'profile2'")

	// Create a new manager to read from disk
	manager, err = NewConfigManager()
	require.NoError(t, err)

	// Verify active profile changed
	_, name, err = manager.GetCurrentProfile()
	require.NoError(t, err)
	assert.Equal(t, "profile2", name)

	// Test switching to non-existent profile
	cmd, _ = setupTestCmd()
	outputText, err = captureOutputFromAction(func() error {
		return UseProfile(cmd, []string{"non-existent"}, false)
	})
	assert.Error(t, err)
	assert.Contains(t, outputText, "not found")
}

func TestViewConfig(t *testing.T) {
	_, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Create a profile and set defaults
	manager, err := NewConfigManager()
	require.NoError(t, err)
	err = manager.CreateProfile("test-profile", "test-access", "test-secret", "production", "Test description")
	require.NoError(t, err)
	err = manager.UseProfile("test-profile")
	require.NoError(t, err)
	err = manager.SetDefault("output", "json")
	require.NoError(t, err)

	// Test view config - this writes to cmd.OutOrStdout()
	cmd, cmdOut := setupTestCmd()
	err = ViewConfig(cmd, nil, false)
	require.NoError(t, err)

	output := cmdOut.String()
	assert.Contains(t, output, "Active Profile: test-profile")
	assert.Contains(t, output, "Environment: production")
	assert.Contains(t, output, "Description: Test description")
	assert.Contains(t, output, "output: json")

	// Test with no active profile by creating a new config
	manager, err = NewConfigManager()
	require.NoError(t, err)
	manager.config.ActiveProfile = "" // Force no active profile
	err = manager.Save()
	require.NoError(t, err)

	cmd, cmdOut = setupTestCmd()
	err = ViewConfig(cmd, nil, false)
	require.NoError(t, err)

	output = cmdOut.String()
	assert.Contains(t, output, "No active profile set")
}

func TestDeleteProfile_CMD(t *testing.T) {
	_, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Setup utility function to automatically confirm deletion
	oldConfirmPrompt := utils.ConfirmPrompt
	utils.ConfirmPrompt = func(message string, noColor bool) bool {
		return true // Auto-confirm
	}
	defer func() { utils.ConfirmPrompt = oldConfirmPrompt }()

	// Create profiles
	manager, err := NewConfigManager()
	require.NoError(t, err)
	err = manager.CreateProfile("test-profile", "access", "secret", "production", "")
	require.NoError(t, err)
	err = manager.CreateProfile("active-profile", "access2", "secret2", "production", "")
	require.NoError(t, err)

	// Set active profile and save
	err = manager.UseProfile("active-profile")
	require.NoError(t, err)

	// Test deleting non-active profile
	cmd, _ := setupTestCmd()
	outputText, err := captureOutputFromAction(func() error {
		return DeleteProfile(cmd, []string{"test-profile"}, false)
	})
	require.NoError(t, err)
	assert.Contains(t, outputText, "Profile 'test-profile' deleted successfully")

	// Very important: Create a new manager instance to read from disk
	manager, err = NewConfigManager()
	require.NoError(t, err)

	// Verify the profile was deleted
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)
	_, exists := profiles["test-profile"]
	assert.False(t, exists, "Profile should be deleted")

	// Test deleting active profile (should fail)
	cmd, _ = setupTestCmd()
	// For this test, don't use output capture since it returns the error but doesn't print it
	err = DeleteProfile(cmd, []string{"active-profile"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete active profile") // Check the error message directly
}

func TestExportImportConfig(t *testing.T) {
	configDir, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Create a profile first
	manager, err := NewConfigManager()
	require.NoError(t, err)
	err = manager.CreateProfile("test-profile", "test-access", "test-secret", "production", "Test profile")
	require.NoError(t, err)
	err = manager.UseProfile("test-profile")
	require.NoError(t, err)
	err = manager.SetDefault("output", "json")
	require.NoError(t, err)

	// Export config to file - use absolute path to avoid any path resolution issues
	exportPath := filepath.Join(configDir, "export.json")
	cmd, _ := setupTestCmd()
	cmd.Flags().String("file", exportPath, "")

	// Mark flag as changed by parsing args
	err = cmd.ParseFlags([]string{"--file=" + exportPath})
	require.NoError(t, err)

	outputText, err := captureOutputFromAction(func() error {
		return ExportConfig(cmd, nil, false)
	})
	require.NoError(t, err)
	assert.Contains(t, outputText, "Configuration exported")

	// Verify file exists
	_, err = os.Stat(exportPath)
	require.NoError(t, err)

	// Read exported content to verify it - this also ensures file is fully written
	exportContent, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	require.NotEmpty(t, exportContent, "Exported content should not be empty")
	t.Logf("Export content: %s", string(exportContent))

	// Verify that export has redacted credentials
	assert.Contains(t, string(exportContent), "[REDACTED]", "Export should have redacted credentials")
	assert.Contains(t, string(exportContent), "test-profile", "Export should include profile name")
	assert.Contains(t, string(exportContent), "production", "Export should include environment setting")
	assert.Contains(t, string(exportContent), "Test profile", "Export should include description")
	assert.Contains(t, string(exportContent), "\"output\": \"json\"", "Export should include defaults")

	// *** SPECIAL TEST HANDLING ***
	// Since the import can't restore profiles with redacted secrets (by design),
	// we'll create a test-specific export file that doesn't have redacted secrets
	// This simulates a user exporting from one machine and importing to another
	// where they manually edit the export file to use real credentials
	manualExportPath := filepath.Join(configDir, "manual-export.json")
	manualExportContent := `{
        "version": 2,
        "active_profile": "test-profile",
        "profiles": {
            "test-profile": {
                "access_key": "test-access",
                "secret_key": "test-secret",
                "environment": "production",
                "description": "Test profile"
            }
        },
        "defaults": {
            "output": "json"
        }
    }`
	err = os.WriteFile(manualExportPath, []byte(manualExportContent), 0644)
	require.NoError(t, err)

	// *** Important: Create a completely new config directory ***
	_, newCleanup := setupTestConfigEnv(t)
	defer newCleanup()

	// Import to the new environment using the manual export file
	cmd, _ = setupTestCmd()
	cmd.Flags().String("file", manualExportPath, "")

	// Mark flag as changed by parsing args
	err = cmd.ParseFlags([]string{"--file=" + manualExportPath})
	require.NoError(t, err)

	// Mock confirmation
	oldConfirmPrompt := utils.ConfirmPrompt
	utils.ConfirmPrompt = func(message string, noColor bool) bool {
		return true // Auto-confirm
	}
	defer func() { utils.ConfirmPrompt = oldConfirmPrompt }()

	outputText, err = captureOutputFromAction(func() error {
		return ImportConfig(cmd, nil, false)
	})

	require.NoError(t, err)
	assert.Contains(t, outputText, "Configuration imported successfully")

	// Create a new manager to read from disk
	newManager, err := NewConfigManager()
	require.NoError(t, err, "Failed to create new config manager")

	profiles, err := newManager.ListProfiles()
	require.NoError(t, err)

	profile, exists := profiles["test-profile"]
	assert.True(t, exists, "Profile should exist after import")
	if exists {
		assert.Equal(t, "production", profile.Environment)
		assert.Equal(t, "Test profile", profile.Description)
		assert.Equal(t, "test-access", profile.AccessKey)
		assert.Equal(t, "test-secret", profile.SecretKey)
	}

	// Verify defaults were imported
	val, exists := newManager.GetDefault("output")
	assert.True(t, exists)
	if exists {
		assert.Equal(t, "json", val)
	}
}

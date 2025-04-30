package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestConfig(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "megaport-config-test")
	require.NoError(t, err)

	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

	return tempDir, func() {
		os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)
		os.RemoveAll(tempDir)
	}
}

func TestNewConfigManager(t *testing.T) {
	tempDir, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)
	require.NotNil(t, manager)

	configPath := filepath.Join(tempDir, "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "Config file should exist")

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config ConfigFile
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	assert.Equal(t, ConfigVersion, config.Version)
	assert.Empty(t, config.ActiveProfile)
	assert.Empty(t, config.Profiles)
	assert.NotNil(t, config.Defaults)
}

func TestCreateProfile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "Test profile")
	require.NoError(t, err)
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)

	profile, exists := profiles["test-profile"]
	assert.True(t, exists, "Profile should exist")
	assert.Equal(t, "access123", profile.AccessKey)
	assert.Equal(t, "secret123", profile.SecretKey)
	assert.Equal(t, "production", profile.Environment)
	assert.Equal(t, "Test profile", profile.Description)
}

func TestUpdateProfile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "Original")
	require.NoError(t, err)

	err = manager.UpdateProfile("test-profile", "", "", "staging", false, "")
	require.NoError(t, err)
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)

	profile, exists := profiles["test-profile"]
	assert.True(t, exists, "Profile should exist")
	assert.Equal(t, "access123", profile.AccessKey, "AccessKey should remain unchanged")
	assert.Equal(t, "secret123", profile.SecretKey, "SecretKey should remain unchanged")
	assert.Equal(t, "staging", profile.Environment, "Environment should be updated")
	assert.Equal(t, "Original", profile.Description, "Description should remain unchanged")

	err = manager.UpdateProfile("test-profile", "newaccess", "", "", true, "Updated desc")
	require.NoError(t, err)

	profiles, err = manager.ListProfiles()
	require.NoError(t, err)

	profile, exists = profiles["test-profile"]
	assert.True(t, exists, "Profile should exist")
	assert.Equal(t, "newaccess", profile.AccessKey, "AccessKey should be updated")
	assert.Equal(t, "secret123", profile.SecretKey, "SecretKey should remain unchanged")
	assert.Equal(t, "staging", profile.Environment, "Environment should remain unchanged")
	assert.Equal(t, "Updated desc", profile.Description, "Description should be updated")

	err = manager.UpdateProfile("non-existent", "foo", "bar", "production", false, "")
	assert.Error(t, err)
	assert.Equal(t, ErrProfileNotFound, err)
}

func TestDeleteProfile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "")
	require.NoError(t, err)

	err = manager.UseProfile("test-profile")
	require.NoError(t, err)

	err = manager.DeleteProfile("test-profile")
	assert.Error(t, err, "Should not allow deleting active profile")

	err = manager.CreateProfile("other-profile", "access456", "secret456", "staging", "")
	require.NoError(t, err)

	err = manager.DeleteProfile("other-profile")
	require.NoError(t, err)

	profiles, err := manager.ListProfiles()
	require.NoError(t, err)
	_, exists := profiles["other-profile"]
	assert.False(t, exists, "Profile should be deleted")

	err = manager.DeleteProfile("non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrProfileNotFound, err)
}

func TestGetCurrentProfile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	_, _, err = manager.GetCurrentProfile()
	assert.Equal(t, ErrProfileNotFound, err)

	err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "")
	require.NoError(t, err)

	err = manager.UseProfile("test-profile")
	require.NoError(t, err)

	profile, name, err := manager.GetCurrentProfile()
	require.NoError(t, err)
	assert.Equal(t, "test-profile", name)
	assert.Equal(t, "access123", profile.AccessKey)
	assert.Equal(t, "secret123", profile.SecretKey)
	assert.Equal(t, "production", profile.Environment)
}

func TestDefaultSettings(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	val, exists := manager.GetDefault("output")
	assert.False(t, exists)
	assert.Nil(t, val)

	err = manager.SetDefault("output", "json")
	require.NoError(t, err)

	err = manager.SetDefault("no-color", true)
	require.NoError(t, err)

	val, exists = manager.GetDefault("output")
	assert.True(t, exists)
	assert.Equal(t, "json", val)

	val, exists = manager.GetDefault("no-color")
	assert.True(t, exists)
	assert.Equal(t, true, val)

	// Remove a default
	err = manager.RemoveDefault("output")
	require.NoError(t, err)

	val, exists = manager.GetDefault("output")
	assert.False(t, exists)
	assert.Nil(t, val)

	// Clear all defaults
	err = manager.ClearDefaults()
	require.NoError(t, err)

	val, exists = manager.GetDefault("no-color")
	assert.False(t, exists)
	assert.Nil(t, val)
}

func TestExportConfig(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("test-profile", "access123", "secret456", "production", "Test desc")
	require.NoError(t, err)

	err = manager.UseProfile("test-profile")
	require.NoError(t, err)

	err = manager.SetDefault("output", "json")
	require.NoError(t, err)

	exported, err := manager.Export()
	require.NoError(t, err)

	assert.Equal(t, "test-profile", exported.ActiveProfile)
	assert.Equal(t, "json", exported.Defaults["output"])

	exportedProfile, exists := exported.Profiles["test-profile"]
	assert.True(t, exists)
	assert.Equal(t, "[REDACTED]", exportedProfile.AccessKey)
	assert.Equal(t, "[REDACTED]", exportedProfile.SecretKey)
	assert.Equal(t, "production", exportedProfile.Environment)
	assert.Equal(t, "Test desc", exportedProfile.Description)
}

func TestListProfiles(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	profiles, err := manager.ListProfiles()
	require.NoError(t, err)
	assert.Empty(t, profiles)

	err = manager.CreateProfile("profile1", "access1", "secret1", "production", "")
	require.NoError(t, err)

	err = manager.CreateProfile("profile2", "access2", "secret2", "staging", "")
	require.NoError(t, err)

	profiles, err = manager.ListProfiles()
	require.NoError(t, err)
	assert.Len(t, profiles, 2)

	_, exists := profiles["profile1"]
	assert.True(t, exists)

	_, exists = profiles["profile2"]
	assert.True(t, exists)
}

func TestUseProfile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("profile1", "access1", "secret1", "production", "")
	require.NoError(t, err)

	err = manager.CreateProfile("profile2", "access2", "secret2", "staging", "")
	require.NoError(t, err)

	err = manager.UseProfile("profile1")
	require.NoError(t, err)
	assert.Equal(t, "profile1", manager.config.ActiveProfile)

	err = manager.UseProfile("profile2")
	require.NoError(t, err)
	assert.Equal(t, "profile2", manager.config.ActiveProfile)

	err = manager.UseProfile("non-existent")
	assert.Error(t, err)
	assert.Equal(t, ErrProfileNotFound, err)
	assert.Equal(t, "profile2", manager.config.ActiveProfile, "Active profile should remain unchanged")
}

func TestNilSafety(t *testing.T) {
	var m *ConfigManager = nil
	profiles, err := m.ListProfiles()
	assert.NoError(t, err)
	assert.Empty(t, profiles)
}

func TestConfigPersistence(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	{
		manager, err := NewConfigManager()
		require.NoError(t, err)

		err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "")
		require.NoError(t, err)

		err = manager.UseProfile("test-profile")
		require.NoError(t, err)
	}

	{
		manager, err := NewConfigManager()
		require.NoError(t, err)

		profile, name, err := manager.GetCurrentProfile()
		require.NoError(t, err)
		assert.Equal(t, "test-profile", name)
		assert.Equal(t, "access123", profile.AccessKey)
	}
}

func TestCorruptedConfigFile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)
	err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "")
	require.NoError(t, err)

	configPath, err := GetConfigFilePath()
	require.NoError(t, err)
	err = os.WriteFile(configPath, []byte("{this is not valid json"), 0644)
	require.NoError(t, err)

	manager, err = NewConfigManager()
	require.NoError(t, err)

	profiles, err := manager.ListProfiles()
	require.NoError(t, err)
	assert.Empty(t, profiles, "Corrupted config should be replaced with default empty config")

	err = manager.CreateProfile("new-profile", "access123", "secret123", "production", "")
	require.NoError(t, err)
}

func TestSpecialProfileNames(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("", "access123", "secret123", "production", "")
	assert.Error(t, err, "Empty profile name should be rejected")

	err = manager.CreateProfile("   ", "access123", "secret123", "production", "")
	assert.Error(t, err, "Whitespace-only profile name should be rejected")

	specialName := "test!@#$%^&*()_+-=[]{}|;':,./<>?"
	err = manager.CreateProfile(specialName, "access123", "secret123", "production", "")
	require.NoError(t, err)

	unicodeName := "उपयोगकर्ता-परीक्षण"
	err = manager.CreateProfile(unicodeName, "access123", "secret123", "production", "")
	require.NoError(t, err)

	// Verify profiles were created
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)

	_, exists := profiles[specialName]
	assert.True(t, exists, "Profile with special characters should exist")

	_, exists = profiles[unicodeName]
	assert.True(t, exists, "Profile with Unicode characters should exist")

	// Test we can use these profiles
	err = manager.UseProfile(specialName)
	require.NoError(t, err)

	err = manager.UseProfile(unicodeName)
	require.NoError(t, err)
}

func TestDuplicateProfiles(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	// Create initial profile
	err = manager.CreateProfile("test-profile", "access123", "secret123", "production", "Original")
	require.NoError(t, err)

	// Create profile with same name - should overwrite
	err = manager.CreateProfile("test-profile", "newaccess", "newsecret", "staging", "Updated")
	require.NoError(t, err)

	// Verify profile was overwritten
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)

	profile, exists := profiles["test-profile"]
	assert.True(t, exists, "Profile should exist")
	assert.Equal(t, "newaccess", profile.AccessKey, "Access key should be updated")
	assert.Equal(t, "newsecret", profile.SecretKey, "Secret key should be updated")
	assert.Equal(t, "staging", profile.Environment, "Environment should be updated")
	assert.Equal(t, "Updated", profile.Description, "Description should be updated")
}

func TestConfigVersionHandling(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	// Create a config file with older version
	oldConfig := ConfigFile{
		Version:       1, // Older version than current
		ActiveProfile: "",
		Profiles:      make(map[string]*Profile),
		Defaults:      make(map[string]interface{}),
	}

	oldConfig.Profiles["old-profile"] = &Profile{
		AccessKey:   "old-access",
		SecretKey:   "old-secret",
		Environment: "production",
	}

	// Write this to the config file
	configPath, err := GetConfigFilePath()
	require.NoError(t, err)

	data, err := json.MarshalIndent(oldConfig, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Now try to load the config - should handle version migration
	manager, err := NewConfigManager()
	require.NoError(t, err)

	// Verify old profile data was retained
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)

	profile, exists := profiles["old-profile"]
	assert.True(t, exists, "Old profile should be migrated")
	assert.Equal(t, "old-access", profile.AccessKey)

	// Verify config was migrated to current version
	assert.Equal(t, ConfigVersion, manager.config.Version, "Config should be upgraded to current version")
}
func TestReadOnlyConfigFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	tempDir, err := os.MkdirTemp("", "megaport-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	// Create the config directory first
	configDir, err := GetConfigDir()
	require.NoError(t, err)

	// Create a version 1 config file (to trigger migration)
	configPath := filepath.Join(configDir, "config.json")
	oldConfig := ConfigFile{
		Version:  1, // Old version to force migration
		Profiles: make(map[string]*Profile),
		Defaults: make(map[string]interface{}),
	}
	data, err := json.MarshalIndent(oldConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Make the config file read-only
	err = os.Chmod(configPath, 0444)
	require.NoError(t, err)

	// Try to load config - should fail due to version migration
	// attempting to save to read-only file
	_, err = NewConfigManager()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestProfileNameCaseSensitivity(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	// Create profiles with names differing only by case
	err = manager.CreateProfile("test-profile", "access1", "secret1", "production", "")
	require.NoError(t, err)

	err = manager.CreateProfile("Test-Profile", "access2", "secret2", "staging", "")
	require.NoError(t, err)

	// Verify both profiles exist
	profiles, err := manager.ListProfiles()
	require.NoError(t, err)
	assert.Len(t, profiles, 2, "Should have two distinct profiles with case-sensitive names")

	// Verify we can access each profile correctly
	profile1, exists := profiles["test-profile"]
	assert.True(t, exists)
	assert.Equal(t, "access1", profile1.AccessKey)

	profile2, exists := profiles["Test-Profile"]
	assert.True(t, exists)
	assert.Equal(t, "access2", profile2.AccessKey)

	// Verify case-sensitivity in profile selection
	err = manager.UseProfile("test-profile")
	require.NoError(t, err)
	assert.Equal(t, "test-profile", manager.config.ActiveProfile)

	err = manager.UseProfile("Test-Profile")
	require.NoError(t, err)
	assert.Equal(t, "Test-Profile", manager.config.ActiveProfile)
}

func TestVeryLongPaths(t *testing.T) {
	// Create a deeply nested temporary directory
	tempBase, err := os.MkdirTemp("", "megaport-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempBase)

	// Create a very long path that stays within OS limits but is unusually deep
	longPath := tempBase
	for i := 0; i < 15; i++ {
		longPath = filepath.Join(longPath, fmt.Sprintf("nested_dir_%d", i))
	}

	// Set as config dir
	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", longPath)
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	// Should still work
	manager, err := NewConfigManager()
	require.NoError(t, err)

	// Verify we can create and use profiles
	err = manager.CreateProfile("test", "access", "secret", "production", "")
	require.NoError(t, err)

	// Verify proper path creation
	_, err = os.Stat(filepath.Join(longPath, "config.json"))
	require.NoError(t, err)
}

func TestSymlinkConfigDir(t *testing.T) {
	// Create real dir and symlink dir
	realDir, err := os.MkdirTemp("", "megaport-real")
	require.NoError(t, err)
	defer os.RemoveAll(realDir)

	symlinkDir, err := os.MkdirTemp("", "megaport-symlink")
	require.NoError(t, err)
	defer os.RemoveAll(symlinkDir)

	// Remove the symlink target
	os.RemoveAll(symlinkDir)

	// Create symlink
	err = os.Symlink(realDir, symlinkDir)
	require.NoError(t, err)

	// Use symlink as config dir
	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", symlinkDir)
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	// Should work with symlink
	manager, err := NewConfigManager()
	require.NoError(t, err)

	// Verify we can create profiles
	err = manager.CreateProfile("test", "access", "secret", "production", "")
	require.NoError(t, err)

	// Config file should exist in the real directory
	_, err = os.Stat(filepath.Join(realDir, "config.json"))
	require.NoError(t, err)
}

func TestImportWithMissingFields(t *testing.T) {
	_, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Create a minimal import file with missing fields
	minimalConfig := `{
        "version": 2,
        "active_profile": "minimal-profile",
        "profiles": {
            "minimal-profile": {
                "access_key": "access",
                "secret_key": "secret"
            }
        },
        "defaults": {}
    }`

	importPath := filepath.Join(os.TempDir(), "minimal-config.json")
	err := os.WriteFile(importPath, []byte(minimalConfig), 0644)
	require.NoError(t, err)
	defer os.Remove(importPath)

	// Import this minimal config
	cmd, _ := setupTestCmd()
	cmd.Flags().String("file", importPath, "")
	err = cmd.ParseFlags([]string{"--file=" + importPath})
	require.NoError(t, err)

	// Mock confirmation
	oldConfirmPrompt := utils.ConfirmPrompt
	utils.ConfirmPrompt = func(message string, noColor bool) bool {
		return true
	}
	defer func() { utils.ConfirmPrompt = oldConfirmPrompt }()

	_, err = captureOutputFromAction(func() error {
		return ImportConfig(cmd, nil, false)
	})
	require.NoError(t, err)

	// Verify the import worked and defaulted missing fields
	manager, err := NewConfigManager()
	require.NoError(t, err)

	profile, _, err := manager.GetCurrentProfile()
	require.NoError(t, err)
	assert.Equal(t, "minimal-profile", manager.config.ActiveProfile)
	assert.Equal(t, "access", profile.AccessKey)
	assert.Equal(t, "secret", profile.SecretKey)
	assert.NotEmpty(t, profile.Environment, "Environment should have a default value")
}

func TestExportWithMaxProfiles(t *testing.T) {
	_, cleanup := setupTestConfigEnv(t)
	defer cleanup()

	// Create many profiles (testing export with large dataset)
	manager, err := NewConfigManager()
	require.NoError(t, err)

	// Create 100 profiles
	for i := 0; i < 100; i++ {
		profileName := fmt.Sprintf("profile-%d", i)
		err = manager.CreateProfile(profileName, "access", "secret", "production", "")
		require.NoError(t, err)
	}

	// Export all these profiles
	exportPath := filepath.Join(os.TempDir(), "many-profiles-export.json")
	cmd, _ := setupTestCmd()
	cmd.Flags().String("file", exportPath, "")
	err = cmd.ParseFlags([]string{"--file=" + exportPath})
	require.NoError(t, err)

	_, err = captureOutputFromAction(func() error {
		return ExportConfig(cmd, nil, false)
	})
	require.NoError(t, err)

	// Verify export worked
	_, err = os.Stat(exportPath)
	require.NoError(t, err)
	defer os.Remove(exportPath)

	// Read the export file and verify it contains all profiles
	data, err := os.ReadFile(exportPath)
	require.NoError(t, err)

	var exportedConfig ConfigFile
	err = json.Unmarshal(data, &exportedConfig)
	require.NoError(t, err)

	assert.Equal(t, 100, len(exportedConfig.Profiles), "Export should contain all 100 profiles")
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "megaport-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	configDir, err := GetConfigDir()
	require.NoError(t, err)
	assert.Equal(t, tempDir, configDir)

	os.Setenv("MEGAPORT_CONFIG_DIR", "")
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	configDir, err = GetConfigDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(homeDir, ".megaport"), configDir)

	_, err = os.Stat(configDir)
	assert.NoError(t, err)
}

func TestGetConfigFilePath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "megaport-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	configPath, err := GetConfigFilePath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, "config.json"), configPath)
}

func TestPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	parentDir, err := os.MkdirTemp("", "megaport-config-test-parent")
	require.NoError(t, err)
	defer os.RemoveAll(parentDir)
	tempDir := filepath.Join(parentDir, "config-subdir")
	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	err = os.Chmod(parentDir, 0500)
	require.NoError(t, err)
	_, err = GetConfigDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config directory")
}

func TestEmptyConfigFile(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	configPath, err := GetConfigFilePath()
	require.NoError(t, err)

	err = os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	manager, err := NewConfigManager()
	require.NoError(t, err)

	assert.Equal(t, ConfigVersion, manager.config.Version)
	assert.NotNil(t, manager.config.Profiles)
}

func TestVeryLargeValues(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	largeDesc := strings.Repeat("This is a very long description. ", 2500)

	err = manager.CreateProfile("large-profile", "access", "secret", "production", largeDesc)
	require.NoError(t, err)

	newManager, err := NewConfigManager()
	require.NoError(t, err)

	profiles, err := newManager.ListProfiles()
	require.NoError(t, err)

	profile, exists := profiles["large-profile"]
	assert.True(t, exists)
	assert.Equal(t, largeDesc, profile.Description)
}

func TestMalformedButValidJSON(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	configPath, err := GetConfigFilePath()
	require.NoError(t, err)

	malformedConfig := `{
        "version": 2,
        "active_profile": "test",
        "profiles": "not an object but a string",
        "defaults": {}
    }`

	err = os.WriteFile(configPath, []byte(malformedConfig), 0644)
	require.NoError(t, err)

	manager, err := NewConfigManager()
	require.NoError(t, err)

	assert.NotNil(t, manager.config.Profiles)
	assert.Empty(t, manager.config.Profiles)
}

func TestConfigFilePermissions(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	err = manager.CreateProfile("secure-profile", "sensitive-access", "sensitive-secret", "production", "")
	require.NoError(t, err)

	configPath, err := GetConfigFilePath()
	require.NoError(t, err)

	info, err := os.Stat(configPath)
	require.NoError(t, err)

	expectedMode := os.FileMode(0600)
	assert.Equal(t, expectedMode, info.Mode().Perm(),
		"Config file should have 0600 permissions")
}

func TestSecretHandling(t *testing.T) {
	_, cleanup := setupTestConfig(t)
	defer cleanup()

	manager, err := NewConfigManager()
	require.NoError(t, err)

	secretKey := "SECRET_do_not_share_123!"
	err = manager.CreateProfile("secure-profile", "access", secretKey, "production", "")
	require.NoError(t, err)

	exported, err := manager.Export()
	require.NoError(t, err)

	exportedProfile, exists := exported.Profiles["secure-profile"]
	assert.True(t, exists)
	assert.Equal(t, "[REDACTED]", exportedProfile.SecretKey, "Secret key should be redacted in export")
	assert.NotEqual(t, secretKey, exportedProfile.SecretKey, "Raw secret should never appear in export")
}

func TestChangingConfigDirMidway(t *testing.T) {
	dir1, err := os.MkdirTemp("", "megaport-config-test1")
	require.NoError(t, err)
	defer os.RemoveAll(dir1)

	dir2, err := os.MkdirTemp("", "megaport-config-test2")
	require.NoError(t, err)
	defer os.RemoveAll(dir2)

	oldConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")
	defer os.Setenv("MEGAPORT_CONFIG_DIR", oldConfigDir)

	os.Setenv("MEGAPORT_CONFIG_DIR", dir1)

	manager1, err := NewConfigManager()
	require.NoError(t, err)

	err = manager1.CreateProfile("profile1", "access1", "secret1", "production", "")
	require.NoError(t, err)

	os.Setenv("MEGAPORT_CONFIG_DIR", dir2)

	manager2, err := NewConfigManager()
	require.NoError(t, err)

	configPath1 := filepath.Join(dir1, "config.json")
	configPath2 := filepath.Join(dir2, "config.json")

	err = manager2.CreateProfile("profile2", "access2", "secret2", "staging", "")
	require.NoError(t, err)

	err = manager1.UpdateProfile("profile1", "updated", "", "", false, "")
	require.NoError(t, err)

	configData1, err := os.ReadFile(configPath1)
	require.NoError(t, err)
	var config1 ConfigFile
	err = json.Unmarshal(configData1, &config1)
	require.NoError(t, err)

	profile1, exists := config1.Profiles["profile1"]
	require.True(t, exists, "Profile1 should exist in first config")
	assert.Equal(t, "updated", profile1.AccessKey, "Profile1 should be updated in first config")

	configData2, err := os.ReadFile(configPath2)
	require.NoError(t, err)
	var config2 ConfigFile
	err = json.Unmarshal(configData2, &config2)
	require.NoError(t, err)

	profile2, exists := config2.Profiles["profile2"]
	require.True(t, exists, "Profile2 should exist in second config")
	assert.Equal(t, "access2", profile2.AccessKey, "Profile2 should have correct access key in second config")

	_, exists = config2.Profiles["profile1"]
	assert.False(t, exists, "Profile1 should not exist in second config")

	_, exists = config1.Profiles["profile2"]
	assert.False(t, exists, "Profile2 should not exist in first config")
}

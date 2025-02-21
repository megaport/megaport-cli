package cmd

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testConfigFile = filepath.Join(os.TempDir(), ".megaport-cli-config-test.json")
)

func init() {
	configFile = testConfigFile
}

func cleanup(_ *testing.T) {
	os.Remove(testConfigFile)
}

func TestEncryptDecrypt(t *testing.T) {
	password := "test-password"
	data := []byte("test-data")

	// Generate salt
	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	assert.NoError(t, err)

	// Encrypt data
	encrypted, err := encrypt(password, salt, data)
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	// Decrypt data
	decrypted, err := decrypt(password, salt, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestWriteReadConfig(t *testing.T) {
	cleanup(t)
	defer cleanup(t)

	cfg := Config{
		Environment:     "staging",
		EncryptedAccess: "encrypted-access",
		EncryptedSecret: "encrypted-secret",
		Salt:            "salt",
	}

	// Write config to file
	err := writeConfigFile(cfg)
	assert.NoError(t, err)

	// Read config from file
	loadedCfg, err := loadConfig()
	assert.NoError(t, err)
	assert.Equal(t, cfg, *loadedCfg)
}

func TestPromptPassword(t *testing.T) {
	// Save original reader
	original := passwordReader
	defer func() { passwordReader = original }()

	// Use test implementation
	passwordReader = &TestPasswordReader{input: "test-password"}

	password, err := promptPassword("Enter password: ")
	assert.NoError(t, err)
	assert.Equal(t, "test-password", password)
}
func TestLoadConfig_FileNotFound(t *testing.T) {
	cleanup(t)
	defer cleanup(t)

	_, err := loadConfig()
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestWriteConfigFile_Error(t *testing.T) {
	// Simulate error by setting an invalid config file path
	invalidConfigFile := string([]byte{0})
	configFile = invalidConfigFile
	defer func() { configFile = testConfigFile }()

	cfg := Config{
		Environment:     "staging",
		EncryptedAccess: "encrypted-access",
		EncryptedSecret: "encrypted-secret",
		Salt:            "salt",
	}

	err := writeConfigFile(cfg)
	assert.Error(t, err)
}

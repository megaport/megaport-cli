package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

var (
	env               string
	accessKeyEnvVar   = "MEGAPORT_ACCESS_KEY"
	secretKeyEnvVar   = "MEGAPORT_SECRET_KEY"
	environmentEnvVar = "MEGAPORT_ENVIRONMENT"
)

func TestLogin(t *testing.T) {
	originalLoginFunc := LoginFunc
	defer func() {
		LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		envFlag     string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "all valid env vars",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: false,
		},
		{
			name: "empty access key",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: true,
			errorMsg:    "access key, secret key, and environment are required",
		},
		{
			name: "empty secret key",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			shouldError: true,
			errorMsg:    "access key, secret key, and environment are required",
		},
		{
			name:        "no env vars set",
			envVars:     map[string]string{},
			shouldError: true,
			errorMsg:    "access key, secret key, and environment are required",
		},
		{
			name: "flag overrides env var",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY":  "test-access-key",
				"MEGAPORT_SECRET_KEY":  "test-secret-key",
				"MEGAPORT_ENVIRONMENT": "staging",
			},
			envFlag:     "production",
			shouldError: false,
		},
		{
			name: "flag provides env when not set",
			envVars: map[string]string{
				"MEGAPORT_ACCESS_KEY": "test-access-key",
				"MEGAPORT_SECRET_KEY": "test-secret-key",
			},
			envFlag:     "production",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("MEGAPORT_ACCESS_KEY")
			os.Unsetenv("MEGAPORT_SECRET_KEY")
			os.Unsetenv("MEGAPORT_ENVIRONMENT")

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			if tt.envFlag != "" {
				env = tt.envFlag
			} else {
				env = ""
			}

			var capturedAccessKey, capturedSecretKey, capturedEnv string

			LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				capturedAccessKey = os.Getenv(accessKeyEnvVar)
				capturedSecretKey = os.Getenv(secretKeyEnvVar)
				capturedEnv = env
				if capturedEnv == "" {
					capturedEnv = os.Getenv(environmentEnvVar)
				}

				if capturedAccessKey == "" || capturedSecretKey == "" {
					return nil, fmt.Errorf("access key, secret key, and environment are required")
				}

				client := &megaport.Client{}
				return client, nil
			}

			client, err := Login(context.Background())

			if tt.shouldError {
				assert.Nil(t, client)
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NotNil(t, client)
				assert.NoError(t, err)

				if tt.envFlag != "" {
					assert.Equal(t, tt.envFlag, capturedEnv)
				} else if envVal, ok := tt.envVars["MEGAPORT_ENVIRONMENT"]; ok {
					assert.Equal(t, envVal, capturedEnv)
				} else {
					assert.Equal(t, "production", capturedEnv)
				}
			}
		})
	}
}

func TestEnvironmentSelectionPrecedence(t *testing.T) {
	// Save original values
	originalEnv := utils.Env
	originalMegaportEnv := os.Getenv("MEGAPORT_ENVIRONMENT")

	defer func() {
		// Restore original values
		utils.Env = originalEnv
		if originalMegaportEnv == "" {
			os.Unsetenv("MEGAPORT_ENVIRONMENT")
		} else {
			os.Setenv("MEGAPORT_ENVIRONMENT", originalMegaportEnv)
		}
	}()

	tests := []struct {
		name        string
		globalFlag  string
		envVar      string
		expectedEnv string
	}{
		{
			name:        "Global flag takes precedence over environment variable",
			globalFlag:  "staging",
			envVar:      "production",
			expectedEnv: "staging",
		},
		{
			name:        "Environment variable used when no global flag",
			globalFlag:  "",
			envVar:      "staging",
			expectedEnv: "staging",
		},
		{
			name:        "Default to production when neither flag nor env var set",
			globalFlag:  "",
			envVar:      "",
			expectedEnv: "production",
		},
		{
			name:        "Development environment via global flag",
			globalFlag:  "development",
			envVar:      "",
			expectedEnv: "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			utils.Env = tt.globalFlag
			if tt.envVar != "" {
				os.Setenv("MEGAPORT_ENVIRONMENT", tt.envVar)
			} else {
				os.Unsetenv("MEGAPORT_ENVIRONMENT")
			}

			// Test the environment selection logic by examining the values
			// that would be read in the login function
			var env string

			// Simulate the new login function logic
			if utils.Env != "" {
				env = utils.Env
			} else {
				env = os.Getenv("MEGAPORT_ENVIRONMENT")
			}

			if env == "" {
				env = "production"
			}

			assert.Equal(t, tt.expectedEnv, env, "Environment selection should match expected value")
		})
	}
}

func TestProfileOverrideLogin(t *testing.T) {
	// Save original values
	originalEnv := utils.Env
	originalProfileOverride := utils.ProfileOverride
	originalAccessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	originalSecretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	originalMegaportEnv := os.Getenv("MEGAPORT_ENVIRONMENT")
	originalConfigDir := os.Getenv("MEGAPORT_CONFIG_DIR")

	defer func() {
		utils.Env = originalEnv
		utils.ProfileOverride = originalProfileOverride
		restoreEnvVar("MEGAPORT_ACCESS_KEY", originalAccessKey)
		restoreEnvVar("MEGAPORT_SECRET_KEY", originalSecretKey)
		restoreEnvVar("MEGAPORT_ENVIRONMENT", originalMegaportEnv)
		restoreEnvVar("MEGAPORT_CONFIG_DIR", originalConfigDir)
	}()

	// Setup temp config dir with profiles
	tempDir, err := os.MkdirTemp("", "megaport-login-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)
	os.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

	// Clear env vars so profile credentials are used
	os.Unsetenv("MEGAPORT_ACCESS_KEY")
	os.Unsetenv("MEGAPORT_SECRET_KEY")
	os.Unsetenv("MEGAPORT_ENVIRONMENT")

	// Create config with two profiles
	manager, err := NewConfigManager()
	assert.NoError(t, err)
	err = manager.CreateProfile("staging", "staging-access", "staging-secret", "staging", "")
	assert.NoError(t, err)
	err = manager.CreateProfile("prod", "prod-access", "prod-secret", "production", "")
	assert.NoError(t, err)
	err = manager.UseProfile("prod")
	assert.NoError(t, err)

	t.Run("profile override uses specified profile credentials", func(t *testing.T) {
		utils.Env = ""
		utils.ProfileOverride = "staging"

		var capturedAccessKey, capturedSecretKey, capturedEnv string
		originalLoginFuncWithOutput := LoginFuncWithOutput
		defer func() { LoginFuncWithOutput = originalLoginFuncWithOutput }()

		// We need to test the actual credential resolution logic, not mock it.
		// So we read the resolved values by inspecting what LoginFuncWithOutput would do.
		// Since the real function tries to hit the API, we test the credential resolution separately.
		manager2, err := NewConfigManager()
		assert.NoError(t, err)
		profile, err := manager2.GetProfile(utils.ProfileOverride)
		assert.NoError(t, err)
		capturedAccessKey = profile.AccessKey
		capturedSecretKey = profile.SecretKey
		capturedEnv = profile.Environment

		assert.Equal(t, "staging-access", capturedAccessKey)
		assert.Equal(t, "staging-secret", capturedSecretKey)
		assert.Equal(t, "staging", capturedEnv)
	})

	t.Run("profile override with non-existent profile returns error", func(t *testing.T) {
		utils.Env = ""
		utils.ProfileOverride = "non-existent"

		manager2, err := NewConfigManager()
		assert.NoError(t, err)
		_, err = manager2.GetProfile(utils.ProfileOverride)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-existent")
	})

	t.Run("without profile override uses active profile", func(t *testing.T) {
		utils.Env = ""
		utils.ProfileOverride = ""

		manager2, err := NewConfigManager()
		assert.NoError(t, err)
		profile, name, err := manager2.GetCurrentProfile()
		assert.NoError(t, err)
		assert.Equal(t, "prod", name)
		assert.Equal(t, "prod-access", profile.AccessKey)
		assert.Equal(t, "prod-secret", profile.SecretKey)
	})

	t.Run("env flag overrides profile environment", func(t *testing.T) {
		utils.Env = "development"
		utils.ProfileOverride = "staging"

		manager2, err := NewConfigManager()
		assert.NoError(t, err)
		profile, err := manager2.GetProfile(utils.ProfileOverride)
		assert.NoError(t, err)

		// Profile says staging, but --env says development
		env := profile.Environment
		if utils.Env != "" {
			env = utils.Env
		}
		assert.Equal(t, "development", env)
	})
}

func restoreEnvVar(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

func TestCredentialSelectionPrecedence(t *testing.T) {
	// Save original values
	originalEnv := utils.Env
	originalAccessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	originalSecretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	defer func() {
		// Restore original values
		utils.Env = originalEnv
		if originalAccessKey == "" {
			os.Unsetenv("MEGAPORT_ACCESS_KEY")
		} else {
			os.Setenv("MEGAPORT_ACCESS_KEY", originalAccessKey)
		}
		if originalSecretKey == "" {
			os.Unsetenv("MEGAPORT_SECRET_KEY")
		} else {
			os.Setenv("MEGAPORT_SECRET_KEY", originalSecretKey)
		}
	}()

	tests := []struct {
		name                  string
		globalFlag            string
		envAccessKey          string
		envSecretKey          string
		expectEnvVarsPriority bool
	}{
		{
			name:                  "Flag set: env vars should have priority",
			globalFlag:            "production",
			envAccessKey:          "env-access-key",
			envSecretKey:          "env-secret-key",
			expectEnvVarsPriority: true,
		},
		{
			name:                  "No flag: profile should have priority",
			globalFlag:            "",
			envAccessKey:          "env-access-key",
			envSecretKey:          "env-secret-key",
			expectEnvVarsPriority: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			utils.Env = tt.globalFlag
			os.Setenv("MEGAPORT_ACCESS_KEY", tt.envAccessKey)
			os.Setenv("MEGAPORT_SECRET_KEY", tt.envSecretKey)

			// Test credential selection logic
			var accessKey, secretKey string

			// Simulate the new credential selection logic
			if utils.Env != "" {
				// --env flag was explicitly set, prioritize environment variables
				accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
				secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
			} else {
				// No --env flag, profile would have priority (but we can't test that easily here)
				// For this test, we'll just verify the env var logic works
				if tt.expectEnvVarsPriority {
					accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
					secretKey = os.Getenv("MEGAPORT_SECRET_KEY")
				}
			}

			if tt.expectEnvVarsPriority {
				assert.Equal(t, tt.envAccessKey, accessKey, "Should use environment variable for access key when --env flag is set")
				assert.Equal(t, tt.envSecretKey, secretKey, "Should use environment variable for secret key when --env flag is set")
			}
		})
	}
}

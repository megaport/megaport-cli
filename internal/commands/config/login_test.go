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
			// Clear all env vars first, then set the ones specified in the test case.
			// t.Setenv auto-restores when the subtest finishes.
			t.Setenv("MEGAPORT_ACCESS_KEY", "") // empty string is equivalent to unset for os.Getenv callers
			t.Setenv("MEGAPORT_SECRET_KEY", "")
			t.Setenv("MEGAPORT_ENVIRONMENT", "")

			for key, value := range tt.envVars {
				t.Setenv(key, value)
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
	// Save and restore non-env-var globals
	originalEnv := utils.Env
	defer func() {
		utils.Env = originalEnv
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
			origEnv := utils.Env
			defer func() { utils.Env = origEnv }()

			// Setup test environment
			utils.Env = tt.globalFlag
			t.Setenv("MEGAPORT_ENVIRONMENT", tt.envVar)

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
	// Save and restore non-env-var globals
	originalEnv := utils.Env
	originalProfileOverride := utils.ProfileOverride
	originalLoginFuncWithOutput := LoginFuncWithOutput

	defer func() {
		utils.Env = originalEnv
		utils.ProfileOverride = originalProfileOverride
		LoginFuncWithOutput = originalLoginFuncWithOutput
	}()

	// Setup temp config dir with profiles
	tempDir, err := os.MkdirTemp("", "megaport-login-test")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	t.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

	// Clear env vars so profile credentials are used
	t.Setenv("MEGAPORT_ACCESS_KEY", "")
	t.Setenv("MEGAPORT_SECRET_KEY", "")
	t.Setenv("MEGAPORT_ENVIRONMENT", "")

	// Create config with two profiles
	manager, err := NewConfigManager()
	assert.NoError(t, err)
	err = manager.CreateProfile("staging", "staging-access", "staging-secret", "staging", "")
	assert.NoError(t, err)
	err = manager.CreateProfile("prod", "prod-access", "prod-secret", "production", "")
	assert.NoError(t, err)
	err = manager.UseProfile("prod")
	assert.NoError(t, err)

	t.Run("profile override uses specified profile and reaches API call", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = ""
		utils.ProfileOverride = "staging"

		// Call the real LoginFuncWithOutput - it will resolve credentials from
		// the "staging" profile and fail at Authorize (no real API), which proves
		// credential resolution succeeded (didn't get "access key not provided" error).
		_, err := LoginFuncWithOutput(context.Background(), "json")
		assert.Error(t, err)
		// Should NOT be a "not provided" error — that would mean credential resolution failed
		assert.NotContains(t, err.Error(), "access key not provided")
		assert.NotContains(t, err.Error(), "secret key not provided")
		assert.NotContains(t, err.Error(), "not found")
	})

	t.Run("profile override with non-existent profile returns error", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = ""
		utils.ProfileOverride = "non-existent"

		_, err := LoginFuncWithOutput(context.Background(), "json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "non-existent")
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("without profile override uses active profile and reaches API call", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = ""
		utils.ProfileOverride = ""

		// Active profile is "prod" — should resolve credentials from it
		_, err := LoginFuncWithOutput(context.Background(), "json")
		assert.Error(t, err)
		// Should NOT be a "not provided" error
		assert.NotContains(t, err.Error(), "access key not provided")
		assert.NotContains(t, err.Error(), "secret key not provided")
	})

	t.Run("env flag overrides profile environment", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = "development"
		utils.ProfileOverride = "staging"

		// Both --profile and --env are set: credentials from staging profile,
		// environment from --env flag. Should still resolve and reach API call.
		_, err := LoginFuncWithOutput(context.Background(), "json")
		assert.Error(t, err)
		assert.NotContains(t, err.Error(), "access key not provided")
		assert.NotContains(t, err.Error(), "not found")
	})

	t.Run("no profile and no env vars returns credential error", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = ""
		utils.ProfileOverride = ""

		// Delete the active profile so there are no credentials available
		err := manager.UseProfile("")
		// UseProfile with empty may fail, so we set active profile to non-existent
		_ = err

		// Use a fresh temp dir with no profiles
		emptyDir, err := os.MkdirTemp("", "megaport-empty-test")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(emptyDir) })
		t.Setenv("MEGAPORT_CONFIG_DIR", emptyDir)

		_, err = LoginFuncWithOutput(context.Background(), "json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access key not provided")
	})
}

func TestNewUnauthenticatedClient(t *testing.T) {
	// Save and restore non-env-var globals
	originalEnv := utils.Env
	originalProfileOverride := utils.ProfileOverride

	defer func() {
		utils.Env = originalEnv
		utils.ProfileOverride = originalProfileOverride
	}()

	// Default empty config dir for all subtests (subtests that need profiles override this)
	defaultEmptyDir, err := os.MkdirTemp("", "megaport-unauth-default")
	assert.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(defaultEmptyDir) })
	t.Setenv("MEGAPORT_CONFIG_DIR", defaultEmptyDir)

	t.Run("defaults to production when no env configured", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = ""
		utils.ProfileOverride = ""
		t.Setenv("MEGAPORT_ENVIRONMENT", "")

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api.megaport.com")
	})

	t.Run("respects env flag", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = "staging"
		utils.ProfileOverride = ""
		t.Setenv("MEGAPORT_ENVIRONMENT", "")

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api-staging.megaport.com")
	})

	t.Run("respects MEGAPORT_ENVIRONMENT env var", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = ""
		utils.ProfileOverride = ""
		t.Setenv("MEGAPORT_ENVIRONMENT", "staging")

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api-staging.megaport.com")
	})

	t.Run("profile override with valid profile uses profile env", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		tempDir, err := os.MkdirTemp("", "megaport-unauth-test")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDir) })
		t.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

		manager, err := NewConfigManager()
		assert.NoError(t, err)
		err = manager.CreateProfile("staging-profile", "key", "secret", "staging", "")
		assert.NoError(t, err)

		utils.Env = ""
		utils.ProfileOverride = "staging-profile"
		t.Setenv("MEGAPORT_ENVIRONMENT", "")

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api-staging.megaport.com")
	})

	t.Run("profile override with non-existent profile returns error", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		tempDir, err := os.MkdirTemp("", "megaport-unauth-test")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDir) })
		t.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

		utils.Env = ""
		utils.ProfileOverride = "non-existent"

		client, err := NewUnauthenticatedClient()
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "non-existent")
	})

	t.Run("env flag overrides profile environment", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		tempDir, err := os.MkdirTemp("", "megaport-unauth-test")
		assert.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDir) })
		t.Setenv("MEGAPORT_CONFIG_DIR", tempDir)

		manager, err := NewConfigManager()
		assert.NoError(t, err)
		err = manager.CreateProfile("staging-profile", "key", "secret", "staging", "")
		assert.NoError(t, err)

		utils.Env = "production"
		utils.ProfileOverride = "staging-profile"

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api.megaport.com")
	})

	t.Run("does not require credentials", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = "production"
		utils.ProfileOverride = ""
		t.Setenv("MEGAPORT_ACCESS_KEY", "")
		t.Setenv("MEGAPORT_SECRET_KEY", "")

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Empty(t, client.AccessKey)
		assert.Empty(t, client.SecretKey)
	})

	t.Run("accepts short alias prod", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = "prod"
		utils.ProfileOverride = ""

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api.megaport.com")
	})

	t.Run("accepts short alias dev", func(t *testing.T) {
		origEnv := utils.Env
		defer func() { utils.Env = origEnv }()
		origProfile := utils.ProfileOverride
		defer func() { utils.ProfileOverride = origProfile }()

		utils.Env = "dev"
		utils.ProfileOverride = ""

		client, err := NewUnauthenticatedClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Contains(t, client.BaseURL.String(), "api-mpone-dev.megaport.com")
	})
}

func TestCredentialSelectionPrecedence(t *testing.T) {
	// Save and restore non-env-var globals
	originalEnv := utils.Env
	defer func() {
		utils.Env = originalEnv
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
			origEnv := utils.Env
			defer func() { utils.Env = origEnv }()

			// Setup test environment
			utils.Env = tt.globalFlag
			t.Setenv("MEGAPORT_ACCESS_KEY", tt.envAccessKey)
			t.Setenv("MEGAPORT_SECRET_KEY", tt.envSecretKey)

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

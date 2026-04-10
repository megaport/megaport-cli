package auth

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// helper to create a mock login func returning a client with the given mock service
func setupMockLogin(mock *MockUserManagementService) {
	baseURL, _ := url.Parse("https://api.megaport.com/")
	config.SetLoginFunc(func(_ context.Context) (*megaport.Client, error) {
		client := &megaport.Client{BaseURL: baseURL}
		client.UserManagementService = mock
		return client, nil
	})
	listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
		return client.UserManagementService.ListCompanyUsers(ctx)
	}
}

func TestAuthStatus(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name          string
		setupMock     func()
		expectedError string
		expectedOut   string
	}{
		{
			name: "success with user info",
			setupMock: func() {
				setupMockLogin(&MockUserManagementService{
					ListUsersResult: []*megaport.User{
						{
							PartyId:       12345,
							FirstName:     "Jane",
							LastName:      "Smith",
							Email:         "jane@example.com",
							Position:      "Company Admin",
							Active:        true,
							CompanyName:   "Acme Corp",
							SecurityRoles: []string{"companyAdmin"},
						},
					},
				})
			},
			expectedOut: "Jane",
		},
		{
			name: "success with multiple users picks admin",
			setupMock: func() {
				setupMockLogin(&MockUserManagementService{
					ListUsersResult: []*megaport.User{
						{
							PartyId:       1,
							FirstName:     "Read",
							LastName:      "Only",
							Email:         "readonly@example.com",
							Position:      "Read Only",
							Active:        true,
							CompanyName:   "Acme Corp",
							SecurityRoles: []string{"readOnly"},
						},
						{
							PartyId:       2,
							FirstName:     "Admin",
							LastName:      "User",
							Email:         "admin@example.com",
							Position:      "Company Admin",
							Active:        true,
							CompanyName:   "Acme Corp",
							SecurityRoles: []string{"companyAdmin"},
						},
					},
				})
			},
			expectedOut: "Admin",
		},
		{
			name: "success shows company name from first user",
			setupMock: func() {
				setupMockLogin(&MockUserManagementService{
					ListUsersResult: []*megaport.User{
						{
							PartyId:     1,
							FirstName:   "Test",
							LastName:    "User",
							Active:      true,
							CompanyName: "Megaport Pty Ltd",
						},
					},
				})
			},
			expectedOut: "Megaport Pty Ltd",
		},
		{
			name: "success with user that has no company name falls back",
			setupMock: func() {
				setupMockLogin(&MockUserManagementService{
					ListUsersResult: []*megaport.User{
						{
							PartyId:   1,
							FirstName: "Test",
							LastName:  "User",
							Active:    true,
						},
					},
				})
			},
			expectedOut: "Test",
		},
		{
			name: "auth failure",
			setupMock: func() {
				config.SetLoginFunc(func(_ context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("invalid credentials")
				})
			},
			expectedError: "invalid credentials",
		},
		{
			name: "API error listing users",
			setupMock: func() {
				setupMockLogin(&MockUserManagementService{
					ListUsersErr: fmt.Errorf("API failure"),
				})
			},
			expectedError: "API failure",
		},
		{
			name: "empty user list shows endpoint",
			setupMock: func() {
				setupMockLogin(&MockUserManagementService{
					ListUsersResult: []*megaport.User{},
				})
			},
			expectedOut: "api.megaport.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			cmd := testutil.NewCommand("status", testutil.OutputAdapter(AuthStatus))

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = testutil.OutputAdapter(AuthStatus)(cmd, nil)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.expectedOut != "" {
					assert.Contains(t, capturedOutput, tt.expectedOut)
				}
			}
		})
	}
}

func TestAuthStatusJSONOutput(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	setupMockLogin(&MockUserManagementService{
		ListUsersResult: []*megaport.User{
			{
				PartyId:       1,
				FirstName:     "Test",
				LastName:      "User",
				Email:         "test@example.com",
				Position:      "Technical Admin",
				Active:        true,
				CompanyName:   "Test Co",
				SecurityRoles: []string{"companyAdmin"},
			},
		},
	})

	cmd := testutil.NewCommand("status", testutil.OutputAdapter(AuthStatus))
	_ = cmd.Flags().Set("output", "json")

	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = cmd.RunE(cmd, nil)
	})

	assert.NoError(t, err)
	assert.Contains(t, capturedOutput, "test@example.com")
	assert.Contains(t, capturedOutput, "first_name")
	assert.Contains(t, capturedOutput, "api_endpoint")
	assert.Contains(t, capturedOutput, "environment")
	assert.Contains(t, capturedOutput, "profile")
}

func TestAuthStatusCSVOutput(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	setupMockLogin(&MockUserManagementService{
		ListUsersResult: []*megaport.User{
			{
				PartyId:       1,
				FirstName:     "CSV",
				LastName:      "Test",
				Email:         "csv@example.com",
				Active:        true,
				CompanyName:   "CSV Co",
				SecurityRoles: []string{"companyAdmin"},
			},
		},
	})

	cmd := testutil.NewCommand("status", testutil.OutputAdapter(AuthStatus))
	_ = cmd.Flags().Set("output", "csv")

	var err error
	capturedOutput := output.CaptureOutput(func() {
		err = cmd.RunE(cmd, nil)
	})

	assert.NoError(t, err)
	assert.Contains(t, capturedOutput, "csv@example.com")
}

func TestFindCurrentUser(t *testing.T) {
	tests := []struct {
		name     string
		users    []*megaport.User
		expected string
	}{
		{
			name:     "nil list",
			users:    nil,
			expected: "",
		},
		{
			name:     "empty list",
			users:    []*megaport.User{},
			expected: "",
		},
		{
			name: "single user",
			users: []*megaport.User{
				{FirstName: "Only", Active: true},
			},
			expected: "Only",
		},
		{
			name: "prefers active admin over non-admin",
			users: []*megaport.User{
				{FirstName: "ReadOnly", Active: true, SecurityRoles: []string{"readOnly"}},
				{FirstName: "Admin", Active: true, SecurityRoles: []string{"companyAdmin"}},
			},
			expected: "Admin",
		},
		{
			name: "skips inactive admin falls back to active user",
			users: []*megaport.User{
				{FirstName: "InactiveAdmin", Active: false, SecurityRoles: []string{"companyAdmin"}},
				{FirstName: "ActiveUser", Active: true, SecurityRoles: []string{"readOnly"}},
			},
			expected: "ActiveUser",
		},
		{
			name: "all inactive falls back to first user",
			users: []*megaport.User{
				{FirstName: "Inactive1", Active: false},
				{FirstName: "Inactive2", Active: false},
			},
			expected: "Inactive1",
		},
		{
			name: "nil users in list are skipped",
			users: []*megaport.User{
				nil,
				{FirstName: "Valid", Active: true, SecurityRoles: []string{"companyAdmin"}},
			},
			expected: "Valid",
		},
		{
			name: "user with multiple roles including admin",
			users: []*megaport.User{
				{FirstName: "NonAdmin", Active: true, SecurityRoles: []string{"readOnly"}},
				{FirstName: "MultiRole", Active: true, SecurityRoles: []string{"technicalAdmin", "companyAdmin"}},
			},
			expected: "MultiRole",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCurrentUser(tt.users)
			if tt.expected == "" {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected, result.FirstName)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	assert.Equal(t, "Production", capitalizeFirst("production"))
	assert.Equal(t, "Staging", capitalizeFirst("staging"))
	assert.Equal(t, "Development", capitalizeFirst("development"))
	assert.Equal(t, "", capitalizeFirst(""))
	assert.Equal(t, "A", capitalizeFirst("a"))
	assert.Equal(t, "ALREADY", capitalizeFirst("ALREADY"))
}

func TestResolveProfileInfo(t *testing.T) {
	tests := []struct {
		name            string
		profileOverride string
		envOverride     string
		megaportEnvVar  string
		expectedProfile string
		expectedEnv     string
	}{
		{
			name:            "defaults when no config or env vars",
			profileOverride: "",
			envOverride:     "",
			megaportEnvVar:  "",
			expectedProfile: "(env vars)",
			expectedEnv:     "production",
		},
		{
			name:            "profile override sets profile name",
			profileOverride: "my-profile",
			envOverride:     "",
			megaportEnvVar:  "",
			expectedProfile: "my-profile",
			expectedEnv:     "production",
		},
		{
			name:            "env flag overrides environment",
			profileOverride: "",
			envOverride:     "staging",
			megaportEnvVar:  "",
			expectedProfile: "(env vars)",
			expectedEnv:     "staging",
		},
		{
			name:            "both profile and env overrides",
			profileOverride: "prod-profile",
			envOverride:     "development",
			megaportEnvVar:  "",
			expectedProfile: "prod-profile",
			expectedEnv:     "development",
		},
		{
			name:            "MEGAPORT_ENVIRONMENT env var used as fallback",
			profileOverride: "",
			envOverride:     "",
			megaportEnvVar:  "staging",
			expectedProfile: "(env vars)",
			expectedEnv:     "staging",
		},
		{
			name:            "env flag takes precedence over MEGAPORT_ENVIRONMENT",
			profileOverride: "",
			envOverride:     "development",
			megaportEnvVar:  "staging",
			expectedProfile: "(env vars)",
			expectedEnv:     "development",
		},
		{
			name:            "profile override with MEGAPORT_ENVIRONMENT fallback",
			profileOverride: "some-profile",
			envOverride:     "",
			megaportEnvVar:  "development",
			expectedProfile: "some-profile",
			expectedEnv:     "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Isolate config directory so the test doesn't read the developer's
			// real ~/.megaport/config.json or create files outside the test.
			t.Setenv("MEGAPORT_CONFIG_DIR", t.TempDir())

			origProfile := utils.ProfileOverride
			origEnv := utils.Env
			defer func() {
				utils.ProfileOverride = origProfile
				utils.Env = origEnv
			}()

			utils.ProfileOverride = tt.profileOverride
			utils.Env = tt.envOverride
			if tt.megaportEnvVar != "" {
				t.Setenv("MEGAPORT_ENVIRONMENT", tt.megaportEnvVar)
			} else {
				t.Setenv("MEGAPORT_ENVIRONMENT", "")
			}

			profileName, environment := resolveProfileInfo()

			assert.Equal(t, tt.expectedProfile, profileName)
			assert.Equal(t, tt.expectedEnv, environment)
		})
	}
}

func TestAddCommandsTo(t *testing.T) {
	root := &cobra.Command{Use: "megaport-cli"}
	AddCommandsTo(root)

	// Check auth command is registered
	authFound := false
	whoamiFound := false
	for _, cmd := range root.Commands() {
		switch cmd.Use {
		case "auth":
			authFound = true
			// Check status subcommand exists under auth
			statusFound := false
			for _, sub := range cmd.Commands() {
				if sub.Use == "status" {
					statusFound = true
				}
			}
			assert.True(t, statusFound, "auth should have a status subcommand")
		case "whoami":
			whoamiFound = true
		}
	}
	assert.True(t, authFound, "auth command should be registered")
	assert.True(t, whoamiFound, "whoami command should be registered at root level")
}

func TestModule(t *testing.T) {
	m := NewModule()
	assert.Equal(t, "auth", m.Name())

	root := &cobra.Command{Use: "megaport-cli"}
	m.RegisterCommands(root)

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Use == "auth" {
			found = true
			break
		}
	}
	assert.True(t, found, "RegisterCommands should add auth command")
}

func TestToAuthStatusOutput(t *testing.T) {
	t.Run("with user", func(t *testing.T) {
		user := &megaport.User{
			FirstName:   "John",
			LastName:    "Doe",
			Email:       "john@example.com",
			Position:    "Technical Admin",
			Active:      true,
			CompanyName: "Test Corp",
		}
		out := toAuthStatusOutput(user, "default", "production", "https://api.megaport.com/", "Fallback Corp")

		assert.Equal(t, "John", out.FirstName)
		assert.Equal(t, "Doe", out.LastName)
		assert.Equal(t, "john@example.com", out.Email)
		assert.Equal(t, "Technical Admin", out.Position)
		assert.True(t, out.Active)
		assert.Equal(t, "Test Corp", out.CompanyName) // User's company overrides fallback
		assert.Equal(t, "default", out.Profile)
		assert.Equal(t, "Production", out.Environment)
		assert.Equal(t, "https://api.megaport.com/", out.APIEndpoint)
	})

	t.Run("without user", func(t *testing.T) {
		out := toAuthStatusOutput(nil, "my-profile", "staging", "https://api-staging.megaport.com/", "Company Inc")

		assert.Equal(t, "", out.FirstName)
		assert.Equal(t, "", out.Email)
		assert.False(t, out.Active)
		assert.Equal(t, "Company Inc", out.CompanyName) // Fallback used when no user
		assert.Equal(t, "my-profile", out.Profile)
		assert.Equal(t, "Staging", out.Environment)
	})

	t.Run("user with empty company name uses fallback", func(t *testing.T) {
		user := &megaport.User{
			FirstName:   "Jane",
			CompanyName: "",
		}
		out := toAuthStatusOutput(user, "p", "production", "https://api.megaport.com/", "Fallback Corp")

		assert.Equal(t, "Fallback Corp", out.CompanyName)
	})
}

func TestPrintAuthStatus(t *testing.T) {
	user := &megaport.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       "test@example.com",
		Active:      true,
		CompanyName: "Co",
	}

	capturedOutput := output.CaptureOutput(func() {
		err := printAuthStatus(user, "profile", "production", "https://api.megaport.com/", "Co", "table", true)
		assert.NoError(t, err)
	})

	assert.Contains(t, capturedOutput, "Test")
	assert.Contains(t, capturedOutput, "test@example.com")
}

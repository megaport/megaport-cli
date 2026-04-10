package auth

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

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
				baseURL, _ := url.Parse("https://api.megaport.com/")
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{BaseURL: baseURL}
					client.UserManagementService = &MockUserManagementService{
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
					}
					return client, nil
				})
				listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
					return client.UserManagementService.ListCompanyUsers(ctx)
				}
			},
			expectedOut: "Jane",
		},
		{
			name: "success with multiple users",
			setupMock: func() {
				baseURL, _ := url.Parse("https://api.megaport.com/")
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{BaseURL: baseURL}
					client.UserManagementService = &MockUserManagementService{
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
					}
					return client, nil
				})
				listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
					return client.UserManagementService.ListCompanyUsers(ctx)
				}
			},
			expectedOut: "Admin",
		},
		{
			name: "auth failure",
			setupMock: func() {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("invalid credentials")
				})
			},
			expectedError: "invalid credentials",
		},
		{
			name: "API error listing users",
			setupMock: func() {
				baseURL, _ := url.Parse("https://api.megaport.com/")
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{BaseURL: baseURL}
					client.UserManagementService = &MockUserManagementService{
						ListUsersErr: fmt.Errorf("API failure"),
					}
					return client, nil
				})
				listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
					return client.UserManagementService.ListCompanyUsers(ctx)
				}
			},
			expectedError: "API failure",
		},
		{
			name: "empty user list",
			setupMock: func() {
				baseURL, _ := url.Parse("https://api.megaport.com/")
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{BaseURL: baseURL}
					client.UserManagementService = &MockUserManagementService{
						ListUsersResult: []*megaport.User{},
					}
					return client, nil
				})
				listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
					return client.UserManagementService.ListCompanyUsers(ctx)
				}
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

	baseURL, _ := url.Parse("https://api.megaport.com/")
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{BaseURL: baseURL}
		client.UserManagementService = &MockUserManagementService{
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
		}
		return client, nil
	})
	listCompanyUsersFunc = func(ctx context.Context, client *megaport.Client) ([]*megaport.User, error) {
		return client.UserManagementService.ListCompanyUsers(ctx)
	}

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
			name: "prefers admin",
			users: []*megaport.User{
				{FirstName: "ReadOnly", Active: true, SecurityRoles: []string{"readOnly"}},
				{FirstName: "Admin", Active: true, SecurityRoles: []string{"companyAdmin"}},
			},
			expected: "Admin",
		},
		{
			name: "skips inactive admin",
			users: []*megaport.User{
				{FirstName: "InactiveAdmin", Active: false, SecurityRoles: []string{"companyAdmin"}},
				{FirstName: "ActiveUser", Active: true, SecurityRoles: []string{"readOnly"}},
			},
			expected: "ActiveUser",
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
	assert.Equal(t, "", capitalizeFirst(""))
	assert.Equal(t, "A", capitalizeFirst("a"))
}

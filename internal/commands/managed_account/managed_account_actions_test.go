package managed_account

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListManagedAccounts(t *testing.T) {
	output.SetTerminalWidthForTesting(200)
	defer output.SetTerminalWidthForTesting(0)
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	testAccounts := []*megaport.ManagedAccount{
		{
			AccountName: "Acme Corp",
			AccountRef:  "REF-001",
			CompanyUID:  "company-uid-1",
		},
		{
			AccountName: "Beta Inc",
			AccountRef:  "REF-002",
			CompanyUID:  "company-uid-2",
		},
		{
			AccountName: "Acme Subsidiary",
			AccountRef:  "ACME-SUB",
			CompanyUID:  "company-uid-3",
		},
	}

	tests := []struct {
		name             string
		flags            map[string]string
		setupMock        func(*MockManagedAccountService)
		expectedError    string
		expectedAccounts []string
		unexpectedAccts  []string
		outputFormat     string
	}{
		{
			name: "list all accounts",
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp", "Beta Inc", "Acme Subsidiary"},
			outputFormat:     "table",
		},
		{
			name: "filter by account name",
			flags: map[string]string{
				"account-name": "Acme",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp", "Acme Subsidiary"},
			unexpectedAccts:  []string{"Beta Inc"},
			outputFormat:     "table",
		},
		{
			name: "filter by account ref",
			flags: map[string]string{
				"account-ref": "REF",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp", "Beta Inc"},
			unexpectedAccts:  []string{"Acme Subsidiary"},
			outputFormat:     "table",
		},
		{
			name: "combined filters",
			flags: map[string]string{
				"account-name": "Acme",
				"account-ref":  "REF",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp"},
			unexpectedAccts:  []string{"Beta Inc", "Acme Subsidiary"},
			outputFormat:     "table",
		},
		{
			name: "JSON output format",
			flags: map[string]string{
				"account-name": "Acme Corp",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp"},
			unexpectedAccts:  []string{"Beta Inc"},
			outputFormat:     "json",
		},
		{
			name: "CSV output format",
			flags: map[string]string{
				"account-name": "Acme Corp",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp"},
			unexpectedAccts:  []string{"Beta Inc"},
			outputFormat:     "csv",
		},
		{
			name: "empty result",
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{}
			},
			expectedAccounts: []string{},
			outputFormat:     "table",
		},
		{
			name: "nil API result",
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = nil
			},
			expectedAccounts: []string{},
			outputFormat:     "table",
		},
		{
			name: "API error",
			setupMock: func(m *MockManagedAccountService) {
				m.listErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
			outputFormat:  "table",
		},
		{
			name: "login error",
			setupMock: func(m *MockManagedAccountService) {
				// Will be handled by overriding LoginFunc to return error
			},
			expectedError: "login failed",
			outputFormat:  "table",
		},
		{
			name: "limit results",
			flags: map[string]string{
				"limit": "2",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp", "Beta Inc"},
			unexpectedAccts:  []string{"Acme Subsidiary"},
			outputFormat:     "table",
		},
		{
			name: "negative limit returns error",
			flags: map[string]string{
				"limit": "-1",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedError: "--limit must be a non-negative integer",
			outputFormat:  "table",
		},
		{
			name: "invalid output format",
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedError: "invalid output format",
			outputFormat:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.name == "login error" {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("login failed")
				})
			} else {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.ManagedAccountService = mockService
					return client, nil
				})
			}

			cmd := testutil.NewCommand("list", func(cmd *cobra.Command, args []string) error {
				return ListManagedAccounts(cmd, args, true, tt.outputFormat)
			})

			cmd.Flags().String("account-name", "", "Filter by account name")
			cmd.Flags().String("account-ref", "", "Filter by account ref")
			cmd.Flags().Int("limit", 0, "Maximum number of results to display")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			var stderrOutput string
			stdoutOutput := output.CaptureOutput(func() {
				stderrOutput = captureStderr(t, func() {
					err = cmd.RunE(cmd, []string{})
				})
			})
			capturedOutput := stdoutOutput + stderrOutput

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)

				for _, expected := range tt.expectedAccounts {
					assert.Contains(t, capturedOutput, expected,
						"Expected account '%s' should be in output", expected)
				}

				for _, unexpected := range tt.unexpectedAccts {
					assert.NotContains(t, capturedOutput, unexpected,
						"Unexpected account '%s' should NOT be in output", unexpected)
				}

				switch tt.outputFormat {
				case "json":
					if len(tt.expectedAccounts) > 0 {
						assert.Contains(t, capturedOutput, "\"account_name\":")
						assert.Contains(t, capturedOutput, "\"company_uid\":")
					}
				case "table":
					if len(tt.expectedAccounts) > 0 {
						assert.Contains(t, capturedOutput, "ACCOUNT NAME")
						assert.Contains(t, capturedOutput, "COMPANY UID")
					}
				}

				if len(tt.expectedAccounts) == 0 && tt.expectedError == "" {
					assert.Contains(t, capturedOutput, "No managed accounts found.")
				}
			}
		})
	}
}

func TestGetManagedAccount(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalGetFunc := getManagedAccountFunc
	defer func() {
		getManagedAccountFunc = originalGetFunc
	}()

	tests := []struct {
		name          string
		companyUID    string
		accountName   string
		setupMock     func(*MockManagedAccountService)
		expectedError string
		outputFormat  string
		expectedOut   []string
	}{
		{
			name:        "success table format",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *MockManagedAccountService) {
				m.getResult = &megaport.ManagedAccount{
					AccountName: "Acme Corp",
					AccountRef:  "REF-001",
					CompanyUID:  "company-uid-1",
				}
			},
			outputFormat: "table",
			expectedOut:  []string{"Acme Corp", "REF-001", "company-uid-1"},
		},
		{
			name:        "success JSON format",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *MockManagedAccountService) {
				m.getResult = &megaport.ManagedAccount{
					AccountName: "Acme Corp",
					AccountRef:  "REF-001",
					CompanyUID:  "company-uid-1",
				}
			},
			outputFormat: "json",
			expectedOut:  []string{"Acme Corp", "REF-001", "company-uid-1"},
		},
		{
			name:        "success CSV format",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *MockManagedAccountService) {
				m.getResult = &megaport.ManagedAccount{
					AccountName: "Acme Corp",
					AccountRef:  "REF-001",
					CompanyUID:  "company-uid-1",
				}
			},
			outputFormat: "csv",
			expectedOut:  []string{"Acme Corp", "REF-001", "company-uid-1"},
		},
		{
			name:        "nil result without error",
			companyUID:  "company-uid-1",
			accountName: "Nonexistent",
			setupMock: func(m *MockManagedAccountService) {
				m.getResult = nil
			},
			expectedError: "invalid managed account: nil value",
			outputFormat:  "table",
		},
		{
			name:        "invalid output format",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *MockManagedAccountService) {
				m.getResult = &megaport.ManagedAccount{
					AccountName: "Acme Corp",
					AccountRef:  "REF-001",
					CompanyUID:  "company-uid-1",
				}
			},
			expectedError: "invalid output format",
			outputFormat:  "invalid",
		},
		{
			name:        "API error",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *MockManagedAccountService) {
				m.getErr = fmt.Errorf("managed account not found")
			},
			expectedError: "managed account not found",
			outputFormat:  "table",
		},
		{
			name:        "login error",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *MockManagedAccountService) {
				// Will be handled by overriding LoginFunc to return error
			},
			expectedError: "login failed",
			outputFormat:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.name == "login error" {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("login failed")
				})
			} else {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.ManagedAccountService = mockService
					return client, nil
				})
			}

			getManagedAccountFunc = func(ctx context.Context, client *megaport.Client, companyUID string, name string) (*megaport.ManagedAccount, error) {
				return mockService.GetManagedAccount(ctx, companyUID, name)
			}

			cmd := testutil.NewCommand("get", func(cmd *cobra.Command, args []string) error {
				return GetManagedAccount(cmd, args, true, tt.outputFormat)
			})

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.companyUID, tt.accountName})
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				for _, expected := range tt.expectedOut {
					assert.Contains(t, capturedOutput, expected)
				}
			}

			// Verify captured args on the mock for successful calls
			if tt.expectedError == "" || (tt.name != "login error" && tt.name != "API error") {
				if mockService.capturedGetCompanyUID != "" {
					assert.Equal(t, tt.companyUID, mockService.capturedGetCompanyUID)
					assert.Equal(t, tt.accountName, mockService.capturedGetAccountName)
				}
			}
		})
	}
}

func TestCreateManagedAccount(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalCreateFunc := createManagedAccountFunc
	originalPrompt := utils.GetResourcePrompt()
	defer func() {
		createManagedAccountFunc = originalCreateFunc
		utils.SetResourcePrompt(originalPrompt)
	}()

	tests := []struct {
		name             string
		flags            map[string]string
		setupMock        func(*MockManagedAccountService)
		prompts          []string
		expectedError    string
		expectedOut      []string
		validateCaptured func(t *testing.T, m *MockManagedAccountService)
	}{
		{
			name: "flag mode",
			flags: map[string]string{
				"account-name": "Test Account",
				"account-ref":  "REF-001",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.createResult = &megaport.ManagedAccount{
					AccountName: "Test Account",
					AccountRef:  "REF-001",
					CompanyUID:  "new-company-uid",
				}
			},
			expectedOut: []string{"new-company-uid"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.NotNil(t, m.capturedCreateReq)
				assert.Equal(t, "Test Account", m.capturedCreateReq.AccountName)
				assert.Equal(t, "REF-001", m.capturedCreateReq.AccountRef)
			},
		},
		{
			name: "interactive mode",
			flags: map[string]string{
				"interactive": "true",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.createResult = &megaport.ManagedAccount{
					AccountName: "Interactive Account",
					AccountRef:  "INT-REF",
					CompanyUID:  "interactive-uid",
				}
			},
			prompts:     []string{"Interactive Account", "INT-REF"},
			expectedOut: []string{"interactive-uid"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.NotNil(t, m.capturedCreateReq)
				assert.Equal(t, "Interactive Account", m.capturedCreateReq.AccountName)
				assert.Equal(t, "INT-REF", m.capturedCreateReq.AccountRef)
			},
		},
		{
			name: "JSON string mode",
			flags: map[string]string{
				"json": `{"accountName":"JSON Account","accountRef":"JSON-REF"}`,
			},
			setupMock: func(m *MockManagedAccountService) {
				m.createResult = &megaport.ManagedAccount{
					AccountName: "JSON Account",
					AccountRef:  "JSON-REF",
					CompanyUID:  "json-uid",
				}
			},
			expectedOut: []string{"json-uid"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.NotNil(t, m.capturedCreateReq)
				assert.Equal(t, "JSON Account", m.capturedCreateReq.AccountName)
				assert.Equal(t, "JSON-REF", m.capturedCreateReq.AccountRef)
			},
		},
		{
			name: "API error",
			flags: map[string]string{
				"account-name": "Test Account",
				"account-ref":  "REF-001",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.createErr = fmt.Errorf("API error: creation failed")
			},
			expectedError: "API error: creation failed",
		},
		{
			name:          "no input",
			flags:         map[string]string{},
			setupMock:     func(m *MockManagedAccountService) {},
			expectedError: "no input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ManagedAccountService = mockService
				return client, nil
			})

			createManagedAccountFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
				return mockService.CreateManagedAccount(ctx, req)
			}

			if tt.prompts != nil {
				promptIndex := 0
				utils.SetResourcePrompt(func(_, msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				})
			}

			cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateManagedAccount))

			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			var stderrOutput string
			stdoutOutput := output.CaptureOutput(func() {
				stderrOutput = captureStderr(t, func() {
					err = cmd.RunE(cmd, []string{})
				})
			})
			capturedOutput := stdoutOutput + stderrOutput

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				for _, expected := range tt.expectedOut {
					assert.Contains(t, capturedOutput, expected)
				}
			}

			if tt.validateCaptured != nil {
				tt.validateCaptured(t, mockService)
			}
		})
	}
}

func TestCreateManagedAccount_NilResponse(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalCreateFunc := createManagedAccountFunc
	defer func() { createManagedAccountFunc = originalCreateFunc }()

	// createResult left nil so the mock returns (nil, nil).
	mockService := &MockManagedAccountService{}
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ManagedAccountService = mockService
		return client, nil
	})
	createManagedAccountFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
		return mockService.CreateManagedAccount(ctx, req)
	}

	cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateManagedAccount))
	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	testutil.SetFlags(t, cmd, map[string]string{
		"account-name": "Test Account",
		"account-ref":  "REF-001",
	})

	var err error
	require.NotPanics(t, func() {
		output.CaptureOutput(func() {
			err = cmd.RunE(cmd, []string{})
		})
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty response")
}

func TestCreateManagedAccount_InvalidJSON(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ManagedAccountService = &MockManagedAccountService{}
	})
	defer cleanup()
	originalCreateFunc := createManagedAccountFunc
	defer func() {
		createManagedAccountFunc = originalCreateFunc
	}()

	cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateManagedAccount))

	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")

	testutil.SetFlags(t, cmd, map[string]string{"json": "{invalid json}"})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")

	var cliErr *exitcodes.CLIError
	require.True(t, errors.As(err, &cliErr))
	assert.Equal(t, exitcodes.Usage, cliErr.Code)
}

func TestCreateManagedAccount_LoginError(t *testing.T) {
	cleanup := testutil.SetupLoginError(fmt.Errorf("login failed: invalid credentials"))
	defer cleanup()

	cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateManagedAccount))

	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")

	testutil.SetFlags(t, cmd, map[string]string{
		"account-name": "Test Account",
		"account-ref":  "REF-001",
	})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

func TestUpdateManagedAccount(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()
	originalUpdateFunc := updateManagedAccountFunc
	originalPrompt := utils.GetResourcePrompt()
	defer func() {
		updateManagedAccountFunc = originalUpdateFunc
		utils.SetResourcePrompt(originalPrompt)
	}()

	tests := []struct {
		name             string
		companyUID       string
		flags            map[string]string
		setupMock        func(*MockManagedAccountService)
		prompts          []string
		expectedError    string
		expectedOut      []string
		validateCaptured func(t *testing.T, m *MockManagedAccountService)
	}{
		{
			name:       "flag mode - update name",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "Updated Name",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Old Name",
						AccountRef:  "REF-001",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "Updated Name",
					AccountRef:  "REF-001",
					CompanyUID:  "company-uid-1",
				}
			},
			expectedOut: []string{"Account Name:", "Old Name", "Updated Name"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "Updated Name", m.capturedUpdateReq.AccountName)
				// --account-name alone must not blank the existing ref.
				assert.Equal(t, "REF-001", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "flag mode - update both fields",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
				"account-ref":  "NEW-REF",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Old Name",
						AccountRef:  "OLD-REF",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "New Name",
					AccountRef:  "NEW-REF",
					CompanyUID:  "company-uid-1",
				}
			},
			expectedOut: []string{"Account Name:", "Account Ref:"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "New Name", m.capturedUpdateReq.AccountName)
				assert.Equal(t, "NEW-REF", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "flag mode - update ref only",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-ref": "UPDATED-REF",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Same Name",
						AccountRef:  "OLD-REF",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "Same Name",
					AccountRef:  "UPDATED-REF",
					CompanyUID:  "company-uid-1",
				}
			},
			expectedOut: []string{"Account Ref:", "OLD-REF", "UPDATED-REF"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "UPDATED-REF", m.capturedUpdateReq.AccountRef)
				// --account-ref alone must not blank the existing name.
				assert.Equal(t, "Same Name", m.capturedUpdateReq.AccountName)
			},
		},
		{
			name:       "interactive mode",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"interactive": "true",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Old Name",
						AccountRef:  "OLD-REF",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "Interactive Name",
					AccountRef:  "OLD-REF",
					CompanyUID:  "company-uid-1",
				}
			},
			prompts:     []string{"Interactive Name", ""},
			expectedOut: []string{"Account Name:", "Old Name", "Interactive Name"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "Interactive Name", m.capturedUpdateReq.AccountName)
				// Empty ref prompt must not blank the existing ref.
				assert.Equal(t, "OLD-REF", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "JSON string mode",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"json": `{"accountName":"JSON Name","accountRef":"JSON-REF"}`,
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Old Name",
						AccountRef:  "OLD-REF",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "JSON Name",
					AccountRef:  "JSON-REF",
					CompanyUID:  "company-uid-1",
				}
			},
			expectedOut: []string{"Account Name:", "Account Ref:"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "JSON Name", m.capturedUpdateReq.AccountName)
				assert.Equal(t, "JSON-REF", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "JSON partial mode preserves omitted ref",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"json": `{"accountName":"JSON Name"}`,
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Old Name",
						AccountRef:  "OLD-REF",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "JSON Name",
					AccountRef:  "OLD-REF",
					CompanyUID:  "company-uid-1",
				}
			},
			expectedOut: []string{"Account Name:", "JSON Name"},
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "JSON Name", m.capturedUpdateReq.AccountName)
				// accountRef omitted from the JSON body must not be blanked.
				assert.Equal(t, "OLD-REF", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "no fields provided",
			companyUID: "company-uid-1",
			flags:      map[string]string{},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{}
			},
			expectedError: "at least one field must be updated",
		},
		{
			// Empty JSON must be rejected like the flag/interactive modes, and
			// before any account lookup (no listResult is seeded here).
			name:       "empty JSON object rejected",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"json": `{}`,
			},
			setupMock:     func(m *MockManagedAccountService) {},
			expectedError: "at least one field must be updated",
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Nil(t, m.capturedUpdateReq)
			},
		},
		{
			name:       "API error",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Old Name",
						AccountRef:  "REF-001",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateErr = fmt.Errorf("API error: update failed")
			},
			expectedError: "API error: update failed",
		},
		{
			name:       "login error",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock:     func(m *MockManagedAccountService) {},
			expectedError: "login failed",
		},
		{
			// The current account is required to merge unspecified fields, so a
			// list failure must abort rather than risk clobbering.
			name:       "list error during current fetch - update aborts",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listErr = fmt.Errorf("list failed temporarily")
			},
			expectedError: "list failed temporarily",
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Nil(t, m.capturedUpdateReq, "update must not be attempted when the current fetch fails")
			},
		},
		{
			name:       "current account not found - update aborts",
			companyUID: "company-uid-999",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock: func(m *MockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Other Account",
						AccountRef:  "OTHER-REF",
						CompanyUID:  "company-uid-1",
					},
				}
			},
			expectedError: "not found",
			validateCaptured: func(t *testing.T, m *MockManagedAccountService) {
				assert.Nil(t, m.capturedUpdateReq, "update must not be attempted when the account is missing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.name == "login error" {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("login failed")
				})
			} else {
				config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.ManagedAccountService = mockService
					return client, nil
				})
			}

			updateManagedAccountFunc = func(ctx context.Context, client *megaport.Client, companyUID string, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
				return mockService.UpdateManagedAccount(ctx, companyUID, req)
			}

			if tt.prompts != nil {
				promptIndex := 0
				utils.SetResourcePrompt(func(_, msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				})
			}

			cmd := testutil.NewCommand("update", testutil.NoColorAdapter(UpdateManagedAccount))

			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			var stderrOutput string
			stdoutOutput := output.CaptureOutput(func() {
				stderrOutput = captureStderr(t, func() {
					err = cmd.RunE(cmd, []string{tt.companyUID})
				})
			})
			capturedOutput := stdoutOutput + stderrOutput

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				for _, expected := range tt.expectedOut {
					assert.Contains(t, capturedOutput, expected)
				}
			}

			if tt.validateCaptured != nil {
				tt.validateCaptured(t, mockService)
			}
		})
	}
}

func TestUpdateManagedAccount_InvalidJSON(t *testing.T) {
	// No account is seeded: malformed JSON must fail before any account lookup.
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {
		c.ManagedAccountService = &MockManagedAccountService{}
	})
	defer cleanup()
	originalUpdateFunc := updateManagedAccountFunc
	defer func() {
		updateManagedAccountFunc = originalUpdateFunc
	}()

	cmd := testutil.NewCommand("update", testutil.NoColorAdapter(UpdateManagedAccount))

	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")

	testutil.SetFlags(t, cmd, map[string]string{"json": "{invalid json}"})

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"company-uid-1"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

// Function variable tests - success cases

func TestCreateManagedAccountFunc(t *testing.T) {
	mockService := &MockManagedAccountService{
		createResult: &megaport.ManagedAccount{
			AccountName: "Test Account",
			AccountRef:  "REF-001",
			CompanyUID:  "new-uid",
		},
	}

	client := &megaport.Client{}
	client.ManagedAccountService = mockService

	req := &megaport.ManagedAccountRequest{
		AccountName: "Test Account",
		AccountRef:  "REF-001",
	}

	result, err := createManagedAccountFunc(context.Background(), client, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Account", result.AccountName)
	assert.Equal(t, "REF-001", result.AccountRef)
	assert.Equal(t, "new-uid", result.CompanyUID)
	assert.Equal(t, req, mockService.capturedCreateReq)
}

func TestUpdateManagedAccountFunc(t *testing.T) {
	mockService := &MockManagedAccountService{
		updateResult: &megaport.ManagedAccount{
			AccountName: "Updated Account",
			AccountRef:  "NEW-REF",
			CompanyUID:  "uid-123",
		},
	}

	client := &megaport.Client{}
	client.ManagedAccountService = mockService

	req := &megaport.ManagedAccountRequest{
		AccountName: "Updated Account",
		AccountRef:  "NEW-REF",
	}

	result, err := updateManagedAccountFunc(context.Background(), client, "uid-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Account", result.AccountName)
	assert.Equal(t, "NEW-REF", result.AccountRef)
	assert.Equal(t, "uid-123", mockService.capturedUpdateUID)
	assert.Equal(t, req, mockService.capturedUpdateReq)
}

func TestGetManagedAccountFunc(t *testing.T) {
	mockService := &MockManagedAccountService{
		getResult: &megaport.ManagedAccount{
			AccountName: "Test Account",
			AccountRef:  "REF-001",
			CompanyUID:  "uid-123",
		},
	}

	client := &megaport.Client{}
	client.ManagedAccountService = mockService

	result, err := getManagedAccountFunc(context.Background(), client, "uid-123", "Test Account")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Account", result.AccountName)
	assert.Equal(t, "REF-001", result.AccountRef)
	assert.Equal(t, "uid-123", mockService.capturedGetCompanyUID)
	assert.Equal(t, "Test Account", mockService.capturedGetAccountName)
}

// Function variable tests - error cases

func TestCreateManagedAccountFunc_Error(t *testing.T) {
	mockService := &MockManagedAccountService{
		createErr: fmt.Errorf("creation failed: duplicate account"),
	}

	client := &megaport.Client{}
	client.ManagedAccountService = mockService

	req := &megaport.ManagedAccountRequest{
		AccountName: "Test Account",
		AccountRef:  "REF-001",
	}

	result, err := createManagedAccountFunc(context.Background(), client, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "creation failed")
	assert.Equal(t, req, mockService.capturedCreateReq)
}

func TestUpdateManagedAccountFunc_Error(t *testing.T) {
	mockService := &MockManagedAccountService{
		updateErr: fmt.Errorf("update failed: not found"),
	}

	client := &megaport.Client{}
	client.ManagedAccountService = mockService

	req := &megaport.ManagedAccountRequest{
		AccountName: "Updated Account",
		AccountRef:  "NEW-REF",
	}

	result, err := updateManagedAccountFunc(context.Background(), client, "uid-123", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "update failed")
	assert.Equal(t, "uid-123", mockService.capturedUpdateUID)
	assert.Equal(t, req, mockService.capturedUpdateReq)
}

func TestGetManagedAccountFunc_Error(t *testing.T) {
	mockService := &MockManagedAccountService{
		getErr: fmt.Errorf("managed account not found"),
	}

	client := &megaport.Client{}
	client.ManagedAccountService = mockService

	result, err := getManagedAccountFunc(context.Background(), client, "uid-123", "Nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "managed account not found")
	assert.Equal(t, "uid-123", mockService.capturedGetCompanyUID)
	assert.Equal(t, "Nonexistent", mockService.capturedGetAccountName)
}

func TestCreateManagedAccount_InteractiveConflict(t *testing.T) {
	tests := []struct {
		name  string
		flags map[string]string
	}{
		{
			name: "interactive with value flag",
			flags: map[string]string{
				"interactive":  "true",
				"account-name": "Test Account",
			},
		},
		{
			name: "interactive with json",
			flags: map[string]string{
				"interactive": "true",
				"json":        `{"accountName":"JSON Account","accountRef":"JSON-REF"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Login must never be reached: the guard returns before it.
			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				t.Fatal("login should not be called when input modes conflict")
				return nil, nil
			})

			cmd := testutil.NewCommand("create", testutil.NoColorAdapter(CreateManagedAccount))
			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be combined with")

			var cliErr *exitcodes.CLIError
			require.True(t, errors.As(err, &cliErr))
			assert.Equal(t, exitcodes.Usage, cliErr.Code)
		})
	}
}

func TestUpdateManagedAccount_InteractiveConflict(t *testing.T) {
	tests := []struct {
		name  string
		flags map[string]string
	}{
		{
			name: "interactive with value flag",
			flags: map[string]string{
				"interactive":  "true",
				"account-name": "Test Account",
			},
		},
		{
			name: "interactive with json",
			flags: map[string]string{
				"interactive": "true",
				"json":        `{"accountName":"JSON Name","accountRef":"JSON-REF"}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				t.Fatal("login should not be called when input modes conflict")
				return nil, nil
			})

			cmd := testutil.NewCommand("update", testutil.NoColorAdapter(UpdateManagedAccount))
			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			testutil.SetFlags(t, cmd, tt.flags)

			var err error
			output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{"acct-uid"})
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot be combined with")

			var cliErr *exitcodes.CLIError
			require.True(t, errors.As(err, &cliErr))
			assert.Equal(t, exitcodes.Usage, cliErr.Code)
		})
	}
}

func captureStderr(t *testing.T, fn func()) (result string) {
	t.Helper()
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	defer func() { os.Stderr = old }()
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() { defer close(done); _, _ = io.Copy(&buf, r) }()
	defer func() { _ = w.Close(); <-done; _ = r.Close(); result = buf.String() }()
	fn()
	return
}

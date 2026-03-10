package managed_account

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func testCommandAdapter(fn func(cmd *cobra.Command, args []string, noColor bool) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return fn(cmd, args, true)
	}
}

func TestListManagedAccounts(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

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
		setupMock        func(*mockManagedAccountService)
		expectedError    string
		expectedAccounts []string
		unexpectedAccts  []string
		outputFormat     string
	}{
		{
			name: "list all accounts",
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedAccounts: []string{"Acme Corp"},
			unexpectedAccts:  []string{"Beta Inc"},
			outputFormat:     "csv",
		},
		{
			name: "empty result",
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{}
			},
			expectedAccounts: []string{},
			outputFormat:     "table",
		},
		{
			name: "nil API result",
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = nil
			},
			expectedAccounts: []string{},
			outputFormat:     "table",
		},
		{
			name: "API error",
			setupMock: func(m *mockManagedAccountService) {
				m.listErr = fmt.Errorf("API error: service unavailable")
			},
			expectedError: "API error: service unavailable",
			outputFormat:  "table",
		},
		{
			name: "login error",
			setupMock: func(m *mockManagedAccountService) {
				// Will be handled by overriding LoginFunc to return error
			},
			expectedError: "login failed",
			outputFormat:  "table",
		},
		{
			name: "invalid output format",
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = testAccounts
			},
			expectedError: "invalid output format",
			outputFormat:  "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.name == "login error" {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("login failed")
				}
			} else {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.ManagedAccountService = mockService
					return client, nil
				}
			}

			cmd := &cobra.Command{
				Use: "list",
				RunE: func(cmd *cobra.Command, args []string) error {
					return ListManagedAccounts(cmd, args, true, tt.outputFormat)
				},
			}

			cmd.Flags().String("account-name", "", "Filter by account name")
			cmd.Flags().String("account-ref", "", "Filter by account ref")

			for flagName, flagValue := range tt.flags {
				err := cmd.Flags().Set(flagName, flagValue)
				if err != nil {
					t.Fatalf("Failed to set %s flag: %v", flagName, err)
				}
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
			})

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
					assert.Contains(t, capturedOutput, "No managed accounts found matching the specified filters")
				}
			}
		})
	}
}

func TestGetManagedAccount(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalGetFunc := getManagedAccountFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		getManagedAccountFunc = originalGetFunc
	}()

	tests := []struct {
		name          string
		companyUID    string
		accountName   string
		setupMock     func(*mockManagedAccountService)
		expectedError string
		outputFormat  string
		expectedOut   []string
	}{
		{
			name:        "success table format",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
				m.getResult = nil
			},
			expectedError: "invalid managed account: nil value",
			outputFormat:  "table",
		},
		{
			name:        "invalid output format",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
				m.getErr = fmt.Errorf("managed account not found")
			},
			expectedError: "managed account not found",
			outputFormat:  "table",
		},
		{
			name:        "login error",
			companyUID:  "company-uid-1",
			accountName: "Acme Corp",
			setupMock: func(m *mockManagedAccountService) {
				// Will be handled by overriding LoginFunc to return error
			},
			expectedError: "login failed",
			outputFormat:  "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.name == "login error" {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("login failed")
				}
			} else {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.ManagedAccountService = mockService
					return client, nil
				}
			}

			getManagedAccountFunc = func(ctx context.Context, client *megaport.Client, companyUID string, name string) (*megaport.ManagedAccount, error) {
				return mockService.GetManagedAccount(ctx, companyUID, name)
			}

			cmd := &cobra.Command{
				Use: "get",
				RunE: func(cmd *cobra.Command, args []string) error {
					return GetManagedAccount(cmd, args, true, tt.outputFormat)
				},
			}

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
	originalLoginFunc := config.LoginFunc
	originalCreateFunc := createManagedAccountFunc
	originalPrompt := utils.ResourcePrompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		createManagedAccountFunc = originalCreateFunc
		utils.ResourcePrompt = originalPrompt
	}()

	tests := []struct {
		name             string
		flags            map[string]string
		setupMock        func(*mockManagedAccountService)
		prompts          []string
		expectedError    string
		expectedOut      []string
		validateCaptured func(t *testing.T, m *mockManagedAccountService)
	}{
		{
			name: "flag mode",
			flags: map[string]string{
				"account-name": "Test Account",
				"account-ref":  "REF-001",
			},
			setupMock: func(m *mockManagedAccountService) {
				m.createResult = &megaport.ManagedAccount{
					AccountName: "Test Account",
					AccountRef:  "REF-001",
					CompanyUID:  "new-company-uid",
				}
			},
			expectedOut: []string{"new-company-uid"},
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
				m.createResult = &megaport.ManagedAccount{
					AccountName: "Interactive Account",
					AccountRef:  "INT-REF",
					CompanyUID:  "interactive-uid",
				}
			},
			prompts:     []string{"Interactive Account", "INT-REF"},
			expectedOut: []string{"interactive-uid"},
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
				m.createResult = &megaport.ManagedAccount{
					AccountName: "JSON Account",
					AccountRef:  "JSON-REF",
					CompanyUID:  "json-uid",
				}
			},
			expectedOut: []string{"json-uid"},
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
				m.createErr = fmt.Errorf("API error: creation failed")
			},
			expectedError: "API error: creation failed",
		},
		{
			name:          "no input",
			flags:         map[string]string{},
			setupMock:     func(m *mockManagedAccountService) {},
			expectedError: "no input provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.ManagedAccountService = mockService
				return client, nil
			}

			createManagedAccountFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
				return mockService.CreateManagedAccount(ctx, req)
			}

			if tt.prompts != nil {
				promptIndex := 0
				utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			cmd := &cobra.Command{
				Use:  "create",
				RunE: testCommandAdapter(CreateManagedAccount),
			}

			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{})
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

			if tt.validateCaptured != nil {
				tt.validateCaptured(t, mockService)
			}
		})
	}
}

func TestCreateManagedAccount_InvalidJSON(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalCreateFunc := createManagedAccountFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		createManagedAccountFunc = originalCreateFunc
	}()

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ManagedAccountService = &mockManagedAccountService{}
		return client, nil
	}

	cmd := &cobra.Command{
		Use:  "create",
		RunE: testCommandAdapter(CreateManagedAccount),
	}

	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")

	_ = cmd.Flags().Set("json", "{invalid json}")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}

func TestCreateManagedAccount_LoginError(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
	}()

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("login failed: invalid credentials")
	}

	cmd := &cobra.Command{
		Use:  "create",
		RunE: testCommandAdapter(CreateManagedAccount),
	}

	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")

	_ = cmd.Flags().Set("account-name", "Test Account")
	_ = cmd.Flags().Set("account-ref", "REF-001")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

func TestUpdateManagedAccount(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalUpdateFunc := updateManagedAccountFunc
	originalPrompt := utils.ResourcePrompt
	defer func() {
		config.LoginFunc = originalLoginFunc
		updateManagedAccountFunc = originalUpdateFunc
		utils.ResourcePrompt = originalPrompt
	}()

	tests := []struct {
		name             string
		companyUID       string
		flags            map[string]string
		setupMock        func(*mockManagedAccountService)
		prompts          []string
		expectedError    string
		expectedOut      []string
		validateCaptured func(t *testing.T, m *mockManagedAccountService)
	}{
		{
			name:       "flag mode - update name",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "Updated Name",
			},
			setupMock: func(m *mockManagedAccountService) {
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
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "Updated Name", m.capturedUpdateReq.AccountName)
			},
		},
		{
			name:       "flag mode - update both fields",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
				"account-ref":  "NEW-REF",
			},
			setupMock: func(m *mockManagedAccountService) {
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
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
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
			setupMock: func(m *mockManagedAccountService) {
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
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "UPDATED-REF", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "interactive mode",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"interactive": "true",
			},
			setupMock: func(m *mockManagedAccountService) {
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
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "Interactive Name", m.capturedUpdateReq.AccountName)
			},
		},
		{
			name:       "JSON string mode",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"json": `{"accountName":"JSON Name","accountRef":"JSON-REF"}`,
			},
			setupMock: func(m *mockManagedAccountService) {
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
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
				assert.Equal(t, "JSON Name", m.capturedUpdateReq.AccountName)
				assert.Equal(t, "JSON-REF", m.capturedUpdateReq.AccountRef)
			},
		},
		{
			name:       "no fields provided",
			companyUID: "company-uid-1",
			flags:      map[string]string{},
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{}
			},
			expectedError: "at least one field must be updated",
		},
		{
			name:       "API error",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{}
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
			setupMock:     func(m *mockManagedAccountService) {},
			expectedError: "login failed",
		},
		{
			name:       "list error during original fetch - update still succeeds",
			companyUID: "company-uid-1",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock: func(m *mockManagedAccountService) {
				m.listErr = fmt.Errorf("list failed temporarily")
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "New Name",
					AccountRef:  "REF-001",
					CompanyUID:  "company-uid-1",
				}
			},
			// Update should succeed even if list fails — no originalAccount for change display
			expectedOut: []string{"company-uid-1"},
			validateCaptured: func(t *testing.T, m *mockManagedAccountService) {
				assert.Equal(t, "company-uid-1", m.capturedUpdateUID)
				assert.NotNil(t, m.capturedUpdateReq)
			},
		},
		{
			name:       "original account not found in list - update still succeeds",
			companyUID: "company-uid-999",
			flags: map[string]string{
				"account-name": "New Name",
			},
			setupMock: func(m *mockManagedAccountService) {
				m.listResult = []*megaport.ManagedAccount{
					{
						AccountName: "Other Account",
						AccountRef:  "OTHER-REF",
						CompanyUID:  "company-uid-1",
					},
				}
				m.updateResult = &megaport.ManagedAccount{
					AccountName: "New Name",
					AccountRef:  "REF-999",
					CompanyUID:  "company-uid-999",
				}
			},
			expectedOut: []string{"company-uid-999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockManagedAccountService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			if tt.name == "login error" {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					return nil, fmt.Errorf("login failed")
				}
			} else {
				config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
					client := &megaport.Client{}
					client.ManagedAccountService = mockService
					return client, nil
				}
			}

			updateManagedAccountFunc = func(ctx context.Context, client *megaport.Client, companyUID string, req *megaport.ManagedAccountRequest) (*megaport.ManagedAccount, error) {
				return mockService.UpdateManagedAccount(ctx, companyUID, req)
			}

			if tt.prompts != nil {
				promptIndex := 0
				utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
					if promptIndex < len(tt.prompts) {
						response := tt.prompts[promptIndex]
						promptIndex++
						return response, nil
					}
					return "", fmt.Errorf("unexpected prompt call")
				}
			}

			cmd := &cobra.Command{
				Use:  "update",
				RunE: testCommandAdapter(UpdateManagedAccount),
			}

			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = cmd.RunE(cmd, []string{tt.companyUID})
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

			if tt.validateCaptured != nil {
				tt.validateCaptured(t, mockService)
			}
		})
	}
}

func TestUpdateManagedAccount_InvalidJSON(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalUpdateFunc := updateManagedAccountFunc
	defer func() {
		config.LoginFunc = originalLoginFunc
		updateManagedAccountFunc = originalUpdateFunc
	}()

	config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.ManagedAccountService = &mockManagedAccountService{}
		return client, nil
	}

	cmd := &cobra.Command{
		Use:  "update",
		RunE: testCommandAdapter(UpdateManagedAccount),
	}

	cmd.Flags().String("account-name", "", "")
	cmd.Flags().String("account-ref", "", "")
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")

	_ = cmd.Flags().Set("json", "{invalid json}")

	var err error
	output.CaptureOutput(func() {
		err = cmd.RunE(cmd, []string{"company-uid-1"})
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing JSON")
}

// Function variable tests - success cases

func TestCreateManagedAccountFunc(t *testing.T) {
	mockService := &mockManagedAccountService{
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
	mockService := &mockManagedAccountService{
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
	mockService := &mockManagedAccountService{
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
	mockService := &mockManagedAccountService{
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
	mockService := &mockManagedAccountService{
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
	mockService := &mockManagedAccountService{
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

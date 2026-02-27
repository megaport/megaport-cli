package managed_account

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var noColor = true

var testAccounts = []*megaport.ManagedAccount{
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
}

func TestPrintManagedAccounts_Table(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "table", noColor)
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "ACCOUNT NAME")
	assert.Contains(t, out, "ACCOUNT REF")
	assert.Contains(t, out, "COMPANY UID")

	assert.Contains(t, out, "Acme Corp")
	assert.Contains(t, out, "REF-001")
	assert.Contains(t, out, "company-uid-1")

	assert.Contains(t, out, "Beta Inc")
	assert.Contains(t, out, "REF-002")
	assert.Contains(t, out, "company-uid-2")

	assert.Contains(t, out, "┌")
	assert.Contains(t, out, "┐")
	assert.Contains(t, out, "└")
	assert.Contains(t, out, "┘")
	assert.Contains(t, out, "├")
	assert.Contains(t, out, "┤")
	assert.Contains(t, out, "│")
	assert.Contains(t, out, "─")
}

func TestPrintManagedAccounts_JSON(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "json", noColor)
		assert.NoError(t, err)
	})

	expected := `[
  {
	"account_name": "Acme Corp",
	"account_ref": "REF-001",
	"company_uid": "company-uid-1"
  },
  {
	"account_name": "Beta Inc",
	"account_ref": "REF-002",
	"company_uid": "company-uid-2"
  }
]`
	assert.JSONEq(t, expected, out)
}

func TestPrintManagedAccounts_CSV(t *testing.T) {
	out := output.CaptureOutput(func() {
		err := printManagedAccounts(testAccounts, "csv", noColor)
		assert.NoError(t, err)
	})

	expected := `account_name,account_ref,company_uid
Acme Corp,REF-001,company-uid-1
Beta Inc,REF-002,company-uid-2
`
	assert.Equal(t, expected, out)
}

func TestPrintManagedAccounts_Invalid(t *testing.T) {
	var err error
	out := output.CaptureOutput(func() {
		err = printManagedAccounts(testAccounts, "invalid", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, out)
}

func TestPrintManagedAccounts_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		accounts []*megaport.ManagedAccount
		format   string
	}{
		{
			name:     "empty slice table format",
			accounts: []*megaport.ManagedAccount{},
			format:   "table",
		},
		{
			name:     "empty slice csv format",
			accounts: []*megaport.ManagedAccount{},
			format:   "csv",
		},
		{
			name:     "empty slice json format",
			accounts: []*megaport.ManagedAccount{},
			format:   "json",
		},
		{
			name:     "nil slice table format",
			accounts: nil,
			format:   "table",
		},
		{
			name:     "nil slice csv format",
			accounts: nil,
			format:   "csv",
		},
		{
			name:     "nil slice json format",
			accounts: nil,
			format:   "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := output.CaptureOutput(func() {
				err := printManagedAccounts(tt.accounts, tt.format, noColor)
				assert.NoError(t, err)
			})

			switch tt.format {
			case "table":
				assert.Contains(t, out, "ACCOUNT NAME")
				assert.Contains(t, out, "ACCOUNT REF")
				assert.Contains(t, out, "COMPANY UID")
				assert.Contains(t, out, "┌")
				assert.Contains(t, out, "┐")
				assert.Contains(t, out, "└")
				assert.Contains(t, out, "┘")
				assert.Contains(t, out, "│")
				assert.Contains(t, out, "─")
			case "csv":
				expected := "account_name,account_ref,company_uid\n"
				assert.Equal(t, expected, out)
			case "json":
				assert.Equal(t, "[]\n", out)
			}
		})
	}
}

func TestToManagedAccountOutput_EdgeCases(t *testing.T) {
	t.Run("nil account", func(t *testing.T) {
		_, err := ToManagedAccountOutput(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid managed account: nil value")
	})

	t.Run("zero value account", func(t *testing.T) {
		account := &megaport.ManagedAccount{}
		out, err := ToManagedAccountOutput(account)
		assert.NoError(t, err)
		assert.Equal(t, "", out.AccountName)
		assert.Equal(t, "", out.AccountRef)
		assert.Equal(t, "", out.CompanyUID)
	})

	t.Run("full account", func(t *testing.T) {
		account := &megaport.ManagedAccount{
			AccountName: "Test Account",
			AccountRef:  "REF-123",
			CompanyUID:  "uid-456",
		}
		out, err := ToManagedAccountOutput(account)
		assert.NoError(t, err)
		assert.Equal(t, "Test Account", out.AccountName)
		assert.Equal(t, "REF-123", out.AccountRef)
		assert.Equal(t, "uid-456", out.CompanyUID)
	})
}

func TestFilterManagedAccounts(t *testing.T) {
	activeAccounts := []*megaport.ManagedAccount{
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
			AccountName: "Gamma LLC",
			AccountRef:  "GAMMA-REF",
			CompanyUID:  "company-uid-3",
		},
		{
			AccountName: "Acme Subsidiary",
			AccountRef:  "ACME-SUB",
			CompanyUID:  "company-uid-4",
		},
	}

	tests := []struct {
		name         string
		accounts     []*megaport.ManagedAccount
		accountName  string
		accountRef   string
		expected     int
		expectedUIDs []string
	}{
		{
			name:         "no filters",
			accounts:     activeAccounts,
			expected:     4,
			expectedUIDs: []string{"company-uid-1", "company-uid-2", "company-uid-3", "company-uid-4"},
		},
		{
			name:         "filter by name (partial match)",
			accounts:     activeAccounts,
			accountName:  "Acme",
			expected:     2,
			expectedUIDs: []string{"company-uid-1", "company-uid-4"},
		},
		{
			name:         "filter by name (case insensitive)",
			accounts:     activeAccounts,
			accountName:  "acme",
			expected:     2,
			expectedUIDs: []string{"company-uid-1", "company-uid-4"},
		},
		{
			name:         "filter by account ref",
			accounts:     activeAccounts,
			accountRef:   "REF-00",
			expected:     2,
			expectedUIDs: []string{"company-uid-1", "company-uid-2"},
		},
		{
			name:         "filter by account ref (case insensitive)",
			accounts:     activeAccounts,
			accountRef:   "gamma",
			expected:     1,
			expectedUIDs: []string{"company-uid-3"},
		},
		{
			name:         "combined filters",
			accounts:     activeAccounts,
			accountName:  "Acme",
			accountRef:   "REF",
			expected:     1,
			expectedUIDs: []string{"company-uid-1"},
		},
		{
			name:         "non-matching filters",
			accounts:     activeAccounts,
			accountName:  "nonexistent",
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "nil slice",
			accounts:     nil,
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "empty slice",
			accounts:     []*megaport.ManagedAccount{},
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "nil elements in slice",
			accounts:     []*megaport.ManagedAccount{nil, activeAccounts[0], nil, activeAccounts[1]},
			expected:     2,
			expectedUIDs: []string{"company-uid-1", "company-uid-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterManagedAccounts(tt.accounts, tt.accountName, tt.accountRef)

			assert.Equal(t, tt.expected, len(filtered), "Filtered account count should match expected")

			if len(tt.expectedUIDs) > 0 {
				actualUIDs := make([]string, len(filtered))
				for i, account := range filtered {
					actualUIDs[i] = account.CompanyUID
				}
				assert.ElementsMatch(t, tt.expectedUIDs, actualUIDs, "Filtered account UIDs should match expected")
			}
		})
	}
}

func TestDisplayManagedAccountChanges(t *testing.T) {
	tests := []struct {
		name        string
		original    *megaport.ManagedAccount
		updated     *megaport.ManagedAccount
		expectedOut []string
	}{
		{
			name: "name changed",
			original: &megaport.ManagedAccount{
				AccountName: "Old Name",
				AccountRef:  "REF-001",
			},
			updated: &megaport.ManagedAccount{
				AccountName: "New Name",
				AccountRef:  "REF-001",
			},
			expectedOut: []string{"Account Name:", "Old Name", "New Name"},
		},
		{
			name: "ref changed",
			original: &megaport.ManagedAccount{
				AccountName: "Same Name",
				AccountRef:  "OLD-REF",
			},
			updated: &megaport.ManagedAccount{
				AccountName: "Same Name",
				AccountRef:  "NEW-REF",
			},
			expectedOut: []string{"Account Ref:", "OLD-REF", "NEW-REF"},
		},
		{
			name: "both changed",
			original: &megaport.ManagedAccount{
				AccountName: "Old Name",
				AccountRef:  "OLD-REF",
			},
			updated: &megaport.ManagedAccount{
				AccountName: "New Name",
				AccountRef:  "NEW-REF",
			},
			expectedOut: []string{"Account Name:", "Account Ref:"},
		},
		{
			name: "no changes",
			original: &megaport.ManagedAccount{
				AccountName: "Same",
				AccountRef:  "SAME",
			},
			updated: &megaport.ManagedAccount{
				AccountName: "Same",
				AccountRef:  "SAME",
			},
			expectedOut: []string{"No changes detected"},
		},
		{
			name:        "nil original",
			original:    nil,
			updated:     &megaport.ManagedAccount{AccountName: "Test"},
			expectedOut: []string{},
		},
		{
			name:        "nil updated",
			original:    &megaport.ManagedAccount{AccountName: "Test"},
			updated:     nil,
			expectedOut: []string{},
		},
		{
			name:        "both nil",
			original:    nil,
			updated:     nil,
			expectedOut: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capturedOutput := output.CaptureOutput(func() {
				displayManagedAccountChanges(tt.original, tt.updated, true)
			})

			for _, expected := range tt.expectedOut {
				assert.Contains(t, capturedOutput, expected)
			}
		})
	}
}

func TestBuildManagedAccountRequestFromFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]string
		validate func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name: "both flags provided",
			flags: map[string]string{
				"account-name": "Test Account",
				"account-ref":  "REF-001",
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Test Account", req.AccountName)
				assert.Equal(t, "REF-001", req.AccountRef)
			},
		},
		{
			name: "name only",
			flags: map[string]string{
				"account-name": "Test Account",
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Test Account", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name:  "no flags (defaults)",
			flags: map[string]string{},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			req, err := buildManagedAccountRequestFromFlags(cmd)
			assert.NoError(t, err)
			assert.NotNil(t, req)
			tt.validate(t, req)
		})
	}
}

func TestBuildManagedAccountRequestFromJSON(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		setupFile     func(t *testing.T) string
		expectedError string
		validate      func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"accountName":"JSON Account","accountRef":"JSON-REF"}`,
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "JSON Account", req.AccountName)
				assert.Equal(t, "JSON-REF", req.AccountRef)
			},
		},
		{
			name:    "valid JSON string with partial fields",
			jsonStr: `{"accountName":"Partial Account"}`,
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Partial Account", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name:          "invalid JSON syntax",
			jsonStr:       `{invalid json}`,
			expectedError: "error parsing JSON",
		},
		{
			name:    "empty JSON object",
			jsonStr: `{}`,
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name: "valid JSON file",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "account.json")
				err := os.WriteFile(path, []byte(`{"accountName":"File Account","accountRef":"FILE-REF"}`), 0644)
				assert.NoError(t, err)
				return path
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "File Account", req.AccountName)
				assert.Equal(t, "FILE-REF", req.AccountRef)
			},
		},
		{
			name:          "JSON file not found",
			jsonFile:      "/nonexistent/path/account.json",
			expectedError: "error reading JSON file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.setupFile != nil {
				jsonFile = tt.setupFile(t)
			}

			req, err := buildManagedAccountRequestFromJSON(tt.jsonStr, jsonFile)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				tt.validate(t, req)
			}
		})
	}
}

func TestBuildUpdateManagedAccountRequestFromFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]string
		validate func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name: "account-name only",
			flags: map[string]string{
				"account-name": "Updated Account",
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name: "account-ref only",
			flags: map[string]string{
				"account-ref": "NEW-REF",
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name: "both flags",
			flags: map[string]string{
				"account-name": "Updated Account",
				"account-ref":  "NEW-REF",
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name:  "no flags",
			flags: map[string]string{},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("account-name", "", "")
			cmd.Flags().String("account-ref", "", "")

			for k, v := range tt.flags {
				_ = cmd.Flags().Set(k, v)
			}

			req, err := buildUpdateManagedAccountRequestFromFlags(cmd)
			assert.NoError(t, err)
			assert.NotNil(t, req)
			tt.validate(t, req)
		})
	}
}

func TestBuildUpdateManagedAccountRequestFromJSON(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		setupFile     func(t *testing.T) string
		expectedError string
		validate      func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"accountName":"Updated Account","accountRef":"NEW-REF"}`,
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name:    "valid JSON string with partial fields",
			jsonStr: `{"accountName":"Updated Account"}`,
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name:          "invalid JSON syntax",
			jsonStr:       `{invalid}`,
			expectedError: "error parsing JSON",
		},
		{
			name: "valid JSON file",
			setupFile: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "update.json")
				err := os.WriteFile(path, []byte(`{"accountName":"File Updated","accountRef":"FILE-REF"}`), 0644)
				assert.NoError(t, err)
				return path
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "File Updated", req.AccountName)
				assert.Equal(t, "FILE-REF", req.AccountRef)
			},
		},
		{
			name:          "JSON file not found",
			jsonFile:      "/nonexistent/path/update.json",
			expectedError: "error reading JSON file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.setupFile != nil {
				jsonFile = tt.setupFile(t)
			}

			req, err := buildUpdateManagedAccountRequestFromJSON(tt.jsonStr, jsonFile)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				tt.validate(t, req)
			}
		})
	}
}

func TestBuildManagedAccountRequestFromPrompt(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() {
		utils.ResourcePrompt = originalPrompt
	}()

	tests := []struct {
		name          string
		prompts       []string
		expectedError string
		validate      func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name: "all prompts answered successfully",
			prompts: []string{
				"Test Account",
				"REF-001",
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Test Account", req.AccountName)
				assert.Equal(t, "REF-001", req.AccountRef)
			},
		},
		{
			name: "empty account name",
			prompts: []string{
				"",
			},
			expectedError: "account name is required",
		},
		{
			name: "empty account ref",
			prompts: []string{
				"Test Account",
				"",
			},
			expectedError: "account reference is required",
		},
		{
			name: "prompt error",
			prompts: []string{
				"ERROR",
			},
			expectedError: "prompt failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					if response == "ERROR" {
						return "", fmt.Errorf("prompt failed")
					}
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			req, err := buildManagedAccountRequestFromPrompt(true)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				tt.validate(t, req)
			}
		})
	}
}

func TestBuildUpdateManagedAccountRequestFromPrompt(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() {
		utils.ResourcePrompt = originalPrompt
	}()

	tests := []struct {
		name          string
		prompts       []string
		expectedError string
		validate      func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name: "update name only",
			prompts: []string{
				"Updated Account", // name
				"",                // ref
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name: "update ref only",
			prompts: []string{
				"",        // name
				"NEW-REF", // ref
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name: "update both fields",
			prompts: []string{
				"Updated Account", // name
				"NEW-REF",         // ref
			},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name: "no fields updated",
			prompts: []string{
				"", // name
				"", // ref
			},
			expectedError: "at least one field must be updated",
		},
		{
			name: "prompt error",
			prompts: []string{
				"ERROR",
			},
			expectedError: "prompt failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptIndex := 0
			utils.ResourcePrompt = func(_, msg string, _ bool) (string, error) {
				if promptIndex < len(tt.prompts) {
					response := tt.prompts[promptIndex]
					promptIndex++
					if response == "ERROR" {
						return "", fmt.Errorf("prompt failed")
					}
					return response, nil
				}
				return "", fmt.Errorf("unexpected prompt call")
			}

			req, err := buildUpdateManagedAccountRequestFromPrompt(true)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				tt.validate(t, req)
			}
		})
	}
}

func TestPrintManagedAccounts_NilAccountInSlice(t *testing.T) {
	accounts := []*megaport.ManagedAccount{
		{
			AccountName: "Test Account",
			AccountRef:  "REF-001",
			CompanyUID:  "uid-001",
		},
		nil,
	}

	var err error
	output.CaptureOutput(func() {
		err = printManagedAccounts(accounts, "table", noColor)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid managed account: nil value")
}

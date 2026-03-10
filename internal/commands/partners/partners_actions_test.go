package partners

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
	"github.com/stretchr/testify/require"
)

func TestFindPartners(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalPrompt := utils.Prompt
	origPrintPartnersFunc := printPartnersFunc

	defer func() {
		printPartnersFunc = origPrintPartnersFunc
		utils.Prompt = originalPrompt
		config.LoginFunc = originalLoginFunc
	}()

	tests := []struct {
		name          string
		prompts       []string
		expectedError string
		setupMock     func(*testing.T, *mockPartnerService)
		expectedCount int
	}{
		{
			name: "successful search with all filters",
			prompts: []string{
				"Test Product",
				"AWS",
				"Amazon",
				"123",
				"blue",
				"table",
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockPartnerService) {
				m.listPartnersResponse = []*megaport.PartnerMegaport{
					{
						ProductName:   "Test Product",
						ConnectType:   "AWS",
						CompanyName:   "Amazon",
						LocationId:    123,
						DiversityZone: "blue",
					},
					{
						ProductName:   "Other Product",
						ConnectType:   "AZURE",
						CompanyName:   "Microsoft",
						LocationId:    456,
						DiversityZone: "red",
					},
				}
				m.listPartnersErr = nil
			},
			expectedCount: 1,
		},
		{
			name: "search with no filters",
			prompts: []string{
				"",
				"",
				"",
				"",
				"",
				"json",
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockPartnerService) {
				m.listPartnersResponse = []*megaport.PartnerMegaport{
					{
						ProductName:   "Test Product",
						ConnectType:   "AWS",
						CompanyName:   "Amazon",
						LocationId:    123,
						DiversityZone: "blue",
					},
					{
						ProductName:   "Other Product",
						ConnectType:   "AZURE",
						CompanyName:   "Microsoft",
						LocationId:    456,
						DiversityZone: "red",
					},
				}
				m.listPartnersErr = nil
			},
			expectedCount: 2,
		},
		{
			name: "invalid location ID format",
			prompts: []string{
				"",
				"",
				"",
				"not-a-number",
				"",
				"table",
			},
			expectedError: "invalid location ID format",
			setupMock: func(t *testing.T, m *mockPartnerService) {
				m.listPartnersResponse = []*megaport.PartnerMegaport{}
				m.listPartnersErr = nil
			},
			expectedCount: 0,
		},
		{
			name: "API error",
			prompts: []string{
				"",
				"",
			},
			expectedError: "error listing partners",
			setupMock: func(t *testing.T, m *mockPartnerService) {
				m.listPartnersResponse = nil
				m.listPartnersErr = fmt.Errorf("API connection failure")
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockPartnerService{}
			if tt.setupMock != nil {
				tt.setupMock(t, mockService)
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.expectedError == "error logging in" {
					return nil, fmt.Errorf("login failure")
				}
				return &megaport.Client{
					PartnerService: mockService,
				}, nil
			}

			promptIndex := 0
			utils.Prompt = func(message string, noColor bool) (string, error) {
				if promptIndex >= len(tt.prompts) {
					return "", fmt.Errorf("unexpected additional prompt: %s", message)
				}
				response := tt.prompts[promptIndex]
				promptIndex++
				return response, nil
			}

			var capturedPartners []*megaport.PartnerMegaport
			printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string, noColor bool) error {
				capturedPartners = partners
				return nil
			}

			cmd := &cobra.Command{}
			err := FindPartners(cmd, []string{}, false)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.prompts), promptIndex, "not all prompts were used")
				assert.Equal(t, tt.expectedCount, len(capturedPartners), "incorrect number of filtered partners")
			}
		})
	}
}

func TestListPartners(t *testing.T) {
	testPartners := []*megaport.PartnerMegaport{
		{
			ProductName:   "AWS Direct Connect",
			ConnectType:   "AWS",
			CompanyName:   "Amazon",
			LocationId:    123,
			DiversityZone: "blue",
		},
		{
			ProductName:   "Azure ExpressRoute",
			ConnectType:   "AZURE",
			CompanyName:   "Microsoft",
			LocationId:    456,
			DiversityZone: "red",
		},
		{
			ProductName:   "Google Cloud Interconnect",
			ConnectType:   "GOOGLE",
			CompanyName:   "Google",
			LocationId:    789,
			DiversityZone: "blue",
		},
	}

	tests := []struct {
		name          string
		partners      []*megaport.PartnerMegaport
		partnersErr   error
		loginErr      error
		flags         map[string]string
		expectedErr   string
		expectedCount int
		expectWarning bool
	}{
		{
			name:          "filter by product-name",
			partners:      testPartners,
			flags:         map[string]string{"product-name": "AWS Direct Connect"},
			expectedCount: 1,
		},
		{
			name:          "filter by connect-type",
			partners:      testPartners,
			flags:         map[string]string{"connect-type": "AZURE"},
			expectedCount: 1,
		},
		{
			name:          "empty result",
			partners:      testPartners,
			flags:         map[string]string{"product-name": "NonExistent"},
			expectedCount: 0,
			expectWarning: true,
		},
		{
			name:        "API error",
			partners:    nil,
			partnersErr: fmt.Errorf("API connection failure"),
			expectedErr: "error listing partners",
		},
		{
			name:        "login error",
			loginErr:    fmt.Errorf("login failure"),
			expectedErr: "error logging in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalLoginFunc := config.LoginFunc
			origPrintPartnersFunc := printPartnersFunc
			defer func() {
				config.LoginFunc = originalLoginFunc
				printPartnersFunc = origPrintPartnersFunc
			}()

			mockService := &mockPartnerService{
				listPartnersResponse: tt.partners,
				listPartnersErr:      tt.partnersErr,
			}

			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.loginErr != nil {
					return nil, tt.loginErr
				}
				return &megaport.Client{
					PartnerService: mockService,
				}, nil
			}

			var capturedPartners []*megaport.PartnerMegaport
			printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string, noColor bool) error {
				capturedPartners = partners
				return nil
			}

			cmd := &cobra.Command{
				Use: "list",
			}
			cmd.Flags().String("product-name", "", "")
			cmd.Flags().String("connect-type", "", "")
			cmd.Flags().String("company-name", "", "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().String("diversity-zone", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			var err error
			capturedOutput := output.CaptureOutput(func() {
				err = ListPartners(cmd, []string{}, true, "table")
			})

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(capturedPartners))
				if tt.expectWarning {
					assert.Contains(t, capturedOutput, "No partner ports found")
				}
			}
		})
	}
}

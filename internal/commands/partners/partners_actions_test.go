package partners

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestFindPartners(t *testing.T) {
	originalLoginFunc := config.LoginFunc
	originalPrompt := utils.Prompt
	// Save original functions
	origPrintPartnersFunc := printPartnersFunc

	// Restore originals after tests
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
				"Test Product", // Product name
				"AWS",          // Connect type
				"Amazon",       // Company name
				"123",          // Location ID
				"blue",         // Diversity zone
				"table",        // Output format
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockPartnerService) {
				// Set up mock partners data
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
				"",     // Product name (empty)
				"",     // Connect type (empty)
				"",     // Company name (empty)
				"",     // Location ID (empty)
				"",     // Diversity zone (empty)
				"json", // Output format
			},
			expectedError: "",
			setupMock: func(t *testing.T, m *mockPartnerService) {
				// Set up mock partners data - all should be returned
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
				"",             // Product name
				"",             // Connect type
				"",             // Company name
				"not-a-number", // Invalid Location ID
				"",             // Diversity zone
				"table",        // Output format
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
				"", // Product name
				"", // Connect type - won't get past this due to API error
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
			// Create mock service
			mockService := &mockPartnerService{}
			if tt.setupMock != nil {
				tt.setupMock(t, mockService)
			}

			// Override the package-level login function with our test version
			config.LoginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.expectedError == "error logging in" {
					return nil, fmt.Errorf("login failure")
				}
				return &megaport.Client{
					PartnerService: mockService,
				}, nil
			}

			// Set up the prompt mock to return test values
			promptIndex := 0
			utils.Prompt = func(message string, noColor bool) (string, error) {
				if promptIndex >= len(tt.prompts) {
					return "", fmt.Errorf("unexpected additional prompt: %s", message)
				}
				response := tt.prompts[promptIndex]
				promptIndex++
				return response, nil
			}

			// Capture filtered partners for count verification
			var capturedPartners []*megaport.PartnerMegaport
			printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string, noColor bool) error {
				capturedPartners = partners
				return nil
			}

			// Execute function with noColor=false (default for tests)
			cmd := &cobra.Command{}
			err := FindPartners(cmd, []string{}, false)

			// Verify results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				// Check that all prompts were used
				assert.Equal(t, len(tt.prompts), promptIndex, "not all prompts were used")
				// Verify filtered partner count
				assert.Equal(t, tt.expectedCount, len(capturedPartners), "incorrect number of filtered partners")
			}
		})
	}
}

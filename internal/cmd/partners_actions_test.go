package cmd

import (
	"context"
	"fmt"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestFindPartners(t *testing.T) {
	// Save original functions
	originalLoginFunc := loginFunc
	originalPrompt := prompt
	originalPrintPartners := printPartnersFunc

	// Restore originals after tests
	defer func() {
		loginFunc = originalLoginFunc
		prompt = originalPrompt
		printPartnersFunc = originalPrintPartners
	}()

	// Mock printPartners to avoid actual output during tests
	printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string) error {
		return nil
	}

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

			// Mock the login function
			loginFunc = func(ctx context.Context) (*megaport.Client, error) {
				if tt.expectedError == "error logging in" {
					return nil, fmt.Errorf("login failure")
				}
				return &megaport.Client{
					PartnerService: mockService,
				}, nil
			}

			// Mock the prompt function to return predefined responses
			promptIndex := 0
			prompt = func(message string) (string, error) {
				if promptIndex >= len(tt.prompts) {
					return "", fmt.Errorf("unexpected additional prompt: %s", message)
				}
				response := tt.prompts[promptIndex]
				promptIndex++
				return response, nil
			}

			// Capture filtered partners for count verification
			var capturedPartners []*megaport.PartnerMegaport
			printPartnersFunc = func(partners []*megaport.PartnerMegaport, format string) error {
				capturedPartners = partners
				return nil
			}

			// Execute function
			cmd := &cobra.Command{}
			err := FindPartners(cmd, []string{})

			// Verify results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				// Check that all prompts were consumed
				assert.Equal(t, len(tt.prompts), promptIndex, "not all prompts were used")
				// Verify filtered partner count
				assert.Equal(t, tt.expectedCount, len(capturedPartners), "incorrect number of filtered partners")
			}
		})
	}
}

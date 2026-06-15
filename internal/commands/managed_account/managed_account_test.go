package managed_account

import (
	"testing"

	megaport "github.com/megaport/megaportgo"
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

package mcr

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func mockPromptSequence(responses []string) func(string, string, bool) (string, error) {
	idx := 0
	return func(resourceType, msg string, noColor bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		val := responses[idx]
		idx++
		return val, nil
	}
}

func TestPromptForMCRDetails_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	originalTagsPrompt := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(originalPrompt)
		utils.SetResourceTagsPrompt(originalTagsPrompt)
	}()

	// name, term, portSpeed, locationID, asn, diversityZone, costCentre, promoCode
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"Test MCR", "12", "5000", "1", "65000", "blue", "IT-2024", "PROMO",
	}))
	utils.SetResourceTagsPrompt(func(noColor bool) (map[string]string, error) {
		return map[string]string{"env": "test"}, nil
	})

	req, err := promptForMCRDetails(true)
	assert.NoError(t, err)
	assert.Equal(t, "Test MCR", req.Name)
	assert.Equal(t, 12, req.Term)
	assert.Equal(t, 5000, req.PortSpeed)
	assert.Equal(t, 1, req.LocationID)
	assert.Equal(t, 65000, req.MCRAsn)
	assert.Equal(t, "blue", req.DiversityZone)
	assert.Equal(t, "IT-2024", req.CostCentre)
	assert.Equal(t, "PROMO", req.PromoCode)
}

func TestPromptForMCRDetails_EmptyName(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, err := promptForMCRDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestPromptForMCRDetails_InvalidTerm(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"MCR", "abc"}))

	_, err := promptForMCRDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid term")
}

func TestPromptForMCRDetails_InvalidPortSpeed(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"MCR", "12", "999"}))

	_, err := promptForMCRDetails(true)
	assert.Error(t, err)
}

func TestPromptForMCRDetails_InvalidLocationID(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"MCR", "12", "5000", "abc"}))

	_, err := promptForMCRDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

func TestPromptForUpdateMCRDetails_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// name, costCentre, marketplaceVisibility(yes/no), visibilityValue, term
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"Updated MCR", "IT-New", "yes", "true", "24",
	}))

	req, err := promptForUpdateMCRDetails("mcr-123", true)
	assert.NoError(t, err)
	assert.Equal(t, "Updated MCR", req.Name)
	assert.Equal(t, "IT-New", req.CostCentre)
	assert.NotNil(t, req.MarketplaceVisibility)
	assert.True(t, *req.MarketplaceVisibility)
	assert.NotNil(t, req.ContractTermMonths)
	assert.Equal(t, 24, *req.ContractTermMonths)
}

func TestPromptForUpdateMCRDetails_NoChanges(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"", "", "", ""}))

	_, err := promptForUpdateMCRDetails("mcr-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be updated")
}

func TestPromptForUpdateMCRDetails_InvalidTerm(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"", "", "", "99"}))

	_, err := promptForUpdateMCRDetails("mcr-123", true)
	assert.Error(t, err)
}

func TestPromptPrefixFilterEntry_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{
		"192.168.0.0/24", "permit", "16", "24",
	}))

	entry, err := promptPrefixFilterEntry(true)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, "192.168.0.0/24", entry.Prefix)
	assert.Equal(t, "permit", entry.Action)
	assert.Equal(t, 16, entry.Ge)
	assert.Equal(t, 24, entry.Le)
}

func TestPromptPrefixFilterEntry_EmptyPrefix(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	entry, err := promptPrefixFilterEntry(true)
	assert.NoError(t, err)
	assert.Nil(t, entry)
}

func TestPromptPrefixFilterEntry_InvalidAction(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"10.0.0.0/8", "allow"}))

	_, err := promptPrefixFilterEntry(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid action")
}

func TestPromptPrefixFilterEntry_InvalidGE(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"10.0.0.0/8", "permit", "abc"}))

	_, err := promptPrefixFilterEntry(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid GE value")
}

func TestPromptAddNewPrefixEntries_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// First entry, then empty prefix to stop
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"10.0.0.0/8", "permit", "", "",
		"", // empty prefix to stop
	}))

	entries, err := promptAddNewPrefixEntries(true)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestPromptForPrefixFilterListDetails_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// description, addressFamily, then one entry + empty to stop
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"My PFL", "IPv4",
		"10.0.0.0/8", "permit", "", "",
		"", // stop adding entries
	}))

	req, err := promptForPrefixFilterListDetails("mcr-123", true)
	assert.NoError(t, err)
	assert.Equal(t, "My PFL", req.PrefixFilterList.Description)
	assert.Equal(t, "IPv4", req.PrefixFilterList.AddressFamily)
	assert.Len(t, req.PrefixFilterList.Entries, 1)
}

func TestPromptForPrefixFilterListDetails_EmptyDescription(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, err := promptForPrefixFilterListDetails("mcr-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description is required")
}

func TestPromptForPrefixFilterListDetails_InvalidAddressFamily(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"My PFL", "IPv5"}))

	_, err := promptForPrefixFilterListDetails("mcr-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid address family")
}

func TestPromptForPrefixFilterListDetails_NoEntries(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{
		"My PFL", "IPv4",
		"", // empty prefix immediately = no entries
	}))

	_, err := promptForPrefixFilterListDetails("mcr-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one entry is required")
}

func TestPromptForIPSecTunnelCount(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		wantErr     bool
		errContains string
	}{
		{"empty input uses default", "", 0, false, ""},
		{"valid count 10", "10", 10, false, ""},
		{"valid count 20", "20", 20, false, ""},
		{"valid count 30", "30", 30, false, ""},
		{"non-numeric input", "abc", 0, true, "invalid tunnel count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalPrompt := utils.GetResourcePrompt()
			defer func() { utils.SetResourcePrompt(originalPrompt) }()

			utils.SetResourcePrompt(mockPromptSequence([]string{tt.input}))

			count, err := promptForIPSecTunnelCount(true)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}
		})
	}
}

func TestPromptForIPSecTunnelCountUpdate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int
		wantErr     bool
		errContains string
	}{
		{"valid count 10", "10", 10, false, ""},
		{"valid count 20", "20", 20, false, ""},
		{"valid count 30", "30", 30, false, ""},
		{"zero disables IPSec", "0", 0, false, ""},
		{"empty input requires value", "", 0, true, "tunnel count is required"},
		{"non-numeric input", "abc", 0, true, "invalid tunnel count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalPrompt := utils.GetResourcePrompt()
			defer func() { utils.SetResourcePrompt(originalPrompt) }()

			utils.SetResourcePrompt(mockPromptSequence([]string{tt.input}))

			count, err := promptForIPSecTunnelCountUpdate(true)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}
		})
	}
}

func TestPromptUpdateExistingEntries_KeepUnmodified(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	current := []*megaport.MCRPrefixListEntry{
		{Prefix: "10.0.0.0/8", Action: "permit", Ge: 16, Le: 24},
	}

	// keep=yes, modify=no
	utils.SetResourcePrompt(mockPromptSequence([]string{"yes", "no"}))

	entries, err := promptUpdateExistingEntries(current, true)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "10.0.0.0/8", entries[0].Prefix)
}

func TestPromptUpdateExistingEntries_ModifyEntry(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	current := []*megaport.MCRPrefixListEntry{
		{Prefix: "10.0.0.0/8", Action: "permit", Ge: 16, Le: 24},
	}

	// keep=yes, modify=yes, new prefix, new action, new ge, new le
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"yes", "yes", "192.168.0.0/16", "deny", "20", "28",
	}))

	entries, err := promptUpdateExistingEntries(current, true)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "192.168.0.0/16", entries[0].Prefix)
	assert.Equal(t, "deny", entries[0].Action)
	assert.Equal(t, 20, entries[0].Ge)
	assert.Equal(t, 28, entries[0].Le)
}

func TestPromptUpdateExistingEntries_DeleteEntry(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	current := []*megaport.MCRPrefixListEntry{
		{Prefix: "10.0.0.0/8", Action: "permit"},
	}

	// keep=no (delete)
	utils.SetResourcePrompt(mockPromptSequence([]string{"no"}))

	entries, err := promptUpdateExistingEntries(current, true)
	assert.NoError(t, err)
	assert.Len(t, entries, 0)
}

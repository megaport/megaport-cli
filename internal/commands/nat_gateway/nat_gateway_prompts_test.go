package nat_gateway

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockNGPromptSequence(responses []string) func(string, string, bool) (string, error) {
	idx := 0
	return func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		val := responses[idx]
		idx++
		if val == "ERROR" {
			return "", fmt.Errorf("simulated prompt error")
		}
		return val, nil
	}
}

func noopTagsPrompt(_ bool) (map[string]string, error) {
	return map[string]string{}, nil
}

// promptForCreateNATGatewayDetails: name, locationID, speed, term, sessionCount,
// asn, diversityZone, autoRenew, promoCode, serviceLevelRef, ResourceTagsPrompt.

func TestPromptForCreateNATGatewayDetails_Success(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	utils.SetResourcePrompt(mockNGPromptSequence([]string{
		"My NAT GW", // name
		"1",         // locationID
		"1000",      // speed
		"12",        // term
		"100",       // sessionCount
		"65000",     // asn
		"blue",      // diversityZone
		"yes",       // autoRenew
		"PROMO",     // promoCode
		"SLR-1",     // serviceLevelRef
	}))
	utils.SetResourceTagsPrompt(noopTagsPrompt)

	req, err := promptForCreateNATGatewayDetails(true)
	require.NoError(t, err)
	assert.Equal(t, "My NAT GW", req.ProductName)
	assert.Equal(t, 1, req.LocationID)
	assert.Equal(t, 1000, req.Speed)
	assert.Equal(t, 12, req.Term)
	assert.Equal(t, 100, req.Config.SessionCount)
	assert.Equal(t, 65000, req.Config.ASN)
	assert.Equal(t, "blue", req.Config.DiversityZone)
	assert.True(t, req.AutoRenewTerm)
	assert.Equal(t, "PROMO", req.PromoCode)
	assert.Equal(t, "SLR-1", req.ServiceLevelReference)
}

func TestPromptForCreateNATGatewayDetails_AutoRenewNo(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	utils.SetResourcePrompt(mockNGPromptSequence([]string{
		"GW", "1", "1000", "12", "", "", "", "no", "", "",
	}))
	utils.SetResourceTagsPrompt(noopTagsPrompt)

	req, err := promptForCreateNATGatewayDetails(true)
	require.NoError(t, err)
	assert.False(t, req.AutoRenewTerm)
}

func TestPromptForCreateNATGatewayDetails_InvalidASN(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "1", "1000", "12", "", "blue"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ASN")
}

func TestPromptForCreateNATGatewayDetails_OutOfRangeASN(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "1", "1000", "12", "", "0"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid ASN")
}

func TestPromptForCreateNATGatewayDetails_EmptyName(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	// Provide valid responses for all prompts except name so validation can run.
	utils.SetResourcePrompt(mockNGPromptSequence([]string{"", "1", "1000", "12", "", "", "", "", "", ""}))
	utils.SetResourceTagsPrompt(noopTagsPrompt)

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestPromptForCreateNATGatewayDetails_InvalidLocationID(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "notanumber"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

func TestPromptForCreateNATGatewayDetails_ZeroLocationID(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "0"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

func TestPromptForCreateNATGatewayDetails_InvalidSpeed(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "1", "bad"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid speed")
}

func TestPromptForCreateNATGatewayDetails_ZeroSpeed(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "1", "0"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid speed")
}

func TestPromptForCreateNATGatewayDetails_InvalidTerm(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "1", "1000", "abc"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid term")
}

func TestPromptForCreateNATGatewayDetails_InvalidSessionCount(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"GW", "1", "1000", "12", "abc"}))

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session count")
}

func TestPromptForCreateNATGatewayDetails_TagsError(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	utils.SetResourcePrompt(mockNGPromptSequence([]string{
		"GW", "1", "1000", "12", "", "", "", "", "", "",
	}))
	utils.SetResourceTagsPrompt(func(_ bool) (map[string]string, error) {
		return nil, fmt.Errorf("tags error")
	})

	_, err := promptForCreateNATGatewayDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tags error")
}

// promptForUpdateNATGatewayDetails: name, locationID, speed, term, sessionCount,
// diversityZone, autoRenew, promoCode, serviceLevelRef, ResourceTagsPrompt.

func TestPromptForUpdateNATGatewayDetails_Success(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	utils.SetResourcePrompt(mockNGPromptSequence([]string{
		"Updated GW", // name
		"2",          // locationID
		"2000",       // speed
		"24",         // term
		"200",        // sessionCount
		"red",        // diversityZone
		"y",          // autoRenew
		"SAVE",       // promoCode
		"SLR-2",      // serviceLevelRef
	}))
	utils.SetResourceTagsPrompt(noopTagsPrompt)

	req, explicit, err := promptForUpdateNATGatewayDetails("uid-upd", true)
	require.NoError(t, err)
	assert.Equal(t, "uid-upd", req.ProductUID)
	assert.Equal(t, "Updated GW", req.ProductName)
	assert.Equal(t, 2, req.LocationID)
	assert.Equal(t, 2000, req.Speed)
	assert.Equal(t, 24, req.Term)
	assert.Equal(t, 200, req.Config.SessionCount)
	assert.True(t, explicit.SessionCount)
	assert.Equal(t, "red", req.Config.DiversityZone)
	assert.True(t, explicit.DiversityZone)
	assert.True(t, req.AutoRenewTerm)
	assert.True(t, explicit.AutoRenewTerm)
}

func TestPromptForUpdateNATGatewayDetails_AllEmpty(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	utils.SetResourcePrompt(mockNGPromptSequence([]string{
		"", "", "", "", "", "", "", "", "",
	}))
	utils.SetResourceTagsPrompt(noopTagsPrompt)

	req, explicit, err := promptForUpdateNATGatewayDetails("uid-empty", true)
	require.NoError(t, err)
	assert.Equal(t, "uid-empty", req.ProductUID)
	assert.False(t, explicit.AutoRenewTerm)
	assert.False(t, explicit.SessionCount)
	assert.False(t, explicit.DiversityZone)
}

func TestPromptForUpdateNATGatewayDetails_AutoRenewNo(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	origTags := utils.GetResourceTagsPrompt()
	defer func() {
		utils.SetResourcePrompt(origPrompt)
		utils.SetResourceTagsPrompt(origTags)
	}()

	utils.SetResourcePrompt(mockNGPromptSequence([]string{
		"", "", "", "", "", "", "no", "", "",
	}))
	utils.SetResourceTagsPrompt(noopTagsPrompt)

	req, explicit, err := promptForUpdateNATGatewayDetails("uid-no", true)
	require.NoError(t, err)
	assert.False(t, req.AutoRenewTerm)
	assert.True(t, explicit.AutoRenewTerm)
}

func TestPromptForUpdateNATGatewayDetails_InvalidLocationID(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"", "bad"}))

	_, _, err := promptForUpdateNATGatewayDetails("uid-1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

func TestPromptForUpdateNATGatewayDetails_InvalidSpeed(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"", "", "bad"}))

	_, _, err := promptForUpdateNATGatewayDetails("uid-1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid speed")
}

func TestPromptForUpdateNATGatewayDetails_InvalidTerm(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"", "", "", "abc"}))

	_, _, err := promptForUpdateNATGatewayDetails("uid-1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid term")
}

func TestPromptForUpdateNATGatewayDetails_InvalidSessionCount(t *testing.T) {
	origPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(origPrompt)

	utils.SetResourcePrompt(mockNGPromptSequence([]string{"", "", "", "", "bad"}))

	_, _, err := promptForUpdateNATGatewayDetails("uid-1", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid session count")
}

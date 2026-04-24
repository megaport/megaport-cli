package mve

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func mockPromptSequence(responses []string) func(string, string, bool) (string, error) {
	idx := 0
	return func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		val := responses[idx]
		idx++
		return val, nil
	}
}

// promptMVEBaseDetails tests

func TestPromptMVEBaseDetails_Success(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// name, term, locationID, diversityZone, promoCode, costCentre, vendor, imageID, productSize, mveLabel
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"Test MVE", "12", "1", "blue", "PROMO", "IT-2024", "cisco", "42", "LARGE", "my-label",
	}))

	req, vendorStr, imageID, productSize, mveLabel, err := promptMVEBaseDetails(true)
	assert.NoError(t, err)
	assert.Equal(t, "Test MVE", req.Name)
	assert.Equal(t, 12, req.Term)
	assert.Equal(t, 1, req.LocationID)
	assert.Equal(t, "blue", req.DiversityZone)
	assert.Equal(t, "PROMO", req.PromoCode)
	assert.Equal(t, "IT-2024", req.CostCentre)
	assert.Equal(t, "cisco", vendorStr)
	assert.Equal(t, 42, imageID)
	assert.Equal(t, "LARGE", productSize)
	assert.Equal(t, "my-label", mveLabel)
}

func TestPromptMVEBaseDetails_EmptyName(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, _, _, _, _, err := promptMVEBaseDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestPromptMVEBaseDetails_InvalidTerm(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"MVE", "abc"}))

	_, _, _, _, _, err := promptMVEBaseDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid term")
}

func TestPromptMVEBaseDetails_InvalidLocationID(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"MVE", "12", "abc"}))

	_, _, _, _, _, err := promptMVEBaseDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location ID")
}

func TestPromptMVEBaseDetails_InvalidImageID(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// name, term, locationID, diversityZone, promoCode, costCentre, vendor, imageID (invalid)
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"MVE", "12", "1", "", "", "", "cisco", "abc",
	}))

	_, _, _, _, _, err := promptMVEBaseDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid image ID")
}

// promptForUpdateMVEDetails tests

func TestPromptForUpdateMVEDetails_Success(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// name, costCentre, contractTerm
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"NewName", "", "",
	}))

	req, err := promptForUpdateMVEDetails("mve-123", true)
	assert.NoError(t, err)
	assert.Equal(t, "mve-123", req.MVEID)
	assert.Equal(t, "NewName", req.Name)
	assert.Equal(t, "", req.CostCentre)
	assert.Nil(t, req.ContractTermMonths)
}

func TestPromptForUpdateMVEDetails_NoChanges(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"", "", ""}))

	_, err := promptForUpdateMVEDetails("mve-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be provided for update")
}

func TestPromptForUpdateMVEDetails_InvalidContractTerm(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"", "", "abc"}))

	_, err := promptForUpdateMVEDetails("mve-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid contract term")
}

// promptMVEVnics tests

func TestPromptMVEVnics_NoVnics(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// empty description immediately to finish, then error because 0 VNICs
	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, err := promptMVEVnics(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one VNIC is required")
}

func TestPromptMVEVnics_OneVnic(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// description, vlan, then empty description to finish
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"eth0", "100",
		"", // empty description to stop
	}))

	vnics, err := promptMVEVnics(true)
	assert.NoError(t, err)
	assert.Len(t, vnics, 1)
	assert.Equal(t, "eth0", vnics[0].Description)
	assert.Equal(t, 100, vnics[0].VLAN)
}

func TestPromptMVEVnics_MultipleVnics(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// first VNIC: description, vlan; second VNIC: description, vlan; then empty to stop
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"eth0", "100",
		"eth1", "200",
		"", // empty description to stop
	}))

	vnics, err := promptMVEVnics(true)
	assert.NoError(t, err)
	assert.Len(t, vnics, 2)
	assert.Equal(t, "eth0", vnics[0].Description)
	assert.Equal(t, 100, vnics[0].VLAN)
	assert.Equal(t, "eth1", vnics[1].Description)
	assert.Equal(t, 200, vnics[1].VLAN)
}

func TestPromptMVEVnics_InvalidVLAN(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"eth0", "abc"}))

	_, err := promptMVEVnics(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid VLAN ID")
}

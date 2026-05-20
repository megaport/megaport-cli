package mve

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
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

// promptMVEVendorConfig tests — cisco/palo_alto admin password handling

func TestPromptMVEVendorConfig_Cisco_WithAdminPassword(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origSecret := utils.GetSecretResourcePrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetSecretResourcePrompt(origSecret)
	}()

	// Order: manageLocally, adminSSHPublicKey, sshPublicKey, cloudInit,
	// fmcIPAddress, fmcRegistrationKey, fmcNatID, then SECRET adminPassword
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"true", "admin-ssh", "ssh-key", "cloud-init",
		"10.0.0.1", "fmc-reg", "fmc-nat",
	}))
	utils.SetSecretResourcePrompt(mockPromptSequence([]string{"s3cr3t!"}))

	cfg, err := promptMVEVendorConfig("cisco", 99, "LARGE", "lbl", true)
	assert.NoError(t, err)
	cisco, ok := cfg.(*megaport.CiscoConfig)
	assert.True(t, ok)
	assert.Equal(t, "cisco", cisco.Vendor)
	assert.Equal(t, 99, cisco.ImageID)
	assert.Equal(t, "LARGE", cisco.ProductSize)
	assert.Equal(t, "lbl", cisco.MVELabel)
	assert.True(t, cisco.ManageLocally)
	assert.Equal(t, "admin-ssh", cisco.AdminSSHPublicKey)
	assert.Equal(t, "ssh-key", cisco.SSHPublicKey)
	assert.Equal(t, "cloud-init", cisco.CloudInit)
	assert.Equal(t, "10.0.0.1", cisco.FMCIPAddress)
	assert.Equal(t, "fmc-reg", cisco.FMCRegistrationKey)
	assert.Equal(t, "fmc-nat", cisco.FMCNatID)
	assert.Equal(t, "s3cr3t!", cisco.AdminPassword)
}

func TestPromptMVEVendorConfig_PaloAlto_PlaintextOnly(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origSecret := utils.GetSecretResourcePrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetSecretResourcePrompt(origSecret)
	}()

	// Order: sshPublicKey, SECRET adminPassword, SECRET adminPasswordHash, licenseData
	utils.SetResourcePrompt(mockPromptSequence([]string{"ssh-key", "license"}))
	utils.SetSecretResourcePrompt(mockPromptSequence([]string{"p4ssw0rd", ""}))

	cfg, err := promptMVEVendorConfig("palo_alto", 42, "MEDIUM", "lbl", true)
	assert.NoError(t, err)
	pa, ok := cfg.(*megaport.PaloAltoConfig)
	assert.True(t, ok)
	assert.Equal(t, "palo_alto", pa.Vendor)
	assert.Equal(t, "ssh-key", pa.SSHPublicKey)
	assert.Equal(t, "p4ssw0rd", pa.AdminPassword)
	assert.Equal(t, "", pa.AdminPasswordHash)
	assert.Equal(t, "license", pa.LicenseData)
}

func TestPromptMVEVendorConfig_PaloAlto_HashOnly(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origSecret := utils.GetSecretResourcePrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetSecretResourcePrompt(origSecret)
	}()

	utils.SetResourcePrompt(mockPromptSequence([]string{"ssh-key", "license"}))
	utils.SetSecretResourcePrompt(mockPromptSequence([]string{"", "hashed-value"}))

	cfg, err := promptMVEVendorConfig("palo_alto", 42, "MEDIUM", "lbl", true)
	assert.NoError(t, err)
	pa, ok := cfg.(*megaport.PaloAltoConfig)
	assert.True(t, ok)
	assert.Equal(t, "", pa.AdminPassword)
	assert.Equal(t, "hashed-value", pa.AdminPasswordHash)
}

func TestPromptMVEVendorConfig_PaloAlto_BothBlank(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origSecret := utils.GetSecretResourcePrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetSecretResourcePrompt(origSecret)
	}()

	utils.SetResourcePrompt(mockPromptSequence([]string{"ssh-key"}))
	utils.SetSecretResourcePrompt(mockPromptSequence([]string{"", ""}))

	_, err := promptMVEVendorConfig("palo_alto", 42, "MEDIUM", "lbl", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "either admin password or admin password hash is required")
}

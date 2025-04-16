package validation

import (
	"fmt"
	"strings"
)

// MVE Specific Validation
// This file contains validation functions for MVE-specific fields.
// These functions are used to validate requests and responses related to MVEs.

// Supported MVE vendors
var (
	ValidMVEVendors = []string{
		"6wind",
		"aruba",
		"aviatrix",
		"cisco",
		"fortinet",
		"palo_alto",
		"prisma",
		"versa",
		"vmware",
		"meraki",
	}
)

// ValidateMVERequest validates a request to create an MVE
func ValidateMVERequest(name string, term int, locationID int) error {
	if name == "" {
		return NewValidationError("MVE name", name, "cannot be empty")
	}

	if len(name) > 64 {
		return NewValidationError("MVE name", name, "cannot exceed 64 characters")
	}

	if err := ValidateContractTerm(term); err != nil {
		return err
	}

	if locationID <= 0 {
		return NewValidationError("location ID", locationID, "must be a positive integer")
	}

	return nil
}

// ValidateMVEVendor validates an MVE vendor name
func ValidateMVEVendor(vendor string) error {
	normalizedVendor := strings.ToLower(vendor)
	for _, validVendor := range ValidMVEVendors {
		if normalizedVendor == validVendor {
			return nil
		}
	}
	return NewValidationError("MVE vendor", vendor,
		fmt.Sprintf("must be one of: %v", ValidMVEVendors))
}

// ValidateMVENetworkInterface validates a single network interface configuration
func ValidateMVENetworkInterface(description string) error {
	// Only checking description for now as that's what the API validates
	if description == "" {
		return NewValidationError("network interface description", description, "cannot be empty")
	}
	return nil
}

// ValidateMVENetworkInterfaces validates the list of network interfaces
func ValidateMVENetworkInterfaces(vnics []string) error {
	if len(vnics) > 5 {
		return NewValidationError("network interfaces", len(vnics), "cannot exceed 5 vNICs")
	}

	for i, description := range vnics {
		if err := ValidateMVENetworkInterface(description); err != nil {
			return NewValidationError(fmt.Sprintf("network interface %d", i+1), description, err.Error())
		}
	}

	return nil
}

// ValidateSixwindVSRConfig validates 6WIND VSR configuration
func ValidateSixwindVSRConfig(imageID int, productSize string, mveLabel string, sshPublicKey string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if sshPublicKey == "" {
		return NewValidationError("SSH public key", sshPublicKey, "cannot be empty")
	}

	return nil
}

// ValidateArubaConfig validates Aruba configuration
func ValidateArubaConfig(imageID int, productSize string, mveLabel string, accountName string, accountKey string, systemTag string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if accountName == "" {
		return NewValidationError("account name", accountName, "cannot be empty")
	}

	if accountKey == "" {
		return NewValidationError("account key", accountKey, "cannot be empty")
	}

	if systemTag == "" {
		return NewValidationError("system tag", systemTag, "cannot be empty")
	}

	return nil
}

// ValidateAviatrixConfig validates Aviatrix configuration
func ValidateAviatrixConfig(imageID int, productSize string, mveLabel string, cloudInit string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if cloudInit == "" {
		return NewValidationError("cloud init", cloudInit, "cannot be empty")
	}

	return nil
}

// ValidateCiscoConfig validates Cisco configuration
func ValidateCiscoConfig(imageID int, productSize string, mveLabel string, adminSSHPublicKey string, sshPublicKey string, cloudInit string,
	manageLocally bool, fmcIPAddress string, fmcRegistrationKey string, fmcNatID string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if adminSSHPublicKey == "" {
		return NewValidationError("admin SSH public key", adminSSHPublicKey, "cannot be empty")
	}

	if sshPublicKey == "" {
		return NewValidationError("SSH public key", sshPublicKey, "cannot be empty")
	}

	// Only validate FMC settings if manageLocally is false
	if !manageLocally {
		if fmcIPAddress == "" {
			return NewValidationError("FMC IP address", fmcIPAddress, "cannot be empty when not managing locally")
		}

		if fmcRegistrationKey == "" {
			return NewValidationError("FMC registration key", fmcRegistrationKey, "cannot be empty when not managing locally")
		}

		if fmcNatID == "" {
			return NewValidationError("FMC NAT ID", fmcNatID, "cannot be empty when not managing locally")
		}
	}

	return nil
}

// ValidateFortinetConfig validates Fortinet configuration
func ValidateFortinetConfig(imageID int, productSize string, mveLabel string, adminSSHPublicKey string, sshPublicKey string, licenseData string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if adminSSHPublicKey == "" {
		return NewValidationError("admin SSH public key", adminSSHPublicKey, "cannot be empty")
	}

	if sshPublicKey == "" {
		return NewValidationError("SSH public key", sshPublicKey, "cannot be empty")
	}

	if licenseData == "" {
		return NewValidationError("license data", licenseData, "cannot be empty")
	}

	return nil
}

// ValidatePaloAltoConfig validates Palo Alto configuration
func ValidatePaloAltoConfig(imageID int, productSize string, mveLabel string, sshPublicKey string, adminPasswordHash string, licenseData string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if sshPublicKey == "" {
		return NewValidationError("SSH public key", sshPublicKey, "cannot be empty")
	}

	if adminPasswordHash == "" {
		return NewValidationError("admin password hash", adminPasswordHash, "cannot be empty")
	}

	if licenseData == "" {
		return NewValidationError("license data", licenseData, "cannot be empty")
	}

	return nil
}

// ValidatePrismaConfig validates Prisma configuration
func ValidatePrismaConfig(imageID int, productSize string, mveLabel string, ionKey string, secretKey string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if ionKey == "" {
		return NewValidationError("ION key", ionKey, "cannot be empty")
	}

	if secretKey == "" {
		return NewValidationError("secret key", secretKey, "cannot be empty")
	}

	return nil
}

// ValidateVersaConfig validates Versa configuration
func ValidateVersaConfig(imageID int, productSize string, mveLabel string, directorAddress string, controllerAddress string,
	localAuth string, remoteAuth string, serialNumber string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if directorAddress == "" {
		return NewValidationError("director address", directorAddress, "cannot be empty")
	}

	if controllerAddress == "" {
		return NewValidationError("controller address", controllerAddress, "cannot be empty")
	}

	if localAuth == "" {
		return NewValidationError("local auth", localAuth, "cannot be empty")
	}

	if remoteAuth == "" {
		return NewValidationError("remote auth", remoteAuth, "cannot be empty")
	}

	if serialNumber == "" {
		return NewValidationError("serial number", serialNumber, "cannot be empty")
	}

	return nil
}

// ValidateVmwareConfig validates VMware configuration
func ValidateVmwareConfig(imageID int, productSize string, mveLabel string, adminSSHPublicKey string, sshPublicKey string,
	vcoAddress string, vcoActivationCode string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if adminSSHPublicKey == "" {
		return NewValidationError("admin SSH public key", adminSSHPublicKey, "cannot be empty")
	}

	if sshPublicKey == "" {
		return NewValidationError("SSH public key", sshPublicKey, "cannot be empty")
	}

	if vcoAddress == "" {
		return NewValidationError("VCO address", vcoAddress, "cannot be empty")
	}

	if vcoActivationCode == "" {
		return NewValidationError("VCO activation code", vcoActivationCode, "cannot be empty")
	}

	return nil
}

// ValidateMerakiConfig validates Meraki configuration
func ValidateMerakiConfig(imageID int, productSize string, mveLabel string, token string) error {
	if imageID <= 0 {
		return NewValidationError("image ID", imageID, "must be a positive integer")
	}

	if err := ValidateMVEProductSize(productSize); err != nil {
		return err
	}

	if token == "" {
		return NewValidationError("token", token, "cannot be empty")
	}

	return nil
}

// ValidateMVEVendorConfig validates the vendor configuration based on the vendor type
func ValidateMVEVendorConfig(vendor string, config map[string]interface{}) error {
	if err := ValidateMVEVendor(vendor); err != nil {
		return err
	}

	// Check common required fields
	imageID, ok := config["image_id"].(int)
	if !ok {
		return NewValidationError("image ID", config["image_id"], "must be a valid integer")
	}

	productSize, ok := config["product_size"].(string)
	if !ok {
		return NewValidationError("product size", config["product_size"], "must be a valid string")
	}

	mveLabel, _ := config["mve_label"].(string)
	// mveLabel is optional

	// Validate based on vendor type
	normalizedVendor := strings.ToLower(vendor)
	switch normalizedVendor {
	case "6wind":
		sshPublicKey, _ := config["ssh_public_key"].(string)
		return ValidateSixwindVSRConfig(imageID, productSize, mveLabel, sshPublicKey)
	case "aruba":
		accountName, _ := config["account_name"].(string)
		accountKey, _ := config["account_key"].(string)
		systemTag, _ := config["system_tag"].(string)
		return ValidateArubaConfig(imageID, productSize, mveLabel, accountName, accountKey, systemTag)
	case "aviatrix":
		cloudInit, _ := config["cloud_init"].(string)
		return ValidateAviatrixConfig(imageID, productSize, mveLabel, cloudInit)
	case "cisco":
		adminSSHPublicKey, _ := config["admin_ssh_public_key"].(string)
		sshPublicKey, _ := config["ssh_public_key"].(string)
		cloudInit, _ := config["cloud_init"].(string)
		manageLocally, _ := config["manage_locally"].(bool)
		fmcIPAddress, _ := config["fmc_ip_address"].(string)
		fmcRegistrationKey, _ := config["fmc_registration_key"].(string)
		fmcNatID, _ := config["fmc_nat_id"].(string)
		return ValidateCiscoConfig(imageID, productSize, mveLabel, adminSSHPublicKey, sshPublicKey, cloudInit,
			manageLocally, fmcIPAddress, fmcRegistrationKey, fmcNatID)
	case "fortinet":
		adminSSHPublicKey, _ := config["admin_ssh_public_key"].(string)
		sshPublicKey, _ := config["ssh_public_key"].(string)
		licenseData, _ := config["license_data"].(string)
		return ValidateFortinetConfig(imageID, productSize, mveLabel, adminSSHPublicKey, sshPublicKey, licenseData)
	case "palo_alto":
		sshPublicKey, _ := config["ssh_public_key"].(string)
		adminPasswordHash, _ := config["admin_password_hash"].(string)
		licenseData, _ := config["license_data"].(string)
		return ValidatePaloAltoConfig(imageID, productSize, mveLabel, sshPublicKey, adminPasswordHash, licenseData)
	case "prisma":
		ionKey, _ := config["ion_key"].(string)
		secretKey, _ := config["secret_key"].(string)
		return ValidatePrismaConfig(imageID, productSize, mveLabel, ionKey, secretKey)
	case "versa":
		directorAddress, _ := config["director_address"].(string)
		controllerAddress, _ := config["controller_address"].(string)
		localAuth, _ := config["local_auth"].(string)
		remoteAuth, _ := config["remote_auth"].(string)
		serialNumber, _ := config["serial_number"].(string)
		return ValidateVersaConfig(imageID, productSize, mveLabel, directorAddress, controllerAddress,
			localAuth, remoteAuth, serialNumber)
	case "vmware":
		adminSSHPublicKey, _ := config["admin_ssh_public_key"].(string)
		sshPublicKey, _ := config["ssh_public_key"].(string)
		vcoAddress, _ := config["vco_address"].(string)
		vcoActivationCode, _ := config["vco_activation_code"].(string)
		return ValidateVmwareConfig(imageID, productSize, mveLabel, adminSSHPublicKey, sshPublicKey,
			vcoAddress, vcoActivationCode)
	case "meraki":
		token, _ := config["token"].(string)
		return ValidateMerakiConfig(imageID, productSize, mveLabel, token)
	default:
		return NewValidationError("vendor", vendor, "is not supported")
	}
}

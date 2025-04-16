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

	if len(name) > MaxMVENameLength {
		return NewValidationError("MVE name", name, fmt.Sprintf("cannot exceed %d characters", MaxMVENameLength))
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

	// Check common required fields using helper functions
	imageID, ok := GetIntFromInterface(config["image_id"])
	if !ok {
		return NewValidationError("image ID", config["image_id"], "must be a valid integer")
	}

	productSize, ok := GetStringFromInterface(config["product_size"])
	if !ok {
		return NewValidationError("product size", config["product_size"], "must be a valid string")
	}

	mveLabel, _ := GetStringFromInterface(config["mve_label"]) // Optional field

	// Validate based on vendor type
	normalizedVendor := strings.ToLower(vendor)
	switch normalizedVendor {
	case "6wind":
		sshPublicKey, _ := GetStringFromInterface(config["ssh_public_key"])
		return ValidateSixwindVSRConfig(imageID, productSize, mveLabel, sshPublicKey)
	case "aruba":
		accountName, _ := GetStringFromInterface(config["account_name"])
		accountKey, _ := GetStringFromInterface(config["account_key"])
		systemTag, _ := GetStringFromInterface(config["system_tag"])
		return ValidateArubaConfig(imageID, productSize, mveLabel, accountName, accountKey, systemTag)
	case "aviatrix":
		cloudInit, _ := GetStringFromInterface(config["cloud_init"])
		return ValidateAviatrixConfig(imageID, productSize, mveLabel, cloudInit)
	case "cisco":
		adminSSHPublicKey, _ := GetStringFromInterface(config["admin_ssh_public_key"])
		sshPublicKey, _ := GetStringFromInterface(config["ssh_public_key"])
		cloudInit, _ := GetStringFromInterface(config["cloud_init"])
		manageLocally, _ := GetBoolFromInterface(config["manage_locally"]) // Default is false if not present or invalid
		fmcIPAddress, _ := GetStringFromInterface(config["fmc_ip_address"])
		fmcRegistrationKey, _ := GetStringFromInterface(config["fmc_registration_key"])
		fmcNatID, _ := GetStringFromInterface(config["fmc_nat_id"])
		return ValidateCiscoConfig(imageID, productSize, mveLabel, adminSSHPublicKey, sshPublicKey, cloudInit,
			manageLocally, fmcIPAddress, fmcRegistrationKey, fmcNatID)
	case "fortinet":
		adminSSHPublicKey, _ := GetStringFromInterface(config["admin_ssh_public_key"])
		sshPublicKey, _ := GetStringFromInterface(config["ssh_public_key"])
		licenseData, _ := GetStringFromInterface(config["license_data"])
		return ValidateFortinetConfig(imageID, productSize, mveLabel, adminSSHPublicKey, sshPublicKey, licenseData)
	case "palo_alto":
		sshPublicKey, _ := GetStringFromInterface(config["ssh_public_key"])
		adminPasswordHash, _ := GetStringFromInterface(config["admin_password_hash"])
		licenseData, _ := GetStringFromInterface(config["license_data"])
		return ValidatePaloAltoConfig(imageID, productSize, mveLabel, sshPublicKey, adminPasswordHash, licenseData)
	case "prisma":
		ionKey, _ := GetStringFromInterface(config["ion_key"])
		secretKey, _ := GetStringFromInterface(config["secret_key"])
		return ValidatePrismaConfig(imageID, productSize, mveLabel, ionKey, secretKey)
	case "versa":
		directorAddress, _ := GetStringFromInterface(config["director_address"])
		controllerAddress, _ := GetStringFromInterface(config["controller_address"])
		localAuth, _ := GetStringFromInterface(config["local_auth"])
		remoteAuth, _ := GetStringFromInterface(config["remote_auth"])
		serialNumber, _ := GetStringFromInterface(config["serial_number"])
		return ValidateVersaConfig(imageID, productSize, mveLabel, directorAddress, controllerAddress,
			localAuth, remoteAuth, serialNumber)
	case "vmware":
		adminSSHPublicKey, _ := GetStringFromInterface(config["admin_ssh_public_key"])
		sshPublicKey, _ := GetStringFromInterface(config["ssh_public_key"])
		vcoAddress, _ := GetStringFromInterface(config["vco_address"])
		vcoActivationCode, _ := GetStringFromInterface(config["vco_activation_code"])
		return ValidateVmwareConfig(imageID, productSize, mveLabel, adminSSHPublicKey, sshPublicKey,
			vcoAddress, vcoActivationCode)
	case "meraki":
		token, _ := GetStringFromInterface(config["token"])
		return ValidateMerakiConfig(imageID, productSize, mveLabel, token)
	default:
		// This case should ideally not be reached due to ValidateMVEVendor check, but included for completeness
		return NewValidationError("vendor", vendor, "is not supported")
	}
}

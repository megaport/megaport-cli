package validation

import (
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

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

func ValidateMVENetworkInterfaces(vnics []megaport.MVENetworkInterface) error {
	if len(vnics) > 5 {
		return NewValidationError("network interfaces", len(vnics), "cannot exceed 5 vNICs")
	}
	for i, vnic := range vnics {
		if vnic.Description == "" {
			return NewValidationError(fmt.Sprintf("network interface %d", i+1), vnic.Description, "description cannot be empty")
		}
	}
	return nil
}

func ValidateSixwindVSRConfig(config *megaport.SixwindVSRConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.SSHPublicKey == "" {
		return NewValidationError("SSH public key", config.SSHPublicKey, "cannot be empty")
	}
	return nil
}

func ValidateArubaConfig(config *megaport.ArubaConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.AccountName == "" {
		return NewValidationError("account name", config.AccountName, "cannot be empty")
	}
	if config.AccountKey == "" {
		return NewValidationError("account key", config.AccountKey, "cannot be empty")
	}
	if config.SystemTag == "" {
		return NewValidationError("system tag", config.SystemTag, "cannot be empty")
	}
	return nil
}

func ValidateAviatrixConfig(config *megaport.AviatrixConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.CloudInit == "" {
		return NewValidationError("cloud init", config.CloudInit, "cannot be empty")
	}
	return nil
}

func ValidateCiscoConfig(config *megaport.CiscoConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.AdminSSHPublicKey == "" {
		return NewValidationError("admin SSH public key", config.AdminSSHPublicKey, "cannot be empty")
	}
	if config.SSHPublicKey == "" {
		return NewValidationError("SSH public key", config.SSHPublicKey, "cannot be empty")
	}
	if !config.ManageLocally {
		if config.FMCIPAddress == "" {
			return NewValidationError("FMC IP address", config.FMCIPAddress, "cannot be empty when not managing locally")
		}
		if config.FMCRegistrationKey == "" {
			return NewValidationError("FMC registration key", config.FMCRegistrationKey, "cannot be empty when not managing locally")
		}
		if config.FMCNatID == "" {
			return NewValidationError("FMC NAT ID", config.FMCNatID, "cannot be empty when not managing locally")
		}
	}
	return nil
}

func ValidateVmwareConfig(config *megaport.VmwareConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.AdminSSHPublicKey == "" {
		return NewValidationError("admin SSH public key", config.AdminSSHPublicKey, "cannot be empty")
	}
	if config.SSHPublicKey == "" {
		return NewValidationError("SSH public key", config.SSHPublicKey, "cannot be empty")
	}
	if config.VcoAddress == "" {
		return NewValidationError("VCO address", config.VcoAddress, "cannot be empty")
	}
	if config.VcoActivationCode == "" {
		return NewValidationError("VCO activation code", config.VcoActivationCode, "cannot be empty")
	}
	return nil
}

func ValidateMerakiConfig(config *megaport.MerakiConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.Token == "" {
		return NewValidationError("token", config.Token, "cannot be empty")
	}
	return nil
}

func ValidateFortinetConfig(config *megaport.FortinetConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.AdminSSHPublicKey == "" {
		return NewValidationError("admin SSH public key", config.AdminSSHPublicKey, "cannot be empty")
	}
	if config.SSHPublicKey == "" {
		return NewValidationError("SSH public key", config.SSHPublicKey, "cannot be empty")
	}
	if config.LicenseData == "" {
		return NewValidationError("license data", config.LicenseData, "cannot be empty")
	}
	return nil
}

func ValidatePaloAltoConfig(config *megaport.PaloAltoConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.SSHPublicKey == "" {
		return NewValidationError("SSH public key", config.SSHPublicKey, "cannot be empty")
	}
	if config.AdminPasswordHash == "" {
		return NewValidationError("admin password hash", config.AdminPasswordHash, "cannot be empty")
	}
	if config.LicenseData == "" {
		return NewValidationError("license data", config.LicenseData, "cannot be empty")
	}
	return nil
}

func ValidatePrismaConfig(config *megaport.PrismaConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.IONKey == "" {
		return NewValidationError("ION key", config.IONKey, "cannot be empty")
	}
	if config.SecretKey == "" {
		return NewValidationError("secret key", config.SecretKey, "cannot be empty")
	}
	return nil
}

func ValidateVersaConfig(config *megaport.VersaConfig) error {
	if config.ImageID <= 0 {
		return NewValidationError("image ID", config.ImageID, "must be a positive integer")
	}
	if err := ValidateMVEProductSize(config.ProductSize); err != nil {
		return err
	}
	if config.DirectorAddress == "" {
		return NewValidationError("director address", config.DirectorAddress, "cannot be empty")
	}
	if config.ControllerAddress == "" {
		return NewValidationError("controller address", config.ControllerAddress, "cannot be empty")
	}
	if config.LocalAuth == "" {
		return NewValidationError("local auth", config.LocalAuth, "cannot be empty")
	}
	if config.RemoteAuth == "" {
		return NewValidationError("remote auth", config.RemoteAuth, "cannot be empty")
	}
	if config.SerialNumber == "" {
		return NewValidationError("serial number", config.SerialNumber, "cannot be empty")
	}
	return nil
}

func ValidateMVEVendorConfig(config megaport.VendorConfig) error {
	if config == nil {
		return NewValidationError("vendor config", nil, "cannot be nil")
	}
	switch cfg := config.(type) {
	case *megaport.SixwindVSRConfig:
		return ValidateSixwindVSRConfig(cfg)
	case *megaport.ArubaConfig:
		return ValidateArubaConfig(cfg)
	case *megaport.AviatrixConfig:
		return ValidateAviatrixConfig(cfg)
	case *megaport.CiscoConfig:
		return ValidateCiscoConfig(cfg)
	case *megaport.FortinetConfig:
		return ValidateFortinetConfig(cfg)
	case *megaport.PaloAltoConfig:
		return ValidatePaloAltoConfig(cfg)
	case *megaport.PrismaConfig:
		return ValidatePrismaConfig(cfg)
	case *megaport.VersaConfig:
		return ValidateVersaConfig(cfg)
	case *megaport.VmwareConfig:
		return ValidateVmwareConfig(cfg)
	case *megaport.MerakiConfig:
		return ValidateMerakiConfig(cfg)
	default:
		return NewValidationError("vendor config", config, fmt.Sprintf("unknown vendor type %T", config))
	}
}

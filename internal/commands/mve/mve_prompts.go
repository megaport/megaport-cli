package mve

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

// Interactive prompting for MVE buy details
func promptForBuyMVEDetails(noColor bool) (*megaport.BuyMVERequest, error) {
	// Initialize request
	req := &megaport.BuyMVERequest{}

	// Prompt for required fields
	name, err := utils.ResourcePrompt("mve", "Enter MVE name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	req.Name = name

	termStr, err := utils.ResourcePrompt("mve", "Enter term (1, 12, 24, or 36 months) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("invalid term: %v", err)
	}
	req.Term = term

	locationIDStr, err := utils.ResourcePrompt("mve", "Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID: %v", err)
	}
	req.LocationID = locationID

	// Prompt for optional fields
	diversityZone, err := utils.ResourcePrompt("mve", "Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.DiversityZone = diversityZone

	promoCode, err := utils.ResourcePrompt("mve", "Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	costCentre, err := utils.ResourcePrompt("mve", "Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	// Prompt for vendor selection
	vendorStr, err := utils.ResourcePrompt("mve", "Enter vendor (6wind, aruba, aviatrix, cisco, fortinet, palo_alto, prisma, versa, vmware, meraki) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if vendorStr == "" {
		return nil, fmt.Errorf("vendor is required")
	}

	// Prompt for image ID
	imageIDStr, err := utils.ResourcePrompt("mve", "Enter image ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid image ID: %v", err)
	}

	// Prompt for product size
	productSize, err := utils.ResourcePrompt("mve", "Enter product size (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if productSize == "" {
		return nil, fmt.Errorf("product size is required")
	}

	// Prompt for MVE label
	mveLabel, err := utils.ResourcePrompt("mve", "Enter MVE label (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	// Configure vendor-specific options based on selected vendor
	var vendorConfig megaport.VendorConfig

	switch vendorStr {
	case "6wind":
		sshPublicKey, err := utils.ResourcePrompt("mve", "Enter SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.SixwindVSRConfig{
			Vendor:       "6wind",
			ImageID:      imageID,
			ProductSize:  productSize,
			MVELabel:     mveLabel,
			SSHPublicKey: sshPublicKey,
		}
	case "aruba":
		accountName, err := utils.ResourcePrompt("mve", "Enter account name (required): ", noColor)
		if err != nil {
			return nil, err
		}
		accountKey, err := utils.ResourcePrompt("mve", "Enter account key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		systemTag, err := utils.ResourcePrompt("mve", "Enter system tag (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.ArubaConfig{
			Vendor:      "aruba",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			AccountName: accountName,
			AccountKey:  accountKey,
			SystemTag:   systemTag,
		}
	case "aviatrix":
		cloudInit, err := utils.ResourcePrompt("mve", "Enter cloud init data (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.AviatrixConfig{
			Vendor:      "aviatrix",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			CloudInit:   cloudInit,
		}
	case "cisco":
		manageLocallyStr, err := utils.ResourcePrompt("mve", "Manage locally (true/false) (required): ", noColor)
		if err != nil {
			return nil, err
		}
		manageLocally := strings.ToLower(manageLocallyStr) == "true"

		adminSSHPublicKey, err := utils.ResourcePrompt("mve", "Enter admin SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		sshPublicKey, err := utils.ResourcePrompt("mve", "Enter SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		cloudInit, err := utils.ResourcePrompt("mve", "Enter cloud init data (required): ", noColor)
		if err != nil {
			return nil, err
		}
		fmcIPAddress, err := utils.ResourcePrompt("mve", "Enter FMC IP address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		fmcRegistrationKey, err := utils.ResourcePrompt("mve", "Enter FMC registration key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		fmcNatID, err := utils.ResourcePrompt("mve", "Enter FMC NAT ID (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.CiscoConfig{
			Vendor:             "cisco",
			ImageID:            imageID,
			ProductSize:        productSize,
			MVELabel:           mveLabel,
			ManageLocally:      manageLocally,
			AdminSSHPublicKey:  adminSSHPublicKey,
			SSHPublicKey:       sshPublicKey,
			CloudInit:          cloudInit,
			FMCIPAddress:       fmcIPAddress,
			FMCRegistrationKey: fmcRegistrationKey,
			FMCNatID:           fmcNatID,
		}
	case "fortinet":
		adminSSHPublicKey, err := utils.ResourcePrompt("mve", "Enter admin SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		sshPublicKey, err := utils.ResourcePrompt("mve", "Enter SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		licenseData, err := utils.ResourcePrompt("mve", "Enter license data (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.FortinetConfig{
			Vendor:            "fortinet",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			AdminSSHPublicKey: adminSSHPublicKey,
			SSHPublicKey:      sshPublicKey,
			LicenseData:       licenseData,
		}
	case "palo_alto":
		sshPublicKey, err := utils.ResourcePrompt("mve", "Enter SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		adminPasswordHash, err := utils.ResourcePrompt("mve", "Enter admin password hash (required): ", noColor)
		if err != nil {
			return nil, err
		}
		licenseData, err := utils.ResourcePrompt("mve", "Enter license data (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.PaloAltoConfig{
			Vendor:            "palo_alto",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			SSHPublicKey:      sshPublicKey,
			AdminPasswordHash: adminPasswordHash,
			LicenseData:       licenseData,
		}
	case "prisma":
		ionKey, err := utils.ResourcePrompt("mve", "Enter ION key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		secretKey, err := utils.ResourcePrompt("mve", "Enter secret key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.PrismaConfig{
			Vendor:      "prisma",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			IONKey:      ionKey,
			SecretKey:   secretKey,
		}
	case "versa":
		directorAddress, err := utils.ResourcePrompt("mve", "Enter director address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		controllerAddress, err := utils.ResourcePrompt("mve", "Enter controller address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		localAuth, err := utils.ResourcePrompt("mve", "Enter local auth (required): ", noColor)
		if err != nil {
			return nil, err
		}
		remoteAuth, err := utils.ResourcePrompt("mve", "Enter remote auth (required): ", noColor)
		if err != nil {
			return nil, err
		}
		serialNumber, err := utils.ResourcePrompt("mve", "Enter serial number (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.VersaConfig{
			Vendor:            "versa",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			DirectorAddress:   directorAddress,
			ControllerAddress: controllerAddress,
			LocalAuth:         localAuth,
			RemoteAuth:        remoteAuth,
			SerialNumber:      serialNumber,
		}
	case "vmware":
		adminSSHPublicKey, err := utils.ResourcePrompt("mve", "Enter admin SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		sshPublicKey, err := utils.ResourcePrompt("mve", "Enter SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vcoAddress, err := utils.ResourcePrompt("mve", "Enter VCO address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vcoActivationCode, err := utils.ResourcePrompt("mve", "Enter VCO activation code (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.VmwareConfig{
			Vendor:            "vmware",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			AdminSSHPublicKey: adminSSHPublicKey,
			SSHPublicKey:      sshPublicKey,
			VcoAddress:        vcoAddress,
			VcoActivationCode: vcoActivationCode,
		}
	case "meraki":
		token, err := utils.ResourcePrompt("mve", "Enter token (required): ", noColor)
		if err != nil {
			return nil, err
		}
		vendorConfig = &megaport.MerakiConfig{
			Vendor:      "meraki",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			Token:       token,
		}
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendorStr)
	}

	req.VendorConfig = vendorConfig

	// Prompt for VNICs
	vnics := []megaport.MVENetworkInterface{}
	for {
		fmt.Println("\nEnter VNIC details (leave description empty to finish):")
		description, err := utils.ResourcePrompt("mve", "Enter VNIC description: ", noColor)
		if err != nil {
			return nil, err
		}
		// If description is empty, we're done with VNICs
		if description == "" {
			break
		}

		vlanStr, err := utils.ResourcePrompt("mve", "Enter VLAN ID: ", noColor)
		if err != nil {
			return nil, err
		}
		vlan := 0
		if vlanStr != "" {
			vlan, err = strconv.Atoi(vlanStr)
			if err != nil {
				return nil, fmt.Errorf("invalid VLAN ID: %v", err)
			}
		}

		vnics = append(vnics, megaport.MVENetworkInterface{
			Description: description,
			VLAN:        vlan,
		})
	}

	if len(vnics) == 0 {
		return nil, fmt.Errorf("at least one VNIC is required")
	}

	req.Vnics = vnics

	// Validate the request
	if err := validation.ValidateBuyMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Interactive prompting for MVE update details
func promptForUpdateMVEDetails(mveUID string, noColor bool) (*megaport.ModifyMVERequest, error) {
	// Initialize request with required MVE UID
	req := &megaport.ModifyMVERequest{
		MVEID: mveUID,
	}

	// Prompt for new name (optional)
	name, err := utils.ResourcePrompt("mve", "Enter new MVE name (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if name != "" {
		req.Name = name
	}

	// Prompt for new cost centre (optional)
	costCentre, err := utils.ResourcePrompt("mve", "Enter new cost centre (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if costCentre != "" {
		req.CostCentre = costCentre
	}

	// Prompt for new contract term (optional)
	contractTermStr, err := utils.ResourcePrompt("mve", "Enter new contract term (1, 12, 24, or 36 months, leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if contractTermStr != "" {
		contractTerm, err := strconv.Atoi(contractTermStr)
		if err != nil {
			return nil, fmt.Errorf("invalid contract term: %v", err)
		}
		req.ContractTermMonths = &contractTerm
	}

	// Validate the request
	if err := validation.ValidateUpdateMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

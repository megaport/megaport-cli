package mve

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

func promptForBuyMVEDetails(noColor bool) (*megaport.BuyMVERequest, error) {
	req, vendorStr, imageID, productSize, mveLabel, err := promptMVEBaseDetails(noColor)
	if err != nil {
		return nil, err
	}

	vendorConfig, err := promptMVEVendorConfig(vendorStr, imageID, productSize, mveLabel, noColor)
	if err != nil {
		return nil, err
	}
	req.VendorConfig = vendorConfig

	vnics, err := promptMVEVnics(noColor)
	if err != nil {
		return nil, err
	}
	req.Vnics = vnics

	if err := validation.ValidateBuyMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

func promptMVEBaseDetails(noColor bool) (*megaport.BuyMVERequest, string, int, string, string, error) {
	req := &megaport.BuyMVERequest{}

	name, err := utils.ResourcePrompt("mve", "Enter MVE name (required): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	if name == "" {
		return nil, "", 0, "", "", fmt.Errorf("name is required")
	}
	req.Name = name

	termStr, err := utils.ResourcePrompt("mve", fmt.Sprintf("Enter term (%s months) (required): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, "", 0, "", "", fmt.Errorf("invalid term: %v", err)
	}
	req.Term = term

	locationIDStr, err := utils.ResourcePrompt("mve", "Enter location ID (required): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, "", 0, "", "", fmt.Errorf("invalid location ID: %v", err)
	}
	req.LocationID = locationID

	diversityZone, err := utils.ResourcePrompt("mve", "Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	req.DiversityZone = diversityZone

	promoCode, err := utils.ResourcePrompt("mve", "Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	req.PromoCode = promoCode

	costCentre, err := utils.ResourcePrompt("mve", "Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	req.CostCentre = costCentre

	vendorStr, err := utils.ResourcePrompt("mve", "Enter vendor (6wind, aruba, aviatrix, cisco, fortinet, palo_alto, prisma, versa, vmware, meraki) (required): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	if vendorStr == "" {
		return nil, "", 0, "", "", fmt.Errorf("vendor is required")
	}

	imageIDStr, err := utils.ResourcePrompt("mve", "Enter image ID (required): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, "", 0, "", "", fmt.Errorf("invalid image ID: %v", err)
	}

	productSize, err := utils.ResourcePrompt("mve", "Enter product size (required): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}
	if productSize == "" {
		return nil, "", 0, "", "", fmt.Errorf("product size is required")
	}

	mveLabel, err := utils.ResourcePrompt("mve", "Enter MVE label (optional): ", noColor)
	if err != nil {
		return nil, "", 0, "", "", err
	}

	return req, vendorStr, imageID, productSize, mveLabel, nil
}

func promptMVEVendorConfig(vendorStr string, imageID int, productSize string, mveLabel string, noColor bool) (megaport.VendorConfig, error) {
	switch vendorStr {
	case "6wind":
		sshPublicKey, err := utils.ResourcePrompt("mve", "Enter SSH public key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		return &megaport.SixwindVSRConfig{
			Vendor:       "6wind",
			ImageID:      imageID,
			ProductSize:  productSize,
			MVELabel:     mveLabel,
			SSHPublicKey: sshPublicKey,
		}, nil
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
		return &megaport.ArubaConfig{
			Vendor:      "aruba",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			AccountName: accountName,
			AccountKey:  accountKey,
			SystemTag:   systemTag,
		}, nil
	case "aviatrix":
		cloudInit, err := utils.ResourcePrompt("mve", "Enter cloud init data (required): ", noColor)
		if err != nil {
			return nil, err
		}
		return &megaport.AviatrixConfig{
			Vendor:      "aviatrix",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			CloudInit:   cloudInit,
		}, nil
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
		return &megaport.CiscoConfig{
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
		}, nil
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
		return &megaport.FortinetConfig{
			Vendor:            "fortinet",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			AdminSSHPublicKey: adminSSHPublicKey,
			SSHPublicKey:      sshPublicKey,
			LicenseData:       licenseData,
		}, nil
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
		return &megaport.PaloAltoConfig{
			Vendor:            "palo_alto",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			SSHPublicKey:      sshPublicKey,
			AdminPasswordHash: adminPasswordHash,
			LicenseData:       licenseData,
		}, nil
	case "prisma":
		ionKey, err := utils.ResourcePrompt("mve", "Enter ION key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		secretKey, err := utils.ResourcePrompt("mve", "Enter secret key (required): ", noColor)
		if err != nil {
			return nil, err
		}
		return &megaport.PrismaConfig{
			Vendor:      "prisma",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			IONKey:      ionKey,
			SecretKey:   secretKey,
		}, nil
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
		return &megaport.VersaConfig{
			Vendor:            "versa",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			DirectorAddress:   directorAddress,
			ControllerAddress: controllerAddress,
			LocalAuth:         localAuth,
			RemoteAuth:        remoteAuth,
			SerialNumber:      serialNumber,
		}, nil
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
		return &megaport.VmwareConfig{
			Vendor:            "vmware",
			ImageID:           imageID,
			ProductSize:       productSize,
			MVELabel:          mveLabel,
			AdminSSHPublicKey: adminSSHPublicKey,
			SSHPublicKey:      sshPublicKey,
			VcoAddress:        vcoAddress,
			VcoActivationCode: vcoActivationCode,
		}, nil
	case "meraki":
		token, err := utils.ResourcePrompt("mve", "Enter token (required): ", noColor)
		if err != nil {
			return nil, err
		}
		return &megaport.MerakiConfig{
			Vendor:      "meraki",
			ImageID:     imageID,
			ProductSize: productSize,
			MVELabel:    mveLabel,
			Token:       token,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendorStr)
	}
}

func promptMVEVnics(noColor bool) ([]megaport.MVENetworkInterface, error) {
	vnics := []megaport.MVENetworkInterface{}
	for {
		fmt.Println("\nEnter VNIC details (leave description empty to finish):")
		description, err := utils.ResourcePrompt("mve", "Enter VNIC description: ", noColor)
		if err != nil {
			return nil, err
		}
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

	return vnics, nil
}

func promptForUpdateMVEDetails(mveUID string, noColor bool) (*megaport.ModifyMVERequest, error) {
	req := &megaport.ModifyMVERequest{
		MVEID: mveUID,
	}

	name, err := utils.ResourcePrompt("mve", "Enter new MVE name (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if name != "" {
		req.Name = name
	}

	costCentre, err := utils.ResourcePrompt("mve", "Enter new cost centre (leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	if costCentre != "" {
		req.CostCentre = costCentre
	}

	contractTermStr, err := utils.ResourcePrompt("mve", fmt.Sprintf("Enter new contract term (%s months, leave empty to keep current): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
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

	if err := validation.ValidateUpdateMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

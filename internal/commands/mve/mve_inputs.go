package mve

import (
	"encoding/json"
	"fmt"
	"os"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// Process JSON input (either from string or file) for buying MVE
func processJSONBuyMVEInput(jsonStr, jsonFilePath string) (*megaport.BuyMVERequest, error) {
	var jsonData map[string]interface{}
	var err error

	if jsonStr != "" {
		// Parse JSON from string
		err = json.Unmarshal([]byte(jsonStr), &jsonData)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON string: %v", err)
		}
	} else if jsonFilePath != "" {
		// Read and parse JSON from file
		jsonBytes, err := os.ReadFile(jsonFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
		err = json.Unmarshal(jsonBytes, &jsonData)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON from file: %v", err)
		}
	}

	// Create and populate the request with the parsed JSON data
	req := &megaport.BuyMVERequest{}

	// Map the JSON fields to the request struct
	if name, ok := jsonData["name"].(string); ok && name != "" {
		req.Name = name
	}

	if term, ok := jsonData["term"].(float64); ok && term > 0 {
		req.Term = int(term)
	}

	if locationID, ok := jsonData["locationId"].(float64); ok && locationID > 0 {
		req.LocationID = int(locationID)
	}

	if diversityZone, ok := jsonData["diversityZone"].(string); ok {
		req.DiversityZone = diversityZone
	}

	if promoCode, ok := jsonData["promoCode"].(string); ok {
		req.PromoCode = promoCode
	}

	if costCentre, ok := jsonData["costCentre"].(string); ok {
		req.CostCentre = costCentre
	}

	// Process vendor config
	if vendorConfigMap, ok := jsonData["vendorConfig"].(map[string]interface{}); ok {
		vendorConfig, err := parseVendorConfig(vendorConfigMap)
		if err != nil {
			return nil, err
		}
		req.VendorConfig = vendorConfig
	}

	// Process vnics
	if vnicsData, ok := jsonData["vnics"].([]interface{}); ok {
		vnics := make([]megaport.MVENetworkInterface, 0, len(vnicsData))
		for _, vnicData := range vnicsData {
			if vnicMap, ok := vnicData.(map[string]interface{}); ok {
				vnic := megaport.MVENetworkInterface{}

				if description, ok := vnicMap["description"].(string); ok {
					vnic.Description = description
				}

				if vlan, ok := vnicMap["vlan"].(float64); ok {
					vnic.VLAN = int(vlan)
				}

				vnics = append(vnics, vnic)
			}
		}
		req.Vnics = vnics
	}

	// Validate the request
	if err := validateBuyMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process flag-based input for buying MVE
func processFlagBuyMVEInput(cmd *cobra.Command) (*megaport.BuyMVERequest, error) {
	// Get flag values
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	locationID, _ := cmd.Flags().GetInt("location-id")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	vendorConfigStr, _ := cmd.Flags().GetString("vendor-config")
	vnicsStr, _ := cmd.Flags().GetString("vnics")

	// Parse vendor config JSON string
	var vendorConfig megaport.VendorConfig
	if vendorConfigStr != "" {
		var vendorConfigMap map[string]interface{}
		err := json.Unmarshal([]byte(vendorConfigStr), &vendorConfigMap)
		if err != nil {
			return nil, fmt.Errorf("error parsing vendor-config JSON string: %v", err)
		}
		vendorConfig, err = parseVendorConfig(vendorConfigMap)
		if err != nil {
			return nil, err
		}
	}

	// Parse VNics JSON string
	var vnics []megaport.MVENetworkInterface
	if vnicsStr != "" {
		var vnicsData []interface{}
		err := json.Unmarshal([]byte(vnicsStr), &vnicsData)
		if err != nil {
			return nil, fmt.Errorf("error parsing vnics JSON string: %v", err)
		}

		vnics = make([]megaport.MVENetworkInterface, 0, len(vnicsData))
		for _, vnicData := range vnicsData {
			if vnicMap, ok := vnicData.(map[string]interface{}); ok {
				vnic := megaport.MVENetworkInterface{}

				if description, ok := vnicMap["description"].(string); ok {
					vnic.Description = description
				}

				if vlan, ok := vnicMap["vlan"].(float64); ok {
					vnic.VLAN = int(vlan)
				}

				vnics = append(vnics, vnic)
			}
		}
	}

	// Build the request
	req := &megaport.BuyMVERequest{
		Name:          name,
		Term:          term,
		LocationID:    locationID,
		DiversityZone: diversityZone,
		PromoCode:     promoCode,
		CostCentre:    costCentre,
		VendorConfig:  vendorConfig,
		Vnics:         vnics,
	}

	// Validate the request
	if err := validateBuyMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// parseVendorConfig parses a vendor config map to the appropriate VendorConfig type
func parseVendorConfig(vendorConfigMap map[string]interface{}) (megaport.VendorConfig, error) {
	vendor, ok := vendorConfigMap["vendor"].(string)
	if !ok {
		return nil, fmt.Errorf("vendor field is required in vendor config")
	}

	switch vendor {
	case "6wind":
		return parseSixwindConfig(vendorConfigMap)
	case "aruba":
		return parseArubaConfig(vendorConfigMap)
	case "aviatrix":
		return parseAviatrixConfig(vendorConfigMap)
	case "cisco":
		return parseCiscoConfig(vendorConfigMap)
	case "fortinet":
		return parseFortinetConfig(vendorConfigMap)
	case "palo_alto":
		return parsePaloAltoConfig(vendorConfigMap)
	case "prisma":
		return parsePrismaConfig(vendorConfigMap)
	case "versa":
		return parseVersaConfig(vendorConfigMap)
	case "vmware":
		return parseVmwareConfig(vendorConfigMap)
	case "meraki":
		return parseMerakiConfig(vendorConfigMap)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// Parse vendor config functions
func parseSixwindConfig(config map[string]interface{}) (*megaport.SixwindVSRConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for 6WIND configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for 6WIND configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	sshPublicKey, ok := getStringFromMap(config, "sshPublicKey")
	if !ok {
		return nil, fmt.Errorf("sshPublicKey is required for 6WIND configuration")
	}

	return &megaport.SixwindVSRConfig{
		Vendor:       "6wind",
		ImageID:      imageID,
		ProductSize:  productSize,
		MVELabel:     mveLabel,
		SSHPublicKey: sshPublicKey,
	}, nil
}

func parseArubaConfig(config map[string]interface{}) (*megaport.ArubaConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Aruba configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Aruba configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	accountName, ok := getStringFromMap(config, "accountName")
	if !ok {
		return nil, fmt.Errorf("accountName is required for Aruba configuration")
	}

	accountKey, ok := getStringFromMap(config, "accountKey")
	if !ok {
		return nil, fmt.Errorf("accountKey is required for Aruba configuration")
	}

	systemTag, _ := getStringFromMap(config, "systemTag")

	return &megaport.ArubaConfig{
		Vendor:      "aruba",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		AccountName: accountName,
		AccountKey:  accountKey,
		SystemTag:   systemTag,
	}, nil
}

func parseAviatrixConfig(config map[string]interface{}) (*megaport.AviatrixConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Aviatrix configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Aviatrix configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	cloudInit, ok := getStringFromMap(config, "cloudInit")
	if !ok {
		return nil, fmt.Errorf("cloudInit is required for Aviatrix configuration")
	}

	return &megaport.AviatrixConfig{
		Vendor:      "aviatrix",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		CloudInit:   cloudInit,
	}, nil
}

func parseCiscoConfig(config map[string]interface{}) (*megaport.CiscoConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Cisco configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Cisco configuration")
	}

	mveLabel, ok := getStringFromMap(config, "mveLabel")
	if !ok {
		return nil, fmt.Errorf("mveLabel is required for Cisco configuration")
	}

	manageLocally, ok := getBoolFromMap(config, "manageLocally")
	if !ok {
		return nil, fmt.Errorf("manageLocally is required for Cisco configuration")
	}

	adminSSHPublicKey, ok := getStringFromMap(config, "adminSshPublicKey")
	if !ok {
		return nil, fmt.Errorf("adminSshPublicKey is required for Cisco configuration")
	}

	sshPublicKey, ok := getStringFromMap(config, "sshPublicKey")
	if !ok {
		return nil, fmt.Errorf("sshPublicKey is required for Cisco configuration")
	}

	cloudInit, ok := getStringFromMap(config, "cloudInit")
	if !ok {
		return nil, fmt.Errorf("cloudInit is required for Cisco configuration")
	}

	fmcIPAddress, ok := getStringFromMap(config, "fmcIpAddress")
	if !ok {
		return nil, fmt.Errorf("fmcIpAddress is required for Cisco configuration")
	}

	fmcRegistrationKey, ok := getStringFromMap(config, "fmcRegistrationKey")
	if !ok {
		return nil, fmt.Errorf("fmcRegistrationKey is required for Cisco configuration")
	}

	fmcNatID, ok := getStringFromMap(config, "fmcNatId")
	if !ok {
		return nil, fmt.Errorf("fmcNatId is required for Cisco configuration")
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
}

func parseFortinetConfig(config map[string]interface{}) (*megaport.FortinetConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Fortinet configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Fortinet configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	adminSSHPublicKey, ok := getStringFromMap(config, "adminSshPublicKey")
	if !ok {
		return nil, fmt.Errorf("adminSshPublicKey is required for Fortinet configuration")
	}

	sshPublicKey, ok := getStringFromMap(config, "sshPublicKey")
	if !ok {
		return nil, fmt.Errorf("sshPublicKey is required for Fortinet configuration")
	}

	licenseData, ok := getStringFromMap(config, "licenseData")
	if !ok {
		return nil, fmt.Errorf("licenseData is required for Fortinet configuration")
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
}

func parsePaloAltoConfig(config map[string]interface{}) (*megaport.PaloAltoConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for PaloAlto configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for PaloAlto configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	sshPublicKey, ok := getStringFromMap(config, "sshPublicKey")
	if !ok {
		return nil, fmt.Errorf("sshPublicKey is required for PaloAlto configuration")
	}

	adminPasswordHash, ok := getStringFromMap(config, "adminPasswordHash")
	if !ok {
		return nil, fmt.Errorf("adminPasswordHash is required for PaloAlto configuration")
	}

	licenseData, ok := getStringFromMap(config, "licenseData")
	if !ok {
		return nil, fmt.Errorf("licenseData is required for PaloAlto configuration")
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
}

func parsePrismaConfig(config map[string]interface{}) (*megaport.PrismaConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Prisma configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Prisma configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	ionKey, ok := getStringFromMap(config, "ionKey")
	if !ok {
		return nil, fmt.Errorf("ionKey is required for Prisma configuration")
	}

	secretKey, ok := getStringFromMap(config, "secretKey")
	if !ok {
		return nil, fmt.Errorf("secretKey is required for Prisma configuration")
	}

	return &megaport.PrismaConfig{
		Vendor:      "prisma",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		IONKey:      ionKey,
		SecretKey:   secretKey,
	}, nil
}

func parseVersaConfig(config map[string]interface{}) (*megaport.VersaConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Versa configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Versa configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	directorAddress, ok := getStringFromMap(config, "directorAddress")
	if !ok {
		return nil, fmt.Errorf("directorAddress is required for Versa configuration")
	}

	controllerAddress, ok := getStringFromMap(config, "controllerAddress")
	if !ok {
		return nil, fmt.Errorf("controllerAddress is required for Versa configuration")
	}

	localAuth, ok := getStringFromMap(config, "localAuth")
	if !ok {
		return nil, fmt.Errorf("localAuth is required for Versa configuration")
	}

	remoteAuth, ok := getStringFromMap(config, "remoteAuth")
	if !ok {
		return nil, fmt.Errorf("remoteAuth is required for Versa configuration")
	}

	serialNumber, ok := getStringFromMap(config, "serialNumber")
	if !ok {
		return nil, fmt.Errorf("serialNumber is required for Versa configuration")
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
}

func parseVmwareConfig(config map[string]interface{}) (*megaport.VmwareConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for VMware configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for VMware configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	adminSSHPublicKey, ok := getStringFromMap(config, "adminSshPublicKey")
	if !ok {
		return nil, fmt.Errorf("adminSshPublicKey is required for VMware configuration")
	}

	sshPublicKey, ok := getStringFromMap(config, "sshPublicKey")
	if !ok {
		return nil, fmt.Errorf("sshPublicKey is required for VMware configuration")
	}

	vcoAddress, ok := getStringFromMap(config, "vcoAddress")
	if !ok {
		return nil, fmt.Errorf("vcoAddress is required for VMware configuration")
	}

	vcoActivationCode, ok := getStringFromMap(config, "vcoActivationCode")
	if !ok {
		return nil, fmt.Errorf("vcoActivationCode is required for VMware configuration")
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
}

func parseMerakiConfig(config map[string]interface{}) (*megaport.MerakiConfig, error) {
	imageID, ok := getImageIDFromMap(config)
	if !ok {
		return nil, fmt.Errorf("imageId is required for Meraki configuration")
	}

	productSize, ok := getStringFromMap(config, "productSize")
	if !ok {
		return nil, fmt.Errorf("productSize is required for Meraki configuration")
	}

	mveLabel, _ := getStringFromMap(config, "mveLabel")

	token, ok := getStringFromMap(config, "token")
	if !ok {
		return nil, fmt.Errorf("token is required for Meraki configuration")
	}

	return &megaport.MerakiConfig{
		Vendor:      "meraki",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		Token:       token,
	}, nil
}

// Helper functions for parsing maps
func getStringFromMap(m map[string]interface{}, key string) (string, bool) {
	val, ok := m[key]
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func getImageIDFromMap(m map[string]interface{}) (int, bool) {
	val, ok := m["imageId"].(float64)
	if !ok {
		return 0, false
	}
	return int(val), true
}

func getBoolFromMap(m map[string]interface{}, key string) (bool, bool) {
	val, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// Process JSON input (either from string or file) for updating MVE
func processJSONUpdateMVEInput(jsonStr, jsonFilePath, mveUID string) (*megaport.ModifyMVERequest, error) {
	var jsonData map[string]interface{}
	var err error

	if jsonStr != "" {
		// Parse JSON from string
		err = json.Unmarshal([]byte(jsonStr), &jsonData)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON string: %v", err)
		}
	} else if jsonFilePath != "" {
		// Read and parse JSON from file
		jsonBytes, err := os.ReadFile(jsonFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
		err = json.Unmarshal(jsonBytes, &jsonData)
		if err != nil {
			return nil, fmt.Errorf("error parsing JSON from file: %v", err)
		}
	}

	// Create and populate the request with the parsed JSON data
	req := &megaport.ModifyMVERequest{
		MVEID: mveUID,
	}

	// Map the JSON fields to the request struct
	if name, ok := jsonData["name"].(string); ok && name != "" {
		req.Name = name
	}

	if costCentre, ok := jsonData["costCentre"].(string); ok && costCentre != "" {
		req.CostCentre = costCentre
	}

	if contractTermMonths, ok := jsonData["contractTermMonths"].(float64); ok && contractTermMonths > 0 {
		termMonths := int(contractTermMonths)
		req.ContractTermMonths = &termMonths
	}

	// Validate the request
	if err := validateUpdateMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process flag-based input for updating MVE
func processFlagUpdateMVEInput(cmd *cobra.Command, mveUID string) (*megaport.ModifyMVERequest, error) {
	// Get flag values
	name, _ := cmd.Flags().GetString("name")
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	contractTerm, _ := cmd.Flags().GetInt("contract-term")

	// Build the request
	req := &megaport.ModifyMVERequest{
		MVEID: mveUID,
	}

	if name != "" {
		req.Name = name
	}

	if costCentre != "" {
		req.CostCentre = costCentre
	}

	if contractTerm > 0 {
		req.ContractTermMonths = &contractTerm
	}

	// Validate the request
	if err := validateUpdateMVERequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

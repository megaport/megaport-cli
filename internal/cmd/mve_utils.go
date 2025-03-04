package cmd

import (
	"context"
	"fmt"
	"strconv"

	megaport "github.com/megaport/megaportgo"
)

// filterMVEImages filters the provided MVE images based on the given filters.
func filterMVEImages(images []*megaport.MVEImage, vendor, productCode string, id int, version string, releaseImage bool) []*megaport.MVEImage {
	var filtered []*megaport.MVEImage
	for _, image := range images {
		if vendor != "" && image.Vendor != vendor {
			continue
		}
		if productCode != "" && image.ProductCode != productCode {
			continue
		}
		if id != 0 && image.ID != id {
			continue
		}
		if version != "" && image.Version != version {
			continue
		}
		if releaseImage && !image.ReleaseImage {
			continue
		}
		filtered = append(filtered, image)
	}
	return filtered
}

// MVEOutput represents the desired fields for JSON output.
type MVEOutput struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
}

// ToMVEOutput converts an MVE to an MVEOutput.
func ToMVEOutput(m *megaport.MVE) (MVEOutput, error) {
	if m == nil {
		return MVEOutput{}, fmt.Errorf("invalid MVE: nil value")
	}

	return MVEOutput{
		UID:        m.UID,
		Name:       m.Name,
		LocationID: m.LocationID,
	}, nil
}

// printMVEs prints the MVEs in the specified output format.
func printMVEs(mves []*megaport.MVE, format string) error {
	if mves == nil {
		mves = []*megaport.MVE{}
	}

	outputs := make([]MVEOutput, 0, len(mves))
	for _, mve := range mves {
		output, err := ToMVEOutput(mve)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}

// buyMVEFunc allows you to purchase an MVE by providing the necessary details.
var buyMVEFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	return client.MVEService.BuyMVE(ctx, req)
}

// Prompts for MVE Vendor Configs

func promptSixwindConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	sshPublicKey, err := prompt("Enter SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.SixwindVSRConfig{
		Vendor:       "6WIND",
		ImageID:      imageID,
		ProductSize:  productSize,
		MVELabel:     mveLabel,
		SSHPublicKey: sshPublicKey,
	}, nil
}

func promptArubaConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	accountName, err := prompt("Enter Account Name (required): ")
	if err != nil {
		return nil, err
	}

	accountKey, err := prompt("Enter Account Key (required): ")
	if err != nil {
		return nil, err
	}

	systemTag, err := prompt("Enter System Tag (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.ArubaConfig{
		Vendor:      "Aruba",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		AccountName: accountName,
		AccountKey:  accountKey,
		SystemTag:   systemTag,
	}, nil
}

func promptAviatrixConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (required): ")
	if err != nil {
		return nil, err
	}

	cloudInit, err := prompt("Enter Cloud Init (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.AviatrixConfig{
		Vendor:      "Aviatrix",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		CloudInit:   cloudInit,
	}, nil
}

func promptCiscoConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	manageLocallyStr, err := prompt("Manage Locally (true/false) (required): ")
	if err != nil {
		return nil, err
	}
	manageLocally, err := strconv.ParseBool(manageLocallyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid value for Manage Locally")
	}

	adminSSHPublicKey, err := prompt("Enter Admin SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	sshPublicKey, err := prompt("Enter SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	cloudInit, err := prompt("Enter Cloud Init (required): ")
	if err != nil {
		return nil, err
	}

	fmcIPAddress, err := prompt("Enter FMC IP Address (required): ")
	if err != nil {
		return nil, err
	}

	fmcRegistrationKey, err := prompt("Enter FMC Registration Key (required): ")
	if err != nil {
		return nil, err
	}

	fmcNatID, err := prompt("Enter FMC NAT ID (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.CiscoConfig{
		Vendor:             "Cisco",
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

func promptFortinetConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	adminSSHPublicKey, err := prompt("Enter Admin SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	sshPublicKey, err := prompt("Enter SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	licenseData, err := prompt("Enter License Data (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.FortinetConfig{
		Vendor:            "Fortinet",
		ImageID:           imageID,
		ProductSize:       productSize,
		MVELabel:          mveLabel,
		AdminSSHPublicKey: adminSSHPublicKey,
		SSHPublicKey:      sshPublicKey,
		LicenseData:       licenseData,
	}, nil
}

func promptPaloAltoConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (optional): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	adminSSHPublicKey, err := prompt("Enter Admin SSH Public Key (optional): ")
	if err != nil {
		return nil, err
	}

	sshPublicKey, err := prompt("Enter SSH Public Key (optional): ")
	if err != nil {
		return nil, err
	}

	adminPasswordHash, err := prompt("Enter Admin Password Hash (optional): ")
	if err != nil {
		return nil, err
	}

	licenseData, err := prompt("Enter License Data (optional): ")
	if err != nil {
		return nil, err
	}

	return &megaport.PaloAltoConfig{
		Vendor:            "PaloAlto",
		ImageID:           imageID,
		ProductSize:       productSize,
		MVELabel:          mveLabel,
		AdminSSHPublicKey: adminSSHPublicKey,
		SSHPublicKey:      sshPublicKey,
		AdminPasswordHash: adminPasswordHash,
		LicenseData:       licenseData,
	}, nil
}

func promptPrismaConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (required): ")
	if err != nil {
		return nil, err
	}

	ionKey, err := prompt("Enter ION Key (required): ")
	if err != nil {
		return nil, err
	}

	secretKey, err := prompt("Enter Secret Key (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.PrismaConfig{
		Vendor:      "Prisma",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		IONKey:      ionKey,
		SecretKey:   secretKey,
	}, nil
}

func promptVersaConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	directorAddress, err := prompt("Enter Director Address (required): ")
	if err != nil {
		return nil, err
	}

	controllerAddress, err := prompt("Enter Controller Address (required): ")
	if err != nil {
		return nil, err
	}

	localAuth, err := prompt("Enter Local Auth (required): ")
	if err != nil {
		return nil, err
	}

	remoteAuth, err := prompt("Enter Remote Auth (required): ")
	if err != nil {
		return nil, err
	}

	serialNumber, err := prompt("Enter Serial Number (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VersaConfig{
		Vendor:            "Versa",
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

func promptVmwareConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	adminSSHPublicKey, err := prompt("Enter Admin SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	sshPublicKey, err := prompt("Enter SSH Public Key (required): ")
	if err != nil {
		return nil, err
	}

	vcoAddress, err := prompt("Enter VCO Address (required): ")
	if err != nil {
		return nil, err
	}

	vcoActivationCode, err := prompt("Enter VCO Activation Code (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VmwareConfig{
		Vendor:            "VMware",
		ImageID:           imageID,
		ProductSize:       productSize,
		MVELabel:          mveLabel,
		AdminSSHPublicKey: adminSSHPublicKey,
		SSHPublicKey:      sshPublicKey,
		VcoAddress:        vcoAddress,
		VcoActivationCode: vcoActivationCode,
	}, nil
}

func promptMerakiConfig() (megaport.VendorConfig, error) {
	imageIDStr, err := prompt("Enter Image ID (required): ")
	if err != nil {
		return nil, err
	}
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Image ID")
	}

	productSize, err := prompt("Enter Product Size (required): ")
	if err != nil {
		return nil, err
	}

	mveLabel, err := prompt("Enter MVE Label (optional): ")
	if err != nil {
		return nil, err
	}

	token, err := prompt("Enter Token (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.MerakiConfig{
		Vendor:      "Meraki",
		ImageID:     imageID,
		ProductSize: productSize,
		MVELabel:    mveLabel,
		Token:       token,
	}, nil
}

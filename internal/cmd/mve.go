package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// mveCmd is the base command for all Megaport Virtual Edge (MVE) operations.
var mveCmd = &cobra.Command{
	Use:   "mve",
	Short: "Manage MVEs in the Megaport API",
	Long: `Manage MVEs in the Megaport API.

This command groups all operations related to Megaport Virtual Edge devices (MVEs).
Use the "megaport mve get [mveUID]" command to fetch details for a specific MVE identified by its UID.
`,
}

var buyMVEFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyMVERequest) (*megaport.BuyMVEResponse, error) {
	return client.MVEService.BuyMVE(ctx, req)
}

// buyMVECmd allows you to purchase an MVE by providing the necessary details.
var buyMVECmd = &cobra.Command{
	Use:   "buy",
	Short: "Buy an MVE through the Megaport API",
	Long: `Buy an MVE through the Megaport API.

This command allows you to purchase an MVE by providing the necessary details.
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the MVE.
  - term: The term of the MVE (1, 12, 24, or 36 months).
  - location_id: The ID of the location where the MVE will be provisioned.
  - vendor: The vendor of the MVE.

Example usage:

  megaport mve buy
`,
	RunE: BuyMVE,
}

func BuyMVE(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Prompt for required fields
	name, err := prompt("Enter MVE name (required): ")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("MVE name is required")
	}

	termStr, err := prompt("Enter term (1, 12, 24, 36) (required): ")
	if err != nil {
		return err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}

	locationIDStr, err := prompt("Enter location ID (required): ")
	if err != nil {
		return err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return fmt.Errorf("invalid location ID")
	}

	vendor, err := prompt("Enter vendor (required): ")
	if err != nil {
		return err
	}
	if vendor == "" {
		return fmt.Errorf("vendor is required")
	}

	// Prompt for vendor-specific configuration
	var vendorConfig megaport.VendorConfig
	switch strings.ToLower(vendor) {
	case "6wind":
		vendorConfig, err = promptSixwindConfig()
	case "aruba":
		vendorConfig, err = promptArubaConfig()
	case "aviatrix":
		vendorConfig, err = promptAviatrixConfig()
	case "cisco":
		vendorConfig, err = promptCiscoConfig()
	case "fortinet":
		vendorConfig, err = promptFortinetConfig()
	case "paloalto":
		vendorConfig, err = promptPaloAltoConfig()
	case "prisma":
		vendorConfig, err = promptPrismaConfig()
	case "versa":
		vendorConfig, err = promptVersaConfig()
	case "vmware":
		vendorConfig, err = promptVmwareConfig()
	case "meraki":
		vendorConfig, err = promptMerakiConfig()
	default:
		return fmt.Errorf("unsupported vendor: %s", vendor)
	}
	if err != nil {
		return err
	}

	// Create the BuyMVERequest
	req := &megaport.BuyMVERequest{
		Name:         name,
		Term:         term,
		LocationID:   locationID,
		VendorConfig: vendorConfig,
	}

	// Call the BuyMVE method
	client, err := Login(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Buying MVE...")
	if err := client.MVEService.ValidateMVEOrder(ctx, req); err != nil {
		return fmt.Errorf("validation failed: %v", err)
	}

	resp, err := buyMVEFunc(ctx, client, req)
	if err != nil {
		return err
	}

	fmt.Printf("MVE purchased successfully - UID: %s\n", resp.TechnicalServiceUID)
	return nil
}

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

// getMVECmd retrieves details for a single MVE.
var getMVECmd = &cobra.Command{
	Use:   "get [mveUID]",
	Short: "Get details for a single MVE",
	Args:  cobra.ExactArgs(1),
	RunE:  GetMVE,
}

func GetMVE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]
	if mveUID == "" {
		return fmt.Errorf("MVE UID cannot be empty")
	}

	mve, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		return fmt.Errorf("error getting MVE: %v", err)
	}

	if mve == nil {
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func init() {
	mveCmd.AddCommand(buyMVECmd)
	mveCmd.AddCommand(getMVECmd)
	rootCmd.AddCommand(mveCmd)
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

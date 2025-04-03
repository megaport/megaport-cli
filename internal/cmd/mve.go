package cmd

import (
	"github.com/spf13/cobra"
)

var (
	vendorFilter       string
	productCodeFilter  string
	idFilter           int
	versionFilter      string
	releaseImageFilter bool
)

// mveCmd is the base command for all Megaport Virtual Edge (MVE) operations.
var mveCmd = &cobra.Command{
	Use:   "mve",
	Short: "Manage Megaport Virtual Edge (MVE) devices",
	Long: `Manage Megaport Virtual Edge (MVE) devices.

This command groups all operations related to Megaport Virtual Edge devices (MVEs).
MVEs are virtual networking appliances that run in the Megaport network, providing 
software-defined networking capabilities from various vendors.

With MVEs you can:
- Deploy virtual networking appliances without physical hardware
- Create secure connections between cloud services
- Implement SD-WAN solutions across multiple regions
- Run vendor-specific networking software in Megaport's infrastructure

Available operations:
- list: List all MVEs in your account
- get: Retrieve details for a specific MVE
- buy: Purchase a new MVE with vendor-specific configuration
- update: Modify an existing MVE's properties
- delete: Remove an MVE from your account
- list-images: View available MVE software images
- list-sizes: View available MVE hardware configurations

Examples:
  # List all MVEs
  megaport-cli mve list

  # Get details for a specific MVE
  megaport-cli mve get [mveUID]

  # Buy a new MVE
  megaport-cli mve buy

  # Update an existing MVE
  megaport-cli mve update [mveUID]

  # Delete an existing MVE
  megaport-cli mve delete [mveUID]
`,
}

// buyMVECmd allows you to purchase an MVE by providing the necessary details.
var buyMVECmd = &cobra.Command{
	Use:   "buy",
	Short: "Purchase a new Megaport Virtual Edge (MVE) device",
	Long: `Purchase a new Megaport Virtual Edge (MVE) device through the Megaport API.

This command allows you to purchase an MVE by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (with --interactive):
 The command will prompt you for each required and optional field.

2. Flag Mode:
 Provide all required fields as flags:
 --name, --term, --location-id, --vendor-config, --vnics

3. JSON Mode:
 Provide a JSON string or file with all required fields:
 --json <json-string> or --json-file <path>

Required fields:
- name: The name of the MVE.
- term: The term of the MVE (1, 12, 24, or 36 months).
- location_id: The ID of the location where the MVE will be provisioned.
- vendor-config: JSON string with vendor-specific configuration (for flag mode)
- vnics: JSON array of network interfaces (for flag mode)

Vendor-specific configuration details:
--------------------------------------

6WIND (SixwindVSRConfig):
- vendor: Must be "6wind"
- imageId: The ID of the 6WIND image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: Custom label for the MVE
- sshPublicKey: SSH public key for access

Aruba (ArubaConfig):
- vendor: Must be "aruba"
- imageId: The ID of the Aruba image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- accountName: Aruba account name
- accountKey: Aruba authentication key
- systemTag: System tag for pre-configuration

Aviatrix (AviatrixConfig):
- vendor: Must be "aviatrix"
- imageId: The ID of the Aviatrix image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: Custom label for the MVE
- cloudInit: Cloud-init configuration script

Cisco (CiscoConfig):
- vendor: Must be "cisco"
- imageId: The ID of the Cisco image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- manageLocally: Boolean flag to manage locally (true/false)
- adminSshPublicKey: Admin SSH public key
- sshPublicKey: User SSH public key
- cloudInit: Cloud-init configuration script
- fmcIpAddress: Firewall Management Center IP address
- fmcRegistrationKey: Registration key for FMC
- fmcNatId: NAT ID for FMC

Fortinet (FortinetConfig):
- vendor: Must be "fortinet"
- imageId: The ID of the Fortinet image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- adminSshPublicKey: Admin SSH public key
- sshPublicKey: User SSH public key
- licenseData: License data for the Fortinet instance

PaloAlto (PaloAltoConfig):
- vendor: Must be "paloalto"
- imageId: The ID of the PaloAlto image to use
- productSize: (Optional) Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- adminSshPublicKey: (Optional) Admin SSH public key
- sshPublicKey: (Optional) SSH public key for access
- adminPasswordHash: (Optional) Hashed admin password
- licenseData: (Optional) License data for the PaloAlto instance

Prisma (PrismaConfig):
- vendor: Must be "prisma"
- imageId: The ID of the Prisma image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: Custom label for the MVE
- ionKey: ION key for authentication
- secretKey: Secret key for authentication

Versa (VersaConfig):
- vendor: Must be "versa"
- imageId: The ID of the Versa image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- directorAddress: Versa director address
- controllerAddress: Versa controller address
- localAuth: Local authentication string
- remoteAuth: Remote authentication string
- serialNumber: Serial number for the device

VMware (VmwareConfig):
- vendor: Must be "vmware"
- imageId: The ID of the VMware image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- adminSshPublicKey: Admin SSH public key
- sshPublicKey: User SSH public key
- vcoAddress: VCO address for configuration
- vcoActivationCode: Activation code for VCO

Meraki (MerakiConfig):
- vendor: Must be "meraki"
- imageId: The ID of the Meraki image to use
- productSize: Size of the virtual machine (SMALL, MEDIUM, LARGE)
- mveLabel: (Optional) Custom label for the MVE
- token: Authentication token

Example usage:

# Interactive mode
megaport-cli mve buy --interactive

# JSON mode - Complete example with full schema
megaport-cli mve buy --json '{
  "name": "My MVE Display Name",
  "term": 12,
  "locationId": 123,
  "diversityZone": "zone-1",
  "promoCode": "PROMO2023",
  "costCentre": "Marketing Dept",
  "vendorConfig": {
    "vendor": "cisco",
    "imageId": 123,
    "productSize": "MEDIUM",
    "mveLabel": "custom-label",
    "manageLocally": true,
    "adminSshPublicKey": "ssh-rsa AAAA...",
    "sshPublicKey": "ssh-rsa AAAA...",
    "cloudInit": "#cloud-config\npackages:\n - nginx\n",
    "fmcIpAddress": "10.0.0.1",
    "fmcRegistrationKey": "key123",
    "fmcNatId": "natid123"
  },
  "vnics": [
    {"description": "Data Plane", "vlan": 100},
    {"description": "Management", "vlan": 200}
  ]
}'

Notes:
- For production deployments, you may want to use a JSON file to manage complex configurations
- To list available images and their IDs, use: megaport-cli mve list-images
- To list available sizes, use: megaport-cli mve list-sizes
- Location IDs can be retrieved with: megaport-cli locations list
`,
	RunE: WrapRunE(BuyMVE),
}

// getMVECmd retrieves details for a single MVE.
var getMVECmd = &cobra.Command{
	Use:   "get [mveUID]",
	Short: "Get details for a single MVE",
	Long: `Get details for a single MVE from the Megaport API.

This command fetches and displays detailed information about a specific MVE.
You need to provide the UID of the MVE as an argument.

Example usage:

  megaport-cli mve get [mveUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(GetMVE),
}

// updateMVECmd updates an existing Megaport Virtual Edge (MVE).
var updateMVECmd = &cobra.Command{
	Use:   "update [mveUID]",
	Short: "Update an existing MVE",
	Long: `Update an existing Megaport Virtual Edge (MVE).

This command allows you to update the details of an existing MVE.
You can provide details in one of three ways:

1. Interactive Mode (default):
 The command will prompt you for each field you can update.

2. Flag Mode:
 Provide the fields you want to update as flags:
   --name, --cost-centre, --contract-term

3. JSON Mode:
 Provide a JSON string or file with the fields you want to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- name: The new name of the MVE.
- cost_centre: The new cost center for the MVE.
- contract_term_months: The new contract term in months (1, 12, 24, or 36).

Example usage:

# Interactive mode (default)
megaport-cli mve update [mveUID]

# Flag mode
megaport-cli mve update [mveUID] --name "New MVE Name" --cost-centre "New Cost Centre" --contract-term 24

# JSON mode
megaport-cli mve update [mveUID] --json '{"name": "New MVE Name", "costCentre": "New Cost Centre", "contractTermMonths": 24}'
megaport-cli mve update [mveUID] --json-file ./mve-update.json
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(UpdateMVE),
}

// deleteMVECmd deletes an existing Megaport Virtual Edge (MVE).
var deleteMVECmd = &cobra.Command{
	Use:   "delete [mveUID]",
	Short: "Delete an existing MVE",
	Long: `Delete an existing Megaport Virtual Edge (MVE).

This command allows you to delete an existing MVE by providing its UID.

Example usage:

  megaport-cli mve delete [mveUID]
`,
	Args: cobra.ExactArgs(1),
	RunE: WrapRunE(DeleteMVE),
}

// listMVEImagesCmd lists all available MVE images.
var listMVEImagesCmd = &cobra.Command{
	Use:   "list-images",
	Short: "List all available MVE images",
	Long: `List all available MVE images from the Megaport API.

This command fetches and displays a list of all available MVE images with details such as
image ID, version, product, and vendor. You can filter the images based on vendor, product code, ID, version, or release image.

Available filters:
  - vendor: Filter images by vendor.
  - product-code: Filter images by product code.
  - id: Filter images by ID.
  - version: Filter images by version.
  - release-image: Filter images by release image.

Example usage:

  megaport-cli mve list-images --vendor "Cisco" --product-code "CISCO123" --id 1 --version "1.0" --release-image true
`,
	RunE: WrapRunE(ListMVEImages),
}

// listAvailableMVESizesCmd lists all available MVE sizes.
var listAvailableMVESizesCmd = &cobra.Command{
	Use:   "list-sizes",
	Short: "List all available MVE sizes",
	Long: `List all available MVE sizes from the Megaport API.

This command fetches and displays a list of all available MVE sizes with details such as
size, label, CPU core count, and RAM.

Example usage:

  megaport-cli mve list-sizes
`,
	RunE: WrapRunE(ListAvailableMVESizes),
}

func init() {
	buyMVECmd.Flags().BoolP("interactive", "i", false, "Use interactive mode with prompts")
	buyMVECmd.Flags().String("name", "", "MVE name")
	buyMVECmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
	buyMVECmd.Flags().Int("location-id", 0, "Location ID where the MVE will be provisioned")
	buyMVECmd.Flags().String("diversity-zone", "", "Diversity zone for the MVE")
	buyMVECmd.Flags().String("promo-code", "", "Promotional code for discounts")
	buyMVECmd.Flags().String("cost-centre", "", "Cost centre for billing")
	buyMVECmd.Flags().String("vendor-config", "", "JSON string containing vendor-specific configuration")
	buyMVECmd.Flags().String("vnics", "", "JSON array of network interfaces")
	buyMVECmd.Flags().String("json", "", "JSON string containing MVE configuration")
	buyMVECmd.Flags().String("json-file", "", "Path to JSON file containing MVE configuration")

	updateMVECmd.Flags().BoolP("interactive", "i", true, "Use interactive mode with prompts")
	updateMVECmd.Flags().String("name", "", "New MVE name")
	updateMVECmd.Flags().String("cost-centre", "", "New cost centre")
	updateMVECmd.Flags().Int("contract-term", 0, "New contract term in months (1, 12, 24, or 36)")
	updateMVECmd.Flags().String("json", "", "JSON string containing MVE update configuration")
	updateMVECmd.Flags().String("json-file", "", "Path to JSON file containing MVE update configuration")

	listMVEImagesCmd.Flags().StringVar(&vendorFilter, "vendor", "", "Filter images by vendor")
	listMVEImagesCmd.Flags().StringVar(&productCodeFilter, "product-code", "", "Filter images by product code")
	listMVEImagesCmd.Flags().IntVar(&idFilter, "id", 0, "Filter images by ID")
	listMVEImagesCmd.Flags().StringVar(&versionFilter, "version", "", "Filter images by version")
	listMVEImagesCmd.Flags().BoolVar(&releaseImageFilter, "release-image", false, "Filter images by release image")

	mveCmd.AddCommand(buyMVECmd)
	mveCmd.AddCommand(getMVECmd)
	mveCmd.AddCommand(updateMVECmd)
	mveCmd.AddCommand(deleteMVECmd)
	mveCmd.AddCommand(listMVEImagesCmd)
	mveCmd.AddCommand(listAvailableMVESizesCmd)
	rootCmd.AddCommand(mveCmd)
}

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

// listMVEsCmd lists all Megaport Virtual Edge (MVE) devices.
var listMVEsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all MVEs",
	Long: `List all Megaport Virtual Edge (MVE) devices associated with your Megaport account.

This command retrieves all MVEs from the Megaport API and displays them in the specified format.
By default, inactive MVEs are excluded. Use the --inactive flag to include them.

Example usage:

# List all active MVEs
megaport-cli mve list

# List all MVEs including inactive ones
megaport-cli mve list --inactive

# List all MVEs in JSON format
megaport-cli mve list --output json
`,
	RunE: WrapRunE(ListMVEs),
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

This command allows you to update specific properties of an existing MVE without
disrupting its service or connectivity. Updates apply immediately but may take
a few minutes to fully propagate in the Megaport system.

You can provide update details in one of three ways:

1. Interactive Mode (default):
   The command will prompt you for each updatable field, showing current values
   and allowing you to make changes. Press ENTER to keep the current value.

2. Flag Mode:
   Provide only the fields you want to update as flags. Fields not specified
   will remain unchanged:
   --name, --cost-centre, --contract-term

3. JSON Mode:
   Provide a JSON string or file with the fields you want to update:
   --json <json-string> or --json-file <path>

Fields that can be updated:
- name: The new name of the MVE (1-64 characters)
- cost_centre: The new cost center for billing purposes (optional)
- contract_term_months: The new contract term in months (1, 12, 24, or 36)

Important notes:
- The MVE UID cannot be changed
- Vendor configuration cannot be changed after provisioning
- Technical specifications (size, location) cannot be modified
- Connectivity (VXCs) will not be affected by these changes
- Changing the contract term may affect billing immediately

Example usage:

# Interactive mode (default)
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p

# Flag mode
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name "Edge Router West" --cost-centre "IT-Network-2023" --contract-term 24

# JSON mode with string
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{"name": "Edge Router West", "costCentre": "IT-Network-2023", "contractTermMonths": 24}'

# JSON mode with file
megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./mve-update.json

JSON format example (mve-update.json):
{
  "name": "Edge Router West",
  "costCentre": "IT-Network-2023",
  "contractTermMonths": 24
}

Note the JSON property names differ from flag names:
- Flag: --name             → JSON: "name"
- Flag: --cost-centre      → JSON: "costCentre"
- Flag: --contract-term    → JSON: "contractTermMonths"

Example successful output:
  MVE updated successfully:
  UID:          1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p
  Name:         Edge Router West (previously "Edge Router")  
  Cost Centre:  IT-Network-2023 (previously "IT-Network")
  Term:         24 months (previously 12 months)
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

This command fetches and displays a list of all available MVE images with details
about each one. These images are used when creating new MVEs with the 'buy' command.

The output includes:
- ID: Unique identifier required for the 'buy' command
- Vendor: The network function vendor (e.g., Cisco, Fortinet, Palo Alto)
- Product: Specific product name (e.g., C8000, FortiGate-VM, VM-Series)
- Version: Software version of the image
- Release: Whether this is a production release image (true) or development/beta (false)
- Sizes: Available instance sizes (SMALL, MEDIUM, LARGE, X_LARGE_12)
- Description: Additional vendor-specific information when available

Available filters:
  --vendor string        Filter images by vendor name (e.g., "Cisco", "Fortinet")
  --product-code string  Filter images by product code
  --id int               Filter images by exact image ID
  --version string       Filter images by version string
  --release-image        Only show official release images (excludes beta/development)

Example usage:

  # List all available images
  megaport-cli mve list-images

  # List only Cisco images
  megaport-cli mve list-images --vendor "Cisco"

  # List only release (production) images for Fortinet
  megaport-cli mve list-images --vendor "Fortinet" --release-image

Example output:
  +-----+----------+----------------------------------+--------------+-------+----------------------+-------------------------+
  | ID  |  VENDOR  |            PRODUCT               |   VERSION    | RELEAS |        SIZES         |      DESCRIPTION        |
  +-----+----------+----------------------------------+--------------+-------+----------------------+-------------------------+
  | 83  | Cisco    | C8000                            | 17.15.01a    | true  | SMALL,MEDIUM,LARGE   |                         |
  | 78  | Cisco    | Secure Firewall Threat Defense   | 7.4.2-172    | true  | MEDIUM,LARGE         |                         |
  | 57  | Fortinet | FortiGate-VM                     | 7.0.14       | true  | SMALL,MEDIUM,LARGE   |                         |
  | 65  | Palo Alto| VM-Series                        | 10.2.9-h1    | true  | SMALL,MEDIUM,LARGE   |                         |
  | 88  | Palo Alto| Prisma SD-WAN 310xv              | vION 3102v-  | true  | SMALL                | Requires MVE Size 2/8   |
  | 62  | Meraki   | vMX                              | 20231214     | false | SMALL,MEDIUM,LARGE   | Engineering Build - Not |
  +-----+----------+----------------------------------+--------------+-------+----------------------+-------------------------+
`,
	RunE: WrapRunE(ListMVEImages),
}

// listAvailableMVESizesCmd lists all available MVE sizes.
var listAvailableMVESizesCmd = &cobra.Command{
	Use:   "list-sizes",
	Short: "List all available MVE sizes",
	Long: `List all available MVE sizes from the Megaport API.

This command fetches and displays details about all available MVE instance sizes.
The size you select determines the MVE's capabilities and compute resources.

Each size includes the following specifications:
- Size: Size identifier used when creating an MVE (e.g., SMALL, MEDIUM, LARGE)
- Label: Human-readable name (e.g., "MVE 2/8", "MVE 4/16")
- CPU: Number of virtual CPU cores
- RAM: Amount of memory in GB
- Max CPU Count: Maximum CPU cores available for the size

Standard MVE sizes available across most vendors:
- SMALL: 2 vCPU, 8GB RAM
- MEDIUM: 4 vCPU, 16GB RAM
- LARGE: 8 vCPU, 32GB RAM
- X_LARGE_12: 12 vCPU, 48GB RAM

Note: Not all sizes are available for all vendor images. Some vendors or specific
products may have restrictions on which sizes can be used. Check the image details
using 'megaport-cli mve list-images' for size compatibility.

Example usage:

  megaport-cli mve list-sizes
  
Example output:
  +------------+------------+----------+---------+
  |    SIZE    |   LABEL    |   CPU    |   RAM   |
  +------------+------------+----------+---------+
  | SMALL      | MVE 2/8    | 2 vCPU   | 8 GB    |
  | MEDIUM     | MVE 4/16   | 4 vCPU   | 16 GB   |
  | LARGE      | MVE 8/32   | 8 vCPU   | 32 GB   |
  | X_LARGE_12 | MVE 12/48  | 12 vCPU  | 48 GB   |
  +------------+------------+----------+---------+
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

	listMVEsCmd.Flags().Bool("inactive", false, "Include inactive MVEs in the list")

	mveCmd.AddCommand(listMVEsCmd)
	mveCmd.AddCommand(buyMVECmd)
	mveCmd.AddCommand(getMVECmd)
	mveCmd.AddCommand(updateMVECmd)
	mveCmd.AddCommand(deleteMVECmd)
	mveCmd.AddCommand(listMVEImagesCmd)
	mveCmd.AddCommand(listAvailableMVESizesCmd)
	rootCmd.AddCommand(mveCmd)
}

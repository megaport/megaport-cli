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
You can use this command to list, get details, buy, update, and delete MVEs.

Examples:
  # List all MVEs
  megaport mve list

  # Get details for a specific MVE
  megaport mve get [mveUID]

  # Buy a new MVE
  megaport mve buy

  # Update an existing MVE
  megaport mve update [mveUID]

  # Delete an existing MVE
  megaport mve delete [mveUID]
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

Vendor-specific fields:
- 6WIND:
- image_id (required)
- product_size (required)
- mve_label (optional)
- ssh_public_key (required)
- Aruba:
- image_id (required)
- product_size (required)
- mve_label (optional)
- account_name (required)
- account_key (required)
- system_tag (optional)
- Aviatrix:
- image_id (required)
- product_size (required)
- mve_label (optional)
- cloud_init (required)
- Cisco:
- image_id (required)
- product_size (required)
- mve_label (required)
- manage_locally (required, true/false)
- admin_ssh_public_key (required)
- ssh_public_key (required)
- cloud_init (required)
- fmc_ip_address (required)
- fmc_registration_key (required)
- fmc_nat_id (required)
- Fortinet:
- image_id (required)
- product_size (required)
- mve_label (optional)
- admin_ssh_public_key (required)
- ssh_public_key (required)
- ha_license (required for HA deployments)
- license_data (required for non-HA deployments)
- PaloAlto:
- image_id (required)
- product_size (required)
- mve_label (optional)
- ssh_public_key (required)
- admin_password_hash (required)
- license_data (required)
- Prisma:
- image_id (required)
- product_size (required)
- mve_label (optional)
- ion_key (required)
- secret_key (required)
- Versa:
- image_id (required)
- product_size (required)
- mve_label (optional)
- director_address (required)
- controller_address (required)
- local_auth (required)
- remote_auth (required)
- serial_number (required)
- VMware:
- image_id (required)
- product_size (required)
- mve_label (optional)
- admin_ssh_public_key (required)
- ssh_public_key (required)
- vco_address (required)
- vco_activation_code (required)
- Meraki:
- image_id (required)
- product_size (required)
- mve_label (optional)
- token (required)

Example usage:

# Interactive mode
megaport mve buy --interactive

# Flag mode - Cisco example
megaport mve buy --name "My Cisco MVE" --term 12 --location-id 123 \
  --vendor-config '{"vendor":"cisco","imageId":1,"productSize":"large","mveLabel":"cisco-mve",
                   "manageLocally":true,"adminSshPublicKey":"ssh-rsa AAAA...","sshPublicKey":"ssh-rsa AAAA...",
                   "cloudInit":"#cloud-config\npackages:\n - nginx\n","fmcIpAddress":"10.0.0.1",
                   "fmcRegistrationKey":"key123","fmcNatId":"natid123"}' \
  --vnics '[{"description":"Data Plane","vlan":100}]'

# Flag mode - Aruba example
megaport mve buy --name "Megaport MVE Example" --term 1 --location-id 123 \
  --vendor-config '{"vendor":"aruba","imageId":23,"productSize":"MEDIUM",
                   "accountName":"Aruba Test Account","accountKey":"12345678",
                   "systemTag":"Preconfiguration-aruba-test-1"}' \
  --vnics '[{"description":"Data Plane"},{"description":"Control Plane"},{"description":"Management Plane"}]'

# Flag mode - Versa example
megaport mve buy --name "Megaport Versa MVE Example" --term 1 --location-id 123 \
  --vendor-config '{"vendor":"versa","imageId":20,"productSize":"MEDIUM",
                   "directorAddress":"director1.versa.com","controllerAddress":"controller1.versa.com",
                   "localAuth":"SDWAN-Branch@Versa.com","remoteAuth":"Controller-1-staging@Versa.com",
                   "serialNumber":"Megaport-Hub1"}' \
  --vnics '[{"description":"Data Plane"}]'


# JSON mode - Cisco example
megaport mve buy --json '{
"name": "My Cisco MVE",
"term": 12,
"locationId": 67,
"vendorConfig": {
  "vendor": "cisco",
  "imageId": 1,
  "productSize": "large",
  "mveLabel": "cisco-mve",
  "manageLocally": true,
  "adminSshPublicKey": "ssh-rsa AAAA...",
  "sshPublicKey": "ssh-rsa AAAA...",
  "cloudInit": "#cloud-config\npackages:\n - nginx\n",
  "fmcIpAddress": "10.0.0.1",
  "fmcRegistrationKey": "key123",
  "fmcNatId": "natid123"
},
"vnics": [
  {"description": "Data Plane", "vlan": 100}
]
}'

# JSON mode - Aruba example
megaport mve buy --json '{
"name": "Megaport MVE Example",
"term": 1,
"locationId": 67,
"vendorConfig": {
  "vendor": "aruba",
  "imageId": 23,
  "productSize": "MEDIUM",
  "accountName": "Aruba Test Account",
  "accountKey": "12345678",
  "systemTag": "Preconfiguration-aruba-test-1"
},
"vnics": [
  {"description": "Data Plane"},
  {"description": "Control Plane"},
  {"description": "Management Plane"}
]
}'

# JSON mode - Versa example
megaport mve buy --json '{
"name": "Megaport Versa MVE Example",
"term": 1,
"locationId": 67,
"vendorConfig": {
  "vendor": "versa",
  "imageId": 20,
  "productSize": "MEDIUM",
  "directorAddress": "director1.versa.com",
  "controllerAddress": "controller1.versa.com",
  "localAuth": "SDWAN-Branch@Versa.com",
  "remoteAuth": "Controller-1-staging@Versa.com",
  "serialNumber": "Megaport-Hub1"
},
"vnics": [
  {"description": "Data Plane"}
]
}'

# JSON from file
megaport mve buy --json-file ./mve-config.json
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

  megaport mve get [mveUID]
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
megaport mve update [mveUID]

# Flag mode
megaport mve update [mveUID] --name "New MVE Name" --cost-centre "New Cost Centre" --contract-term 24

# JSON mode
megaport mve update [mveUID] --json '{"name": "New MVE Name", "costCentre": "New Cost Centre", "contractTermMonths": 24}'
megaport mve update [mveUID] --json-file ./mve-update.json
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

  megaport mve delete [mveUID]
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

  megaport mve list-images --vendor "Cisco" --product-code "CISCO123" --id 1 --version "1.0" --release-image true
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

  megaport mve list-sizes
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
	buyMVECmd.Flags().String("resource-tags", "", "JSON string of key-value resource tags")

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

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
You will be prompted to enter the required and optional fields.

Required fields:
  - name: The name of the MVE.
  - term: The term of the MVE (1, 12, 24, or 36 months).
  - location_id: The ID of the location where the MVE will be provisioned.
  - vendor: The vendor of the MVE. Available values are:
    - 6WIND
    - Aruba
    - Aviatrix
    - Cisco
    - Fortinet
    - PaloAlto
    - Prisma
    - Versa
    - VMware
    - Meraki

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
    - license_data (required)
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

  # Purchase a new MVE
  megaport mve buy
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
You will be prompted to enter the new values for the fields you want to update.

Fields that can be updated:
  - name: The new name of the MVE.
  - cost_center: The new cost center for the MVE.
  - contract_term_months: The new contract term in months.

Example usage:

  megaport mve update [mveUID]
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

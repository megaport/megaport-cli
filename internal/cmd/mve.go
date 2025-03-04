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
	Short: "Manage MVEs in the Megaport API",
	Long: `Manage MVEs in the Megaport API.

This command groups all operations related to Megaport Virtual Edge devices (MVEs).
Use the "megaport mve get [mveUID]" command to fetch details for a specific MVE identified by its UID.

Examples:
  megaport mve list
  megaport mve get [mveUID]
  megaport mve buy
`,
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

Example usage:

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
	mveCmd.AddCommand(listMVEImagesCmd)
	mveCmd.AddCommand(listAvailableMVESizesCmd)
	rootCmd.AddCommand(mveCmd)
}

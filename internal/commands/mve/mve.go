package mve

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the mve commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create mve parent command
	mveCmd := cmdbuilder.NewCommand("mve", "Manage Megaport Virtual Edge (MVE) devices").
		WithLongDesc("Manage Megaport Virtual Edge (MVE) devices.\n\nThis command groups all operations related to Megaport Virtual Edge devices (MVEs). MVEs are virtual networking appliances that run in the Megaport network, providing software-defined networking capabilities from various vendors.").
		WithExample("mve list").
		WithExample("mve get [mveUID]").
		WithExample("mve buy").
		WithExample("mve update [mveUID]").
		WithExample("mve delete [mveUID]").
		WithImportantNote("With MVEs you can deploy virtual networking appliances without physical hardware").
		WithImportantNote("Create secure connections between cloud services").
		WithImportantNote("Run vendor-specific networking software in Megaport's infrastructure").
		WithRootCmd(rootCmd).
		Build()

	// Create buy MVE command
	buyMVECmd := cmdbuilder.NewCommand("buy", "Purchase a new Megaport Virtual Edge (MVE) device").
		WithColorAwareRunFunc(BuyMVE).
		WithBoolFlagP("interactive", "i", false, "Use interactive mode with prompts").
		WithMVECreateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Purchase a new Megaport Virtual Edge (MVE) device through the Megaport API.\n\nThis command allows you to purchase an MVE by providing the necessary details.").
		WithExample("buy --interactive").
		WithExample("buy --json '{\"name\":\"My MVE\",\"term\":12,\"locationId\":123,\"vendorConfig\":{\"vendor\":\"cisco\",\"imageId\":123,\"productSize\":\"MEDIUM\"},\"vnics\":[{\"description\":\"Data Plane\",\"vlan\":100}]}'").
		WithExample("buy --json-file ./mve-config.json").
		WithJSONExample(`{
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
    "cloudInit": "#cloud-config\npackages:\n - nginx\n"
  },
  "vnics": [
    {"description": "Data Plane", "vlan": 100},
    {"description": "Management", "vlan": 200}
  ]
}`).
		WithImportantNote("For production deployments, you may want to use a JSON file to manage complex configurations").
		WithImportantNote("To list available images and their IDs, use: megaport-cli mve list-images").
		WithImportantNote("To list available sizes, use: megaport-cli mve list-sizes").
		WithImportantNote("Location IDs can be retrieved with: megaport-cli locations list").
		WithRootCmd(rootCmd).
		Build()

	// Create get MVE command
	getMVECmd := cmdbuilder.NewCommand("get", "Get details for a single MVE").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(GetMVE).
		WithLongDesc("Get details for a single MVE from the Megaport API.\n\nThis command fetches and displays detailed information about a specific MVE. You need to provide the UID of the MVE as an argument.").
		WithExample("get [mveUID]").
		WithRootCmd(rootCmd).
		Build()

	// Create update MVE command
	updateMVECmd := cmdbuilder.NewCommand("update", "Update an existing MVE").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateMVE).
		WithBoolFlagP("interactive", "i", true, "Use interactive mode with prompts").
		WithMVEUpdateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Update an existing Megaport Virtual Edge (MVE).\n\nThis command allows you to update specific properties of an existing MVE without disrupting its service or connectivity. Updates apply immediately but may take a few minutes to fully propagate in the Megaport system.").
		WithOptionalFlag("name", "The new name of the MVE (1-64 characters)").
		WithOptionalFlag("cost-centre", "The new cost centre for billing purposes (optional)").
		WithOptionalFlag("contract-term", "The new contract term in months (1, 12, 24, or 36)").
		WithExample("update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithExample("update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name \"Edge Router West\" --cost-centre \"IT-Network-2023\" --contract-term 24").
		WithExample("update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{\"name\": \"Edge Router West\", \"costCentre\": \"IT-Network-2023\", \"contractTermMonths\": 24}'").
		WithExample("update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./mve-update.json").
		WithJSONExample(`{
  "name": "Edge Router West",
  "costCentre": "IT-Network-2023",
  "contractTermMonths": 24
}`).
		WithImportantNote("The MVE UID cannot be changed").
		WithImportantNote("Vendor configuration cannot be changed after provisioning").
		WithImportantNote("Technical specifications (size, location) cannot be modified").
		WithImportantNote("Connectivity (VXCs) will not be affected by these changes").
		WithImportantNote("Changing the contract term may affect billing immediately").
		WithRootCmd(rootCmd).
		Build()

	// Create delete MVE command
	deleteMVECmd := cmdbuilder.NewCommand("delete", "Delete an existing MVE").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(DeleteMVE).
		WithDeleteFlags().
		WithLongDesc("Delete an existing Megaport Virtual Edge (MVE).\n\nThis command allows you to delete an existing MVE by providing its UID.").
		WithExample("delete [mveUID]").
		WithExample("delete [mveUID] --force").
		WithExample("delete [mveUID] --now").
		WithRootCmd(rootCmd).
		Build()

	// Create list MVE images command
	listMVEImagesCmd := cmdbuilder.NewCommand("list-images", "List all available MVE images").
		WithOutputFormatRunFunc(ListMVEImages).
		WithMVEImageFilterFlags().
		WithLongDesc("List all available MVE images from the Megaport API.\n\nThis command fetches and displays a list of all available MVE images with details about each one. These images are used when creating new MVEs with the 'buy' command.").
		WithOptionalFlag("vendor", "Filter images by vendor name (e.g., \"Cisco\", \"Fortinet\")").
		WithOptionalFlag("product-code", "Filter images by product code").
		WithOptionalFlag("id", "Filter images by exact image ID").
		WithOptionalFlag("version", "Filter images by version string").
		WithOptionalFlag("release-image", "Only show official release images (excludes beta/development)").
		WithExample("list-images").
		WithExample("list-images --vendor \"Cisco\"").
		WithExample("list-images --vendor \"Fortinet\" --release-image").
		WithImportantNote("The output includes the image ID, vendor, product, version, release status, available sizes, and description").
		WithImportantNote("The ID field is required when specifying an image in the 'buy' command").
		WithRootCmd(rootCmd).
		Build()

	// Create list MVE sizes command
	listAvailableMVESizesCmd := cmdbuilder.NewCommand("list-sizes", "List all available MVE sizes").
		WithOutputFormatRunFunc(ListAvailableMVESizes).
		WithLongDesc("List all available MVE sizes from the Megaport API.\n\nThis command fetches and displays details about all available MVE instance sizes. The size you select determines the MVE's capabilities and compute resources.").
		WithExample("list-sizes").
		WithImportantNote("Standard MVE sizes available across most vendors: SMALL (2 vCPU, 8GB RAM), MEDIUM (4 vCPU, 16GB RAM), LARGE (8 vCPU, 32GB RAM), X_LARGE_12 (12 vCPU, 48GB RAM)").
		WithImportantNote("Not all sizes are available for all vendor images. Check the image details using 'megaport-cli mve list-images' for size compatibility").
		WithRootCmd(rootCmd).
		Build()

	// Add commands to their parents
	mveCmd.AddCommand(
		buyMVECmd,
		getMVECmd,
		updateMVECmd,
		deleteMVECmd,
		listMVEImagesCmd,
		listAvailableMVESizesCmd,
	)
	rootCmd.AddCommand(mveCmd)
}

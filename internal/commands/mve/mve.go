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
		WithExample("megaport-cli mve list").
		WithExample("megaport-cli mve get [mveUID]").
		WithExample("megaport-cli mve buy").
		WithExample("megaport-cli mve update [mveUID]").
		WithExample("megaport-cli mve delete [mveUID]").
		WithImportantNote("With MVEs you can deploy virtual networking appliances without physical hardware").
		WithImportantNote("Create secure connections between cloud services").
		WithImportantNote("Run vendor-specific networking software in Megaport's infrastructure").
		WithRootCmd(rootCmd).
		Build()

		// Create buy MVE command
	buyMVECmd := cmdbuilder.NewCommand("buy", "Purchase a new Megaport Virtual Edge (MVE) device").
		WithColorAwareRunFunc(BuyMVE).
		WithInteractiveFlag().
		WithMVECreateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Purchase a new Megaport Virtual Edge (MVE) device through the Megaport API.\n\nThis command allows you to purchase an MVE by providing the necessary details.").
		WithDocumentedRequiredFlag("name", "The name of the MVE").
		WithDocumentedRequiredFlag("term", "The term of the MVE (1, 12, 24, or 36 months)").
		WithDocumentedRequiredFlag("location-id", "The ID of the location where the MVE will be provisioned").
		WithDocumentedRequiredFlag("vendor-config", "JSON string with vendor-specific configuration (for flag mode)").
		WithDocumentedRequiredFlag("vnics", "JSON array of network interfaces (for flag mode)").
		WithOptionalFlag("diversity-zone", "The diversity zone for the MVE").
		WithOptionalFlag("promo-code", "Promotional code for discounts").
		WithOptionalFlag("cost-centre", "Cost centre for billing").
		WithExample("megaport-cli mve buy --interactive").
		WithExample("megaport-cli mve buy --name \"My MVE\" --term 12 --location-id 123 --vendor-config '{\"vendor\":\"cisco\",\"imageId\":123,\"productSize\":\"MEDIUM\"}' --vnics '[{\"description\":\"Data Plane\",\"vlan\":100}]' --resource-tags '{\"env\":\"prod\",\"owner\":\"netops\"}'").
		WithExample("megaport-cli mve buy --json '{\"name\":\"My MVE\",\"term\":12,\"locationId\":123,\"vendorConfig\":{\"vendor\":\"cisco\",\"imageId\":123,\"productSize\":\"MEDIUM\"},\"vnics\":[{\"description\":\"Data Plane\",\"vlan\":100}],\"resourceTags\":{\"env\":\"prod\",\"owner\":\"netops\"}}'").
		WithExample("megaport-cli mve buy --json-file ./mve-config.json").
		WithJSONExample(`{
  "name": "My MVE Display Name",
  "term": 12,
  "locationId": 123,
  "diversityZone": "blue",
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
  ],
  "resourceTags": { // Add resourceTags to JSON example
    "environment": "production",
    "billing_code": "BC12345",
    "owner_team": "network-operations"
  }
}`).
		WithImportantNote("For production deployments, you may want to use a JSON file to manage complex configurations").
		WithImportantNote("To list available images and their IDs, use: megaport-cli mve list-images").
		WithImportantNote("To list available sizes, use: megaport-cli mve list-sizes").
		WithImportantNote("Location IDs can be retrieved with: megaport-cli locations list").
		WithImportantNote("Resource tags are key-value pairs used for organizing and managing resources.").
		WithRootCmd(rootCmd).
		WithConditionalRequirements("name", "term", "location-id", "vendor-config", "vnics").
		Build()

	// Create get MVE command
	getMVECmd := cmdbuilder.NewCommand("get", "Get details for a single MVE").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(GetMVE).
		WithLongDesc("Get details for a single MVE from the Megaport API.\n\nThis command retrieves and displays detailed information for a single Megaport Virtual Edge (MVE). You must provide the unique identifier (UID) of the MVE you wish to retrieve.").
		WithExample("megaport-cli mve get a1b2c3d4-e5f6-7890-1234-567890abcdef").
		WithImportantNote("The output includes the MVE's UID, name, vendor, version, status, and connectivity details").
		WithRootCmd(rootCmd).
		Build()

	// Create update MVE command
	updateMVECmd := cmdbuilder.NewCommand("update", "Update an existing MVE").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateMVE).
		WithInteractiveFlag().
		WithMVEUpdateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Update an existing Megaport Virtual Edge (MVE).\n\nThis command allows you to update specific properties of an existing MVE without disrupting its service or connectivity. Updates apply immediately but may take a few minutes to fully propagate in the Megaport system.").
		WithOptionalFlag("name", "The new name of the MVE (1-64 characters)").
		WithOptionalFlag("cost-centre", "The new cost centre for billing purposes").
		WithOptionalFlag("contract-term", "The new contract term in months (1, 12, 24, or 36)").
		WithExample("megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithExample("megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name \"Edge Router West\" --cost-centre \"IT-Network-2023\" --contract-term 24").
		WithExample("megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{\"name\": \"Edge Router West\", \"costCentre\": \"IT-Network-2023\", \"contractTermMonths\": 24}'").
		WithExample("megaport-cli mve update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./mve-update.json").
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
		WithExample("megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p").
		WithExample("megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --force").
		WithExample("megaport-cli mve delete 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --now").
		WithImportantNote("Deletion is final and cannot be undone").
		WithImportantNote("Billing for the MVE stops at the end of the current billing period unless --now is specified").
		WithImportantNote("All associated VXCs will be automatically terminated").
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
		WithExample("megaport-cli mve list-images").
		WithExample("megaport-cli mve list-images --vendor \"Cisco\"").
		WithExample("megaport-cli mve list-images --vendor \"Fortinet\" --release-image").
		WithImportantNote("The output includes the image ID, vendor, product, version, release status, available sizes, and description").
		WithImportantNote("The ID field is required when specifying an image in the 'buy' command").
		WithRootCmd(rootCmd).
		Build()

	// Create list MVE sizes command
	listAvailableMVESizesCmd := cmdbuilder.NewCommand("list-sizes", "List all available MVE sizes").
		WithOutputFormatRunFunc(ListAvailableMVESizes).
		WithLongDesc("List all available MVE sizes from the Megaport API.\n\nThis command fetches and displays details about all available MVE instance sizes. The size you select determines the MVE's capabilities and compute resources.").
		WithExample("megaport-cli mve list-sizes").
		WithImportantNote("Standard MVE sizes available across most vendors: SMALL (2 vCPU, 8GB RAM), MEDIUM (4 vCPU, 16GB RAM), LARGE (8 vCPU, 32GB RAM), X_LARGE_12 (12 vCPU, 48GB RAM)").
		WithImportantNote("Not all sizes are available for all vendor images. Check the image details using 'megaport-cli mve list-images' for size compatibility").
		WithRootCmd(rootCmd).
		Build()

		// Create list MVEs command
	listMVEsCmd := cmdbuilder.NewCommand("list", "List all MVEs with optional filters").
		WithOutputFormatRunFunc(ListMVEs).
		WithMVEFilterFlags().
		WithLongDesc("List all MVEs available in the Megaport API.\n\nThis command fetches and displays a list of MVEs with details such as MVE ID, name, location, vendor, and status. By default, only active MVEs are shown.").
		WithExample("megaport-cli mve list").
		WithExample("megaport-cli mve list --location-id 123").
		WithExample("megaport-cli mve list --vendor \"Cisco\"").
		WithExample("megaport-cli mve list --name \"Edge Router\"").
		WithExample("megaport-cli mve list --include-inactive").
		WithExample("megaport-cli mve list --location-id 123 --vendor \"Cisco\" --name \"Edge\"").
		WithRootCmd(rootCmd).
		Build()

	// Add list-tags command
	listTagsCmd := cmdbuilder.NewCommand("list-tags", "List resource tags on a specific MVE").
		WithLongDesc("Lists all resource tags associated with a specific MVE").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(ListMVEResourceTags).
		WithExample("megaport-cli mve list-tags mve-abc123").
		Build()

	// Add update-tags command
	updateTagsCmd := cmdbuilder.NewCommand("update-tags", "Update resource tags on a specific MVE").
		WithLongDesc("Update resource tags associated with a specific MVE. Tags can be provided via interactive prompts, JSON string, or JSON file.").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateMVEResourceTags).
		WithStandardInputFlags().
		WithExample("megaport-cli mve update-tags mve-abc123 --interactive").
		WithExample("megaport-cli mve update-tags mve-abc123 --json '{\"env\":\"production\",\"team\":\"network\"}'").
		WithExample("megaport-cli mve update-tags mve-abc123 --json-file ./tags.json").
		WithImportantNote("All existing tags will be replaced with the provided tags. To clear all tags, provide an empty tag set.").
		Build()

	// Add commands to their parents
	mveCmd.AddCommand(
		buyMVECmd,
		getMVECmd,
		updateMVECmd,
		deleteMVECmd,
		listMVEImagesCmd,
		listAvailableMVESizesCmd,
		listMVEsCmd,
		listTagsCmd,   // Add list-tags
		updateTagsCmd, // Add update-tags
	)
	rootCmd.AddCommand(mveCmd)
}

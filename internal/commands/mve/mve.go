package mve

import (
	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/utils"
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
}

// buyMVECmd allows you to purchase an MVE by providing the necessary details.
var buyMVECmd = &cobra.Command{
	Use:   "buy",
	Short: "Purchase a new Megaport Virtual Edge (MVE) device",
	RunE:  utils.WrapColorAwareRunE(BuyMVE),
}

// getMVECmd retrieves details for a single MVE.
var getMVECmd = &cobra.Command{
	Use:   "get [mveUID]",
	Short: "Get details for a single MVE",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapOutputFormatRunE(GetMVE),
}

// updateMVECmd updates an existing Megaport Virtual Edge (MVE).
var updateMVECmd = &cobra.Command{
	Use:   "update [mveUID]",
	Short: "Update an existing MVE",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(UpdateMVE),
}

// deleteMVECmd deletes an existing Megaport Virtual Edge (MVE).
var deleteMVECmd = &cobra.Command{
	Use:   "delete [mveUID]",
	Short: "Delete an existing MVE",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(DeleteMVE),
}

// listMVEImagesCmd lists all available MVE images.
var listMVEImagesCmd = &cobra.Command{
	Use:   "list-images",
	Short: "List all available MVE images",
	RunE:  utils.WrapOutputFormatRunE(ListMVEImages),
}

// listAvailableMVESizesCmd lists all available MVE sizes.
var listAvailableMVESizesCmd = &cobra.Command{
	Use:   "list-sizes",
	Short: "List all available MVE sizes",
	RunE:  utils.WrapOutputFormatRunE(ListAvailableMVESizes),
}

func init() {

}

func AddCommandsTo(rootCmd *cobra.Command) {
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

	// Set up help builders for commands

	// mve command help
	mveHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve",
		ShortDesc:   "Manage Megaport Virtual Edge (MVE) devices",
		LongDesc:    "Manage Megaport Virtual Edge (MVE) devices.\n\nThis command groups all operations related to Megaport Virtual Edge devices (MVEs). MVEs are virtual networking appliances that run in the Megaport network, providing software-defined networking capabilities from various vendors.",
		Examples: []string{
			"mve list",
			"mve get [mveUID]",
			"mve buy",
			"mve update [mveUID]",
			"mve delete [mveUID]",
		},
		ImportantNotes: []string{
			"With MVEs you can deploy virtual networking appliances without physical hardware",
			"Create secure connections between cloud services",
			"Implement SD-WAN solutions across multiple regions",
			"Run vendor-specific networking software in Megaport's infrastructure",
		},
	}
	mveCmd.Long = mveHelp.Build(rootCmd)

	// buy MVE help
	buyMVEHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve buy",
		ShortDesc:   "Purchase a new Megaport Virtual Edge (MVE) device",
		LongDesc:    "Purchase a new Megaport Virtual Edge (MVE) device through the Megaport API.\n\nThis command allows you to purchase an MVE by providing the necessary details.",
		RequiredFlags: map[string]string{
			"name":          "The name of the MVE",
			"term":          "The term of the MVE (1, 12, 24, or 36 months)",
			"location-id":   "The ID of the location where the MVE will be provisioned",
			"vendor-config": "JSON string with vendor-specific configuration (for flag mode)",
			"vnics":         "JSON array of network interfaces (for flag mode)",
		},
		OptionalFlags: map[string]string{
			"diversity-zone": "The diversity zone for the MVE",
			"promo-code":     "Promotional code for discounts",
			"cost-centre":    "Cost centre for billing",
		},
		Examples: []string{
			"buy --interactive",
			"buy --json '{\"name\":\"My MVE\",\"term\":12,\"locationId\":123,\"vendorConfig\":{\"vendor\":\"cisco\",\"imageId\":123,\"productSize\":\"MEDIUM\"},\"vnics\":[{\"description\":\"Data Plane\",\"vlan\":100}]}'",
			"buy --json-file ./mve-config.json",
		},
		JSONExamples: []string{
			`{
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
}`,
		},
		ImportantNotes: []string{
			"For production deployments, you may want to use a JSON file to manage complex configurations",
			"To list available images and their IDs, use: megaport-cli mve list-images",
			"To list available sizes, use: megaport-cli mve list-sizes",
			"Location IDs can be retrieved with: megaport-cli locations list",
		},
	}
	buyMVECmd.Long = buyMVEHelp.Build(rootCmd)

	// get MVE help
	getMVEHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve get",
		ShortDesc:   "Get details for a single MVE",
		LongDesc:    "Get details for a single MVE from the Megaport API.\n\nThis command fetches and displays detailed information about a specific MVE. You need to provide the UID of the MVE as an argument.",
		Examples: []string{
			"get [mveUID]",
		},
	}
	getMVECmd.Long = getMVEHelp.Build(rootCmd)

	// update MVE help
	updateMVEHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve update",
		ShortDesc:   "Update an existing MVE",
		LongDesc:    "Update an existing Megaport Virtual Edge (MVE).\n\nThis command allows you to update specific properties of an existing MVE without disrupting its service or connectivity. Updates apply immediately but may take a few minutes to fully propagate in the Megaport system.",
		OptionalFlags: map[string]string{
			"name":          "The new name of the MVE (1-64 characters)",
			"cost-centre":   "The new cost center for billing purposes (optional)",
			"contract-term": "The new contract term in months (1, 12, 24, or 36)",
		},
		Examples: []string{
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --name \"Edge Router West\" --cost-centre \"IT-Network-2023\" --contract-term 24",
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json '{\"name\": \"Edge Router West\", \"costCentre\": \"IT-Network-2023\", \"contractTermMonths\": 24}'",
			"update 1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p --json-file ./mve-update.json",
		},
		JSONExamples: []string{
			`{
  "name": "Edge Router West",
  "costCentre": "IT-Network-2023",
  "contractTermMonths": 24
}`,
		},
		ImportantNotes: []string{
			"The MVE UID cannot be changed",
			"Vendor configuration cannot be changed after provisioning",
			"Technical specifications (size, location) cannot be modified",
			"Connectivity (VXCs) will not be affected by these changes",
			"Changing the contract term may affect billing immediately",
		},
	}
	updateMVECmd.Long = updateMVEHelp.Build(rootCmd)

	// delete MVE help
	deleteMVEHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve delete",
		ShortDesc:   "Delete an existing MVE",
		LongDesc:    "Delete an existing Megaport Virtual Edge (MVE).\n\nThis command allows you to delete an existing MVE by providing its UID.",
		Examples: []string{
			"delete [mveUID]",
		},
	}
	deleteMVECmd.Long = deleteMVEHelp.Build(rootCmd)

	// list MVE images help
	listMVEImagesHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve list-images",
		ShortDesc:   "List all available MVE images",
		LongDesc:    "List all available MVE images from the Megaport API.\n\nThis command fetches and displays a list of all available MVE images with details about each one. These images are used when creating new MVEs with the 'buy' command.",
		OptionalFlags: map[string]string{
			"vendor":        "Filter images by vendor name (e.g., \"Cisco\", \"Fortinet\")",
			"product-code":  "Filter images by product code",
			"id":            "Filter images by exact image ID",
			"version":       "Filter images by version string",
			"release-image": "Only show official release images (excludes beta/development)",
		},
		Examples: []string{
			"list-images",
			"list-images --vendor \"Cisco\"",
			"list-images --vendor \"Fortinet\" --release-image",
		},
		ImportantNotes: []string{
			"The output includes the image ID, vendor, product, version, release status, available sizes, and description",
			"The ID field is required when specifying an image in the 'buy' command",
		},
	}
	listMVEImagesCmd.Long = listMVEImagesHelp.Build(rootCmd)

	// list MVE sizes help
	listMVESizesHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli mve list-sizes",
		ShortDesc:   "List all available MVE sizes",
		LongDesc:    "List all available MVE sizes from the Megaport API.\n\nThis command fetches and displays details about all available MVE instance sizes. The size you select determines the MVE's capabilities and compute resources.",
		Examples: []string{
			"list-sizes",
		},
		ImportantNotes: []string{
			"Standard MVE sizes available across most vendors: SMALL (2 vCPU, 8GB RAM), MEDIUM (4 vCPU, 16GB RAM), LARGE (8 vCPU, 32GB RAM), X_LARGE_12 (12 vCPU, 48GB RAM)",
			"Not all sizes are available for all vendor images. Check the image details using 'megaport-cli mve list-images' for size compatibility",
		},
	}
	listAvailableMVESizesCmd.Long = listMVESizesHelp.Build(rootCmd)

	mveCmd.AddCommand(buyMVECmd)
	mveCmd.AddCommand(getMVECmd)
	mveCmd.AddCommand(updateMVECmd)
	mveCmd.AddCommand(deleteMVECmd)
	mveCmd.AddCommand(listMVEImagesCmd)
	mveCmd.AddCommand(listAvailableMVESizesCmd)
	rootCmd.AddCommand(mveCmd)
}

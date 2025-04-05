package vxc

import (
	"github.com/megaport/megaport-cli/internal/base/help"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/spf13/cobra"
)

// vxcCmd is the base command for all operations related to Virtual Cross Connects (VXCs).
var vxcCmd = &cobra.Command{
	Use:   "vxc",
	Short: "Manage VXCs in the Megaport API",
}

// getVXCCmd retrieves detailed information for a single Virtual Cross Connect (VXC).
var getVXCCmd = &cobra.Command{
	Use:   "get [vxcUID]",
	Short: "Get details for a single VXC",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapOutputFormatRunE(GetVXC),
}

// buyVXCCmd purchases a new Virtual Cross Connect (VXC).
var buyVXCCmd = &cobra.Command{
	Use:   "buy",
	Short: "Purchase a new Virtual Cross Connect (VXC)",
	RunE:  utils.WrapColorAwareRunE(BuyVXC),
}

// updateVXCCmd updates an existing Virtual Cross Connect (VXC).
var updateVXCCmd = &cobra.Command{
	Use:   "update [vxcUID]",
	Short: "Update an existing Virtual Cross Connect (VXC)",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(UpdateVXC),
}

// deleteVXCCmd deletes an existing Virtual Cross Connect (VXC) in the Megaport API.
var deleteVXCCmd = &cobra.Command{
	Use:   "delete [vxcUID]",
	Short: "Delete an existing Virtual Cross Connect (VXC)",
	Args:  cobra.ExactArgs(1),
	RunE:  utils.WrapColorAwareRunE(DeleteVXC),
}

func AddCommandsTo(rootCmd *cobra.Command) {
	// Add subcommands to vxc command
	vxcCmd.AddCommand(getVXCCmd)
	vxcCmd.AddCommand(buyVXCCmd)
	vxcCmd.AddCommand(updateVXCCmd)
	vxcCmd.AddCommand(deleteVXCCmd)

	// Add flags to buy command
	buyVXCCmd.Flags().String("a-end-uid", "", "UID of the A-End product")
	buyVXCCmd.Flags().String("b-end-uid", "", "UID of the B-End product")
	buyVXCCmd.Flags().String("name", "", "Name of the VXC")
	buyVXCCmd.Flags().Int("rate-limit", 0, "Bandwidth in Mbps")
	buyVXCCmd.Flags().Int("term", 0, "Contract term in months (1, 12, 24, or 36)")
	buyVXCCmd.Flags().Int("a-end-vlan", 0, "VLAN for A-End (0-4093, except 1)")
	buyVXCCmd.Flags().Int("b-end-vlan", 0, "VLAN for B-End (0-4093, except 1)")
	buyVXCCmd.Flags().Int("a-end-inner-vlan", 0, "Inner VLAN for A-End (-1 or higher)")
	buyVXCCmd.Flags().Int("b-end-inner-vlan", 0, "Inner VLAN for B-End (-1 or higher)")
	buyVXCCmd.Flags().Int("a-end-vnic-index", 0, "vNIC index for A-End MVE")
	buyVXCCmd.Flags().Int("b-end-vnic-index", 0, "vNIC index for B-End MVE")
	buyVXCCmd.Flags().String("promo-code", "", "Promotional code")
	buyVXCCmd.Flags().String("service-key", "", "Service key")
	buyVXCCmd.Flags().String("cost-centre", "", "Cost centre")
	buyVXCCmd.Flags().String("a-end-partner-config", "", "JSON string with A-End partner configuration")
	buyVXCCmd.Flags().String("b-end-partner-config", "", "JSON string with B-End partner configuration")
	buyVXCCmd.Flags().String("json", "", "JSON string with all VXC configuration")
	buyVXCCmd.Flags().String("json-file", "", "Path to JSON file with VXC configuration")
	buyVXCCmd.Flags().Bool("interactive", false, "Use interactive mode")

	// Add flags to update command
	updateVXCCmd.Flags().String("name", "", "New name for the VXC")
	updateVXCCmd.Flags().Int("rate-limit", 0, "New bandwidth in Mbps")
	updateVXCCmd.Flags().Int("term", 0, "New contract term in months (1, 12, 24, or 36)")
	updateVXCCmd.Flags().String("cost-centre", "", "New cost centre for billing")
	updateVXCCmd.Flags().Bool("shutdown", false, "Whether to shut down the VXC")
	updateVXCCmd.Flags().Int("a-end-vlan", 0, "New VLAN for A-End (0-4093, except 1)")
	updateVXCCmd.Flags().Int("b-end-vlan", 0, "New VLAN for B-End (0-4093, except 1)")
	updateVXCCmd.Flags().Int("a-end-inner-vlan", 0, "New inner VLAN for A-End")
	updateVXCCmd.Flags().Int("b-end-inner-vlan", 0, "New inner VLAN for B-End")
	updateVXCCmd.Flags().String("a-end-uid", "", "New A-End product UID")
	updateVXCCmd.Flags().String("b-end-uid", "", "New B-End product UID")
	updateVXCCmd.Flags().String("a-end-partner-config", "", "JSON string with A-End VRouter partner configuration")
	updateVXCCmd.Flags().String("b-end-partner-config", "", "JSON string with B-End VRouter partner configuration")
	updateVXCCmd.Flags().String("json", "", "JSON string with update fields")
	updateVXCCmd.Flags().String("json-file", "", "Path to JSON file with update fields")
	updateVXCCmd.Flags().Bool("interactive", false, "Use interactive mode")

	// Set up help builders for commands

	// vxc command help
	vxcHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli vxc",
		ShortDesc:   "Manage VXCs in the Megaport API",
		LongDesc:    "Manage VXCs in the Megaport API.\n\nThis command groups all operations related to Virtual Cross Connects (VXCs). VXCs are virtual point-to-point connections between two ports or devices on the Megaport network. You can use the subcommands to perform actions such as retrieving details, purchasing, updating, and deleting VXCs.",
		Examples: []string{
			"vxc get [vxcUID]",
			"vxc buy",
			"vxc update [vxcUID]",
			"vxc delete [vxcUID]",
		},
	}
	vxcCmd.Long = vxcHelp.Build(rootCmd)

	// get VXC help
	getVXCHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli vxc get",
		ShortDesc:   "Get details for a single VXC",
		LongDesc:    "Get details for a single VXC through the Megaport API.\n\nThis command retrieves detailed information for a single Virtual Cross Connect (VXC). You must provide the unique identifier (UID) of the VXC you wish to retrieve.",
		Examples: []string{
			"get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		},
		ImportantNotes: []string{
			"The output includes the VXC's UID, name, rate limit, A-End and B-End details, status, and cost centre.",
		},
	}
	getVXCCmd.Long = getVXCHelp.Build(rootCmd)

	// buy VXC help
	buyVXCHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli vxc buy",
		ShortDesc:   "Purchase a new Virtual Cross Connect (VXC)",
		LongDesc:    "Purchase a new Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to purchase a VXC by providing the necessary details.",
		RequiredFlags: map[string]string{
			"a-end-uid":  "UID of the A-End product (Port, MCR, MVE)",
			"name":       "Name of the VXC (1-64 characters)",
			"rate-limit": "Bandwidth in Mbps (50 - 10000)",
			"term":       "Contract term in months (1, 12, 24, or 36)",
			"a-end-vlan": "VLAN for A-End (2-4093, except 4090)",
		},
		OptionalFlags: map[string]string{
			"b-end-uid":            "UID of the B-End product (if connecting to non-partner)",
			"b-end-vlan":           "VLAN for B-End (2-4093, except 4090)",
			"a-end-inner-vlan":     "Inner VLAN for A-End (-1 or higher, only for QinQ)",
			"b-end-inner-vlan":     "Inner VLAN for B-End (-1 or higher, only for QinQ)",
			"a-end-vnic-index":     "vNIC index for A-End MVE (required for MVE A-End)",
			"b-end-vnic-index":     "vNIC index for B-End MVE (required for MVE B-End)",
			"promo-code":           "Promotional code",
			"service-key":          "Service key",
			"cost-centre":          "Cost centre",
			"a-end-partner-config": "JSON string with A-End partner configuration (for VRouter)",
			"b-end-partner-config": "JSON string with B-End partner configuration (for CSPs like AWS, Azure)",
		},
		Examples: []string{
			"buy --interactive",
			"buy --a-end-uid \"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" --b-end-uid \"port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy\" --name \"My VXC\" --rate-limit 1000 --term 12 --a-end-vlan 100 --b-end-vlan 200",
			"buy --a-end-uid \"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" --name \"My AWS VXC\" --rate-limit 1000 --term 12 --a-end-vlan 100 --b-end-partner-config '{\"connectType\":\"AWS\",\"ownerAccount\":\"123456789012\",\"asn\":65000,\"amazonAsn\":64512}'",
			"buy --a-end-uid \"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" --name \"My Azure VXC\" --rate-limit 1000 --term 12 --a-end-vlan 100 --b-end-partner-config '{\"connectType\":\"AZURE\",\"serviceKey\":\"s-abcd1234\"}'",
			"buy --json '{\"aEndUid\":\"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\",\"name\":\"My VXC\",\"rateLimit\":1000,\"term\":12,\"aEndConfiguration\":{\"vlan\":100},\"bEndConfiguration\":{\"productUid\":\"port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy\",\"vlan\":200}}'",
		},
		JSONExamples: []string{
			`{
  "aEndUid": "port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "name": "My VXC",
  "rateLimit": 1000,
  "term": 12,
  "aEndConfiguration": {"vlan": 100},
  "bEndConfiguration": {
    "productUid": "port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "vlan": 200
  }
}`,
		},
		ImportantNotes: []string{
			"For AWS connections, you must provide owner account, ASN, and Amazon ASN in b-end-partner-config",
			"For Azure connections, you must provide a service key in b-end-partner-config",
			"QinQ VLANs require both outer and inner VLANs",
			"MVE connections require specifying vNIC indexes",
		},
	}
	buyVXCCmd.Long = buyVXCHelp.Build(rootCmd)

	// update VXC help
	updateVXCHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli vxc update",
		ShortDesc:   "Update an existing Virtual Cross Connect (VXC)",
		LongDesc:    "Update an existing Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to update an existing VXC by providing the necessary details.",
		OptionalFlags: map[string]string{
			"name":                 "New name for the VXC (1-64 characters)",
			"rate-limit":           "New bandwidth in Mbps (50 - 10000)",
			"term":                 "New contract term in months (1, 12, 24, or 36)",
			"cost-centre":          "New cost centre for billing",
			"shutdown":             "Whether to shut down the VXC (true/false)",
			"a-end-vlan":           "New VLAN for A-End (2-4093, except 4090)",
			"b-end-vlan":           "New VLAN for B-End (2-4093, except 4090)",
			"a-end-inner-vlan":     "New inner VLAN for A-End (-1 or higher, only for QinQ)",
			"b-end-inner-vlan":     "New inner VLAN for B-End (-1 or higher, only for QinQ)",
			"a-end-uid":            "New A-End product UID",
			"b-end-uid":            "New B-End product UID",
			"a-end-partner-config": "JSON string with A-End VRouter partner configuration",
			"b-end-partner-config": "JSON string with B-End VRouter partner configuration",
		},
		Examples: []string{
			"update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive",
			"update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --name \"New VXC Name\" --rate-limit 2000 --cost-centre \"New Cost Centre\"",
			"update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --a-end-vlan 200 --b-end-vlan 300",
			"update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --b-end-partner-config '{\"interfaces\":[{\"vlan\":100,\"ipAddresses\":[\"192.168.1.1/30\"],\"bgpConnections\":[{\"peerAsn\":65000,\"localAsn\":64512,\"localIpAddress\":\"192.168.1.1\",\"peerIpAddress\":\"192.168.1.2\",\"password\":\"bgppassword\",\"shutdown\":false,\"bfdEnabled\":true}]}]}'",
			"update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json '{\"name\":\"Updated VXC Name\",\"rateLimit\":2000,\"costCentre\":\"New Cost Centre\",\"aEndVlan\":200,\"bEndVlan\":300,\"term\":24,\"shutdown\":false}'",
		},
		JSONExamples: []string{
			`{
  "name": "Updated VXC Name",
  "rateLimit": 2000,
  "costCentre": "New Cost Centre",
  "aEndVlan": 200,
  "bEndVlan": 300,
  "term": 24,
  "shutdown": false
}`,
		},
		ImportantNotes: []string{
			"Only VRouter partner configurations can be updated after creation",
			"CSP partner configurations (AWS, Azure, etc.) cannot be changed after creation",
			"Changing the rate limit may result in additional charges",
			"Updating VLANs will cause temporary disruption to the VXC connectivity",
		},
	}
	updateVXCCmd.Long = updateVXCHelp.Build(rootCmd)

	// delete VXC help
	deleteVXCHelp := &help.CommandHelpBuilder{
		CommandName: "megaport-cli vxc delete",
		ShortDesc:   "Delete an existing Virtual Cross Connect (VXC)",
		LongDesc:    "Delete an existing Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to delete an existing VXC by providing its UID.",
		Examples: []string{
			"delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
		},
		ImportantNotes: []string{
			"Deletion is final and cannot be undone",
			"Billing for the VXC stops at the end of the current billing period",
			"The VXC is immediately disconnected upon deletion",
		},
	}
	deleteVXCCmd.Long = deleteVXCHelp.Build(rootCmd)

	rootCmd.AddCommand(vxcCmd)
}

package vxc

import (
	"github.com/megaport/megaport-cli/internal/base/cmdbuilder"
	"github.com/spf13/cobra"
)

// AddCommandsTo builds the vxc commands and adds them to the root command
func AddCommandsTo(rootCmd *cobra.Command) {
	// Create vxc parent command
	vxcCmd := cmdbuilder.NewCommand("vxc", "Manage VXCs in the Megaport API").
		WithLongDesc("Manage VXCs in the Megaport API.\n\nThis command groups all operations related to Virtual Cross Connects (VXCs). VXCs are virtual point-to-point connections between two ports or devices on the Megaport network. You can use the subcommands to perform actions such as retrieving details, purchasing, updating, and deleting VXCs.").
		WithExample("megaport-cli vxc get [vxcUID]").
		WithExample("megaport-cli vxc buy").
		WithExample("megaport-cli vxc update [vxcUID]").
		WithExample("megaport-cli vxc delete [vxcUID]").
		WithRootCmd(rootCmd).
		Build()

	// Create get VXC command
	getVXCCmd := cmdbuilder.NewCommand("get", "Get details for a single VXC").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(GetVXC).
		WithLongDesc("Get details for a single VXC through the Megaport API.\n\nThis command retrieves detailed information for a single Virtual Cross Connect (VXC). You must provide the unique identifier (UID) of the VXC you wish to retrieve.").
		WithExample("megaport-cli vxc get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx").
		WithImportantNote("The output includes the VXC's UID, name, rate limit, A-End and B-End details, status, and cost centre.").
		WithRootCmd(rootCmd).
		Build()

		// Create buy VXC command
	buyVXCCmd := cmdbuilder.NewCommand("buy", "Purchase a new VXC").
		WithColorAwareRunFunc(BuyVXC).
		WithInteractiveFlag().
		WithVXCCreateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Purchase a new Megaport Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to create a VXC by providing the necessary details.").
		WithDocumentedRequiredFlag("name", "Name of the VXC").
		WithDocumentedRequiredFlag("rate-limit", "Bandwidth in Mbps").
		WithDocumentedRequiredFlag("term", "Contract term in months (1, 12, 24, or 36)").
		WithDocumentedRequiredFlag("a-end-uid", "UID of the A-End product").
		WithDocumentedRequiredFlag("b-end-uid", "UID of the B-End product (if not using partner configuration)").
		WithDocumentedRequiredFlag("a-end-vlan", "VLAN for A-End (0-4093, except 1)").
		WithDocumentedRequiredFlag("b-end-vlan", "VLAN for B-End (0-4093, except 1)").
		WithExample("megaport-cli vxc buy --interactive").
		WithExample("megaport-cli vxc buy --name \"My VXC\" --rate-limit 1000 --term 12 --a-end-uid port-123 --b-end-uid port-456 --a-end-vlan 100 --b-end-vlan 200").
		WithExample("megaport-cli vxc buy --json '{\"vxcName\":\"My VXC\",\"rateLimit\":1000,\"term\":12,\"portUid\":\"port-123\",\"aEndConfiguration\":{\"vlan\":100},\"bEndConfiguration\":{\"productUID\":\"port-456\",\"vlan\":200}}'").
		WithExample("megaport-cli vxc buy --json-file ./vxc-config.json").
		WithRootCmd(rootCmd).
		WithConditionalRequirements("name", "rate-limit", "term", "a-end-uid", "a-end-vlan").
		Build()

	// Create update VXC command
	updateVXCCmd := cmdbuilder.NewCommand("update", "Update an existing Virtual Cross Connect (VXC)").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(UpdateVXC).
		WithBoolFlag("interactive", false, "Use interactive mode").
		WithVXCUpdateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Update an existing Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to update an existing VXC by providing the necessary details.").
		WithOptionalFlag("name", "New name for the VXC (1-64 characters)").
		WithOptionalFlag("rate-limit", "New bandwidth in Mbps (50 - 10000)").
		WithOptionalFlag("term", "New contract term in months (1, 12, 24, or 36)").
		WithOptionalFlag("cost-centre", "New cost centre for billing").
		WithOptionalFlag("shutdown", "Whether to shut down the VXC (true/false)").
		WithOptionalFlag("a-end-vlan", "New VLAN for A-End (2-4093, except 4090)").
		WithOptionalFlag("b-end-vlan", "New VLAN for B-End (2-4093, except 4090)").
		WithOptionalFlag("a-end-inner-vlan", "New inner VLAN for A-End (-1 or higher, only for QinQ)").
		WithOptionalFlag("b-end-inner-vlan", "New inner VLAN for B-End (-1 or higher, only for QinQ)").
		WithOptionalFlag("a-end-uid", "New A-End product UID").
		WithOptionalFlag("b-end-uid", "New B-End product UID").
		WithOptionalFlag("a-end-partner-config", "JSON string with A-End VRouter partner configuration").
		WithOptionalFlag("b-end-partner-config", "JSON string with B-End VRouter partner configuration").
		WithExample("megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive").
		WithExample("megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --name \"New VXC Name\" --rate-limit 2000 --cost-centre \"New Cost Centre\"").
		WithExample("megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --a-end-vlan 200 --b-end-vlan 300").
		WithExample("megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --b-end-partner-config '{\"interfaces\":[{\"vlan\":100,\"ipAddresses\":[\"192.168.1.1/30\"],\"bgpConnections\":[{\"peerAsn\":65000,\"localAsn\":64512,\"localIpAddress\":\"192.168.1.1\",\"peerIpAddress\":\"192.168.1.2\",\"password\":\"bgppassword\",\"shutdown\":false,\"bfdEnabled\":true}]}]}'").
		WithExample("megaport-cli vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json '{\"name\":\"Updated VXC Name\",\"rateLimit\":2000,\"costCentre\":\"New Cost Centre\",\"aEndVlan\":200,\"bEndVlan\":300,\"term\":24,\"shutdown\":false}'").
		WithJSONExample(`{
  "name": "Updated VXC Name",
  "rateLimit": 2000,
  "costCentre": "New Cost Centre",
  "aEndVlan": 200,
  "bEndVlan": 300,
  "term": 24,
  "shutdown": false
}`).
		WithImportantNote("Only VRouter partner configurations can be updated after creation").
		WithImportantNote("CSP partner configurations (AWS, Azure, etc.) cannot be changed after creation").
		WithImportantNote("Changing the rate limit may result in additional charges").
		WithImportantNote("Updating VLANs will cause temporary disruption to the VXC connectivity").
		WithRootCmd(rootCmd).
		Build()

	// Create delete VXC command
	deleteVXCCmd := cmdbuilder.NewCommand("delete", "Delete an existing Virtual Cross Connect (VXC)").
		WithArgs(cobra.ExactArgs(1)).
		WithColorAwareRunFunc(DeleteVXC).
		WithDeleteFlags().
		WithLongDesc("Delete an existing Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to delete an existing VXC by providing its UID.").
		WithExample("megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx").
		WithExample("megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --force").
		WithExample("megaport-cli vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --now").
		WithImportantNote("Deletion is final and cannot be undone").
		WithImportantNote("Billing for the VXC stops at the end of the current billing period").
		WithImportantNote("The VXC is immediately disconnected upon deletion").
		WithRootCmd(rootCmd).
		Build()

	// Add commands to their parents
	vxcCmd.AddCommand(
		getVXCCmd,
		buyVXCCmd,
		updateVXCCmd,
		deleteVXCCmd,
	)
	rootCmd.AddCommand(vxcCmd)
}

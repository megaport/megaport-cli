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
		WithExample("vxc get [vxcUID]").
		WithExample("vxc buy").
		WithExample("vxc update [vxcUID]").
		WithExample("vxc delete [vxcUID]").
		WithRootCmd(rootCmd).
		Build()

	// Create get VXC command
	getVXCCmd := cmdbuilder.NewCommand("get", "Get details for a single VXC").
		WithArgs(cobra.ExactArgs(1)).
		WithOutputFormatRunFunc(GetVXC).
		WithLongDesc("Get details for a single VXC through the Megaport API.\n\nThis command retrieves detailed information for a single Virtual Cross Connect (VXC). You must provide the unique identifier (UID) of the VXC you wish to retrieve.").
		WithExample("get vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx").
		WithImportantNote("The output includes the VXC's UID, name, rate limit, A-End and B-End details, status, and cost centre.").
		WithRootCmd(rootCmd).
		Build()

	// Create buy VXC command
	buyVXCCmd := cmdbuilder.NewCommand("buy", "Purchase a new Virtual Cross Connect (VXC)").
		WithColorAwareRunFunc(BuyVXC).
		WithBoolFlag("interactive", false, "Use interactive mode").
		WithVXCCreateFlags().
		WithJSONConfigFlags().
		WithLongDesc("Purchase a new Virtual Cross Connect (VXC) through the Megaport API.\n\nThis command allows you to purchase a VXC by providing the necessary details.").
		WithRequiredFlag("a-end-uid", "UID of the A-End product (Port, MCR, MVE)").
		WithRequiredFlag("name", "Name of the VXC (1-64 characters)").
		WithRequiredFlag("rate-limit", "Bandwidth in Mbps (50 - 10000)").
		WithRequiredFlag("term", "Contract term in months (1, 12, 24, or 36)").
		WithRequiredFlag("a-end-vlan", "VLAN for A-End (2-4093, except 4090)").
		WithOptionalFlag("b-end-uid", "UID of the B-End product (if connecting to non-partner)").
		WithOptionalFlag("b-end-vlan", "VLAN for B-End (2-4093, except 4090)").
		WithOptionalFlag("a-end-inner-vlan", "Inner VLAN for A-End (-1 or higher, only for QinQ)").
		WithOptionalFlag("b-end-inner-vlan", "Inner VLAN for B-End (-1 or higher, only for QinQ)").
		WithOptionalFlag("a-end-vnic-index", "vNIC index for A-End MVE (required for MVE A-End)").
		WithOptionalFlag("b-end-vnic-index", "vNIC index for B-End MVE (required for MVE B-End)").
		WithOptionalFlag("promo-code", "Promotional code").
		WithOptionalFlag("service-key", "Service key").
		WithOptionalFlag("cost-centre", "Cost centre").
		WithOptionalFlag("a-end-partner-config", "JSON string with A-End partner configuration (for VRouter)").
		WithOptionalFlag("b-end-partner-config", "JSON string with B-End partner configuration (for CSPs like AWS, Azure)").
		WithExample("buy --interactive").
		WithExample("buy --a-end-uid \"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" --b-end-uid \"port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy\" --name \"My VXC\" --rate-limit 1000 --term 12 --a-end-vlan 100 --b-end-vlan 200").
		WithExample("buy --a-end-uid \"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" --name \"My AWS VXC\" --rate-limit 1000 --term 12 --a-end-vlan 100 --b-end-partner-config '{\"connectType\":\"AWS\",\"ownerAccount\":\"123456789012\",\"asn\":65000,\"amazonAsn\":64512}'").
		WithExample("buy --a-end-uid \"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\" --name \"My Azure VXC\" --rate-limit 1000 --term 12 --a-end-vlan 100 --b-end-partner-config '{\"connectType\":\"AZURE\",\"serviceKey\":\"s-abcd1234\"}'").
		WithExample("buy --json '{\"aEndUid\":\"port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\",\"name\":\"My VXC\",\"rateLimit\":1000,\"term\":12,\"aEndConfiguration\":{\"vlan\":100},\"bEndConfiguration\":{\"productUid\":\"port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy\",\"vlan\":200}}'").
		WithJSONExample(`{
  "aEndUid": "port-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "name": "My VXC",
  "rateLimit": 1000,
  "term": 12,
  "aEndConfiguration": {"vlan": 100},
  "bEndConfiguration": {
    "productUid": "port-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "vlan": 200
  }
}`).
		WithImportantNote("For AWS connections, you must provide owner account, ASN, and Amazon ASN in b-end-partner-config").
		WithImportantNote("For Azure connections, you must provide a service key in b-end-partner-config").
		WithImportantNote("QinQ VLANs require both outer and inner VLANs").
		WithImportantNote("MVE connections require specifying vNIC indexes").
		WithRootCmd(rootCmd).
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
		WithExample("update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive").
		WithExample("update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --name \"New VXC Name\" --rate-limit 2000 --cost-centre \"New Cost Centre\"").
		WithExample("update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --a-end-vlan 200 --b-end-vlan 300").
		WithExample("update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --b-end-partner-config '{\"interfaces\":[{\"vlan\":100,\"ipAddresses\":[\"192.168.1.1/30\"],\"bgpConnections\":[{\"peerAsn\":65000,\"localAsn\":64512,\"localIpAddress\":\"192.168.1.1\",\"peerIpAddress\":\"192.168.1.2\",\"password\":\"bgppassword\",\"shutdown\":false,\"bfdEnabled\":true}]}]}'").
		WithExample("update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json '{\"name\":\"Updated VXC Name\",\"rateLimit\":2000,\"costCentre\":\"New Cost Centre\",\"aEndVlan\":200,\"bEndVlan\":300,\"term\":24,\"shutdown\":false}'").
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
		WithExample("delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx").
		WithExample("delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --force").
		WithExample("delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --now").
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

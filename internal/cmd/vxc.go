package cmd

import (
	"github.com/spf13/cobra"
)

// vxcCmd is the base command for all operations related to Virtual Cross Connects (VXCs).
// It groups subcommands for managing VXCs in the Megaport API.
// Use the "megaport vxc get [vxcUID]" command to retrieve detailed information about a specific VXC.
var vxcCmd = &cobra.Command{
	Use:   "vxc",
	Short: "Manage VXCs in the Megaport API",
	Long: `Manage VXCs in the Megaport API.

This command groups all operations related to Virtual Cross Connects (VXCs).
You can use the subcommands to perform actions such as retrieving details for a specific VXC.
For example, use the "megaport vxc get [vxcUID]" command to fetch details for the VXC identified by its UID.
`,
}

// getVXCCmd retrieves detailed information for a single Virtual Cross Connect (VXC).
// This command requires exactly one argument: the UID of the VXC.
// It establishes a context with a timeout, logs into the Megaport API, and uses the API client
// to obtain and then display the VXC details using the configured output format (JSON or table).
//
// Example usage:
//
//	megaport vxc get VXC12345
var getVXCCmd = &cobra.Command{
	Use:   "get [vxcUID]",
	Short: "Get details for a single VXC",
	Args:  cobra.ExactArgs(1),
	RunE:  WrapRunE(GetVXC),
}

var buyVXCCmd = &cobra.Command{
	Use:   "buy",
	Short: "Purchase a new Virtual Cross Connect (VXC)",
	Long: `Purchase a new Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to purchase a VXC by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode (default):
   The command will prompt you for each required and optional field.

2. Flag Mode:
   Provide all required fields as flags:
   --a-end-uid, --b-end-uid, --name, --rate-limit, --term,
   --a-end-vlan, --b-end-vlan, --a-end-partner-config, --b-end-partner-config

3. JSON Mode:
   Provide a JSON string or file with all required fields:
   --json <json-string> or --json-file <path>

Required fields:
- a-end-uid: UID of the A-End product
- name: Name of the VXC
- rate-limit: Bandwidth in Mbps
- term: Contract term in months (1, 12, 24, or 36)

Optional fields:
- b-end-uid: UID of the B-End product (if connecting to non-partner)
- a-end-vlan: VLAN for A-End (0-4093, except 1)
- b-end-vlan: VLAN for B-End (0-4093, except 1)
- a-end-inner-vlan: Inner VLAN for A-End (-1 or higher)
- b-end-inner-vlan: Inner VLAN for B-End (-1 or higher)
- a-end-vnic-index: vNIC index for A-End MVE
- b-end-vnic-index: vNIC index for B-End MVE
- promo-code: Promotional code
- service-key: Service key
- cost-centre: Cost centre
- a-end-partner-config: JSON string with A-End partner configuration
- b-end-partner-config: JSON string with B-End partner configuration

Example usage:

# Interactive mode
megaport vxc buy --interactive

# Flag mode - Basic VXC between two ports
megaport vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --b-end-uid "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy" \
  --name "My VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-vlan 200

# Flag mode - VXC to AWS Direct Connect
megaport vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --b-end-uid "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy" \
  --name "My AWS VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-partner-config '{"connectType":"AWS","ownerAccount":"123456789012","asn":65000,"amazonAsn":64512}'

# Flag mode - VXC to Azure ExpressRoute
megaport vxc buy \
  --a-end-uid "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
  --name "My Azure VXC" \
  --rate-limit 1000 \
  --term 12 \
  --a-end-vlan 100 \
  --b-end-partner-config '{"connectType":"AZURE","serviceKey":"s-abcd1234"}'

# JSON mode
megaport vxc buy --json '{
  "portUID": "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "vxcName": "My VXC",
  "rateLimit": 1000,
  "term": 12,
  "aEndConfiguration": {
    "vlan": 100
  },
  "bEndConfiguration": {
    "productUid": "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "vlan": 200
  }
}'

# JSON mode with partner config
megaport vxc buy --json '{
  "portUID": "dcc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "vxcName": "My AWS VXC",
  "rateLimit": 1000,
  "term": 12,
  "aEndConfiguration": {
    "vlan": 100
  },
  "bEndConfiguration": {
    "productUid": "dcc-yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
    "partnerConfig": {
      "connectType": "AWS",
      "ownerAccount": "123456789012",
      "asn": 65000,
      "amazonAsn": 64512,
      "type": "private"
    }
  }
}'

# JSON file
megaport vxc buy --json-file ./vxc-config.json
`,
	RunE: WrapRunE(BuyVXC),
}

var updateVXCCmd = &cobra.Command{
	Use:   "update [vxcUID]",
	Short: "Update an existing Virtual Cross Connect (VXC)",
	Args:  cobra.ExactArgs(1),
	Long: `Update an existing Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to update an existing VXC by providing the necessary details.
You can provide details in one of three ways:

1. Interactive Mode:
   The command will prompt you for each field that can be updated.

2. Flag Mode:
   Provide fields to update using flags:
   --name, --rate-limit, --term, --cost-centre, --shutdown, 
   --a-end-vlan, --b-end-vlan, --a-end-inner-vlan, --b-end-inner-vlan,
   --a-end-uid, --b-end-uid, --a-end-partner-config, --b-end-partner-config

3. JSON Mode:
   Provide a JSON string or file with update fields:
   --json <json-string> or --json-file <path>

Updateable fields:
- name: New name for the VXC
- rate-limit: New bandwidth in Mbps
- term: New contract term in months (1, 12, 24, or 36)
- cost-centre: New cost centre for billing
- shutdown: Whether to shut down the VXC (true/false)
- a-end-vlan: New VLAN for A-End (0-4093, except 1)
- b-end-vlan: New VLAN for B-End (0-4093, except 1)
- a-end-inner-vlan: New inner VLAN for A-End (-1 or higher)
- b-end-inner-vlan: New inner VLAN for B-End (-1 or higher)
- a-end-uid: New A-End product UID
- b-end-uid: New B-End product UID

NOTE: For partner configurations, only VRouter partner configurations can be updated.
Other CSP partner configurations (AWS, Azure, etc.) cannot be changed after creation.

Example usage:

# Interactive mode
megaport vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --interactive

# Flag mode - Basic updates
megaport vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --name "New VXC Name" \
  --rate-limit 2000 \
  --cost-centre "New Cost Centre"

# Flag mode - Update VLANs
megaport vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --a-end-vlan 200 \
  --b-end-vlan 300

# Flag mode - Update with VRouter partner config
megaport vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx \
  --b-end-partner-config '{
    "interfaces": [
      {
        "vlan": 100,
        "ipAddresses": ["192.168.1.1/30"],
        "bgpConnections": [
          {
            "peerAsn": 65000,
            "localAsn": 64512,
            "localIpAddress": "192.168.1.1",
            "peerIpAddress": "192.168.1.2",
            "password": "bgppassword",
            "shutdown": false,
            "bfdEnabled": true
          }
        ]
      }
    ]
  }'

# JSON mode
megaport vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json '{
  "name": "Updated VXC Name",
  "rateLimit": 2000,
  "costCentre": "New Cost Centre",
  "aEndVlan": 200,
  "bEndVlan": 300,
  "term": 24,
  "shutdown": false
}'

# JSON file
megaport vxc update vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx --json-file ./vxc-update.json
`,
	RunE: WrapRunE(UpdateVXC),
}

// deleteVXCCmd deletes an existing Virtual Cross Connect (VXC) in the Megaport API.
var deleteVXCCmd = &cobra.Command{
	Use:   "delete [vxcUID]",
	Short: "Delete an existing Virtual Cross Connect (VXC)",
	Args:  cobra.ExactArgs(1),
	Long: `Delete an existing Virtual Cross Connect (VXC) through the Megaport API.

This command allows you to delete an existing VXC by providing its UID.

Example usage:

# Delete a VXC
megaport vxc delete vxc-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
`,
	RunE: WrapRunE(DeleteVXC),
}

func init() {
	vxcCmd.AddCommand(getVXCCmd)
	vxcCmd.AddCommand(buyVXCCmd)
	vxcCmd.AddCommand(updateVXCCmd)
	vxcCmd.AddCommand(deleteVXCCmd)
	rootCmd.AddCommand(vxcCmd)

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
}

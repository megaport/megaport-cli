package cmd

import (
	"github.com/spf13/cobra"
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

func init() {
	mveCmd.AddCommand(buyMVECmd)
	mveCmd.AddCommand(getMVECmd)
	rootCmd.AddCommand(mveCmd)
}

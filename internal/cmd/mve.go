package cmd

import (
	"context"
	"fmt"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// mveCmd is the base command for all Megaport Virtual Edge (MVE) operations.
var mveCmd = &cobra.Command{
	Use:   "mve",
	Short: "Manage MVEs in the Megaport API",
	Long: `Manage MVEs in the Megaport API.

This command groups all operations related to Megaport Virtual Edge devices (MVEs).
Use the "megaport mve get [mveUID]" command to fetch details for a specific MVE identified by its UID.
`,
}

// getMVECmd retrieves details for a single MVE.
var getMVECmd = &cobra.Command{
	Use:   "get [mveUID]",
	Short: "Get details for a single MVE",
	Args:  cobra.ExactArgs(1),
	RunE:  GetMVE,
}

func GetMVE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Login(ctx)
	if err != nil {
		return fmt.Errorf("error logging in: %v", err)
	}

	mveUID := args[0]
	if mveUID == "" {
		return fmt.Errorf("MVE UID cannot be empty")
	}

	mve, err := client.MVEService.GetMVE(ctx, mveUID)
	if err != nil {
		return fmt.Errorf("error getting MVE: %v", err)
	}

	if mve == nil {
		return fmt.Errorf("no MVE found with UID: %s", mveUID)
	}

	err = printMVEs([]*megaport.MVE{mve}, outputFormat)
	if err != nil {
		return fmt.Errorf("error printing MVEs: %v", err)
	}
	return nil
}

func init() {
	mveCmd.AddCommand(getMVECmd)
	rootCmd.AddCommand(mveCmd)
}

// MVEOutput represents the desired fields for JSON output.
type MVEOutput struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	LocationID int    `json:"location_id"`
}

// ToMVEOutput converts an MVE to an MVEOutput.
func ToMVEOutput(m *megaport.MVE) (MVEOutput, error) {
	if m == nil {
		return MVEOutput{}, fmt.Errorf("invalid MVE: nil value")
	}

	return MVEOutput{
		UID:        m.UID,
		Name:       m.Name,
		LocationID: m.LocationID,
	}, nil
}

// printMVEs prints the MVEs in the specified output format.
func printMVEs(mves []*megaport.MVE, format string) error {
	if mves == nil {
		mves = []*megaport.MVE{}
	}

	outputs := make([]MVEOutput, 0, len(mves))
	for _, mve := range mves {
		output, err := ToMVEOutput(mve)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}

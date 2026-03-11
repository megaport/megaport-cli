package ix

import (
	"fmt"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
)

// IXOutput represents the desired fields for JSON output of IX details.
type IXOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid" header:"UID"`
	Name               string `json:"name" header:"Name"`
	NetworkServiceType string `json:"network_service_type" header:"Network Service Type"`
	ASN                int    `json:"asn" header:"ASN"`
	RateLimit          int    `json:"rate_limit" header:"Rate Limit"`
	VLAN               int    `json:"vlan" header:"VLAN"`
	MACAddress         string `json:"mac_address" header:"MAC Address"`
	Status             string `json:"status" header:"Status"`
}

// IXStatus represents a lightweight status view of an IX.
type IXStatus struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid" header:"UID"`
	Name          string `json:"name" header:"Name"`
	Status        string `json:"status" header:"Status"`
	Type          string `json:"type" header:"Type"`
}

// ToIXOutput converts a *megaport.IX to our IXOutput struct.
func ToIXOutput(ix *megaport.IX) (IXOutput, error) {
	if ix == nil {
		return IXOutput{}, fmt.Errorf("invalid IX: nil value")
	}

	return IXOutput{
		UID:                ix.ProductUID,
		Name:               ix.ProductName,
		NetworkServiceType: ix.NetworkServiceType,
		ASN:                ix.ASN,
		RateLimit:          ix.RateLimit,
		VLAN:               ix.VLAN,
		MACAddress:         ix.MACAddress,
		Status:             ix.ProvisioningStatus,
	}, nil
}

// printIXs prints a list of IXs in the specified format.
func printIXs(ixs []*megaport.IX, format string, noColor bool) error {
	outputs := make([]IXOutput, 0, len(ixs))
	for _, ix := range ixs {
		o, err := ToIXOutput(ix)
		if err != nil {
			return err
		}
		outputs = append(outputs, o)
	}
	return output.PrintOutput(outputs, format, noColor)
}

// displayIXChanges compares the original and updated IX and displays the differences.
func displayIXChanges(original, updated *megaport.IX, noColor bool) {
	if original == nil || updated == nil {
		return
	}

	fmt.Println()
	output.PrintInfo("Changes applied:", noColor)

	changesFound := false

	if original.ProductName != updated.ProductName {
		changesFound = true
		fmt.Printf("  • Name: %s → %s\n",
			output.FormatOldValue(original.ProductName, noColor),
			output.FormatNewValue(updated.ProductName, noColor))
	}

	if original.RateLimit != updated.RateLimit {
		changesFound = true
		fmt.Printf("  • Rate Limit: %s → %s\n",
			output.FormatOldValue(fmt.Sprintf("%d Mbps", original.RateLimit), noColor),
			output.FormatNewValue(fmt.Sprintf("%d Mbps", updated.RateLimit), noColor))
	}

	if original.VLAN != updated.VLAN {
		changesFound = true
		fmt.Printf("  • VLAN: %s → %s\n",
			output.FormatOldValue(fmt.Sprintf("%d", original.VLAN), noColor),
			output.FormatNewValue(fmt.Sprintf("%d", updated.VLAN), noColor))
	}

	if original.MACAddress != updated.MACAddress {
		changesFound = true
		fmt.Printf("  • MAC Address: %s → %s\n",
			output.FormatOldValue(original.MACAddress, noColor),
			output.FormatNewValue(updated.MACAddress, noColor))
	}

	if original.ASN != updated.ASN {
		changesFound = true
		fmt.Printf("  • ASN: %s → %s\n",
			output.FormatOldValue(fmt.Sprintf("%d", original.ASN), noColor),
			output.FormatNewValue(fmt.Sprintf("%d", updated.ASN), noColor))
	}

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}

package cmd

import (
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

// VXCOutput represents the desired fields for JSON output.
type VXCOutput struct {
	output
	UID     string `json:"uid"`
	Name    string `json:"name"`
	AEndUID string `json:"a_end_uid"`
	BEndUID string `json:"b_end_uid"`
}

// ToVXCOutput converts a VXC to a VXCOutput.
func ToVXCOutput(v *megaport.VXC) (VXCOutput, error) {
	if v == nil {
		return VXCOutput{}, fmt.Errorf("invalid VXC: nil value")
	}

	return VXCOutput{
		UID:     v.UID,
		Name:    v.Name,
		AEndUID: v.AEndConfiguration.UID,
		BEndUID: v.BEndConfiguration.UID,
	}, nil
}

// printVXCs prints the VXCs in the specified output format
func printVXCs(vxcs []*megaport.VXC, format string) error {
	if vxcs == nil {
		vxcs = []*megaport.VXC{}
	}

	outputs := make([]VXCOutput, 0, len(vxcs))
	for _, vxc := range vxcs {
		output, err := ToVXCOutput(vxc)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return printOutput(outputs, format)
}

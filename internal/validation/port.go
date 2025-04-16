package validation

import (
	"fmt"
	"slices"

	megaport "github.com/megaport/megaportgo"
)

// Define constants for LAG validation
var (
	ValidLAGPortSpeeds = []int{10000, 100000}
	MinLAGCount        = 1
	MaxLAGCount        = 8
)

// ValidatePortVLANAvailability validates if a VLAN ID is within the range
// typically available for user assignment on a Port (excluding special/reserved values).
func ValidatePortVLANAvailability(vlan int) error {
	// This function specifically checks the range for VLANs that can be assigned,
	// excluding special values like AutoAssign (-1) or Untagged (0), and potentially
	// reserved values like 1 or 4094/4095 depending on the platform.
	if vlan < MinAssignableVLAN || vlan > MaxAssignableVLAN {
		return NewValidationError("VLAN ID", vlan, fmt.Sprintf("must be between %d-%d for VLAN availability check", MinAssignableVLAN, MaxAssignableVLAN))
	}
	return nil
}

// ValidatePortName validates a port name
func ValidatePortName(name string) error {
	if name == "" {
		return NewValidationError("port name", name, "cannot be empty")
	}

	// The spec says names can be up to MaxPortNameLength characters (inclusive)
	if len(name) > MaxPortNameLength {
		return NewValidationError("port name", name, fmt.Sprintf("cannot exceed %d characters", MaxPortNameLength))
	}

	return nil
}

// ValidatePortRequest validates a standard port buy request.
func ValidatePortRequest(req *megaport.BuyPortRequest) error {
	if req.Name == "" {
		return NewValidationError("port name", req.Name, "cannot be empty")
	}
	if len(req.Name) > MaxPortNameLength {
		return NewValidationError("port name", req.Name, fmt.Sprintf("cannot exceed %d characters", MaxPortNameLength))
	}
	if req.LocationId <= 0 {
		return NewValidationError("location ID", req.LocationId, "must be a positive integer")
	}
	if err := ValidatePortSpeed(req.PortSpeed); err != nil {
		return err
	}
	if err := ValidateContractTerm(req.Term); err != nil {
		return err
	}
	return nil
}

// ValidateLAGPortRequest validates a LAG port buy request.
func ValidateLAGPortRequest(req *megaport.BuyPortRequest) error {
	if req.Name == "" {
		return NewValidationError("port name", req.Name, "cannot be empty")
	}
	if len(req.Name) > MaxPortNameLength {
		return NewValidationError("port name", req.Name, fmt.Sprintf("cannot exceed %d characters", MaxPortNameLength))
	}
	if req.LocationId <= 0 {
		return NewValidationError("location ID", req.LocationId, "must be a positive integer")
	}
	// Use the defined constant for LAG port speeds with specific formatting
	if !slices.Contains(ValidLAGPortSpeeds, req.PortSpeed) {
		return NewValidationError("port speed", req.PortSpeed, fmt.Sprintf("must be one of: %v for LAG ports", ValidLAGPortSpeeds))
	}
	// Use the defined constants for LAG count
	if req.LagCount < MinLAGCount || req.LagCount > MaxLAGCount {
		return NewValidationError("LAG count", req.LagCount, fmt.Sprintf("must be between %d and %d", MinLAGCount, MaxLAGCount))
	}
	if err := ValidateContractTerm(req.Term); err != nil {
		return err
	}
	return nil
}

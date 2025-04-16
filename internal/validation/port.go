package validation

import "fmt"

// ValidatePortRequest validates a Port order request
func ValidatePortRequest(name string, term int, portSpeed int, locationID int) error {
	if name == "" {
		return NewValidationError("port name", name, "cannot be empty")
	}

	if err := ValidateContractTerm(term); err != nil {
		return err
	}

	if err := ValidatePortSpeed(portSpeed); err != nil {
		return err
	}

	if locationID <= 0 {
		return NewValidationError("location ID", locationID, "must be a positive integer")
	}

	return nil
}

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

	// The spec says names can be up to 64 characters (inclusive)
	if len(name) > 64 {
		return NewValidationError("port name", name, "cannot exceed 64 characters")
	}

	return nil
}

package validation

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

// ValidatePortVLANAvailability validates if a VLAN is available on a port
func ValidatePortVLANAvailability(vlan int) error {
	if vlan < 2 || vlan > 4093 {
		return NewValidationError("VLAN ID", vlan, "must be between 2-4093 for VLAN availability check")
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

package validation

import "fmt"

// ValidateMCRRequest validates an MCR order request
func ValidateMCRRequest(name string, term int, portSpeed int, locationID int) error {
	if name == "" {
		return NewValidationError("MCR name", name, "cannot be empty")
	}

	// Specialized term validation for MCR with specific error messages expected by tests
	if term == 0 {
		return fmt.Errorf("term is required")
	}
	if term != 1 && term != 12 && term != 24 && term != 36 {
		return fmt.Errorf("invalid term")
	}

	if err := ValidateMCRPortSpeed(portSpeed); err != nil {
		return err
	}

	if locationID <= 0 {
		return NewValidationError("location ID", locationID, "must be a positive integer")
	}

	return nil
}

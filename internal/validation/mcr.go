package validation

// ValidateMCRRequest validates an MCR order request
func ValidateMCRRequest(name string, term int, portSpeed int, locationID int) error {
	if name == "" {
		return NewValidationError("MCR name", name, "cannot be empty")
	}

	// Use standard contract term validation
	if err := ValidateContractTerm(term); err != nil {
		// Return the specific ValidationError from ValidateContractTerm
		return err
	}

	if err := ValidateMCRPortSpeed(portSpeed); err != nil {
		return err
	}

	if locationID <= 0 {
		return NewValidationError("location ID", locationID, "must be a positive integer")
	}

	return nil
}

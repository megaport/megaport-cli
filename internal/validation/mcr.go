package validation

import (
	megaport "github.com/megaport/megaportgo"
)

// ValidateMCRRequest validates a request to buy/provision a new MCR (Megaport Cloud Router).
// This function ensures all parameters meet the requirements for creating a new MCR.
//
// Parameters:
//   - req: The BuyMCRRequest object containing all MCR provisioning parameters
//
// Validation checks:
//   - Name cannot be empty
//   - Contract term must be valid (typically 1, 12, 24, or 36 months)
//   - Port speed must be one of the valid MCR port speeds
//   - Location ID must be a positive integer
//
// Returns:
//   - A ValidationError if any validation check fails
//   - nil if all validation checks pass
func ValidateMCRRequest(req *megaport.BuyMCRRequest) error {
	if req.Name == "" {
		return NewValidationError("MCR name", req.Name, "cannot be empty")
	}
	if err := ValidateContractTerm(req.Term); err != nil {
		return err
	}
	if err := ValidateMCRPortSpeed(req.PortSpeed); err != nil {
		return err
	}
	if req.LocationID <= 0 {
		return NewValidationError("location ID", req.LocationID, "must be a positive integer")
	}
	return nil
}

// ValidatePrefixFilterListRequest validates a request to create a new prefix filter list for an MCR.
// Prefix filter lists are used to control route advertisements in BGP sessions on MCRs.
//
// Parameters:
//   - req: The CreateMCRPrefixFilterListRequest object containing the prefix filter list definition
//
// Validation checks:
//   - Description cannot be empty
//   - Address family must be provided ("IPv4" or "IPv6")
//   - Address family must be a valid value ("IPv4" or "IPv6")
//   - At least one entry must be provided in the prefix filter list
//   - For each entry:
//   - Prefix cannot be empty
//   - Action must be "permit" or "deny"
//
// Returns:
//   - A ValidationError if any validation check fails
//   - nil if all validation checks pass
func ValidatePrefixFilterListRequest(req *megaport.CreateMCRPrefixFilterListRequest) error {
	if req.PrefixFilterList.Description == "" {
		return NewValidationError("description", req.PrefixFilterList.Description, "cannot be empty")
	}
	if req.PrefixFilterList.AddressFamily == "" {
		return NewValidationError("address family", req.PrefixFilterList.AddressFamily, "cannot be empty")
	}
	if req.PrefixFilterList.AddressFamily != "IPv4" && req.PrefixFilterList.AddressFamily != "IPv6" {
		return NewValidationError("address family", req.PrefixFilterList.AddressFamily, "must be IPv4 or IPv6")
	}
	if len(req.PrefixFilterList.Entries) == 0 {
		return NewValidationError("entries", req.PrefixFilterList.Entries, "must contain at least one entry")
	}

	// Validate each entry
	for i, entry := range req.PrefixFilterList.Entries {
		if entry.Prefix == "" {
			return NewValidationError("entry prefix index", i, "prefix cannot be empty")
		}
		if entry.Action != "permit" && entry.Action != "deny" {
			return NewValidationError("entry action", entry.Action, "must be permit or deny")
		}
	}

	return nil
}

// ValidateUpdatePrefixFilterList validates a request to update an existing prefix filter list for an MCR.
// This function ensures that the updated prefix filter list entries meet the requirements.
//
// Parameters:
//   - prefixFilterList: The MCRPrefixFilterList object containing the updated prefix filter list
//
// Validation checks:
//   - If entries are provided:
//   - For each entry:
//   - Prefix cannot be empty
//   - Action must be "permit" or "deny"
//
// Returns:
//   - A ValidationError if any validation check fails
//   - nil if all validation checks pass
func ValidateUpdatePrefixFilterList(prefixFilterList *megaport.MCRPrefixFilterList) error {
	// If entries are provided, validate them
	if len(prefixFilterList.Entries) > 0 {
		// Validate each entry
		for i, entry := range prefixFilterList.Entries {
			if entry.Prefix == "" {
				return NewValidationError("entry prefix index", i, "prefix cannot be empty")
			}
			if entry.Action != "permit" && entry.Action != "deny" {
				return NewValidationError("entry action", entry.Action, "must be permit or deny")
			}
		}
	}

	return nil
}

package validation

import (
	megaport "github.com/megaport/megaportgo"
)

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

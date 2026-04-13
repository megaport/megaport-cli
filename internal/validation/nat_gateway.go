package validation

import (
	megaport "github.com/megaport/megaportgo"
)

// ValidateCreateNATGatewayRequest validates a request to create a NAT Gateway.
func ValidateCreateNATGatewayRequest(req *megaport.CreateNATGatewayRequest) error {
	if req.ProductName == "" {
		return NewValidationError("name", req.ProductName, "cannot be empty")
	}
	if req.LocationID < 1 {
		return NewValidationError("location ID", req.LocationID, "must be a positive integer")
	}
	if req.Speed < 1 {
		return NewValidationError("speed", req.Speed, "must be a positive integer")
	}
	if err := ValidateContractTerm(req.Term); err != nil {
		return err
	}
	if req.Config.SessionCount < 0 {
		return NewValidationError("session count", req.Config.SessionCount, "must be a non-negative integer")
	}
	if req.Config.ASN < 0 {
		return NewValidationError("ASN", req.Config.ASN, "must be a non-negative integer")
	}
	return nil
}

// ValidateUpdateNATGatewayRequest validates a request to update a NAT Gateway.
// Only the product UID is strictly required; other fields should be pre-filled
// from the original resource before calling this function.
func ValidateUpdateNATGatewayRequest(req *megaport.UpdateNATGatewayRequest) error {
	if req.ProductUID == "" {
		return NewValidationError("product UID", req.ProductUID, "cannot be empty")
	}
	// Name is required by the API (no omitempty on the field).
	if req.ProductName == "" {
		return NewValidationError("name", req.ProductName, "cannot be empty")
	}
	if req.LocationID < 1 {
		return NewValidationError("location ID", req.LocationID, "must be a positive integer")
	}
	if req.Speed < 1 {
		return NewValidationError("speed", req.Speed, "must be a positive integer")
	}
	if err := ValidateContractTerm(req.Term); err != nil {
		return err
	}
	if req.Config.SessionCount < 0 {
		return NewValidationError("session count", req.Config.SessionCount, "must be a non-negative integer")
	}
	if req.Config.ASN < 0 {
		return NewValidationError("ASN", req.Config.ASN, "must be a non-negative integer")
	}
	return nil
}

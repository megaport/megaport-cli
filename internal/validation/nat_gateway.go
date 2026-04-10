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
	return nil
}

// ValidateUpdateNATGatewayRequest validates a request to update a NAT Gateway.
func ValidateUpdateNATGatewayRequest(req *megaport.UpdateNATGatewayRequest) error {
	if req.ProductUID == "" {
		return NewValidationError("product UID", req.ProductUID, "cannot be empty")
	}
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
	return nil
}

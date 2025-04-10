package mve

import (
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

// Validate MVE buy request
func validateBuyMVERequest(req *megaport.BuyMVERequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Term == 0 {
		return fmt.Errorf("term is required")
	}

	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, or 36 months")
	}

	if req.LocationID == 0 {
		return fmt.Errorf("location ID is required")
	}

	if req.VendorConfig == nil {
		return fmt.Errorf("vendor config is required")
	}

	return nil
}

// Validate MVE update request
func validateUpdateMVERequest(req *megaport.ModifyMVERequest) error {
	// Check if any update fields are provided
	if req.Name == "" && req.CostCentre == "" && req.ContractTermMonths == nil {
		return fmt.Errorf("at least one field to update must be provided")
	}

	// If contract term is provided, validate it
	if req.ContractTermMonths != nil {
		term := *req.ContractTermMonths
		if term != 1 && term != 12 && term != 24 && term != 36 {
			return fmt.Errorf("invalid contract term, must be one of 1, 12, 24, or 36 months")
		}
	}

	return nil
}

// filterMVEImages filters the provided MVE images based on the given filters.
func filterMVEImages(images []*megaport.MVEImage, vendor, productCode string, id int, version string, releaseImage bool) []*megaport.MVEImage {
	var filtered []*megaport.MVEImage
	for _, image := range images {
		if vendor != "" && image.Vendor != vendor {
			continue
		}
		if productCode != "" && image.ProductCode != productCode {
			continue
		}
		if id != 0 && image.ID != id {
			continue
		}
		if version != "" && image.Version != version {
			continue
		}
		if releaseImage && !image.ReleaseImage {
			continue
		}
		filtered = append(filtered, image)
	}
	return filtered
}

// filterMVEs applies filters to a list of MVEs
func filterMVEs(mves []*megaport.MVE, locationID int, vendor, name string) []*megaport.MVE {
	var filtered []*megaport.MVE

	// Handle nil slice
	if mves == nil {
		return filtered
	}

	for _, mve := range mves {
		// Skip nil MVEs
		if mve == nil {
			continue
		}

		// Apply filters
		if locationID > 0 && mve.LocationID != locationID {
			continue
		}

		// Extract vendor from VendorConfiguration if available
		mveVendor := mve.Vendor

		if vendor != "" && !strings.EqualFold(mveVendor, vendor) {
			continue
		}

		if name != "" && !strings.Contains(strings.ToLower(mve.Name), strings.ToLower(name)) {
			continue
		}

		filtered = append(filtered, mve)
	}

	return filtered
}

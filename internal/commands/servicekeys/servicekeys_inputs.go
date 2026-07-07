package servicekeys

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// parseServiceKeyValidFor builds a ValidFor from YYYY-MM-DD date strings.
// Returns nil if both dates are empty.
func parseServiceKeyValidFor(startDate, endDate string) (*megaport.ValidFor, error) {
	if err := validation.ValidateDateRange(startDate, endDate); err != nil {
		return nil, err
	}
	if startDate == "" || endDate == "" {
		return nil, nil
	}

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date %q: %w", startDate, err)
	}
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date %q: %w", endDate, err)
	}

	return &megaport.ValidFor{
		StartTime: &megaport.Time{Time: startTime},
		EndTime:   &megaport.Time{Time: endTime},
	}, nil
}

func processFlagCreateServiceKeyInput(cmd *cobra.Command) (*megaport.CreateServiceKeyRequest, error) {
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	productUID, _ := cmd.Flags().GetString("product-uid")
	productID, _ := cmd.Flags().GetInt("product-id")
	singleUse, _ := cmd.Flags().GetBool("single-use")
	maxSpeed, _ := cmd.Flags().GetInt("max-speed")
	description, _ := cmd.Flags().GetString("description")
	startDate, _ := cmd.Flags().GetString("start-date")
	endDate, _ := cmd.Flags().GetString("end-date")
	active, _ := cmd.Flags().GetBool("active")
	preApproved, _ := cmd.Flags().GetBool("pre-approved")
	vlan, _ := cmd.Flags().GetInt("vlan")

	if productUID != "" && productID != 0 {
		return nil, fmt.Errorf("--product-uid and --product-id cannot both be set")
	}

	validFor, err := parseServiceKeyValidFor(startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &megaport.CreateServiceKeyRequest{
		ProductUID:  productUID,
		ProductID:   productID,
		SingleUse:   singleUse,
		MaxSpeed:    maxSpeed,
		Description: description,
		ValidFor:    validFor,
		Active:      active,
		PreApproved: preApproved,
		VLAN:        vlan,
	}, nil
}

// processJSONCreateServiceKeyInput parses a create request from JSON. Most
// fields unmarshal directly onto CreateServiceKeyRequest since the SDK's JSON
// tags already match the desired field names; startDate/endDate are read
// separately since the SDK has no user-friendly date fields.
func processJSONCreateServiceKeyInput(jsonStr, jsonFile string) (*megaport.CreateServiceKeyRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	req := &megaport.CreateServiceKeyRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	if req.ProductUID != "" && req.ProductID != 0 {
		return nil, fmt.Errorf("productUid and productId cannot both be set")
	}
	// A raw "validFor" key unmarshals onto OrderValidFor (raw epoch millis),
	// bypassing startDate/endDate validation below. Discard it so startDate/
	// endDate are the only supported way to set the validity window.
	req.OrderValidFor = nil

	var dates struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
	}
	if err := json.Unmarshal(jsonData, &dates); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	validFor, err := parseServiceKeyValidFor(dates.StartDate, dates.EndDate)
	if err != nil {
		return nil, err
	}
	req.ValidFor = validFor

	return req, nil
}

// buildUpdateServiceKeyRequestFromFlags mirrors the flag-only update path
// that shipped with the ESD-1272/ESD-1417 merge fix: SingleUse and Active
// default to the current key's values because the SDK serializes them
// unconditionally (no omitempty), so an unset flag must not clobber them.
func buildUpdateServiceKeyRequestFromFlags(cmd *cobra.Command, key string, current *megaport.ServiceKey) *megaport.UpdateServiceKeyRequest {
	req := &megaport.UpdateServiceKeyRequest{
		Key:       key,
		SingleUse: current.SingleUse,
		Active:    current.Active,
	}
	if cmd.Flags().Changed("single-use") {
		req.SingleUse, _ = cmd.Flags().GetBool("single-use")
	}
	if cmd.Flags().Changed("active") {
		req.Active, _ = cmd.Flags().GetBool("active")
	}
	// The SDK rejects requests with both ProductUID and ProductID set, so
	// preserve the current ProductUID only when the user provided neither.
	switch {
	case cmd.Flags().Changed("product-uid"):
		req.ProductUID, _ = cmd.Flags().GetString("product-uid")
	case cmd.Flags().Changed("product-id"):
		req.ProductID, _ = cmd.Flags().GetInt("product-id")
	default:
		req.ProductUID = current.ProductUID
	}
	return req
}

// buildUpdateServiceKeyRequestFromJSON applies the same current-state merge
// as buildUpdateServiceKeyRequestFromFlags, using presence in the raw JSON
// map to distinguish "not provided" from an explicit false/empty value.
func buildUpdateServiceKeyRequestFromJSON(jsonStr, jsonFile, key string, current *megaport.ServiceKey) (*megaport.UpdateServiceKeyRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	_, hasProductUID := raw["productUid"]
	_, hasProductID := raw["productId"]
	if hasProductUID && hasProductID {
		return nil, fmt.Errorf("productUid and productId cannot both be set")
	}

	req := &megaport.UpdateServiceKeyRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	req.Key = key
	// Update has no supported way to change the validity window (no
	// startDate/endDate flags either), so discard a raw "validFor" key
	// rather than sending it to the API unvalidated.
	req.OrderValidFor = nil
	req.ValidFor = nil

	if _, ok := raw["singleUse"]; !ok {
		req.SingleUse = current.SingleUse
	}
	if _, ok := raw["active"]; !ok {
		req.Active = current.Active
	}
	if !hasProductUID && !hasProductID {
		req.ProductUID = current.ProductUID
	}

	return req, nil
}

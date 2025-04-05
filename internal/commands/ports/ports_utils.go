package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var updatePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.ModifyPortRequest) (*megaport.ModifyPortResponse, error) {
	return client.PortService.ModifyPort(ctx, req)
}

var deletePortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.DeletePortRequest) (*megaport.DeletePortResponse, error) {
	return client.PortService.DeletePort(ctx, req)
}

var restorePortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.RestorePortResponse, error) {
	return client.PortService.RestorePort(ctx, portUID)
}

var lockPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.LockPortResponse, error) {
	return client.PortService.LockPort(ctx, portUID)
}

var unlockPortFunc = func(ctx context.Context, client *megaport.Client, portUID string) (*megaport.UnlockPortResponse, error) {
	return client.PortService.UnlockPort(ctx, portUID)
}

var checkPortVLANAvailabilityFunc = func(ctx context.Context, client *megaport.Client, portUID string, vlan int) (bool, error) {
	return client.PortService.CheckPortVLANAvailability(ctx, portUID, vlan)
}

var buyPortFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyPortRequest) (*megaport.BuyPortResponse, error) {
	return client.PortService.BuyPort(ctx, req)
}

// Process JSON input (either from string or file)
func processJSONPortInput(jsonStr, jsonFile string) (*megaport.BuyPortRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		// Read from file
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		// Use the provided string
		jsonData = []byte(jsonStr)
	}

	// Parse JSON into request
	req := &megaport.BuyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Validate required fields
	if err := validatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process flag-based input
func processFlagPortInput(cmd *cobra.Command) (*megaport.BuyPortRequest, error) {
	// Get required fields
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")

	// Get optional fields
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	promoCode, _ := cmd.Flags().GetString("promo-code")

	req := &megaport.BuyPortRequest{
		Name:                  name,
		Term:                  term,
		PortSpeed:             portSpeed,
		LocationId:            locationID,
		MarketPlaceVisibility: marketplaceVisibility,
		DiversityZone:         diversityZone,
		CostCentre:            costCentre,
		PromoCode:             promoCode,
	}

	// Validate required fields
	if err := validatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Validate port request
func validatePortRequest(req *megaport.BuyPortRequest) error {
	if req.Name == "" {
		return fmt.Errorf("port name is required")
	}
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	if req.PortSpeed != 1000 && req.PortSpeed != 10000 && req.PortSpeed != 100000 {
		return fmt.Errorf("invalid port speed, must be one of 1000, 10000, 100000")
	}
	if req.LocationId == 0 {
		return fmt.Errorf("location ID is required")
	}
	return nil
}

// Extract the existing interactive prompting into a separate function
func promptForPortDetails(noColor bool) (*megaport.BuyPortRequest, error) {
	req := &megaport.BuyPortRequest{}

	// Prompt for required fields
	name, err := utils.Prompt("Enter port name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("port name is required")
	}
	req.Name = name

	termStr, err := utils.Prompt("Enter term (1, 12, 24, 36) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	req.Term = term

	portSpeedStr, err := utils.Prompt("Enter port speed (1000, 10000, 100000) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 1000 && portSpeed != 10000 && portSpeed != 100000) {
		return nil, fmt.Errorf("invalid port speed, must be one of 1000, 10000, 100000")
	}
	req.PortSpeed = portSpeed

	locationIDStr, err := utils.Prompt("Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID")
	}
	req.LocationId = locationID

	marketplaceVisibilityStr, err := utils.Prompt("Enter marketplace visibility (true/false) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return nil, fmt.Errorf("invalid marketplace visibility, must be true or false")
	}
	req.MarketPlaceVisibility = marketplaceVisibility

	// Prompt for optional fields
	diversityZone, err := utils.Prompt("Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.DiversityZone = diversityZone

	costCentre, err := utils.Prompt("Enter cost center (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	promoCode, err := utils.Prompt("Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	return req, nil
}

// Process flag-based input for LAG port
func processFlagLAGPortInput(cmd *cobra.Command) (*megaport.BuyPortRequest, error) {
	// Get required fields
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	portSpeed, _ := cmd.Flags().GetInt("port-speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	lagCount, _ := cmd.Flags().GetInt("lag-count")
	marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")

	// Get optional fields
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	promoCode, _ := cmd.Flags().GetString("promo-code")

	req := &megaport.BuyPortRequest{
		Name:                  name,
		Term:                  term,
		PortSpeed:             portSpeed,
		LocationId:            locationID,
		LagCount:              lagCount,
		MarketPlaceVisibility: marketplaceVisibility,
		DiversityZone:         diversityZone,
		CostCentre:            costCentre,
		PromoCode:             promoCode,
	}

	// Validate required fields
	if err := validateLAGPortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Validate LAG port request
func validateLAGPortRequest(req *megaport.BuyPortRequest) error {
	if req.Name == "" {
		return fmt.Errorf("port name is required")
	}
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	if req.PortSpeed != 10000 && req.PortSpeed != 100000 {
		return fmt.Errorf("invalid port speed, must be one of 10000 or 100000")
	}
	if req.LocationId == 0 {
		return fmt.Errorf("location ID is required")
	}
	if req.LagCount < 1 || req.LagCount > 8 {
		return fmt.Errorf("invalid LAG count, must be between 1 and 8")
	}
	return nil
}

// Extract the existing interactive prompting into a separate function for LAG port
func promptForLAGPortDetails(noColor bool) (*megaport.BuyPortRequest, error) {
	req := &megaport.BuyPortRequest{}

	// Prompt for required fields
	name, err := utils.Prompt("Enter port name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("port name is required")
	}
	req.Name = name

	termStr, err := utils.Prompt("Enter term (1, 12, 24, 36) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	req.Term = term

	portSpeedStr, err := utils.Prompt("Enter port speed (10000 or 100000) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	portSpeed, err := strconv.Atoi(portSpeedStr)
	if err != nil || (portSpeed != 10000 && portSpeed != 100000) {
		return nil, fmt.Errorf("invalid port speed, must be one of 10000 or 100000")
	}
	req.PortSpeed = portSpeed

	locationIDStr, err := utils.Prompt("Enter location ID (required): ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid location ID")
	}
	req.LocationId = locationID

	lagCountStr, err := utils.Prompt("Enter LAG count (1-8) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	lagCount, err := strconv.Atoi(lagCountStr)
	if err != nil || lagCount < 1 || lagCount > 8 {
		return nil, fmt.Errorf("invalid LAG count, must be between 1 and 8")
	}
	req.LagCount = lagCount

	marketplaceVisibilityStr, err := utils.Prompt("Enter marketplace visibility (true/false) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return nil, fmt.Errorf("invalid marketplace visibility, must be true or false")
	}
	req.MarketPlaceVisibility = marketplaceVisibility

	// Prompt for optional fields
	diversityZone, err := utils.Prompt("Enter diversity zone (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.DiversityZone = diversityZone

	costCentre, err := utils.Prompt("Enter cost center (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	promoCode, err := utils.Prompt("Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	return req, nil
}

// Process JSON input (either from string or file) for updating port
func processJSONUpdatePortInput(jsonStr, jsonFile string) (*megaport.ModifyPortRequest, error) {
	var jsonData []byte
	var err error

	if jsonFile != "" {
		// Read from file
		jsonData, err = os.ReadFile(jsonFile)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
	} else {
		// Use the provided string
		jsonData = []byte(jsonStr)
	}

	// Parse JSON into request
	req := &megaport.ModifyPortRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	// Validate required fields
	if err := validateUpdatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Process flag-based input for updating port
func processFlagUpdatePortInput(cmd *cobra.Command, portUID string) (*megaport.ModifyPortRequest, error) {
	// Get required fields
	name, _ := cmd.Flags().GetString("name")
	marketplaceVisibility, _ := cmd.Flags().GetBool("marketplace-visibility")

	// Get optional fields
	costCentre, _ := cmd.Flags().GetString("cost-centre")
	term, _ := cmd.Flags().GetInt("term")

	req := &megaport.ModifyPortRequest{
		PortID:                portUID,
		Name:                  name,
		MarketplaceVisibility: &marketplaceVisibility,
		CostCentre:            costCentre,
	}

	if term != 0 {
		req.ContractTermMonths = &term
	}

	// Validate required fields
	if err := validateUpdatePortRequest(req); err != nil {
		return nil, err
	}

	return req, nil
}

// Validate update port request
func validateUpdatePortRequest(req *megaport.ModifyPortRequest) error {
	if req.Name == "" {
		return fmt.Errorf("port name is required")
	}
	if req.MarketplaceVisibility == nil {
		return fmt.Errorf("marketplace visibility is required")
	}
	if req.ContractTermMonths != nil && *req.ContractTermMonths != 1 && *req.ContractTermMonths != 12 && *req.ContractTermMonths != 24 && *req.ContractTermMonths != 36 {
		return fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
	}
	return nil
}

// Extract the existing interactive prompting into a separate function for updating port
func promptForUpdatePortDetails(portUID string, noColor bool) (*megaport.ModifyPortRequest, error) {
	req := &megaport.ModifyPortRequest{
		PortID: portUID,
	}

	// Prompt for required fields
	name, err := utils.Prompt("Enter new port name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("port name is required")
	}
	req.Name = name

	marketplaceVisibilityStr, err := utils.Prompt("Enter marketplace visibility (true/false) (required): ", noColor)
	if err != nil {
		return nil, err
	}
	marketplaceVisibility, err := strconv.ParseBool(marketplaceVisibilityStr)
	if err != nil {
		return nil, fmt.Errorf("invalid marketplace visibility, must be true or false")
	}
	req.MarketplaceVisibility = &marketplaceVisibility

	// Prompt for optional fields
	costCentre, err := utils.Prompt("Enter cost center (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	termStr, err := utils.Prompt("Enter new term (1, 12, 24, 36) (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	if termStr != "" {
		term, err := strconv.Atoi(termStr)
		if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
			return nil, fmt.Errorf("invalid term, must be one of 1, 12, 24, 36")
		}
		req.ContractTermMonths = &term
	}

	return req, nil
}

// filterPorts filters the provided ports based on the given filters.
func filterPorts(ports []*megaport.Port, locationID int, portSpeed int, portName string) []*megaport.Port {
	if ports == nil {
		return []*megaport.Port{}
	}

	filteredPorts := make([]*megaport.Port, 0)

	for _, port := range ports {
		if port == nil {
			continue
		}

		// Apply location ID filter
		if locationID != 0 && port.LocationID != locationID {
			continue
		}

		// Apply port speed filter
		if portSpeed != 0 && port.PortSpeed != portSpeed {
			continue
		}

		// Apply port name filter
		if portName != "" && port.Name != portName {
			continue
		}

		// Port passed all filters
		filteredPorts = append(filteredPorts, port)
	}

	return filteredPorts
}

// PortOutput represents the desired fields for JSON output.
type PortOutput struct {
	output.Output      `json:"-" header:"-"`
	UID                string `json:"uid"`
	Name               string `json:"name"`
	LocationID         int    `json:"location_id"`
	PortSpeed          int    `json:"port_speed"`
	ProvisioningStatus string `json:"provisioning_status"`
}

// ToPortOutput converts a *megaport.Port to our PortOutput struct.
func ToPortOutput(port *megaport.Port) (PortOutput, error) {
	if port == nil {
		return PortOutput{}, fmt.Errorf("invalid port: nil value")
	}

	return PortOutput{
		UID:                port.UID,
		Name:               port.Name,
		LocationID:         port.LocationID,
		PortSpeed:          port.PortSpeed,
		ProvisioningStatus: port.ProvisioningStatus,
	}, nil
}

func printPorts(ports []*megaport.Port, format string, noColor bool) error {
	outputs := make([]PortOutput, 0, len(ports))
	for _, port := range ports {
		output, err := ToPortOutput(port)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return output.PrintOutput(outputs, format, noColor)
}

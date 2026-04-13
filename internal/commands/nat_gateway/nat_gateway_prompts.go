package nat_gateway

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

func promptForCreateNATGatewayDetails(noColor bool) (*megaport.CreateNATGatewayRequest, error) {
	req := &megaport.CreateNATGatewayRequest{}

	name, err := utils.ResourcePrompt("nat-gateway", "NAT Gateway name: ", noColor)
	if err != nil {
		return nil, err
	}
	req.ProductName = strings.TrimSpace(name)

	locationIDStr, err := utils.ResourcePrompt("nat-gateway", "Location ID: ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(strings.TrimSpace(locationIDStr))
	if err != nil || locationID < 1 {
		return nil, fmt.Errorf("invalid location ID: %s", locationIDStr)
	}
	req.LocationID = locationID

	speedStr, err := utils.ResourcePrompt("nat-gateway", "Speed (Mbps): ", noColor)
	if err != nil {
		return nil, err
	}
	speed, err := strconv.Atoi(strings.TrimSpace(speedStr))
	if err != nil || speed < 1 {
		return nil, fmt.Errorf("invalid speed: %s", speedStr)
	}
	req.Speed = speed

	termStr, err := utils.ResourcePrompt("nat-gateway",
		fmt.Sprintf("Contract term (%s months): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(strings.TrimSpace(termStr))
	if err != nil {
		return nil, fmt.Errorf("invalid term: %s", termStr)
	}
	req.Term = term

	sessionCountStr, err := utils.ResourcePrompt("nat-gateway", "Session count (optional, leave empty for default): ", noColor)
	if err != nil {
		return nil, err
	}
	sessionCountStr = strings.TrimSpace(sessionCountStr)
	if sessionCountStr != "" {
		sc, err := strconv.Atoi(sessionCountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session count: %s", sessionCountStr)
		}
		req.Config.SessionCount = sc
	}

	diversityZone, err := utils.ResourcePrompt("nat-gateway", "Diversity zone (optional, leave empty to skip): ", noColor)
	if err != nil {
		return nil, err
	}
	req.Config.DiversityZone = strings.TrimSpace(diversityZone)

	if err := validation.ValidateCreateNATGatewayRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func promptForUpdateNATGatewayDetails(uid string, noColor bool) (*megaport.UpdateNATGatewayRequest, error) {
	req := &megaport.UpdateNATGatewayRequest{ProductUID: uid}

	name, err := utils.ResourcePrompt("nat-gateway", "NAT Gateway name: ", noColor)
	if err != nil {
		return nil, err
	}
	req.ProductName = strings.TrimSpace(name)

	locationIDStr, err := utils.ResourcePrompt("nat-gateway", "Location ID: ", noColor)
	if err != nil {
		return nil, err
	}
	locationID, err := strconv.Atoi(strings.TrimSpace(locationIDStr))
	if err != nil || locationID < 1 {
		return nil, fmt.Errorf("invalid location ID: %s", locationIDStr)
	}
	req.LocationID = locationID

	speedStr, err := utils.ResourcePrompt("nat-gateway", "Speed (Mbps): ", noColor)
	if err != nil {
		return nil, err
	}
	speed, err := strconv.Atoi(strings.TrimSpace(speedStr))
	if err != nil || speed < 1 {
		return nil, fmt.Errorf("invalid speed: %s", speedStr)
	}
	req.Speed = speed

	termStr, err := utils.ResourcePrompt("nat-gateway",
		fmt.Sprintf("Contract term (%s months): ", validation.FormatIntSlice(validation.ValidContractTerms)), noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(strings.TrimSpace(termStr))
	if err != nil {
		return nil, fmt.Errorf("invalid term: %s", termStr)
	}
	req.Term = term

	sessionCountStr, err := utils.ResourcePrompt("nat-gateway", "Session count (optional, leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	sessionCountStr = strings.TrimSpace(sessionCountStr)
	if sessionCountStr != "" {
		sc, err := strconv.Atoi(sessionCountStr)
		if err != nil {
			return nil, fmt.Errorf("invalid session count: %s", sessionCountStr)
		}
		req.Config.SessionCount = sc
	}

	diversityZone, err := utils.ResourcePrompt("nat-gateway", "Diversity zone (optional, leave empty to keep current): ", noColor)
	if err != nil {
		return nil, err
	}
	req.Config.DiversityZone = strings.TrimSpace(diversityZone)

	if err := validation.ValidateUpdateNATGatewayRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

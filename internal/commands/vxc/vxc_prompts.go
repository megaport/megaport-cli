package vxc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
)

// buildVXCRequestFromPrompt creates a BuyVXCRequest from interactive prompts
var buildVXCRequestFromPrompt = func(ctx context.Context, svc megaport.VXCService, noColor bool) (*megaport.BuyVXCRequest, error) {

	name, err := utils.ResourcePrompt("vxc", "Enter VXC name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, validation.NewValidationError("VXC name", name, "cannot be empty")
	}

	rateLimitStr, err := utils.ResourcePrompt("vxc", "Enter rate limit in Mbps (required): ", noColor)
	if err != nil {
		return nil, err
	}
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil {
		return nil, fmt.Errorf("rate limit must be a valid integer")
	}
	if err := validation.ValidateRateLimit(rateLimit); err != nil {
		return nil, err
	}

	termStr, err := utils.ResourcePrompt("vxc", "Enter term in months (1, 12, 24, or 36, required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil {
		return nil, fmt.Errorf("term must be a valid integer")
	}
	if err := validation.ValidateContractTerm(term); err != nil {
		return nil, err
	}

	// A-End configuration
	aEndVLANStr, err := utils.ResourcePrompt("vxc", "A-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
	if err != nil {
		return nil, err
	}
	var aEndVLAN int
	if aEndVLANStr != "" {
		aEndVLAN, err = strconv.Atoi(aEndVLANStr)
		if err != nil {
			return nil, fmt.Errorf("A-End VLAN must be a valid integer")
		}
		if err := validation.ValidateVXCEndVLAN(aEndVLAN, "A-End"); err != nil {
			return nil, err
		}
	}

	aEndInnerVLANStr, err := utils.ResourcePrompt("vxc", "Enter A-End Inner VLAN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	aEndInnerVLAN := 0
	if aEndInnerVLANStr != "" {
		aEndInnerVLAN, err = strconv.Atoi(aEndInnerVLANStr)
		if err != nil {
			return nil, fmt.Errorf("invalid A-End Inner VLAN")
		}
		if err := validation.ValidateVXCEndInnerVLAN(aEndInnerVLAN, aEndVLAN, "A-End"); err != nil {
			return nil, err
		}
	}

	aEndVNICIndexStr, err := utils.ResourcePrompt("vxc", "Enter A-End vNIC Index (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	aEndVNICIndex := 0
	if aEndVNICIndexStr != "" {
		aEndVNICIndex, err = strconv.Atoi(aEndVNICIndexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid A-End vNIC Index")
		}
	}

	// Ask if A-End has partner config
	hasAEndPartnerConfig, err := utils.ResourcePrompt("vxc", "Do you want to configure A-End partner? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	aEndConfig := megaport.VXCOrderEndpointConfiguration{
		VLAN: aEndVLAN,
	}

	if aEndInnerVLAN != 0 || aEndVNICIndex > 0 {
		aEndConfig.VXCOrderMVEConfig = &megaport.VXCOrderMVEConfig{
			InnerVLAN:             aEndInnerVLAN,
			NetworkInterfaceIndex: aEndVNICIndex,
		}
	}

	// Create the base request
	req := &megaport.BuyVXCRequest{
		VXCName:   name,
		RateLimit: rateLimit,
		Term:      term,
	}

	if strings.ToLower(hasAEndPartnerConfig) == "yes" {
		aEndPartnerConfig, uid, err := promptPartnerConfig("A-End", ctx, svc, noColor)
		if err != nil {
			return nil, err
		}
		aEndConfig.PartnerConfig = aEndPartnerConfig
		if uid != "" {
			req.PortUID = uid
		}
	}

	if req.PortUID == "" {
		// Prompt for the required fields
		aEndUID, err := utils.ResourcePrompt("vxc", "Enter A-End product UID (required): ", noColor)
		if err != nil {
			return nil, err
		}
		if aEndUID == "" {
			return nil, fmt.Errorf("a-end product UID is required")
		}
		req.PortUID = aEndUID
	}
	req.AEndConfiguration = aEndConfig

	bEndConfig := megaport.VXCOrderEndpointConfiguration{}

	bEndVLANStr, err := utils.ResourcePrompt("vxc", "B-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
	if err != nil {
		return nil, err
	}
	var bEndVLAN int
	if bEndVLANStr != "" {
		bEndVLAN, err = strconv.Atoi(bEndVLANStr)
		if err != nil {
			return nil, fmt.Errorf("B-End VLAN must be a valid integer")
		}
		if err := validation.ValidateVXCEndVLAN(bEndVLAN, "B-End"); err != nil {
			return nil, err
		}
		req.BEndConfiguration.VLAN = bEndVLAN
	}

	bEndInnerVLANStr, err := utils.ResourcePrompt("vxc", "Enter B-End Inner VLAN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	bEndInnerVLAN := 0
	if bEndInnerVLANStr != "" {
		bEndInnerVLAN, err = strconv.Atoi(bEndInnerVLANStr)
		if err != nil {
			return nil, fmt.Errorf("invalid B-End Inner VLAN")
		}
		if err := validation.ValidateVXCEndInnerVLAN(bEndInnerVLAN, bEndVLAN, "B-End"); err != nil {
			return nil, err
		}
	}

	bEndVNICIndexStr, err := utils.ResourcePrompt("vxc", "Enter B-End vNIC Index (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	bEndVNICIndex := 0
	if bEndVNICIndexStr != "" {
		bEndVNICIndex, err = strconv.Atoi(bEndVNICIndexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid B-End vNIC Index")
		}
	}

	if bEndInnerVLAN != 0 || bEndVNICIndex > 0 {
		bEndConfig.VXCOrderMVEConfig = &megaport.VXCOrderMVEConfig{
			InnerVLAN:             bEndInnerVLAN,
			NetworkInterfaceIndex: bEndVNICIndex,
		}
	}

	hasBEndPartnerConfig, err := utils.ResourcePrompt("vxc", "Do you want to configure B-End partner? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(hasBEndPartnerConfig) == "yes" {
		bEndPartnerConfig, uid, err := promptPartnerConfig("B-End", ctx, svc, noColor)
		if err != nil {
			return nil, err
		}
		if uid != "" {
			bEndConfig.ProductUID = uid
		}
		bEndConfig.PartnerConfig = bEndPartnerConfig
	}

	if bEndConfig.ProductUID == "" {
		bEndUID, err := utils.ResourcePrompt("vxc", "Enter B-End product UID (required): ", noColor)
		if err != nil {
			return nil, err
		}
		if bEndUID == "" {
			return nil, fmt.Errorf("B-End product UID is required")
		}
		bEndConfig.ProductUID = bEndUID
	}

	req.BEndConfiguration = bEndConfig

	// Optional fields
	promoCode, err := utils.ResourcePrompt("vxc", "Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	serviceKey, err := utils.ResourcePrompt("vxc", "Enter service key (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.ServiceKey = serviceKey

	costCentre, err := utils.ResourcePrompt("vxc", "Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	return req, nil
}

// buildUpdateVXCRequestFromPrompt creates an UpdateVXCRequest from interactive prompts
var buildUpdateVXCRequestFromPrompt = func(vxcUID string, noColor bool) (*megaport.UpdateVXCRequest, error) {
	req := &megaport.UpdateVXCRequest{
		WaitForUpdate: true,
		WaitForTime:   5 * time.Minute,
	}

	// Fetch the current VXC to show current values
	ctx := context.Background()
	client, err := config.Login(ctx)
	if err != nil {
		return nil, err
	}

	fmt.Println("Fetching current VXC details...")
	vxc, err := client.VXCService.GetVXC(ctx, vxcUID)
	if err != nil {
		return nil, fmt.Errorf("error fetching VXC details: %v", err)
	}

	// Name
	fmt.Printf("Current name: %s\n", vxc.Name)
	updateName, err := utils.ResourcePrompt("vxc", "Update name? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateName) == "yes" {
		name, err := utils.ResourcePrompt("vxc", "Enter new name: ", noColor)
		if err != nil {
			return nil, err
		}
		req.Name = &name
	}

	// Rate limit
	fmt.Printf("Current rate limit: %d Mbps\n", vxc.RateLimit)
	updateRateLimit, err := utils.ResourcePrompt("vxc", "Update rate limit? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateRateLimit) == "yes" {
		rateLimitStr, err := utils.ResourcePrompt("vxc", "Enter new rate limit in Mbps: ", noColor)
		if err != nil {
			return nil, err
		}
		rateLimit, err := strconv.Atoi(rateLimitStr)
		if err != nil {
			return nil, fmt.Errorf("rate limit must be a valid integer")
		}
		if err := validation.ValidateRateLimit(rateLimit); err != nil {
			return nil, err
		}
		req.RateLimit = &rateLimit
	}

	// Term
	fmt.Printf("Current term: %d months\n", vxc.ContractTermMonths)
	updateTerm, err := utils.ResourcePrompt("vxc", "Update term? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateTerm) == "yes" {
		termStr, err := utils.ResourcePrompt("vxc", "Enter new term in months (0, 1, 12, 24, or 36): ", noColor)
		if err != nil {
			return nil, err
		}
		term, err := strconv.Atoi(termStr)
		if err != nil {
			return nil, fmt.Errorf("term must be a valid integer")
		}
		// Allow 0 for term updates (no change)
		if term != 0 && validation.ValidateContractTerm(term) != nil {
			return nil, validation.NewValidationError("term", term,
				fmt.Sprintf("must be 0, or one of: %v", validation.ValidContractTerms))
		}
		req.Term = &term
	}

	// Cost centre
	fmt.Printf("Current cost centre: %s\n", vxc.CostCentre)
	updateCostCentre, err := utils.ResourcePrompt("vxc", "Update cost centre? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateCostCentre) == "yes" {
		costCentre, err := utils.ResourcePrompt("vxc", "Enter new cost centre: ", noColor)
		if err != nil {
			return nil, err
		}
		req.CostCentre = &costCentre
	}

	// Shutdown
	shutdownStatus := "No"
	if vxc.AdminLocked {
		shutdownStatus = "Yes"
	}
	fmt.Printf("Current shutdown status: %s\n", shutdownStatus)
	updateShutdown, err := utils.ResourcePrompt("vxc", "Update shutdown status? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateShutdown) == "yes" {
		shutdownStr, err := utils.ResourcePrompt("vxc", "Shut down the VXC? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		shutdown := strings.ToLower(shutdownStr) == "yes"
		req.Shutdown = &shutdown
	}

	// A-End VLAN
	fmt.Printf("Current A-End VLAN: %d\n", vxc.AEndConfiguration.VLAN)
	updateAEndVLAN, err := utils.ResourcePrompt("vxc", "Update A-End VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateAEndVLAN) == "yes" {
		aEndVLANStr, err := utils.ResourcePrompt("vxc", "A-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
		if err != nil {
			return nil, err
		}
		aEndVLAN, err := strconv.Atoi(aEndVLANStr)
		if err != nil {
			return nil, fmt.Errorf("A-End VLAN must be a valid integer")
		}
		if err := validation.ValidateVXCEndVLAN(aEndVLAN, "A-End"); err != nil {
			return nil, err
		}
		req.AEndVLAN = &aEndVLAN
	}

	// B-End VLAN
	fmt.Printf("Current B-End VLAN: %d\n", vxc.BEndConfiguration.VLAN)
	updateBEndVLAN, err := utils.ResourcePrompt("vxc", "Update B-End VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateBEndVLAN) == "yes" {
		bEndVLANStr, err := utils.ResourcePrompt("vxc", "B-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
		if err != nil {
			return nil, err
		}
		bEndVLAN, err := strconv.Atoi(bEndVLANStr)
		if err != nil {
			return nil, fmt.Errorf("B-End VLAN must be a valid integer")
		}
		if err := validation.ValidateVXCEndVLAN(bEndVLAN, "B-End"); err != nil {
			return nil, err
		}
		req.BEndVLAN = &bEndVLAN
	}

	// A-End Inner VLAN
	innerVLANAEnd := 0
	if vxc.AEndConfiguration.InnerVLAN != 0 {
		innerVLANAEnd = vxc.AEndConfiguration.InnerVLAN
	}
	fmt.Printf("Current A-End Inner VLAN: %d\n", innerVLANAEnd)
	updateAEndInnerVLAN, err := utils.ResourcePrompt("vxc", "Update A-End Inner VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateAEndInnerVLAN) == "yes" {
		aEndInnerVLANStr, err := utils.ResourcePrompt("vxc", "Enter new A-End Inner VLAN (-1, 0, or 2-4093): ", noColor)
		if err != nil {
			return nil, err
		}
		aEndInnerVLAN, err := strconv.Atoi(aEndInnerVLANStr)
		if err != nil {
			return nil, fmt.Errorf("A-End Inner VLAN must be a valid integer")
		}

		// Get current A-End VLAN for validation context
		aEndVLAN := vxc.AEndConfiguration.VLAN
		if req.AEndVLAN != nil {
			aEndVLAN = *req.AEndVLAN
		}

		if err := validation.ValidateVXCEndInnerVLAN(aEndInnerVLAN, aEndVLAN, "A-End"); err != nil {
			return nil, err
		}

		req.AEndInnerVLAN = &aEndInnerVLAN
	}

	// B-End Inner VLAN
	innerVLANBEnd := 0
	if vxc.BEndConfiguration.InnerVLAN != 0 {
		innerVLANBEnd = vxc.BEndConfiguration.InnerVLAN
	}
	fmt.Printf("Current B-End Inner VLAN: %d\n", innerVLANBEnd)
	updateBEndInnerVLAN, err := utils.ResourcePrompt("vxc", "Update B-End Inner VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateBEndInnerVLAN) == "yes" {
		bEndInnerVLANStr, err := utils.ResourcePrompt("vxc", "Enter new B-End Inner VLAN (-1, 0, or 2-4093): ", noColor)
		if err != nil {
			return nil, err
		}
		bEndInnerVLAN, err := strconv.Atoi(bEndInnerVLANStr)
		if err != nil {
			return nil, fmt.Errorf("B-End Inner VLAN must be a valid integer")
		}

		// Get current B-End VLAN for validation context
		bEndVLAN := vxc.BEndConfiguration.VLAN
		if req.BEndVLAN != nil {
			bEndVLAN = *req.BEndVLAN
		}

		if err := validation.ValidateVXCEndInnerVLAN(bEndInnerVLAN, bEndVLAN, "B-End"); err != nil {
			return nil, err
		}

		req.BEndInnerVLAN = &bEndInnerVLAN
	}

	// A-End UID
	fmt.Printf("Current A-End UID: %s\n", vxc.AEndConfiguration.UID)
	updateAEndUID, err := utils.ResourcePrompt("vxc", "Update A-End product UID? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateAEndUID) == "yes" {
		aEndUID, err := utils.ResourcePrompt("vxc", "Enter new A-End product UID: ", noColor)
		if err != nil {
			return nil, err
		}
		req.AEndProductUID = &aEndUID
	}

	// B-End UID
	fmt.Printf("Current B-End UID: %s\n", vxc.BEndConfiguration.UID)
	updateBEndUID, err := utils.ResourcePrompt("vxc", "Update B-End product UID? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateBEndUID) == "yes" {
		bEndUID, err := utils.ResourcePrompt("vxc", "Enter new B-End product UID: ", noColor)
		if err != nil {
			return nil, err
		}
		req.BEndProductUID = &bEndUID
	}

	wantsAEndPartnerConfig, err := utils.ResourcePrompt("vxc", "Do you want to configure an A-End VRouter partner configuration? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(wantsAEndPartnerConfig) == "yes" {
		aEndPartnerConfig, err := promptVRouterConfig("A-End", noColor)
		if err != nil {
			return nil, err
		}
		req.BEndPartnerConfig = aEndPartnerConfig
	}

	wantsBEndPartnerConfig, err := utils.ResourcePrompt("vxc", "Do you want to configure a B-End VRouter partner configuration? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}

	if strings.ToLower(wantsBEndPartnerConfig) == "yes" {
		bEndPartnerConfig, err := promptVRouterConfig("B-End", noColor)
		if err != nil {
			return nil, err
		}
		req.BEndPartnerConfig = bEndPartnerConfig
	}

	return req, nil
}

// promptVRouterConfig prompts the user for VRouter-specific configuration details.
func promptVRouterConfig(endpoint string, noColor bool) (*megaport.VXCOrderVrouterPartnerConfig, error) {
	fmt.Printf("\n=== %s VRouter Configuration ===\n", endpoint)

	config := &megaport.VXCOrderVrouterPartnerConfig{
		Interfaces: []megaport.PartnerConfigInterface{},
	}

	// Ask for number of interfaces
	interfaceCountStr, err := utils.ResourcePrompt("vxc", "Number of interfaces to configure: ", noColor)
	if err != nil {
		return nil, err
	}
	interfaceCount, err := strconv.Atoi(interfaceCountStr)
	if err != nil || interfaceCount < 1 {
		return nil, fmt.Errorf("number of interfaces must be a positive integer")
	}

	// Configure each interface
	for i := 0; i < interfaceCount; i++ {
		fmt.Printf("\n--- Interface %d ---\n", i+1)

		iface := megaport.PartnerConfigInterface{}

		// VLAN
		vlanStr, err := utils.ResourcePrompt("vxc", "VLAN (0-4093, except 1, optional - press Enter for no VLAN): ", noColor)
		if err != nil {
			return nil, err
		}

		if vlanStr != "" {
			vlan, err := strconv.Atoi(vlanStr)
			if err != nil {
				return nil, fmt.Errorf("VLAN must be a valid integer")
			}
			// Use custom validation logic for partner interface VLANs
			if vlan < 0 || vlan > 4093 || vlan == 1 {
				return nil, validation.NewValidationError("VRouter interface VLAN", vlan,
					"must be 0 or between 2-4093 (1 is reserved)")
			}
			iface.VLAN = vlan
		} else {
			iface.VLAN = -1
		}

		// IP Addresses
		ipAddrs, err := promptIPAddresses("IP Addresses (CIDR notation, e.g., 192.168.1.1/30)", noColor)
		if err != nil {
			return nil, err
		}
		iface.IpAddresses = ipAddrs

		// IP Routes
		hasRoutes, err := utils.ResourcePrompt("vxc", "Do you want to add IP routes? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasRoutes) == "yes" {
			routes, err := promptIPRoutes(noColor)
			if err != nil {
				return nil, err
			}
			iface.IpRoutes = routes
		}

		// NAT IP Addresses
		hasNatIPs, err := utils.ResourcePrompt("vxc", "Do you want to add NAT IP addresses? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasNatIPs) == "yes" {
			natIPs, err := promptNATIPAddresses(noColor)
			if err != nil {
				return nil, err
			}
			iface.NatIpAddresses = natIPs
		}

		// BFD Configuration
		hasBFD, err := utils.ResourcePrompt("vxc", "Do you want to configure BFD? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasBFD) == "yes" {
			bfd, err := promptBFDConfig(noColor)
			if err != nil {
				return nil, err
			}
			iface.Bfd = bfd
		}

		// BGP Connections
		hasBGP, err := utils.ResourcePrompt("vxc", "Do you want to configure BGP connections? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasBGP) == "yes" {
			bgpConns, err := promptBGPConnections(noColor)
			if err != nil {
				return nil, err
			}
			iface.BgpConnections = bgpConns
		}

		config.Interfaces = append(config.Interfaces, iface)
	}

	return config, nil
}

// promptIPRoutes prompts the user for IP routes
func promptIPRoutes(noColor bool) ([]megaport.IpRoute, error) {
	var routes []megaport.IpRoute

	for {
		addRoute, err := utils.ResourcePrompt("vxc", "Add an IP route? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addRoute) != "yes" {
			break
		}

		prefix, err := utils.ResourcePrompt("vxc", "Enter prefix (e.g., 192.168.0.0/24): ", noColor)
		if err != nil {
			return nil, err
		}

		nextHop, err := utils.ResourcePrompt("vxc", "Enter next hop IP: ", noColor)
		if err != nil {
			return nil, err
		}

		description, err := utils.ResourcePrompt("vxc", "Enter description (optional): ", noColor)
		if err != nil {
			return nil, err
		}

		route := megaport.IpRoute{
			Prefix:      prefix,
			NextHop:     nextHop,
			Description: description,
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// promptIPAddresses prompts the user for IP addresses
func promptIPAddresses(message string, noColor bool) ([]string, error) {
	var addresses []string

	for {
		addIP, err := utils.ResourcePrompt("vxc", fmt.Sprintf("Add %s? (yes/no): ", message), noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addIP) != "yes" {
			break
		}

		ip, err := utils.ResourcePrompt("vxc", "Enter IP address: ", noColor)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, ip)
	}

	return addresses, nil
}

// promptNATIPAddresses prompts the user for NAT IP addresses
func promptNATIPAddresses(noColor bool) ([]string, error) {
	return promptIPAddresses("a NAT IP address", noColor)
}

// promptBFDConfig prompts the user for BFD configuration details
func promptBFDConfig(noColor bool) (megaport.BfdConfig, error) {
	bfd := megaport.BfdConfig{}

	txIntervalStr, err := utils.ResourcePrompt("vxc", "Enter transmit interval in ms (default 300): ", noColor)
	if err != nil {
		return bfd, err
	}
	if txIntervalStr != "" {
		txInterval, err := strconv.Atoi(txIntervalStr)
		if err != nil {
			return bfd, fmt.Errorf("transmit interval must be an integer")
		}
		bfd.TxInterval = txInterval
	} else {
		bfd.TxInterval = 300
	}

	rxIntervalStr, err := utils.ResourcePrompt("vxc", "Enter receive interval in ms (default 300): ", noColor)
	if err != nil {
		return bfd, err
	}
	if rxIntervalStr != "" {
		rxInterval, err := strconv.Atoi(rxIntervalStr)
		if err != nil {
			return bfd, fmt.Errorf("receive interval must be an integer")
		}
		bfd.RxInterval = rxInterval
	} else {
		bfd.RxInterval = 300
	}

	multiplierStr, err := utils.ResourcePrompt("vxc", "Enter multiplier (default 3): ", noColor)
	if err != nil {
		return bfd, err
	}
	if multiplierStr != "" {
		multiplier, err := strconv.Atoi(multiplierStr)
		if err != nil {
			return bfd, fmt.Errorf("multiplier must be an integer")
		}
		bfd.Multiplier = multiplier
	} else {
		bfd.Multiplier = 3
	}

	return bfd, nil
}

// promptBGPConnections prompts the user for BGP connections
func promptBGPConnections(noColor bool) ([]megaport.BgpConnectionConfig, error) {
	var bgpConnections []megaport.BgpConnectionConfig

	for {
		addBGP, err := utils.ResourcePrompt("vxc", "Add a BGP connection? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addBGP) != "yes" {
			break
		}

		bgp := megaport.BgpConnectionConfig{}

		// Required fields
		peerAsnStr, err := utils.ResourcePrompt("vxc", "Enter peer ASN (required): ", noColor)
		if err != nil {
			return nil, err
		}
		peerAsn, err := strconv.Atoi(peerAsnStr)
		if err != nil || peerAsn <= 0 {
			return nil, fmt.Errorf("peer ASN must be a positive integer")
		}
		bgp.PeerAsn = peerAsn

		localIP, err := utils.ResourcePrompt("vxc", "Enter local IP address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		if localIP == "" {
			return nil, fmt.Errorf("local IP address is required")
		}
		bgp.LocalIpAddress = localIP

		peerIP, err := utils.ResourcePrompt("vxc", "Enter peer IP address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		if peerIP == "" {
			return nil, fmt.Errorf("peer IP address is required")
		}
		bgp.PeerIpAddress = peerIP

		// Optional fields
		localAsnStr, err := utils.ResourcePrompt("vxc", "Enter local ASN (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if localAsnStr != "" {
			localAsn, err := strconv.Atoi(localAsnStr)
			if err != nil || localAsn <= 0 {
				return nil, fmt.Errorf("local ASN must be a positive integer")
			}
			bgp.LocalAsn = &localAsn
		}

		password, err := utils.ResourcePrompt("vxc", "Enter password (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.Password = password

		shutdownStr, err := utils.ResourcePrompt("vxc", "Shutdown connection? (yes/no, default: no): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.Shutdown = strings.ToLower(shutdownStr) == "yes"

		description, err := utils.ResourcePrompt("vxc", "Enter description (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.Description = description

		bfdEnabledStr, err := utils.ResourcePrompt("vxc", "Enable BFD? (yes/no, default: no): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.BfdEnabled = strings.ToLower(bfdEnabledStr) == "yes"

		// Added: Export Policy
		exportPolicy, err := utils.ResourcePrompt("vxc", "Enter export policy (permit/deny, optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if exportPolicy != "" && exportPolicy != "permit" && exportPolicy != "deny" {
			return nil, fmt.Errorf("export policy must be 'permit' or 'deny'")
		}
		bgp.ExportPolicy = exportPolicy

		// Added: Peer Type
		peerType, err := utils.ResourcePrompt("vxc", "Enter peer type (NON_CLOUD/PRIV_CLOUD/PUB_CLOUD, optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if peerType != "" && peerType != "NON_CLOUD" && peerType != "PRIV_CLOUD" && peerType != "PUB_CLOUD" {
			return nil, fmt.Errorf("peer type must be NON_CLOUD, PRIV_CLOUD, or PUB_CLOUD")
		}
		bgp.PeerType = peerType

		// Added: MED values
		medInStr, err := utils.ResourcePrompt("vxc", "Enter MED in (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if medInStr != "" {
			medIn, err := strconv.Atoi(medInStr)
			if err != nil {
				return nil, fmt.Errorf("MED in must be an integer")
			}
			bgp.MedIn = medIn
		}

		medOutStr, err := utils.ResourcePrompt("vxc", "Enter MED out (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if medOutStr != "" {
			medOut, err := strconv.Atoi(medOutStr)
			if err != nil {
				return nil, fmt.Errorf("MED out must be an integer")
			}
			bgp.MedOut = medOut
		}

		// Added: AS Path Prepend Count
		asPathPrependStr, err := utils.ResourcePrompt("vxc", "Enter AS path prepend count (0-10, optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if asPathPrependStr != "" {
			asPathPrepend, err := strconv.Atoi(asPathPrependStr)
			if err != nil || asPathPrepend < 0 || asPathPrepend > 10 {
				return nil, fmt.Errorf("AS path prepend count must be between 0 and 10")
			}
			bgp.AsPathPrependCount = asPathPrepend
		}

		// Added: Permit Export To
		hasPermitExportTo, err := utils.ResourcePrompt("vxc", "Add permit export to addresses? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasPermitExportTo) == "yes" {
			for i := 0; i < 17; i++ { // Maximum 17 items
				ipAddress, err := utils.ResourcePrompt("vxc", fmt.Sprintf("Enter IP address to permit export to (or empty to finish) [%d/17]: ", i+1), noColor)
				if err != nil {
					return nil, err
				}
				if ipAddress == "" {
					break
				}
				bgp.PermitExportTo = append(bgp.PermitExportTo, ipAddress)
			}
		}

		// Added: Deny Export To
		hasDenyExportTo, err := utils.ResourcePrompt("vxc", "Add deny export to addresses? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasDenyExportTo) == "yes" {
			for i := 0; i < 17; i++ { // Maximum 17 items
				ipAddress, err := utils.ResourcePrompt("vxc", fmt.Sprintf("Enter IP address to deny export to (or empty to finish) [%d/17]: ", i+1), noColor)
				if err != nil {
					return nil, err
				}
				if ipAddress == "" {
					break
				}
				bgp.DenyExportTo = append(bgp.DenyExportTo, ipAddress)
			}
		}

		// Added: Import/Export Whitelist/Blacklist
		importWhitelistStr, err := utils.ResourcePrompt("vxc", "Enter import whitelist prefix list ID (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if importWhitelistStr != "" {
			importWhitelist, err := strconv.Atoi(importWhitelistStr)
			if err != nil {
				return nil, fmt.Errorf("import whitelist must be an integer")
			}
			bgp.ImportWhitelist = importWhitelist
		}

		importBlacklistStr, err := utils.ResourcePrompt("vxc", "Enter import blacklist prefix list ID (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if importBlacklistStr != "" {
			importBlacklist, err := strconv.Atoi(importBlacklistStr)
			if err != nil {
				return nil, fmt.Errorf("import blacklist must be an integer")
			}
			bgp.ImportBlacklist = importBlacklist
		}

		exportWhitelistStr, err := utils.ResourcePrompt("vxc", "Enter export whitelist prefix list ID (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if exportWhitelistStr != "" {
			exportWhitelist, err := strconv.Atoi(exportWhitelistStr)
			if err != nil {
				return nil, fmt.Errorf("export whitelist must be an integer")
			}
			bgp.ExportWhitelist = exportWhitelist
		}

		exportBlacklistStr, err := utils.ResourcePrompt("vxc", "Enter export blacklist prefix list ID (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if exportBlacklistStr != "" {
			exportBlacklist, err := strconv.Atoi(exportBlacklistStr)
			if err != nil {
				return nil, fmt.Errorf("export blacklist must be an integer")
			}
			bgp.ExportBlacklist = exportBlacklist
		}

		bgpConnections = append(bgpConnections, bgp)
	}

	return bgpConnections, nil
}

func promptPartnerConfig(end string, ctx context.Context, svc megaport.VXCService, noColor bool) (megaport.VXCPartnerConfiguration, string, error) {
	partner, err := utils.ResourcePrompt("vxc", fmt.Sprintf("Enter %s partner (AWS, Azure, Google, Oracle, IBM, VRouter, Transit) (optional): ", end), noColor)
	if err != nil {
		return nil, "", err
	}
	if partner == "" {
		return nil, "", nil
	}

	switch strings.ToLower(partner) {
	case "aws":
		awsPartner, err := promptAWSConfig(noColor)
		if err != nil {
			return nil, "", err
		}
		partnerPortUID, err := utils.ResourcePrompt("vxc", "Enter AWS Partner Port product UID (required): ", noColor)
		if err != nil {
			return nil, "", err
		}
		if partnerPortUID == "" {
			return nil, "", fmt.Errorf("AWS Partner Port product UID is required")
		}
		return awsPartner, partnerPortUID, nil
	case "azure":
		azurePartner, uid, err := promptAzureConfig(ctx, svc, noColor)
		if err != nil {
			return nil, "", err
		}
		return azurePartner, uid, nil
	case "google":
		googlePartner, uid, err := promptGoogleConfig(ctx, svc, noColor)
		if err != nil {
			return nil, "", err
		}
		return googlePartner, uid, nil
	case "oracle":
		oraclePartner, uid, err := promptOracleConfig(ctx, svc, noColor)
		if err != nil {
			return nil, "", err
		}
		return oraclePartner, uid, nil
	case "ibm":
		ibmPartner, err := promptIBMConfig(noColor)
		if err != nil {
			return nil, "", err
		}
		partnerPortUID, err := utils.ResourcePrompt("vxc", "Enter IBM Partner Port product UID (required): ", noColor)
		if err != nil {
			return nil, "", err
		}
		if partnerPortUID == "" {
			return nil, "", fmt.Errorf("IBM Partner Port product UID is required")
		}
		return ibmPartner, partnerPortUID, nil
	case "vrouter":
		vrouterPartner, err := promptVRouterConfig(end, noColor)
		if err != nil {
			return nil, "", err
		}
		return vrouterPartner, "", nil
	case "transit":
		return promptTransitConfig(), "", nil
	default:
		return nil, "", fmt.Errorf("unsupported partner: %s", partner)
	}
}

func promptTransitConfig() *megaport.VXCPartnerConfigTransit {
	return &megaport.VXCPartnerConfigTransit{
		ConnectType: "TRANSIT",
	}
}

func promptAWSConfig(noColor bool) (*megaport.VXCPartnerConfigAWS, error) {
	connectType, err := utils.ResourcePrompt("vxc", "Enter connect type (required - either AWS or AWSHC): ", noColor)
	if err != nil {
		return nil, err
	}

	if connectType != "AWS" && connectType != "AWSHC" {
		return nil, fmt.Errorf("connect type must be AWS or AWSHC")
	}

	ownerAccount, err := utils.ResourcePrompt("vxc", "Enter owner account ID (required): ", noColor)
	if err != nil {
		return nil, err
	}

	connectionName, err := utils.ResourcePrompt("vxc", "Enter connection name (required): ", noColor)
	if err != nil {
		return nil, err
	}

	asnStr, err := utils.ResourcePrompt("vxc", "Enter ASN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	asn, err := strconv.Atoi(asnStr)
	if err != nil {
		asn = 0
	}

	amazonASNStr, err := utils.ResourcePrompt("vxc", "Enter Amazon ASN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	amazonASN, err := strconv.Atoi(amazonASNStr)
	if err != nil {
		amazonASN = 0
	}

	authKey, err := utils.ResourcePrompt("vxc", "Enter auth key (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	prefixes, err := utils.ResourcePrompt("vxc", "Enter prefixes (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	customerIPAddress, err := utils.ResourcePrompt("vxc", "Enter customer IP address (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	amazonIPAddress, err := utils.ResourcePrompt("vxc", "Enter Amazon IP address (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	partnerConfigAWS := &megaport.VXCPartnerConfigAWS{
		ConnectType:       connectType,
		OwnerAccount:      ownerAccount,
		ASN:               asn,
		AmazonASN:         amazonASN,
		AuthKey:           authKey,
		Prefixes:          prefixes,
		CustomerIPAddress: customerIPAddress,
		AmazonIPAddress:   amazonIPAddress,
		ConnectionName:    connectionName,
	}
	if connectType == "AWS" {
		vifType, err := utils.ResourcePrompt("vxc", "Enter VIF type (required - either private or public): ", noColor)
		if err != nil {
			return nil, err
		}
		if vifType != "private" && vifType != "public" {
			return nil, fmt.Errorf("VIF type must be private or public")
		}
		partnerConfigAWS.Type = vifType
	}
	return partnerConfigAWS, nil
}

// promptAzureConfig prompts the user for Azure-specific configuration details.
func promptAzureConfig(ctx context.Context, svc megaport.VXCService, noColor bool) (*megaport.VXCPartnerConfigAzure, string, error) {
	serviceKey, err := utils.ResourcePrompt("vxc", "Enter service key (required): ", noColor)
	if err != nil {
		return nil, "", err
	}

	portChoice, err := utils.ResourcePrompt("vxc", "Enter port choice (primary/secondary, optional, default value is primary): ", noColor)
	if err != nil {
		return nil, "", err
	}
	if portChoice != "" && portChoice != "primary" && portChoice != "secondary" {
		return nil, "", fmt.Errorf("port preference must be primary or secondary")
	}
	if portChoice == "" {
		portChoice = "primary"
	}

	var peers []megaport.PartnerOrderAzurePeeringConfig
	for {
		addPeer, err := utils.ResourcePrompt("vxc", "Add a peering config? (yes/no): ", noColor)
		if err != nil {
			return nil, "", err
		}
		if addPeer != "yes" {
			break
		}

		peerConfig, err := promptAzurePeeringConfig(noColor)
		if err != nil {
			return nil, "", err
		}
		peers = append(peers, peerConfig)
	}

	fmt.Println("Finding Azure partner port...")

	partnerPortRes, err := svc.ListPartnerPorts(ctx, &megaport.ListPartnerPortsRequest{
		Key:     serviceKey,
		Partner: "AZURE",
	})
	if err != nil {
		return nil, "", fmt.Errorf("error looking up partner ports: %v", err)
	}
	var uid string
	// find primary or secondary port
	for _, port := range partnerPortRes.Data.Megaports {
		p := &port
		if p.Type == portChoice {
			uid = p.ProductUID
		}
	}
	if uid == "" {
		return nil, "", fmt.Errorf("could not find azure port with type: %s", portChoice)
	}

	return &megaport.VXCPartnerConfigAzure{
		ConnectType: "AZURE",
		ServiceKey:  serviceKey,
		Peers:       peers,
	}, uid, nil
}

// Helper to prompt for Azure Peering Config
func promptAzurePeeringConfig(noColor bool) (megaport.PartnerOrderAzurePeeringConfig, error) {
	peeringType, err := utils.ResourcePrompt("vxc", "Enter peering type (required): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	peerASN, err := utils.ResourcePrompt("vxc", "Enter Peer ASN (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	primarySubnet, err := utils.ResourcePrompt("vxc", "Enter Primary Subnet (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	secondarySubnet, err := utils.ResourcePrompt("vxc", "Enter Secondary Subnet (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	prefixes, err := utils.ResourcePrompt("vxc", "Enter Prefixes (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	sharedKey, err := utils.ResourcePrompt("vxc", "Enter Shared Key (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	vlanStr, err := utils.ResourcePrompt("vxc", "Enter VLAN (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}
	vlan, err := strconv.Atoi(vlanStr)
	if err != nil {
		vlan = 0
	}

	return megaport.PartnerOrderAzurePeeringConfig{
		Type:            peeringType,
		PeerASN:         peerASN,
		PrimarySubnet:   primarySubnet,
		SecondarySubnet: secondarySubnet,
		Prefixes:        prefixes,
		SharedKey:       sharedKey,
		VLAN:            vlan,
	}, nil
}

func promptGoogleConfig(ctx context.Context, svc megaport.VXCService, noColor bool) (*megaport.VXCPartnerConfigGoogle, string, error) {
	pairingKey, err := utils.ResourcePrompt("vxc", "Enter pairing key (required): ", noColor)
	if err != nil {
		return nil, "", err
	}

	uid, err := getPartnerPortUID(ctx, svc, pairingKey, "GOOGLE")
	if err != nil {
		return nil, "", err
	}

	return &megaport.VXCPartnerConfigGoogle{
		ConnectType: "GOOGLE",
		PairingKey:  pairingKey,
	}, uid, nil
}

func promptOracleConfig(ctx context.Context, svc megaport.VXCService, noColor bool) (*megaport.VXCPartnerConfigOracle, string, error) {
	virtualCircuitId, err := utils.ResourcePrompt("vxc", "Enter virtual circuit ID (required): ", noColor)
	if err != nil {
		return nil, "", err
	}

	uid, err := getPartnerPortUID(ctx, svc, virtualCircuitId, "ORACLE")
	if err != nil {
		return nil, "", err
	}

	return &megaport.VXCPartnerConfigOracle{
		ConnectType:      "ORACLE",
		VirtualCircuitId: virtualCircuitId,
	}, uid, nil
}

func promptIBMConfig(noColor bool) (*megaport.VXCPartnerConfigIBM, error) {
	accountID, err := utils.ResourcePrompt("vxc", "Enter account ID (required): ", noColor)
	if err != nil {
		return nil, err
	}

	name, err := utils.ResourcePrompt("vxc", "Enter name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	var customerASN int

	customerASNStr, err := utils.ResourcePrompt("vxc", "Enter customer ASN (required if opposite end is not an MCR): ", noColor)
	if err != nil {
		return nil, err
	}
	customerASN, err = strconv.Atoi(customerASNStr)
	if err != nil {
		return nil, err
	}

	customerIPAddress, err := utils.ResourcePrompt("vxc", "Enter customer IP address (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	providerIPAddress, err := utils.ResourcePrompt("vxc", "Enter provider IP address (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	partnerConfig := &megaport.VXCPartnerConfigIBM{
		ConnectType:       "IBM",
		AccountID:         accountID,
		CustomerIPAddress: customerIPAddress,
		ProviderIPAddress: providerIPAddress,
		Name:              name,
	}

	if customerASN != 0 {
		partnerConfig.CustomerASN = customerASN
	}

	return partnerConfig, nil
}

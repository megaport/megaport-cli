package vxc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

// buildVXCRequestFromFlags creates a BuyVXCRequest from command flags
var buildVXCRequestFromFlags = func(cmd *cobra.Command, ctx context.Context, svc megaport.VXCService) (*megaport.BuyVXCRequest, error) {
	aEndUID, _ := cmd.Flags().GetString("a-end-uid")

	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	rateLimit, _ := cmd.Flags().GetInt("rate-limit")
	if rateLimit <= 0 {
		return nil, fmt.Errorf("rate-limit must be greater than 0")
	}

	term, _ := cmd.Flags().GetInt("term")
	if term != 1 && term != 12 && term != 24 && term != 36 {
		return nil, fmt.Errorf("term must be 1, 12, 24, or 36")
	}

	// Get optional fields
	aEndVLAN, _ := cmd.Flags().GetInt("a-end-vlan")
	bEndVLAN, _ := cmd.Flags().GetInt("b-end-vlan")
	aEndInnerVLAN, _ := cmd.Flags().GetInt("a-end-inner-vlan")
	bEndInnerVLAN, _ := cmd.Flags().GetInt("b-end-inner-vlan")
	aEndVNICIndex, _ := cmd.Flags().GetInt("a-end-vnic-index")
	bEndVNICIndex, _ := cmd.Flags().GetInt("b-end-vnic-index")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	serviceKey, _ := cmd.Flags().GetString("service-key")
	costCentre, _ := cmd.Flags().GetString("cost-centre")

	// Create the base request
	req := &megaport.BuyVXCRequest{
		VXCName:    name,
		RateLimit:  rateLimit,
		Term:       term,
		PromoCode:  promoCode,
		ServiceKey: serviceKey,
		CostCentre: costCentre,
	}

	// A-End configuration
	aEndConfig := megaport.VXCOrderEndpointConfiguration{
		VLAN: aEndVLAN,
	}

	// Set MVE config if needed
	if aEndInnerVLAN != 0 || aEndVNICIndex > 0 {
		aEndConfig.VXCOrderMVEConfig = &megaport.VXCOrderMVEConfig{
			InnerVLAN:             aEndInnerVLAN,
			NetworkInterfaceIndex: aEndVNICIndex,
		}
	}

	// Parse A-End partner config if provided
	aEndPartnerConfigStr, _ := cmd.Flags().GetString("a-end-partner-config")
	if aEndPartnerConfigStr != "" {
		aEndPartnerConfig, err := parsePartnerConfigFromJSON(aEndPartnerConfigStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing a-end-partner-config: %v", err)
		}
		// If the A End UID is not provided, attempt to look it up from the partner port key
		if aEndUID == "" {
			switch aEndPartnerConfig := aEndPartnerConfig.(type) {
			case *megaport.VXCPartnerConfigAzure:
				if aEndPartnerConfig.ServiceKey == "" {
					return nil, fmt.Errorf("serviceKey is required for Azure configuration")
				}
				uid, err := getPartnerPortUID(ctx, svc, aEndPartnerConfig.ServiceKey, "AZURE")
				if err != nil {
					return nil, fmt.Errorf("error looking up Azure Partner Port: %v", err)
				}
				aEndUID = uid
			case *megaport.VXCPartnerConfigGoogle:
				if aEndPartnerConfig.PairingKey == "" {
					return nil, fmt.Errorf("pairingKey is required for Google configuration")
				}
				uid, err := getPartnerPortUID(ctx, svc, aEndPartnerConfig.PairingKey, "GOOGLE")
				if err != nil {
					return nil, fmt.Errorf("error looking up Google Partner Port: %v", err)
				}
				aEndUID = uid
			case *megaport.VXCPartnerConfigOracle:
				if aEndPartnerConfig.VirtualCircuitId == "" {
					return nil, fmt.Errorf("virtualCircuitId is required for Oracle configuration")
				}
				uid, err := getPartnerPortUID(ctx, svc, aEndPartnerConfig.VirtualCircuitId, "ORACLE")
				if err != nil {
					return nil, fmt.Errorf("error looking up Oracle Partner Port: %v", err)
				}
				aEndUID = uid
				aEndConfig.PartnerConfig = aEndPartnerConfig
			}
		}
	}

	req.AEndConfiguration = aEndConfig

	if aEndUID == "" {
		return nil, fmt.Errorf("a-end-uid was neither specified nor could be looked up")
	}

	req.PortUID = aEndUID

	// B-End configuration
	bEndConfig := megaport.VXCOrderEndpointConfiguration{}

	// Parse B-End partner config if provided
	bEndPartnerConfigStr, _ := cmd.Flags().GetString("b-end-partner-config")
	if bEndPartnerConfigStr != "" {
		bEndPartnerConfig, err := parsePartnerConfigFromJSON(bEndPartnerConfigStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing b-end-partner-config: %v", err)
		}
		bEndConfig.PartnerConfig = bEndPartnerConfig
	}

	bEndUID, _ := cmd.Flags().GetString("b-end-uid")

	// Attempt to look up partner port UID if not provided
	if bEndUID == "" {
		switch bEndPartnerConfig := bEndConfig.PartnerConfig.(type) {
		case *megaport.VXCPartnerConfigAzure:
			if bEndPartnerConfig.ServiceKey == "" {
				return nil, fmt.Errorf("serviceKey is required for Azure configuration")
			}
			uid, err := getPartnerPortUID(ctx, svc, bEndPartnerConfig.ServiceKey, "AZURE")
			if err != nil {
				return nil, fmt.Errorf("error looking up Azure Partner Port: %v", err)
			}
			bEndUID = uid
		case *megaport.VXCPartnerConfigGoogle:
			if bEndPartnerConfig.PairingKey == "" {
				return nil, fmt.Errorf("pairingKey is required for Google configuration")
			}
			uid, err := getPartnerPortUID(ctx, svc, bEndPartnerConfig.PairingKey, "GOOGLE")
			if err != nil {
				return nil, fmt.Errorf("error looking up Google Partner Port: %v", err)
			}
			bEndUID = uid
		case *megaport.VXCPartnerConfigOracle:
			if bEndPartnerConfig.VirtualCircuitId == "" {
				return nil, fmt.Errorf("virtualCircuitId is required for Oracle configuration")
			}
			uid, err := getPartnerPortUID(ctx, svc, bEndPartnerConfig.VirtualCircuitId, "ORACLE")
			if err != nil {
				return nil, fmt.Errorf("error looking up Oracle Partner Port: %v", err)
			}
			bEndUID = uid
		}
	}

	if bEndUID == "" {
		return nil, fmt.Errorf("b-end-uid was neither provided nor could be looked up")
	}

	bEndConfig.ProductUID = bEndUID
	bEndConfig.VLAN = bEndVLAN

	// Set MVE config if needed
	if bEndInnerVLAN != 0 || bEndVNICIndex > 0 {
		bEndConfig.VXCOrderMVEConfig = &megaport.VXCOrderMVEConfig{
			InnerVLAN:             bEndInnerVLAN,
			NetworkInterfaceIndex: bEndVNICIndex,
		}
	}

	req.BEndConfiguration = bEndConfig

	return req, nil
}

// parsePartnerConfigFromJSON parses a JSON string into a VXCPartnerConfiguration
func parsePartnerConfigFromJSON(jsonStr string) (megaport.VXCPartnerConfiguration, error) {
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawConfig); err != nil {
		return nil, err
	}

	connectType, ok := rawConfig["connectType"].(string)
	if !ok {
		return nil, fmt.Errorf("connectType is required and must be a string")
	}

	switch strings.ToUpper(connectType) {
	case "AWS", "AWSHC":
		return parseAWSConfig(rawConfig)
	case "AZURE":
		return parseAzureConfig(rawConfig)
	case "GOOGLE":
		return parseGoogleConfig(rawConfig)
	case "ORACLE":
		return parseOracleConfig(rawConfig)
	case "IBM":
		return parseIBMConfig(rawConfig)
	case "TRANSIT":
		return &megaport.VXCPartnerConfigTransit{
			ConnectType: "TRANSIT",
		}, nil
	case "VROUTER":
		return parseVRouterConfig(rawConfig)
	default:
		return nil, fmt.Errorf("unsupported connect type: %s", connectType)
	}
}

// Parse AWS specific configuration
func parseAWSConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigAWS, error) {
	ownerAccount, _ := config["ownerAccount"].(string)
	if ownerAccount == "" {
		return nil, fmt.Errorf("ownerAccount is required for AWS configuration")
	}

	connectType, ok := config["connectType"].(string)
	if !ok {
		return nil, fmt.Errorf("connectType is required for AWS configuration")
	}

	awsConfig := &megaport.VXCPartnerConfigAWS{
		ConnectType:  connectType,
		OwnerAccount: ownerAccount,
	}

	// Handle optional fields
	if asn, ok := config["asn"].(float64); ok {
		awsConfig.ASN = int(asn)
	}

	if amazonAsn, ok := config["amazonAsn"].(float64); ok {
		awsConfig.AmazonASN = int(amazonAsn)
	}

	if authKey, ok := config["authKey"].(string); ok {
		awsConfig.AuthKey = authKey
	}

	if prefixes, ok := config["prefixes"].(string); ok {
		awsConfig.Prefixes = prefixes
	}

	if customerIP, ok := config["customerIPAddress"].(string); ok {
		awsConfig.CustomerIPAddress = customerIP
	}

	if amazonIP, ok := config["amazonIPAddress"].(string); ok {
		awsConfig.AmazonIPAddress = amazonIP
	}

	if connName, ok := config["connectionName"].(string); ok {
		awsConfig.ConnectionName = connName
	}

	if vpcType, ok := config["type"].(string); ok {
		awsConfig.Type = vpcType
	}

	return awsConfig, nil
}

// Parse Azure specific configuration
func parseAzureConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigAzure, error) {
	serviceKey, _ := config["serviceKey"].(string)
	if serviceKey == "" {
		return nil, fmt.Errorf("serviceKey is required for Azure configuration")
	}

	azureConfig := &megaport.VXCPartnerConfigAzure{
		ConnectType: "AZURE",
		ServiceKey:  serviceKey,
	}

	// Parse peers if available
	if peersRaw, ok := config["peers"].([]interface{}); ok {
		for _, peerRaw := range peersRaw {
			if peerMap, ok := peerRaw.(map[string]interface{}); ok {
				peer := megaport.PartnerOrderAzurePeeringConfig{}

				if pType, ok := peerMap["type"].(string); ok {
					peer.Type = pType
				}

				if peerASN, ok := peerMap["peerASN"].(string); ok {
					peer.PeerASN = peerASN
				}

				if primarySubnet, ok := peerMap["primarySubnet"].(string); ok {
					peer.PrimarySubnet = primarySubnet
				}

				if secondarySubnet, ok := peerMap["secondarySubnet"].(string); ok {
					peer.SecondarySubnet = secondarySubnet
				}

				if prefixes, ok := peerMap["prefixes"].(string); ok {
					peer.Prefixes = prefixes
				}

				if sharedKey, ok := peerMap["sharedKey"].(string); ok {
					peer.SharedKey = sharedKey
				}

				if vlan, ok := peerMap["vlan"].(float64); ok {
					peer.VLAN = int(vlan)
				}

				azureConfig.Peers = append(azureConfig.Peers, peer)
			}
		}
	}

	return azureConfig, nil
}

// Parse Google specific configuration
func parseGoogleConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigGoogle, error) {
	pairingKey, _ := config["pairingKey"].(string)
	if pairingKey == "" {
		return nil, fmt.Errorf("pairingKey is required for Google configuration")
	}

	return &megaport.VXCPartnerConfigGoogle{
		ConnectType: "GOOGLE",
		PairingKey:  pairingKey,
	}, nil
}

// Parse Oracle specific configuration
func parseOracleConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigOracle, error) {
	vcID, _ := config["virtualCircuitId"].(string)
	if vcID == "" {
		return nil, fmt.Errorf("virtualCircuitId is required for Oracle configuration")
	}

	return &megaport.VXCPartnerConfigOracle{
		ConnectType:      "ORACLE",
		VirtualCircuitId: vcID,
	}, nil
}

// Parse IBM specific configuration
func parseIBMConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigIBM, error) {
	accountID, _ := config["accountID"].(string)
	if accountID == "" {
		return nil, fmt.Errorf("accountID is required for IBM configuration")
	}

	ibmConfig := &megaport.VXCPartnerConfigIBM{
		ConnectType: "IBM",
		AccountID:   accountID,
	}

	// Handle optional fields
	if customerASN, ok := config["customerASN"].(float64); ok {
		ibmConfig.CustomerASN = int(customerASN)
	}

	if customerIP, ok := config["customerIPAddress"].(string); ok {
		ibmConfig.CustomerIPAddress = customerIP
	}

	if providerIP, ok := config["providerIPAddress"].(string); ok {
		ibmConfig.ProviderIPAddress = providerIP
	}

	if name, ok := config["name"].(string); ok {
		ibmConfig.Name = name
	}

	return ibmConfig, nil
}

// Parse VRouter specific configuration
func parseVRouterConfig(config map[string]interface{}) (*megaport.VXCOrderVrouterPartnerConfig, error) {
	// Extract interfaces
	var interfaces []megaport.PartnerConfigInterface

	if interfacesRaw, ok := config["interfaces"].([]interface{}); ok {
		for _, ifaceRaw := range interfacesRaw {
			if ifaceMap, ok := ifaceRaw.(map[string]interface{}); ok {
				iface := megaport.PartnerConfigInterface{}

				// Parse VLAN
				if vlan, ok := ifaceMap["vlan"].(float64); ok {
					iface.VLAN = int(vlan)
				} else {
					return nil, fmt.Errorf("vlan is required for vRouter interface")
				}

				// Parse IP addresses
				if ipAddressesRaw, ok := ifaceMap["ipAddresses"].([]interface{}); ok {
					for _, ipRaw := range ipAddressesRaw {
						if ip, ok := ipRaw.(string); ok {
							iface.IpAddresses = append(iface.IpAddresses, ip)
						}
					}
				}

				// Parse IP routes
				if ipRoutesRaw, ok := ifaceMap["ipRoutes"].([]interface{}); ok {
					for _, routeRaw := range ipRoutesRaw {
						if routeMap, ok := routeRaw.(map[string]interface{}); ok {
							route := megaport.IpRoute{}

							if prefix, ok := routeMap["prefix"].(string); ok {
								route.Prefix = prefix
							}

							if description, ok := routeMap["description"].(string); ok {
								route.Description = description
							}

							if nextHop, ok := routeMap["nextHop"].(string); ok {
								route.NextHop = nextHop
							}

							iface.IpRoutes = append(iface.IpRoutes, route)
						}
					}
				}

				// Parse NAT IP addresses
				if natIPsRaw, ok := ifaceMap["natIpAddresses"].([]interface{}); ok {
					for _, ipRaw := range natIPsRaw {
						if ip, ok := ipRaw.(string); ok {
							iface.NatIpAddresses = append(iface.NatIpAddresses, ip)
						}
					}
				}

				// Parse BFD config
				if bfdRaw, ok := ifaceMap["bfd"].(map[string]interface{}); ok {
					bfd := megaport.BfdConfig{}

					if txInterval, ok := bfdRaw["txInterval"].(float64); ok {
						bfd.TxInterval = int(txInterval)
					}

					if rxInterval, ok := bfdRaw["rxInterval"].(float64); ok {
						bfd.RxInterval = int(rxInterval)
					}

					if multiplier, ok := bfdRaw["multiplier"].(float64); ok {
						bfd.Multiplier = int(multiplier)
					}

					iface.Bfd = bfd
				}

				// Parse BGP connections
				if bgpConnsRaw, ok := ifaceMap["bgpConnections"].([]interface{}); ok {
					for _, bgpRaw := range bgpConnsRaw {
						if bgpMap, ok := bgpRaw.(map[string]interface{}); ok {
							bgp := megaport.BgpConnectionConfig{}

							if peerAsn, ok := bgpMap["peerAsn"].(float64); ok {
								bgp.PeerAsn = int(peerAsn)
							}

							if localAsn, ok := bgpMap["localAsn"].(float64); ok {
								localAsnVal := int(localAsn)
								bgp.LocalAsn = &localAsnVal
							}

							if localIP, ok := bgpMap["localIpAddress"].(string); ok {
								bgp.LocalIpAddress = localIP
							}

							if peerIP, ok := bgpMap["peerIpAddress"].(string); ok {
								bgp.PeerIpAddress = peerIP
							}

							if password, ok := bgpMap["password"].(string); ok {
								bgp.Password = password
							}

							if shutdown, ok := bgpMap["shutdown"].(bool); ok {
								bgp.Shutdown = shutdown
							}

							if description, ok := bgpMap["description"].(string); ok {
								bgp.Description = description
							}

							if medIn, ok := bgpMap["medIn"].(float64); ok {
								bgp.MedIn = int(medIn)
							}

							if medOut, ok := bgpMap["medOut"].(float64); ok {
								bgp.MedOut = int(medOut)
							}

							if bfdEnabled, ok := bgpMap["bfdEnabled"].(bool); ok {
								bgp.BfdEnabled = bfdEnabled
							}

							if exportPolicy, ok := bgpMap["exportPolicy"].(string); ok {
								bgp.ExportPolicy = exportPolicy
							}

							if permitExportToRaw, ok := bgpMap["permitExportTo"].([]interface{}); ok {
								for _, permitRaw := range permitExportToRaw {
									if permit, ok := permitRaw.(string); ok {
										bgp.PermitExportTo = append(bgp.PermitExportTo, permit)
									}
								}
							}

							if denyExportToRaw, ok := bgpMap["denyExportTo"].([]interface{}); ok {
								for _, denyRaw := range denyExportToRaw {
									if deny, ok := denyRaw.(string); ok {
										bgp.DenyExportTo = append(bgp.DenyExportTo, deny)
									}
								}
							}

							if importWhitelist, ok := bgpMap["importWhitelist"].(float64); ok {
								bgp.ImportWhitelist = int(importWhitelist)
							}

							if importBlacklist, ok := bgpMap["importBlacklist"].(float64); ok {
								bgp.ImportBlacklist = int(importBlacklist)
							}

							if exportWhitelist, ok := bgpMap["exportWhitelist"].(float64); ok {
								bgp.ExportWhitelist = int(exportWhitelist)
							}

							if exportBlacklist, ok := bgpMap["exportBlacklist"].(float64); ok {
								bgp.ExportBlacklist = int(exportBlacklist)
							}

							if asPathPrependCount, ok := bgpMap["asPathPrependCount"].(float64); ok {
								bgp.AsPathPrependCount = int(asPathPrependCount)
							}

							if peerType, ok := bgpMap["peerType"].(string); ok {
								bgp.PeerType = peerType
							}

							iface.BgpConnections = append(iface.BgpConnections, bgp)
						}
					}
				}

				interfaces = append(interfaces, iface)
			}
		}
	}

	return &megaport.VXCOrderVrouterPartnerConfig{
		Interfaces: interfaces,
	}, nil
}

// buildVXCRequestFromJSON creates a BuyVXCRequest from a JSON string or file
func buildVXCRequestFromJSON(jsonStr string, jsonFilePath string) (*megaport.BuyVXCRequest, error) {
	var jsonData string

	if jsonStr != "" {
		jsonData = jsonStr
	} else if jsonFilePath != "" {
		// Read JSON from file
		data, err := os.ReadFile(jsonFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
		jsonData = string(data)
	} else {
		return nil, fmt.Errorf("either json or json-file must be provided")
	}

	// Unmarshal JSON into request
	var req megaport.BuyVXCRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &req, nil
}

// buildUpdateVXCRequestFromFlags creates an UpdateVXCRequest from command flags
var buildUpdateVXCRequestFromFlags = func(cmd *cobra.Command) (*megaport.UpdateVXCRequest, error) {
	req := &megaport.UpdateVXCRequest{}

	// Handle simple string and int fields
	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		req.Name = &name
	}

	if cmd.Flags().Changed("rate-limit") {
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")
		if rateLimit < 0 {
			return nil, fmt.Errorf("rate-limit must be greater than or equal to 0")
		}
		req.RateLimit = &rateLimit
	}

	if cmd.Flags().Changed("term") {
		term, _ := cmd.Flags().GetInt("term")
		if term != 0 && term != 1 && term != 12 && term != 24 && term != 36 {
			return nil, fmt.Errorf("term must be 0, 1, 12, 24, or 36")
		}
		req.Term = &term
	}

	if cmd.Flags().Changed("cost-centre") {
		costCentre, _ := cmd.Flags().GetString("cost-centre")
		req.CostCentre = &costCentre
	}

	if cmd.Flags().Changed("shutdown") {
		shutdown, _ := cmd.Flags().GetBool("shutdown")
		req.Shutdown = &shutdown
	}

	// Handle VLAN fields
	if cmd.Flags().Changed("a-end-vlan") {
		aEndVLAN, _ := cmd.Flags().GetInt("a-end-vlan")
		if aEndVLAN < 0 || aEndVLAN > 4093 || aEndVLAN == 1 {
			return nil, fmt.Errorf("a-end-vlan must be 0 or between 2-4093")
		}
		req.AEndVLAN = &aEndVLAN
	}

	if cmd.Flags().Changed("b-end-vlan") {
		bEndVLAN, _ := cmd.Flags().GetInt("b-end-vlan")
		if bEndVLAN < 0 || bEndVLAN > 4093 || bEndVLAN == 1 {
			return nil, fmt.Errorf("b-end-vlan must be 0 or between 2-4093")
		}
		req.BEndVLAN = &bEndVLAN
	}

	if cmd.Flags().Changed("a-end-inner-vlan") {
		aEndInnerVLAN, _ := cmd.Flags().GetInt("a-end-inner-vlan")
		if aEndInnerVLAN != -1 && aEndInnerVLAN != 0 && aEndInnerVLAN < 2 {
			return nil, fmt.Errorf("a-end-inner-vlan must be -1, 0, or greater than 1")
		}
		req.AEndInnerVLAN = &aEndInnerVLAN
	}

	if cmd.Flags().Changed("b-end-inner-vlan") {
		bEndInnerVLAN, _ := cmd.Flags().GetInt("b-end-inner-vlan")
		if bEndInnerVLAN != -1 && bEndInnerVLAN != 0 && bEndInnerVLAN < 2 {
			return nil, fmt.Errorf("b-end-inner-vlan must be -1, 0, or greater than 1")
		}
		req.BEndInnerVLAN = &bEndInnerVLAN
	}

	// Handle product UIDs
	if cmd.Flags().Changed("a-end-uid") {
		aEndUID, _ := cmd.Flags().GetString("a-end-uid")
		req.AEndProductUID = &aEndUID
	}

	if cmd.Flags().Changed("b-end-uid") {
		bEndUID, _ := cmd.Flags().GetString("b-end-uid")
		req.BEndProductUID = &bEndUID
	}

	// Handle partner configurations
	if cmd.Flags().Changed("a-end-partner-config") {
		aEndPartnerConfigStr, _ := cmd.Flags().GetString("a-end-partner-config")
		if aEndPartnerConfigStr != "" {
			aEndPartnerConfig, err := parsePartnerConfigFromJSON(aEndPartnerConfigStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing a-end-partner-config: %v", err)
			}

			// Verify it's a VRouter config which is the only updatable partner config
			if _, ok := aEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig); !ok {
				return nil, fmt.Errorf("only VRouter partner configurations can be updated")
			}
			req.AEndPartnerConfig = aEndPartnerConfig
		}
	}

	if cmd.Flags().Changed("b-end-partner-config") {
		bEndPartnerConfigStr, _ := cmd.Flags().GetString("b-end-partner-config")
		if bEndPartnerConfigStr != "" {
			bEndPartnerConfig, err := parsePartnerConfigFromJSON(bEndPartnerConfigStr)
			if err != nil {
				return nil, fmt.Errorf("error parsing b-end-partner-config: %v", err)
			}

			// Verify it's a VRouter config which is the only updatable partner config
			if _, ok := bEndPartnerConfig.(*megaport.VXCOrderVrouterPartnerConfig); !ok {
				return nil, fmt.Errorf("only VRouter partner configurations can be updated")
			}
			req.BEndPartnerConfig = bEndPartnerConfig
		}
	}

	return req, nil
}

// buildUpdateVXCRequestFromJSON creates an UpdateVXCRequest from a JSON string or file
var buildUpdateVXCRequestFromJSON = func(jsonStr string, jsonFilePath string) (*megaport.UpdateVXCRequest, error) {
	var jsonData string

	if jsonStr != "" {
		jsonData = jsonStr
	} else if jsonFilePath != "" {
		// Read JSON from file
		data, err := os.ReadFile(jsonFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading JSON file: %v", err)
		}
		jsonData = string(data)
	} else {
		return nil, fmt.Errorf("either json or json-file must be provided")
	}

	// Parse raw JSON first to handle partner configs
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &rawData); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	req := &megaport.UpdateVXCRequest{}

	if rateLimit, ok := rawData["rateLimit"].(float64); ok {
		rateLimitInt := int(rateLimit)
		if rateLimitInt < 0 {
			return nil, fmt.Errorf("rateLimit must be greater than or equal to 0")
		}
		req.RateLimit = &rateLimitInt
	}

	if term, ok := rawData["term"].(float64); ok {
		termInt := int(term)
		if termInt != 0 && termInt != 1 && termInt != 12 && termInt != 24 && termInt != 36 {
			return nil, fmt.Errorf("term must be 0, 1, 12, 24, or 36")
		}
		req.Term = &termInt
	}

	if costCentre, ok := rawData["costCentre"].(string); ok {
		req.CostCentre = &costCentre
	}

	if shutdown, ok := rawData["shutdown"].(bool); ok {
		req.Shutdown = &shutdown
	}

	// Handle nested configurations in addition to flat fields
	if aEndConfig, ok := rawData["aEndConfiguration"].(map[string]interface{}); ok {
		if vlan, ok := aEndConfig["vlan"].(float64); ok {
			vlanInt := int(vlan)
			if vlanInt != -1 && (vlanInt < 0 || vlanInt > 4093 || vlanInt == 1) {
				return nil, fmt.Errorf("aEndConfiguration.vlan must be -1, 0, or between 2-4093")
			}
			req.AEndVLAN = &vlanInt
		}
	} else {
		if aEndVLAN, ok := rawData["aEndVlan"].(float64); ok {
			aEndVLANInt := int(aEndVLAN)
			if aEndVLANInt != -1 && (aEndVLANInt < 0 || aEndVLANInt > 4093 || aEndVLANInt == 1) {
				return nil, fmt.Errorf("aEndVlan must be -1, 0, or between 2-4093")
			}
			req.AEndVLAN = &aEndVLANInt
		}
	}

	if bEndConfig, ok := rawData["bEndConfiguration"].(map[string]interface{}); ok {
		if vlan, ok := bEndConfig["vlan"].(float64); ok {
			vlanInt := int(vlan)
			if vlanInt != -1 && (vlanInt < 0 || vlanInt > 4093 || vlanInt == 1) {
				return nil, fmt.Errorf("bEndConfiguration.vlan must be -1, 0, or between 2-4093")
			}
			req.BEndVLAN = &vlanInt
		}
	} else {
		if bEndVLAN, ok := rawData["bEndVlan"].(float64); ok {
			bEndVLANInt := int(bEndVLAN)
			if bEndVLANInt != -1 && (bEndVLANInt < 0 || bEndVLANInt > 4093 || bEndVLANInt == 1) {
				return nil, fmt.Errorf("bEndVlan must be -1, 0, or between 2-4093")
			}
			req.BEndVLAN = &bEndVLANInt
		}
	}

	// Handle VXC name field variants
	if name, ok := rawData["name"].(string); ok {
		req.Name = &name
	} else if vxcName, ok := rawData["vxcName"].(string); ok {
		req.Name = &vxcName
	}

	if aEndInnerVLAN, ok := rawData["aEndInnerVlan"].(float64); ok {
		aEndInnerVLANInt := int(aEndInnerVLAN)
		if aEndInnerVLANInt != -1 && aEndInnerVLANInt != 0 && aEndInnerVLANInt < 2 {
			return nil, fmt.Errorf("aEndInnerVlan must be -1, 0, or greater than 1")
		}
		req.AEndInnerVLAN = &aEndInnerVLANInt
	}

	if bEndInnerVLAN, ok := rawData["bEndInnerVlan"].(float64); ok {
		bEndInnerVLANInt := int(bEndInnerVLAN)
		if bEndInnerVLANInt != -1 && bEndInnerVLANInt != 0 && bEndInnerVLANInt < 2 {
			return nil, fmt.Errorf("bEndInnerVlan must be -1, 0, or greater than 1")
		}
		req.BEndInnerVLAN = &bEndInnerVLANInt
	}

	// Handle product UIDs
	if aEndUID, ok := rawData["aEndUid"].(string); ok {
		req.AEndProductUID = &aEndUID
	}

	if bEndUID, ok := rawData["bEndUid"].(string); ok {
		req.BEndProductUID = &bEndUID
	}

	// Handle partner configurations
	if aEndPartnerConfigRaw, ok := rawData["aEndPartnerConfig"].(map[string]interface{}); ok {
		if connectType, ok := aEndPartnerConfigRaw["connectType"].(string); ok && strings.ToUpper(connectType) == "VROUTER" {
			aEndPartnerConfigBytes, err := json.Marshal(aEndPartnerConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("error marshaling A-End partner config: %v", err)
			}

			aEndPartnerConfig, err := parsePartnerConfigFromJSON(string(aEndPartnerConfigBytes))
			if err != nil {
				return nil, fmt.Errorf("error parsing A-End partner config: %v", err)
			}

			req.AEndPartnerConfig = aEndPartnerConfig
		} else {
			return nil, fmt.Errorf("only VRouter partner configurations can be updated")
		}
	}

	if bEndPartnerConfigRaw, ok := rawData["bEndPartnerConfig"].(map[string]interface{}); ok {
		if connectType, ok := bEndPartnerConfigRaw["connectType"].(string); ok && strings.ToUpper(connectType) == "VROUTER" {
			bEndPartnerConfigBytes, err := json.Marshal(bEndPartnerConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("error marshaling B-End partner config: %v", err)
			}

			bEndPartnerConfig, err := parsePartnerConfigFromJSON(string(bEndPartnerConfigBytes))
			if err != nil {
				return nil, fmt.Errorf("error parsing B-End partner config: %v", err)
			}

			req.BEndPartnerConfig = bEndPartnerConfig
		} else {
			return nil, fmt.Errorf("only VRouter partner configurations can be updated")
		}
	}

	// Set wait for update to true with a reasonable timeout
	req.WaitForUpdate = true
	req.WaitForTime = 10 * time.Minute

	return req, nil
}

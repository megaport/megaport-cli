package vxc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/base/config"
	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/utils"
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

var deleteVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.DeleteVXCRequest) error {
	err := client.VXCService.DeleteVXC(ctx, vxcUID, req)
	return err
}

// buildVXCRequestFromPrompt creates a BuyVXCRequest from interactive prompts
var buildVXCRequestFromPrompt = func(ctx context.Context, svc megaport.VXCService, noColor bool) (*megaport.BuyVXCRequest, error) {

	name, err := utils.Prompt("Enter VXC name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	rateLimitStr, err := utils.Prompt("Enter rate limit in Mbps (required): ", noColor)
	if err != nil {
		return nil, err
	}
	rateLimit, err := strconv.Atoi(rateLimitStr)
	if err != nil || rateLimit <= 0 {
		return nil, fmt.Errorf("rate limit must be a positive integer")
	}

	termStr, err := utils.Prompt("Enter term in months (1, 12, 24, or 36, required): ", noColor)
	if err != nil {
		return nil, err
	}
	term, err := strconv.Atoi(termStr)
	if err != nil || (term != 1 && term != 12 && term != 24 && term != 36) {
		return nil, fmt.Errorf("term must be 1, 12, 24, or 36")
	}

	// A-End configuration
	aEndVLANStr, err := utils.Prompt("A-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
	if err != nil {
		return nil, err
	}
	var aEndVLAN int
	if aEndVLANStr != "" {
		aEndVLAN, err := strconv.Atoi(aEndVLANStr)
		if err != nil || (aEndVLAN != -1 && aEndVLAN != 0 && (aEndVLAN < 2 || aEndVLAN > 4093)) {
			return nil, fmt.Errorf("A-End VLAN must be -1 (untagged), 0 (auto-assigned), or between 2-4093")
		}
	}

	aEndInnerVLANStr, err := utils.Prompt("Enter A-End Inner VLAN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	aEndInnerVLAN := 0
	if aEndInnerVLANStr != "" {
		aEndInnerVLAN, err = strconv.Atoi(aEndInnerVLANStr)
		if err != nil {
			return nil, fmt.Errorf("invalid A-End Inner VLAN")
		}
	}

	aEndVNICIndexStr, err := utils.Prompt("Enter A-End vNIC Index (optional): ", noColor)
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
	hasAEndPartnerConfig, err := utils.Prompt("Do you want to configure A-End partner? (yes/no): ", noColor)
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
		aEndUID, err := utils.Prompt("Enter A-End product UID (required): ", noColor)
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

	bEndVLANStr, err := utils.Prompt("B-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
	if err != nil {
		return nil, err
	}
	var bEndVLAN int
	if bEndVLANStr != "" {
		bEndVLAN, err = strconv.Atoi(bEndVLANStr)
		if err != nil || (bEndVLAN != -1 && bEndVLAN != 0 && (bEndVLAN < 2 || bEndVLAN > 4093)) {
			return nil, fmt.Errorf("B-End VLAN must be -1 (untagged), 0 (auto-assigned), or between 2-4093")
		}
		req.BEndConfiguration.VLAN = bEndVLAN
	}

	bEndInnerVLANStr, err := utils.Prompt("Enter B-End Inner VLAN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	bEndInnerVLAN := 0
	if bEndInnerVLANStr != "" {
		bEndInnerVLAN, err = strconv.Atoi(bEndInnerVLANStr)
		if err != nil {
			return nil, fmt.Errorf("invalid B-End Inner VLAN")
		}
	}
	bEndVNICIndexStr, err := utils.Prompt("Enter B-End vNIC Index (optional): ", noColor)
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

	hasBEndPartnerConfig, err := utils.Prompt("Do you want to configure B-End partner? (yes/no): ", noColor)
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
		bEndUID, err := utils.Prompt("Enter B-End product UID (required): ", noColor)
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
	promoCode, err := utils.Prompt("Enter promo code (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.PromoCode = promoCode

	serviceKey, err := utils.Prompt("Enter service key (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.ServiceKey = serviceKey

	costCentre, err := utils.Prompt("Enter cost centre (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	req.CostCentre = costCentre

	return req, nil
}

var buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	return client.VXCService.BuyVXC(ctx, req)
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

	// Handle simple fields
	if name, ok := rawData["name"].(string); ok {
		req.Name = &name
	}

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

	// Handle VLAN fields
	if aEndVLAN, ok := rawData["aEndVlan"].(float64); ok {
		aEndVLANInt := int(aEndVLAN)
		if aEndVLANInt < 0 || aEndVLANInt > 4093 || aEndVLANInt == 1 {
			return nil, fmt.Errorf("aEndVlan must be 0 or between 2-4093")
		}
		req.AEndVLAN = &aEndVLANInt
	}

	if bEndVLAN, ok := rawData["bEndVlan"].(float64); ok {
		bEndVLANInt := int(bEndVLAN)
		if bEndVLANInt < 0 || bEndVLANInt > 4093 || bEndVLANInt == 1 {
			return nil, fmt.Errorf("bEndVlan must be 0 or between 2-4093")
		}
		req.BEndVLAN = &bEndVLANInt
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
	req.WaitForTime = 5 * time.Minute

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
	updateName, err := utils.Prompt("Update name? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateName) == "yes" {
		name, err := utils.Prompt("Enter new name: ", noColor)
		if err != nil {
			return nil, err
		}
		req.Name = &name
	}

	// Rate limit
	fmt.Printf("Current rate limit: %d Mbps\n", vxc.RateLimit)
	updateRateLimit, err := utils.Prompt("Update rate limit? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateRateLimit) == "yes" {
		rateLimitStr, err := utils.Prompt("Enter new rate limit in Mbps: ", noColor)
		if err != nil {
			return nil, err
		}
		rateLimit, err := strconv.Atoi(rateLimitStr)
		if err != nil || rateLimit < 0 {
			return nil, fmt.Errorf("rate limit must be a non-negative integer")
		}
		req.RateLimit = &rateLimit
	}

	// Term
	fmt.Printf("Current term: %d months\n", vxc.ContractTermMonths)
	updateTerm, err := utils.Prompt("Update term? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateTerm) == "yes" {
		termStr, err := utils.Prompt("Enter new term in months (0, 1, 12, 24, or 36): ", noColor)
		if err != nil {
			return nil, err
		}
		term, err := strconv.Atoi(termStr)
		if err != nil || (term != 0 && term != 1 && term != 12 && term != 24 && term != 36) {
			return nil, fmt.Errorf("term must be 0, 1, 12, 24, or 36")
		}
		req.Term = &term
	}

	// Cost centre
	fmt.Printf("Current cost centre: %s\n", vxc.CostCentre)
	updateCostCentre, err := utils.Prompt("Update cost centre? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateCostCentre) == "yes" {
		costCentre, err := utils.Prompt("Enter new cost centre: ", noColor)
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
	updateShutdown, err := utils.Prompt("Update shutdown status? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateShutdown) == "yes" {
		shutdownStr, err := utils.Prompt("Shut down the VXC? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		shutdown := strings.ToLower(shutdownStr) == "yes"
		req.Shutdown = &shutdown
	}

	// A-End VLAN
	fmt.Printf("Current A-End VLAN: %d\n", vxc.AEndConfiguration.VLAN)
	updateAEndVLAN, err := utils.Prompt("Update A-End VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateAEndVLAN) == "yes" {
		aEndVLANStr, err := utils.Prompt("A-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
		if err != nil {
			return nil, err
		}
		aEndVLAN, err := strconv.Atoi(aEndVLANStr)
		if err != nil || (aEndVLAN != -1 && aEndVLAN != 0 && (aEndVLAN < 2 || aEndVLAN > 4093)) {
			return nil, fmt.Errorf("A-End VLAN must be -1 (untagged), 0 (auto-assigned), or between 2-4093")
		}
		req.AEndVLAN = &aEndVLAN
	}

	// B-End VLAN
	fmt.Printf("Current B-End VLAN: %d\n", vxc.BEndConfiguration.VLAN)
	updateBEndVLAN, err := utils.Prompt("Update B-End VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateBEndVLAN) == "yes" {
		bEndVLANStr, err := utils.Prompt("B-End VLAN (-1=untagged, 0=auto-assigned, 2-4093 for specific VLAN): ", noColor)
		if err != nil {
			return nil, err
		}
		bEndVLAN, err := strconv.Atoi(bEndVLANStr)
		if err != nil || (bEndVLAN != -1 && bEndVLAN != 0 && (bEndVLAN < 2 || bEndVLAN > 4093)) {
			return nil, fmt.Errorf("B-End VLAN must be -1 (untagged), 0 (auto-assigned), or between 2-4093")
		}
		req.BEndVLAN = &bEndVLAN
	}

	// A-End Inner VLAN
	innerVLANAEnd := 0
	if vxc.AEndConfiguration.InnerVLAN != 0 {
		innerVLANAEnd = vxc.AEndConfiguration.InnerVLAN
	}
	fmt.Printf("Current A-End Inner VLAN: %d\n", innerVLANAEnd)
	updateAEndInnerVLAN, err := utils.Prompt("Update A-End Inner VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateAEndInnerVLAN) == "yes" {
		aEndInnerVLANStr, err := utils.Prompt("Enter new A-End Inner VLAN (-1, 0, or >1): ", noColor)
		if err != nil {
			return nil, err
		}
		aEndInnerVLAN, err := strconv.Atoi(aEndInnerVLANStr)
		if err != nil || (aEndInnerVLAN != -1 && aEndInnerVLAN != 0 && aEndInnerVLAN < 2) {
			return nil, fmt.Errorf("A-End Inner VLAN must be -1, 0, or greater than 1")
		}
		req.AEndInnerVLAN = &aEndInnerVLAN
	}

	// B-End Inner VLAN
	innerVLANBEnd := 0
	if vxc.BEndConfiguration.InnerVLAN != 0 {
		innerVLANBEnd = vxc.BEndConfiguration.InnerVLAN
	}
	fmt.Printf("Current B-End Inner VLAN: %d\n", innerVLANBEnd)
	updateBEndInnerVLAN, err := utils.Prompt("Update B-End Inner VLAN? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateBEndInnerVLAN) == "yes" {
		bEndInnerVLANStr, err := utils.Prompt("Enter new B-End Inner VLAN (-1, 0, or >1): ", noColor)
		if err != nil {
			return nil, err
		}
		bEndInnerVLAN, err := strconv.Atoi(bEndInnerVLANStr)
		if err != nil || (bEndInnerVLAN != -1 && bEndInnerVLAN != 0 && bEndInnerVLAN < 2) {
			return nil, fmt.Errorf("B-End Inner VLAN must be -1, 0, or greater than 1")
		}
		req.BEndInnerVLAN = &bEndInnerVLAN
	}

	// A-End UID
	fmt.Printf("Current A-End UID: %s\n", vxc.AEndConfiguration.UID)
	updateAEndUID, err := utils.Prompt("Update A-End product UID? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateAEndUID) == "yes" {
		aEndUID, err := utils.Prompt("Enter new A-End product UID: ", noColor)
		if err != nil {
			return nil, err
		}
		req.AEndProductUID = &aEndUID
	}

	// B-End UID
	fmt.Printf("Current B-End UID: %s\n", vxc.BEndConfiguration.UID)
	updateBEndUID, err := utils.Prompt("Update B-End product UID? (yes/no): ", noColor)
	if err != nil {
		return nil, err
	}
	if strings.ToLower(updateBEndUID) == "yes" {
		bEndUID, err := utils.Prompt("Enter new B-End product UID: ", noColor)
		if err != nil {
			return nil, err
		}
		req.BEndProductUID = &bEndUID
	}

	wantsAEndPartnerConfig, err := utils.Prompt("Do you want to configure an A-End VRouter partner configuration? (yes/no): ", noColor)
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

	wantsBEndPartnerConfig, err := utils.Prompt("Do you want to configure a B-End VRouter partner configuration? (yes/no): ", noColor)
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
	interfaceCountStr, err := utils.Prompt("Number of interfaces to configure: ", noColor)
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
		vlanStr, err := utils.Prompt("VLAN (0-4093, except 1, optional - press Enter for no VLAN): ", noColor)
		if err != nil {
			return nil, err
		}

		if vlanStr != "" {
			vlan, err := strconv.Atoi(vlanStr)
			if err != nil || vlan < 0 || vlan > 4093 || vlan == 1 {
				return nil, fmt.Errorf("VLAN must be 0 or between 2-4093")
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
		hasRoutes, err := utils.Prompt("Do you want to add IP routes? (yes/no): ", noColor)
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
		hasNatIPs, err := utils.Prompt("Do you want to add NAT IP addresses? (yes/no): ", noColor)
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
		hasBFD, err := utils.Prompt("Do you want to configure BFD? (yes/no): ", noColor)
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
		hasBGP, err := utils.Prompt("Do you want to configure BGP connections? (yes/no): ", noColor)
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
		addRoute, err := utils.Prompt("Add an IP route? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addRoute) != "yes" {
			break
		}

		prefix, err := utils.Prompt("Enter prefix (e.g., 192.168.0.0/24): ", noColor)
		if err != nil {
			return nil, err
		}

		nextHop, err := utils.Prompt("Enter next hop IP: ", noColor)
		if err != nil {
			return nil, err
		}

		description, err := utils.Prompt("Enter description (optional): ", noColor)
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

// displayVXCChanges compares the original and updated VXC and displays the differences
func displayVXCChanges(original, updated *megaport.VXC, noColor bool) {
	if original == nil || updated == nil {
		return
	}

	fmt.Println() // Empty line before changes
	output.PrintInfo("Changes applied:", noColor)

	// Track if any changes were found
	changesFound := false

	// Compare name
	if original.Name != updated.Name {
		changesFound = true
		oldName := output.FormatOldValue(original.Name, noColor)
		newName := output.FormatNewValue(updated.Name, noColor)
		fmt.Printf("  • Name: %s → %s\n", oldName, newName)
	}

	// Compare rate limit
	if original.RateLimit != updated.RateLimit {
		changesFound = true
		oldRate := output.FormatOldValue(fmt.Sprintf("%d Mbps", original.RateLimit), noColor)
		newRate := output.FormatNewValue(fmt.Sprintf("%d Mbps", updated.RateLimit), noColor)
		fmt.Printf("  • Rate Limit: %s → %s\n", oldRate, newRate)
	}

	// Compare cost centre
	if original.CostCentre != updated.CostCentre {
		changesFound = true
		oldCostCentre := original.CostCentre
		if oldCostCentre == "" {
			oldCostCentre = "(none)"
		}
		newCostCentre := updated.CostCentre
		if newCostCentre == "" {
			newCostCentre = "(none)"
		}
		fmt.Printf("  • Cost Centre: %s → %s\n",
			output.FormatOldValue(oldCostCentre, noColor),
			output.FormatNewValue(newCostCentre, noColor))
	}

	// Compare contract term
	if original.ContractTermMonths != updated.ContractTermMonths {
		changesFound = true
		oldTerm := output.FormatOldValue(fmt.Sprintf("%d months", original.ContractTermMonths), noColor)
		newTerm := output.FormatNewValue(fmt.Sprintf("%d months", updated.ContractTermMonths), noColor)
		fmt.Printf("  • Contract Term: %s → %s\n", oldTerm, newTerm)
	}

	// Compare A-End VLAN - directly compare the VLAN values
	if original.AEndConfiguration.VLAN != updated.AEndConfiguration.VLAN {
		changesFound = true
		oldVlan := output.FormatOldValue(fmt.Sprintf("%d", original.AEndConfiguration.VLAN), noColor)
		newVlan := output.FormatNewValue(fmt.Sprintf("%d", updated.AEndConfiguration.VLAN), noColor)
		fmt.Printf("  • A-End VLAN: %s → %s\n", oldVlan, newVlan)
	}

	// Compare B-End VLAN - directly compare the VLAN values
	if original.BEndConfiguration.VLAN != updated.BEndConfiguration.VLAN {
		changesFound = true
		oldVlan := output.FormatOldValue(fmt.Sprintf("%d", original.BEndConfiguration.VLAN), noColor)
		newVlan := output.FormatNewValue(fmt.Sprintf("%d", updated.BEndConfiguration.VLAN), noColor)
		fmt.Printf("  • B-End VLAN: %s → %s\n", oldVlan, newVlan)
	}

	// Compare locked status
	if original.Locked != updated.Locked {
		changesFound = true
		oldLocked := "No"
		if original.Locked {
			oldLocked = "Yes"
		}
		newLocked := "No"
		if updated.Locked {
			newLocked = "Yes"
		}
		fmt.Printf("  • Locked: %s → %s\n",
			output.FormatOldValue(oldLocked, noColor),
			output.FormatNewValue(newLocked, noColor))
	}

	if !changesFound {
		fmt.Println("  No changes detected")
	}
}

// promptIPAddresses prompts the user for IP addresses
func promptIPAddresses(message string, noColor bool) ([]string, error) {
	var addresses []string

	for {
		addIP, err := utils.Prompt(fmt.Sprintf("Add %s? (yes/no): ", message), noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addIP) != "yes" {
			break
		}

		ip, err := utils.Prompt("Enter IP address: ", noColor)
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

	txIntervalStr, err := utils.Prompt("Enter transmit interval in ms (default 300): ", noColor)
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

	rxIntervalStr, err := utils.Prompt("Enter receive interval in ms (default 300): ", noColor)
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

	multiplierStr, err := utils.Prompt("Enter multiplier (default 3): ", noColor)
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
		addBGP, err := utils.Prompt("Add a BGP connection? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addBGP) != "yes" {
			break
		}

		bgp := megaport.BgpConnectionConfig{}

		// Required fields
		peerAsnStr, err := utils.Prompt("Enter peer ASN (required): ", noColor)
		if err != nil {
			return nil, err
		}
		peerAsn, err := strconv.Atoi(peerAsnStr)
		if err != nil || peerAsn <= 0 {
			return nil, fmt.Errorf("peer ASN must be a positive integer")
		}
		bgp.PeerAsn = peerAsn

		localIP, err := utils.Prompt("Enter local IP address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		if localIP == "" {
			return nil, fmt.Errorf("local IP address is required")
		}
		bgp.LocalIpAddress = localIP

		peerIP, err := utils.Prompt("Enter peer IP address (required): ", noColor)
		if err != nil {
			return nil, err
		}
		if peerIP == "" {
			return nil, fmt.Errorf("peer IP address is required")
		}
		bgp.PeerIpAddress = peerIP

		// Optional fields
		localAsnStr, err := utils.Prompt("Enter local ASN (optional): ", noColor)
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

		password, err := utils.Prompt("Enter password (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.Password = password

		shutdownStr, err := utils.Prompt("Shutdown connection? (yes/no, default: no): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.Shutdown = strings.ToLower(shutdownStr) == "yes"

		description, err := utils.Prompt("Enter description (optional): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.Description = description

		bfdEnabledStr, err := utils.Prompt("Enable BFD? (yes/no, default: no): ", noColor)
		if err != nil {
			return nil, err
		}
		bgp.BfdEnabled = strings.ToLower(bfdEnabledStr) == "yes"

		// Added: Export Policy
		exportPolicy, err := utils.Prompt("Enter export policy (permit/deny, optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if exportPolicy != "" && exportPolicy != "permit" && exportPolicy != "deny" {
			return nil, fmt.Errorf("export policy must be 'permit' or 'deny'")
		}
		bgp.ExportPolicy = exportPolicy

		// Added: Peer Type
		peerType, err := utils.Prompt("Enter peer type (NON_CLOUD/PRIV_CLOUD/PUB_CLOUD, optional): ", noColor)
		if err != nil {
			return nil, err
		}
		if peerType != "" && peerType != "NON_CLOUD" && peerType != "PRIV_CLOUD" && peerType != "PUB_CLOUD" {
			return nil, fmt.Errorf("peer type must be NON_CLOUD, PRIV_CLOUD, or PUB_CLOUD")
		}
		bgp.PeerType = peerType

		// Added: MED values
		medInStr, err := utils.Prompt("Enter MED in (optional): ", noColor)
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

		medOutStr, err := utils.Prompt("Enter MED out (optional): ", noColor)
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
		asPathPrependStr, err := utils.Prompt("Enter AS path prepend count (0-10, optional): ", noColor)
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
		hasPermitExportTo, err := utils.Prompt("Add permit export to addresses? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasPermitExportTo) == "yes" {
			for i := 0; i < 17; i++ { // Maximum 17 items
				ipAddress, err := utils.Prompt(fmt.Sprintf("Enter IP address to permit export to (or empty to finish) [%d/17]: ", i+1), noColor)
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
		hasDenyExportTo, err := utils.Prompt("Add deny export to addresses? (yes/no): ", noColor)
		if err != nil {
			return nil, err
		}
		if strings.ToLower(hasDenyExportTo) == "yes" {
			for i := 0; i < 17; i++ { // Maximum 17 items
				ipAddress, err := utils.Prompt(fmt.Sprintf("Enter IP address to deny export to (or empty to finish) [%d/17]: ", i+1), noColor)
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
		importWhitelistStr, err := utils.Prompt("Enter import whitelist prefix list ID (optional): ", noColor)
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

		importBlacklistStr, err := utils.Prompt("Enter import blacklist prefix list ID (optional): ", noColor)
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

		exportWhitelistStr, err := utils.Prompt("Enter export whitelist prefix list ID (optional): ", noColor)
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

		exportBlacklistStr, err := utils.Prompt("Enter export blacklist prefix list ID (optional): ", noColor)
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

// Function to handle VXC update API calls
var updateVXCFunc = func(ctx context.Context, client *megaport.Client, vxcUID string, req *megaport.UpdateVXCRequest) error {
	_, err := client.VXCService.UpdateVXC(ctx, vxcUID, req)
	return err
}

// VXCOutput represents the desired fields for JSON output.
type VXCOutput struct {
	output.Output `json:"-" header:"-"`
	UID           string `json:"uid"`
	Name          string `json:"name"`
	AEndUID       string `json:"a_end_uid"`
	BEndUID       string `json:"b_end_uid"`
}

// ToVXCOutput converts a VXC to a VXCOutput.
func ToVXCOutput(v *megaport.VXC) (VXCOutput, error) {
	if v == nil {
		return VXCOutput{}, fmt.Errorf("invalid VXC: nil value")
	}

	return VXCOutput{
		UID:     v.UID,
		Name:    v.Name,
		AEndUID: v.AEndConfiguration.UID,
		BEndUID: v.BEndConfiguration.UID,
	}, nil
}

// printVXCs prints the VXCs in the specified output format
func printVXCs(vxcs []*megaport.VXC, format string, noColor bool) error {
	if vxcs == nil {
		vxcs = []*megaport.VXC{}
	}

	outputs := make([]VXCOutput, 0, len(vxcs))
	for _, vxc := range vxcs {
		output, err := ToVXCOutput(vxc)
		if err != nil {
			return err
		}
		outputs = append(outputs, output)
	}
	return output.PrintOutput(outputs, format, noColor)
}

func promptPartnerConfig(end string, ctx context.Context, svc megaport.VXCService, noColor bool) (megaport.VXCPartnerConfiguration, string, error) {
	partner, err := utils.Prompt(fmt.Sprintf("Enter %s partner (AWS, Azure, Google, Oracle, IBM, VRouter, Transit) (optional): ", end), noColor)
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
		partnerPortUID, err := utils.Prompt("Enter AWS Partner Port product UID (required): ", noColor)
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
		partnerPortUID, err := utils.Prompt("Enter IBM Partner Port product UID (required): ", noColor)
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
	connectType, err := utils.Prompt("Enter connect type (required - either AWS or AWSHC): ", noColor)
	if err != nil {
		return nil, err
	}

	if connectType != "AWS" && connectType != "AWSHC" {
		return nil, fmt.Errorf("connect type must be AWS or AWSHC")
	}

	ownerAccount, err := utils.Prompt("Enter owner account ID (required): ", noColor)
	if err != nil {
		return nil, err
	}

	connectionName, err := utils.Prompt("Enter connection name (required): ", noColor)
	if err != nil {
		return nil, err
	}

	asnStr, err := utils.Prompt("Enter ASN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	asn, err := strconv.Atoi(asnStr)
	if err != nil {
		asn = 0
	}

	amazonASNStr, err := utils.Prompt("Enter Amazon ASN (optional): ", noColor)
	if err != nil {
		return nil, err
	}
	amazonASN, err := strconv.Atoi(amazonASNStr)
	if err != nil {
		amazonASN = 0
	}

	authKey, err := utils.Prompt("Enter auth key (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	prefixes, err := utils.Prompt("Enter prefixes (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	customerIPAddress, err := utils.Prompt("Enter customer IP address (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	amazonIPAddress, err := utils.Prompt("Enter Amazon IP address (optional): ", noColor)
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
		vifType, err := utils.Prompt("Enter VIF type (required - either private or public): ", noColor)
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
	serviceKey, err := utils.Prompt("Enter service key (required): ", noColor)
	if err != nil {
		return nil, "", err
	}

	portChoice, err := utils.Prompt("Enter port choice (primary/secondary, optional, default value is primary): ", noColor)
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
		addPeer, err := utils.Prompt("Add a peering config? (yes/no): ", noColor)
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
	peeringType, err := utils.Prompt("Enter peering type (required): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	peerASN, err := utils.Prompt("Enter Peer ASN (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	primarySubnet, err := utils.Prompt("Enter Primary Subnet (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	secondarySubnet, err := utils.Prompt("Enter Secondary Subnet (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	prefixes, err := utils.Prompt("Enter Prefixes (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	sharedKey, err := utils.Prompt("Enter Shared Key (optional): ", noColor)
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	vlanStr, err := utils.Prompt("Enter VLAN (optional): ", noColor)
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
	pairingKey, err := utils.Prompt("Enter pairing key (required): ", noColor)
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
	virtualCircuitId, err := utils.Prompt("Enter virtual circuit ID (required): ", noColor)
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
	accountID, err := utils.Prompt("Enter account ID (required): ", noColor)
	if err != nil {
		return nil, err
	}

	name, err := utils.Prompt("Enter name (required): ", noColor)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	var customerASN int

	customerASNStr, err := utils.Prompt("Enter customer ASN (required if opposite end is not an MCR): ", noColor)
	if err != nil {
		return nil, err
	}
	customerASN, err = strconv.Atoi(customerASNStr)
	if err != nil {
		return nil, err
	}

	customerIPAddress, err := utils.Prompt("Enter customer IP address (optional): ", noColor)
	if err != nil {
		return nil, err
	}

	providerIPAddress, err := utils.Prompt("Enter provider IP address (optional): ", noColor)
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

func getPartnerPortUID(ctx context.Context, svc megaport.VXCService, key, partner string) (string, error) {
	fmt.Println("Finding partner port...")

	partnerPortRes, err := svc.LookupPartnerPorts(ctx, &megaport.LookupPartnerPortsRequest{
		Key:     key,
		Partner: partner,
	})
	if err != nil {
		return "", fmt.Errorf("error looking up partner ports: %v", err)
	}
	return partnerPortRes.ProductUID, nil
}

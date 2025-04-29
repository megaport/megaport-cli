package vxc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

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

func parsePartnerConfigFromJSON(jsonStr string) (megaport.VXCPartnerConfiguration, error) {
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawConfig); err != nil {
		return nil, err
	}

	return parsePartnerConfigFromMap(rawConfig)
}

func parsePartnerConfigFromMap(rawConfig map[string]interface{}) (megaport.VXCPartnerConfiguration, error) {
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

func parseAWSConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigAWS, error) {
	ownerAccount, ok := config["ownerAccount"].(string)
	if !ok {
		return nil, fmt.Errorf("ownerAccount is required for AWS configuration and must be a string")
	}
	if ownerAccount == "" {
		return nil, fmt.Errorf("ownerAccount cannot be empty for AWS configuration")
	}

	connectType, ok := config["connectType"].(string)
	if !ok {
		return nil, fmt.Errorf("connectType is required for AWS configuration and must be a string")
	}
	if connectType == "" {
		return nil, fmt.Errorf("connectType cannot be empty for AWS configuration")
	}

	awsConfig := &megaport.VXCPartnerConfigAWS{
		ConnectType:  connectType,
		OwnerAccount: ownerAccount,
	}

	// Handle optional fields with improved error handling
	if asnVal, exists := config["asn"]; exists {
		asn, ok := asnVal.(float64)
		if !ok {
			return nil, fmt.Errorf("asn must be a number for AWS configuration")
		}
		if asn < 0 {
			return nil, fmt.Errorf("asn cannot be negative for AWS configuration")
		}
		awsConfig.ASN = int(asn)
	}

	if amazonAsnVal, exists := config["amazonAsn"]; exists {
		amazonAsn, ok := amazonAsnVal.(float64)
		if !ok {
			return nil, fmt.Errorf("amazonAsn must be a number for AWS configuration")
		}
		if amazonAsn < 0 {
			return nil, fmt.Errorf("amazonAsn cannot be negative for AWS configuration")
		}
		awsConfig.AmazonASN = int(amazonAsn)
	}

	if authKeyVal, exists := config["authKey"]; exists {
		authKey, ok := authKeyVal.(string)
		if !ok {
			return nil, fmt.Errorf("authKey must be a string for AWS configuration")
		}
		awsConfig.AuthKey = authKey
	}

	if prefixesVal, exists := config["prefixes"]; exists {
		prefixes, ok := prefixesVal.(string)
		if !ok {
			return nil, fmt.Errorf("prefixes must be a string for AWS configuration")
		}
		awsConfig.Prefixes = prefixes
	}

	if customerIPVal, exists := config["customerIPAddress"]; exists {
		customerIP, ok := customerIPVal.(string)
		if !ok {
			return nil, fmt.Errorf("customerIPAddress must be a string for AWS configuration")
		}
		awsConfig.CustomerIPAddress = customerIP
	}

	if amazonIPVal, exists := config["amazonIPAddress"]; exists {
		amazonIP, ok := amazonIPVal.(string)
		if !ok {
			return nil, fmt.Errorf("amazonIPAddress must be a string for AWS configuration")
		}
		awsConfig.AmazonIPAddress = amazonIP
	}

	if connNameVal, exists := config["connectionName"]; exists {
		connName, ok := connNameVal.(string)
		if !ok {
			return nil, fmt.Errorf("connectionName must be a string for AWS configuration")
		}
		awsConfig.ConnectionName = connName
	}

	if vpcTypeVal, exists := config["type"]; exists {
		vpcType, ok := vpcTypeVal.(string)
		if !ok {
			return nil, fmt.Errorf("type must be a string for AWS configuration")
		}
		awsConfig.Type = vpcType
	}

	return awsConfig, nil
}

func parseAzureConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigAzure, error) {
	serviceKeyVal, exists := config["serviceKey"]
	if !exists {
		return nil, fmt.Errorf("serviceKey is required for Azure configuration")
	}

	serviceKey, ok := serviceKeyVal.(string)
	if !ok {
		return nil, fmt.Errorf("serviceKey must be a string for Azure configuration")
	}

	if serviceKey == "" {
		return nil, fmt.Errorf("serviceKey cannot be empty for Azure configuration")
	}

	azureConfig := &megaport.VXCPartnerConfigAzure{
		ConnectType: "AZURE",
		ServiceKey:  serviceKey,
	}

	// Parse peers if available
	if peersRaw, exists := config["peers"]; exists {
		peersList, ok := peersRaw.([]interface{})
		if !ok {
			return nil, fmt.Errorf("peers must be an array for Azure configuration")
		}

		for i, peerRaw := range peersList {
			peerMap, ok := peerRaw.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("peer at index %d must be an object for Azure configuration", i)
			}

			peer := megaport.PartnerOrderAzurePeeringConfig{}

			if pTypeVal, exists := peerMap["type"]; exists {
				pType, ok := pTypeVal.(string)
				if !ok {
					return nil, fmt.Errorf("type must be a string in peer at index %d", i)
				}
				peer.Type = pType
			}

			if peerASNVal, exists := peerMap["peerASN"]; exists {
				peerASN, ok := peerASNVal.(string)
				if !ok {
					return nil, fmt.Errorf("peerASN must be a string in peer at index %d", i)
				}
				peer.PeerASN = peerASN
			}

			if primarySubnetVal, exists := peerMap["primarySubnet"]; exists {
				primarySubnet, ok := primarySubnetVal.(string)
				if !ok {
					return nil, fmt.Errorf("primarySubnet must be a string in peer at index %d", i)
				}
				peer.PrimarySubnet = primarySubnet
			}

			if secondarySubnetVal, exists := peerMap["secondarySubnet"]; exists {
				secondarySubnet, ok := secondarySubnetVal.(string)
				if !ok {
					return nil, fmt.Errorf("secondarySubnet must be a string in peer at index %d", i)
				}
				peer.SecondarySubnet = secondarySubnet
			}

			if prefixesVal, exists := peerMap["prefixes"]; exists {
				prefixes, ok := prefixesVal.(string)
				if !ok {
					return nil, fmt.Errorf("prefixes must be a string in peer at index %d", i)
				}
				peer.Prefixes = prefixes
			}

			if sharedKeyVal, exists := peerMap["sharedKey"]; exists {
				sharedKey, ok := sharedKeyVal.(string)
				if !ok {
					return nil, fmt.Errorf("sharedKey must be a string in peer at index %d", i)
				}
				peer.SharedKey = sharedKey
			}

			if vlanVal, exists := peerMap["vlan"]; exists {
				vlan, ok := vlanVal.(float64)
				if !ok {
					return nil, fmt.Errorf("vlan must be a number in peer at index %d", i)
				}
				peer.VLAN = int(vlan)
			}

			azureConfig.Peers = append(azureConfig.Peers, peer)
		}
	}

	return azureConfig, nil
}

func parseGoogleConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigGoogle, error) {
	pairingKeyVal, exists := config["pairingKey"]
	if !exists {
		return nil, fmt.Errorf("pairingKey is required for Google configuration")
	}

	pairingKey, ok := pairingKeyVal.(string)
	if !ok {
		return nil, fmt.Errorf("pairingKey must be a string for Google configuration")
	}

	if pairingKey == "" {
		return nil, fmt.Errorf("pairingKey cannot be empty for Google configuration")
	}

	return &megaport.VXCPartnerConfigGoogle{
		ConnectType: "GOOGLE",
		PairingKey:  pairingKey,
	}, nil
}

func parseOracleConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigOracle, error) {
	vcIDVal, exists := config["virtualCircuitId"]
	if !exists {
		return nil, fmt.Errorf("virtualCircuitId is required for Oracle configuration")
	}

	vcID, ok := vcIDVal.(string)
	if !ok {
		return nil, fmt.Errorf("virtualCircuitId must be a string for Oracle configuration")
	}

	if vcID == "" {
		return nil, fmt.Errorf("virtualCircuitId cannot be empty for Oracle configuration")
	}

	return &megaport.VXCPartnerConfigOracle{
		ConnectType:      "ORACLE",
		VirtualCircuitId: vcID,
	}, nil
}

func parseIBMConfig(config map[string]interface{}) (*megaport.VXCPartnerConfigIBM, error) {
	accountIDVal, exists := config["accountID"]
	if !exists {
		return nil, fmt.Errorf("accountID is required for IBM configuration")
	}

	accountID, ok := accountIDVal.(string)
	if !ok {
		return nil, fmt.Errorf("accountID must be a string for IBM configuration")
	}

	if accountID == "" {
		return nil, fmt.Errorf("accountID cannot be empty for IBM configuration")
	}

	ibmConfig := &megaport.VXCPartnerConfigIBM{
		ConnectType: "IBM",
		AccountID:   accountID,
	}

	// Handle optional fields with improved error handling
	if customerASNVal, exists := config["customerASN"]; exists {
		customerASN, ok := customerASNVal.(float64)
		if !ok {
			return nil, fmt.Errorf("customerASN must be a number for IBM configuration")
		}
		if customerASN < 0 {
			return nil, fmt.Errorf("customerASN cannot be negative for IBM configuration")
		}
		ibmConfig.CustomerASN = int(customerASN)
	}

	if customerIPVal, exists := config["customerIPAddress"]; exists {
		customerIP, ok := customerIPVal.(string)
		if !ok {
			return nil, fmt.Errorf("customerIPAddress must be a string for IBM configuration")
		}
		ibmConfig.CustomerIPAddress = customerIP
	}

	if providerIPVal, exists := config["providerIPAddress"]; exists {
		providerIP, ok := providerIPVal.(string)
		if !ok {
			return nil, fmt.Errorf("providerIPAddress must be a string for IBM configuration")
		}
		ibmConfig.ProviderIPAddress = providerIP
	}

	if nameVal, exists := config["name"]; exists {
		name, ok := nameVal.(string)
		if !ok {
			return nil, fmt.Errorf("name must be a string for IBM configuration")
		}
		ibmConfig.Name = name
	}

	return ibmConfig, nil
}

func parseVRouterConfig(config map[string]interface{}) (*megaport.VXCOrderVrouterPartnerConfig, error) {
	// Extract interfaces
	var interfaces []megaport.PartnerConfigInterface

	if interfacesRawVal, exists := config["interfaces"]; exists {
		interfacesRaw, ok := interfacesRawVal.([]interface{})
		if !ok {
			return nil, fmt.Errorf("interfaces must be an array for VRouter configuration")
		}

		for i, ifaceRaw := range interfacesRaw {
			ifaceMap, ok := ifaceRaw.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("interface at index %d must be an object for VRouter configuration", i)
			}

			iface := megaport.PartnerConfigInterface{}

			if vlanVal, exists := ifaceMap["vlan"]; exists {
				vlan, ok := vlanVal.(float64)
				if !ok {
					return nil, fmt.Errorf("vlan must be a number in interface at index %d", i)
				}
				iface.VLAN = int(vlan)
			}

			if ipAddressesRawVal, exists := ifaceMap["ipAddresses"]; exists {
				ipAddressesRaw, ok := ipAddressesRawVal.([]interface{})
				if !ok {
					return nil, fmt.Errorf("ipAddresses must be an array in interface at index %d", i)
				}

				for j, ipRaw := range ipAddressesRaw {
					ip, ok := ipRaw.(string)
					if !ok {
						return nil, fmt.Errorf("IP address at index %d in interface %d must be a string", j, i)
					}
					iface.IpAddresses = append(iface.IpAddresses, ip)
				}
			}

			// Parse IP routes with similar careful checking
			if ipRoutesRawVal, exists := ifaceMap["ipRoutes"]; exists {
				ipRoutesRaw, ok := ipRoutesRawVal.([]interface{})
				if !ok {
					return nil, fmt.Errorf("ipRoutes must be an array in interface at index %d", i)
				}

				for j, routeRaw := range ipRoutesRaw {
					routeMap, ok := routeRaw.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("route at index %d in interface %d must be an object", j, i)
					}

					route := megaport.IpRoute{}

					if prefixVal, exists := routeMap["prefix"]; exists {
						prefix, ok := prefixVal.(string)
						if !ok {
							return nil, fmt.Errorf("prefix must be a string in route %d of interface %d", j, i)
						}
						route.Prefix = prefix
					}

					if descriptionVal, exists := routeMap["description"]; exists {
						description, ok := descriptionVal.(string)
						if !ok {
							return nil, fmt.Errorf("description must be a string in route %d of interface %d", j, i)
						}
						route.Description = description
					}

					if nextHopVal, exists := routeMap["nextHop"]; exists {
						nextHop, ok := nextHopVal.(string)
						if !ok {
							return nil, fmt.Errorf("nextHop must be a string in route %d of interface %d", j, i)
						}
						route.NextHop = nextHop
					}

					iface.IpRoutes = append(iface.IpRoutes, route)
				}
			}

			// Parse NAT IP addresses with careful error handling
			if natIPsRawVal, exists := ifaceMap["natIpAddresses"]; exists {
				natIPsRaw, ok := natIPsRawVal.([]interface{})
				if !ok {
					return nil, fmt.Errorf("natIpAddresses must be an array in interface at index %d", i)
				}

				for j, ipRaw := range natIPsRaw {
					ip, ok := ipRaw.(string)
					if !ok {
						return nil, fmt.Errorf("NAT IP address at index %d in interface %d must be a string", j, i)
					}
					iface.NatIpAddresses = append(iface.NatIpAddresses, ip)
				}
			}

			// Parse BFD config with careful error handling
			if bfdRawVal, exists := ifaceMap["bfd"]; exists {
				bfdRaw, ok := bfdRawVal.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("bfd must be an object in interface at index %d", i)
				}

				bfd := megaport.BfdConfig{}

				if txIntervalVal, exists := bfdRaw["txInterval"]; exists {
					txInterval, ok := txIntervalVal.(float64)
					if !ok {
						return nil, fmt.Errorf("txInterval must be a number in bfd config of interface %d", i)
					}
					bfd.TxInterval = int(txInterval)
				}

				if rxIntervalVal, exists := bfdRaw["rxInterval"]; exists {
					rxInterval, ok := rxIntervalVal.(float64)
					if !ok {
						return nil, fmt.Errorf("rxInterval must be a number in bfd config of interface %d", i)
					}
					bfd.RxInterval = int(rxInterval)
				}

				if multiplierVal, exists := bfdRaw["multiplier"]; exists {
					multiplier, ok := multiplierVal.(float64)
					if !ok {
						return nil, fmt.Errorf("multiplier must be a number in bfd config of interface %d", i)
					}
					bfd.Multiplier = int(multiplier)
				}

				iface.Bfd = bfd
			}

			// ... existing code for BGP connections (similar pattern would be applied) ...

			interfaces = append(interfaces, iface)
		}
	}

	return &megaport.VXCOrderVrouterPartnerConfig{
		Interfaces: interfaces,
	}, nil
}

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

	// Parse raw JSON first to handle partner configs correctly
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &rawData); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	portUID, ok := rawData["portUid"].(string)
	if !ok {
		return nil, validation.NewValidationError("portUid", "", "Port UID is required")
	}

	// Create the base request
	req := &megaport.BuyVXCRequest{
		PortUID: portUID,
	}

	// Set simple fields
	if vxcName, ok := rawData["vxcName"].(string); ok {
		req.VXCName = vxcName
	}

	if rateLimit, ok := rawData["rateLimit"].(float64); ok {
		req.RateLimit = int(rateLimit)
	}

	if term, ok := rawData["term"].(float64); ok {
		req.Term = int(term)
	}

	if shutdown, ok := rawData["shutdown"].(bool); ok {
		req.Shutdown = shutdown
	}

	if promoCode, ok := rawData["promoCode"].(string); ok {
		req.PromoCode = promoCode
	}

	if serviceKey, ok := rawData["serviceKey"].(string); ok {
		req.ServiceKey = serviceKey
	}

	if costCentre, ok := rawData["costCentre"].(string); ok {
		req.CostCentre = costCentre
	}

	// Handle resource tags if they exist
	if resourceTags, ok := rawData["resourceTags"].(map[string]interface{}); ok {
		req.ResourceTags = make(map[string]string)
		for k, v := range resourceTags {
			if strValue, ok := v.(string); ok {
				req.ResourceTags[k] = strValue
			}
		}
	}

	// Handle A-End configuration
	if aEndConfigRaw, ok := rawData["aEndConfiguration"].(map[string]interface{}); ok {
		aEndConfig := megaport.VXCOrderEndpointConfiguration{}

		if vlan, ok := aEndConfigRaw["vlan"].(float64); ok {
			aEndConfig.VLAN = int(vlan)
		}

		if diversityZone, ok := aEndConfigRaw["diversityZone"].(string); ok {
			aEndConfig.DiversityZone = diversityZone
		}

		// Handle A-End partner config - directly use map data
		if partnerConfigRaw, ok := aEndConfigRaw["partnerConfig"].(map[string]interface{}); ok {
			partnerConfig, err := parsePartnerConfigFromMap(partnerConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("error parsing A-End partner config: %v", err)
			}

			aEndConfig.PartnerConfig = partnerConfig
		}

		// Handle A-End MVE config
		innerVLAN, hasInnerVLAN := aEndConfigRaw["innerVlan"].(float64)
		vNicIndex, hasVNicIndex := aEndConfigRaw["vNicIndex"].(float64)

		if hasInnerVLAN || hasVNicIndex {
			mveConfig := &megaport.VXCOrderMVEConfig{}

			if hasInnerVLAN {
				mveConfig.InnerVLAN = int(innerVLAN)
			}

			if hasVNicIndex {
				mveConfig.NetworkInterfaceIndex = int(vNicIndex)
			}

			aEndConfig.VXCOrderMVEConfig = mveConfig
		}

		req.AEndConfiguration = aEndConfig
	}

	// Handle B-End configuration
	if bEndConfigRaw, ok := rawData["bEndConfiguration"].(map[string]interface{}); ok {
		bEndConfig := megaport.VXCOrderEndpointConfiguration{}

		if productUID, ok := bEndConfigRaw["productUID"].(string); ok {
			bEndConfig.ProductUID = productUID
		}

		if vlan, ok := bEndConfigRaw["vlan"].(float64); ok {
			bEndConfig.VLAN = int(vlan)
		}

		if diversityZone, ok := bEndConfigRaw["diversityZone"].(string); ok {
			bEndConfig.DiversityZone = diversityZone
		}

		// Handle B-End partner config - directly use map data
		if partnerConfigRaw, ok := bEndConfigRaw["partnerConfig"].(map[string]interface{}); ok {
			partnerConfig, err := parsePartnerConfigFromMap(partnerConfigRaw)
			if err != nil {
				return nil, fmt.Errorf("error parsing B-End partner config: %v", err)
			}

			bEndConfig.PartnerConfig = partnerConfig
		}

		// Handle B-End MVE config
		innerVLAN, hasInnerVLAN := bEndConfigRaw["innerVlan"].(float64)
		vNicIndex, hasVNicIndex := bEndConfigRaw["vNicIndex"].(float64)

		if hasInnerVLAN || hasVNicIndex {
			mveConfig := &megaport.VXCOrderMVEConfig{}

			if hasInnerVLAN {
				mveConfig.InnerVLAN = int(innerVLAN)
			}

			if hasVNicIndex {
				mveConfig.NetworkInterfaceIndex = int(vNicIndex)
			}

			bEndConfig.VXCOrderMVEConfig = mveConfig
		}

		req.BEndConfiguration = bEndConfig
	}

	return req, nil
}

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

	// Handle partner configurations - using direct map access
	if aEndPartnerConfigRaw, ok := rawData["aEndPartnerConfig"].(map[string]interface{}); ok {
		if connectType, ok := aEndPartnerConfigRaw["connectType"].(string); ok && strings.ToUpper(connectType) == "VROUTER" {
			aEndPartnerConfig, err := parsePartnerConfigFromMap(aEndPartnerConfigRaw)
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
			bEndPartnerConfig, err := parsePartnerConfigFromMap(bEndPartnerConfigRaw)
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

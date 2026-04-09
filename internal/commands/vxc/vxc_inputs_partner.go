package vxc

import (
	"encoding/json"
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

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

package validation

import (
	"fmt"
	"strings"
)

// VXC Specific Validation
// This file contains validation functions for VXC-specific fields.
// These functions are used to validate requests and responses related to VXCs.

// Constants for validation
const (
	// BGP validation
	MinASPathPrependCount = 0
	MaxASPathPrependCount = 10

	// BFD validation
	MinBFDInterval   = 300
	MaxBFDInterval   = 30000
	MinBFDMultiplier = 3
	MaxBFDMultiplier = 20

	// MED validation
	MinMED = 0
	MaxMED = 4294967295

	// BGP peer types
	BGPPeerNonCloud  = "NON_CLOUD"
	BGPPeerPrivCloud = "PRIV_CLOUD"
	BGPPeerPubCloud  = "PUB_CLOUD"

	// BGP export policies
	BGPExportPolicyPermit = "permit"
	BGPExportPolicyDeny   = "deny"

	// IBM validation
	MaxIBMNameLength   = 100
	IBMAccountIDLength = 32

	// AWS connect types
	AWSConnectTypeAWS     = "AWS"
	AWSConnectTypeAWSHC   = "AWSHC"
	AWSConnectTypeTransit = "transit"
	AWSConnectTypePrivate = "private"
	AWSConnectTypePublic  = "public"
)

// Helper functions for validation

// ValidateVXCEndVLAN validates the VLAN ID for a VXC endpoint.
func ValidateVXCEndVLAN(vlan int) error {
	// Use the common validation function
	return ValidateVLAN(vlan)
}

// ValidateVXCEndInnerVLAN validates the inner VLAN ID for a VXC endpoint (Q-in-Q).
func ValidateVXCEndInnerVLAN(vlan int) error {
	// Inner VLAN typically follows the same rules as outer VLAN
	// Use the common validation function
	return ValidateVLAN(vlan)
}

// ValidateVXCRequest validates a VXC order request
// Only the A-End UID is required, and B-End UID is only required if no partner config is provided
func ValidateVXCRequest(name string, term int, rateLimit int, aEndUID string, bEndUID string, hasPartnerConfig bool) error {
	if name == "" {
		return NewValidationError("VXC name", name, "cannot be empty")
	}

	if err := ValidateContractTerm(term); err != nil {
		return err
	}

	if err := ValidateRateLimit(rateLimit); err != nil {
		return err
	}

	if aEndUID == "" {
		return NewValidationError("A-End UID", aEndUID, "cannot be empty")
	}

	// B-End UID is only required if there's no partner configuration
	if bEndUID == "" && !hasPartnerConfig {
		return NewValidationError("B-End UID", bEndUID, "cannot be empty when no partner configuration is provided")
	}

	return nil
}

// ValidateAWSPartnerConfig validates AWS partner configuration
func ValidateAWSPartnerConfig(connectType string, ownerAccount string, asn int, amazonAsn int, authKey string, customerIPAddress string, amazonIPAddress string, name string, awsType string) error {
	// Required fields check
	if connectType == "" {
		return NewValidationError("AWS connect type", connectType, "cannot be empty")
	}

	// Validate connect_type - must be 'AWS', 'AWSHC', 'transit', 'private', or 'public'
	if connectType != "AWS" && connectType != "AWSHC" && connectType != "transit" &&
		connectType != "private" && connectType != "public" {
		return NewValidationError("AWS connect type", connectType, "must be 'AWS', 'AWSHC', 'private', or 'public'")
	}

	if ownerAccount == "" {
		return NewValidationError("AWS owner account", ownerAccount, "cannot be empty")
	}

	// Validate IP Addresses if provided
	if customerIPAddress != "" {
		if err := ValidateCIDR(customerIPAddress, "AWS customer IP address"); err != nil {
			return NewValidationError("AWS customer IP address", customerIPAddress, "must be a valid IPv4 CIDR")
		}
	}

	if amazonIPAddress != "" {
		if err := ValidateCIDR(amazonIPAddress, "AWS Amazon IP address"); err != nil {
			return NewValidationError("AWS Amazon IP address", amazonIPAddress, "must be a valid IPv4 CIDR")
		}
	}

	// Validate AWS connection name if provided
	if name != "" {
		if len(name) > MaxAWSConnectionNameLength {
			return NewValidationError("AWS connection name", name, fmt.Sprintf("must be no longer than %d characters", MaxAWSConnectionNameLength))
		}
	}

	// Validate type if provided (typically used with public connect type)
	if awsType != "" && connectType == "AWS" {
		if awsType != "private" && awsType != "public" {
			return NewValidationError("AWS type", awsType, "must be 'private' or 'public' for AWS connect type")
		}
	}

	return nil
}

// ValidateAWSPartnerConfigMap validates AWS partner configuration from a map
func ValidateAWSPartnerConfigMap(config map[string]interface{}) error {
	// Check required fields first
	missingField := ValidateFieldPresence(config, []string{"connect_type", "owner_account"})
	if missingField != "" {
		return NewValidationError(fmt.Sprintf("AWS %s", missingField), "", "cannot be empty")
	}

	// Extract all fields with proper types
	fields := ExtractFieldsWithTypes(config, map[string]string{
		"connect_type":        "string",
		"owner_account":       "string",
		"asn":                 "int",
		"amazon_asn":          "int",
		"auth_key":            "string",
		"customer_ip_address": "string",
		"amazon_ip_address":   "string",
		"name":                "string",
		"type":                "string",
	})

	// Call the original validator with extracted fields
	connectType, ok1 := fields["connect_type"].(string)
	ownerAccount, ok2 := fields["owner_account"].(string)
	asn, ok3 := fields["asn"].(int)
	amazonAsn, ok4 := fields["amazon_asn"].(int)
	authKey, ok5 := fields["auth_key"].(string)
	customerIPAddress, ok6 := fields["customer_ip_address"].(string)
	amazonIPAddress, ok7 := fields["amazon_ip_address"].(string)
	name, ok8 := fields["name"].(string)
	awsType, ok9 := fields["type"].(string)

	// Check if type assertions were successful (should be, due to ExtractFieldsWithTypes)
	if !ok1 {
		return NewValidationError("AWS connect_type", fields["connect_type"], "must be a valid string")
	}
	if !ok2 {
		return NewValidationError("AWS owner_account", fields["owner_account"], "must be a valid string")
	}
	if !ok3 {
		return NewValidationError("AWS asn", fields["asn"], "must be a valid integer")
	}
	if !ok4 {
		return NewValidationError("AWS amazon_asn", fields["amazon_asn"], "must be a valid integer")
	}
	if !ok5 {
		return NewValidationError("AWS auth_key", fields["auth_key"], "must be a valid string")
	}
	if !ok6 {
		return NewValidationError("AWS customer_ip_address", fields["customer_ip_address"], "must be a valid string")
	}
	if !ok7 {
		return NewValidationError("AWS amazon_ip_address", fields["amazon_ip_address"], "must be a valid string")
	}
	if !ok8 {
		return NewValidationError("AWS name", fields["name"], "must be a valid string")
	}
	if !ok9 {
		return NewValidationError("AWS type", fields["type"], "must be a valid string")
	}

	return ValidateAWSPartnerConfig(
		connectType,
		ownerAccount,
		asn,
		amazonAsn,
		authKey,
		customerIPAddress,
		amazonIPAddress,
		name,
		awsType,
	)
}

// ValidateAzurePartnerConfig validates Azure partner configuration
func ValidateAzurePartnerConfig(serviceKey string, peers []map[string]interface{}) error {
	if serviceKey == "" {
		return NewValidationError("Azure service key", serviceKey, "cannot be empty")
	}

	// Validate peers if provided
	if len(peers) > 0 {
		for i, peer := range peers {
			// Validate peer type
			peerType, hasType := GetStringFromInterface(peer["type"])
			if hasType {
				if peerType != "private" && peerType != "microsoft" {
					return NewValidationError(fmt.Sprintf("Azure peer [%d] type", i), peerType, "must be 'private' or 'microsoft'")
				}
			}

			// Validate peer_asn if provided
			if peerASN, ok := GetStringFromInterface(peer["peer_asn"]); ok && peerASN != "" {
				// Azure ASN is stored as string but should be parseable as integer
				// and be within valid ranges
				var asnValue int
				_, err := fmt.Sscanf(peerASN, "%d", &asnValue)
				if err != nil {
					// Consider adding ASN range validation if needed
					return NewValidationError(fmt.Sprintf("Azure peer [%d] ASN", i), peerASN, "must be a valid ASN number")
				}
			}

			// Validate subnets if provided
			if primarySubnet, ok := GetStringFromInterface(peer["primary_subnet"]); ok && primarySubnet != "" {
				if err := ValidateCIDR(primarySubnet, fmt.Sprintf("Azure peer [%d] primary subnet", i)); err != nil {
					return err
				}
			}

			if secondarySubnet, ok := GetStringFromInterface(peer["secondary_subnet"]); ok && secondarySubnet != "" {
				if err := ValidateCIDR(secondarySubnet, fmt.Sprintf("Azure peer [%d] secondary subnet", i)); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ValidateGooglePartnerConfig validates Google partner configuration
func ValidateGooglePartnerConfig(pairingKey string) error {
	if pairingKey == "" {
		return NewValidationError("Google pairing key", pairingKey, "cannot be empty")
	}
	return nil
}

// ValidateOraclePartnerConfig validates Oracle partner configuration
func ValidateOraclePartnerConfig(virtualCircuitID string) error {
	if virtualCircuitID == "" {
		return NewValidationError("Oracle virtual circuit ID", virtualCircuitID, "cannot be empty")
	}
	return nil
}

// ValidateIBMPartnerConfig validates IBM partner configuration
func ValidateIBMPartnerConfig(accountID string, customerASN int, name string, customerIPAddress string, providerIPAddress string) error {
	// Account ID is required and must be a 32 character hexadecimal string
	if accountID == "" {
		return NewValidationError("IBM account ID", accountID, "cannot be empty")
	}

	// Validate account ID format - must be 32 hexadecimal characters
	if len(accountID) != IBMAccountIDLength {
		return NewValidationError("IBM account ID", accountID, fmt.Sprintf("must be exactly %d characters", IBMAccountIDLength))
	}

	// Check if account ID contains only hexadecimal characters
	validHex := true
	for _, c := range accountID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			validHex = false
			break
		}
	}
	if !validHex {
		return NewValidationError("IBM account ID", accountID, "must contain only hexadecimal characters (0-9, a-f, A-F)")
	}

	// Validate name if provided
	if name != "" {
		// Max 100 characters from 0-9 a-z A-Z / - _ ,
		if len(name) > MaxIBMNameLength {
			return NewValidationError("IBM connection name", name,
				fmt.Sprintf("must be no longer than %d characters", MaxIBMNameLength))
		}

		validNameChars := func(c rune) bool {
			return (c >= '0' && c <= '9') ||
				(c >= 'a' && c <= 'z') ||
				(c >= 'A' && c <= 'Z') ||
				c == '/' || c == '-' || c == '_' || c == ','
		}

		for _, c := range name {
			if !validNameChars(c) {
				return NewValidationError("IBM connection name", name,
					"must only contain characters 0-9, a-z, A-Z, /, -, _, or ,")
			}
		}
	}

	// Validate IP addresses if provided
	if customerIPAddress != "" {
		if err := ValidateCIDR(customerIPAddress, "IBM customer IP address"); err != nil {
			return err
		}
	}

	if providerIPAddress != "" {
		if err := ValidateCIDR(providerIPAddress, "IBM provider IP address"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateVXCPartnerConfig validates that the partner configuration is valid
// It ensures that only one partner configuration type is provided, and that the
// configuration for that partner type is valid according to the schema.
func ValidateVXCPartnerConfig(config map[string]interface{}) error {
	// Map raw field names to user-friendly display names
	fieldDisplayNames := map[string]string{
		"connect_type":        "connect type",
		"owner_account":       "owner account",
		"asn":                 "ASN",
		"amazon_asn":          "Amazon ASN",
		"auth_key":            "auth key",
		"customer_ip_address": "customer IP address",
		"amazon_ip_address":   "Amazon IP address",
		"name":                "name",
		"type":                "type",
	}

	// Supported partner types and their config keys
	var partnerValidators = map[string]struct {
		configKey string
		validate  func(map[string]interface{}) error
	}{
		"aws": {
			configKey: "aws_config",
			validate: func(cfg map[string]interface{}) error {
				err := ValidateAWSPartnerConfigMap(cfg)
				if err != nil {
					ve, ok := err.(*ValidationError)
					if !ok {
						// If it's not a ValidationError, return it directly
						return err
					}
					// Prefer "connect type" over "type" for connect_type errors
					if strings.Contains(ve.Field, "connect_type") {
						return NewValidationError("AWS connect type", ve.Value, ve.Reason)
					}
					for rawField, friendlyField := range fieldDisplayNames {
						if strings.Contains(ve.Field, rawField) {
							return NewValidationError(fmt.Sprintf("AWS %s", friendlyField), ve.Value, ve.Reason)
						}
					}
				}
				return err
			},
		},
		"azure": {
			configKey: "azure_config",
			validate: func(cfg map[string]interface{}) error {
				serviceKey, _ := GetStringFromInterface(cfg["service_key"])
				peers, _ := GetSliceMapStringInterfaceFromInterface(cfg["peers"])
				return ValidateAzurePartnerConfig(serviceKey, peers)
			},
		},
		"google": {
			configKey: "google_config",
			validate: func(cfg map[string]interface{}) error {
				pairingKey, _ := GetStringFromInterface(cfg["pairing_key"])
				return ValidateGooglePartnerConfig(pairingKey)
			},
		},
		"oracle": {
			configKey: "oracle_config",
			validate: func(cfg map[string]interface{}) error {
				virtualCircuitID, _ := GetStringFromInterface(cfg["virtual_circuit_id"])
				return ValidateOraclePartnerConfig(virtualCircuitID)
			},
		},
		"ibm": {
			configKey: "ibm_config",
			validate: func(cfg map[string]interface{}) error {
				fields := ExtractFieldsWithTypes(cfg, map[string]string{
					"account_id":          "string",
					"customer_asn":        "int",
					"name":                "string",
					"customer_ip_address": "string",
					"provider_ip_address": "string",
				})
				accountID, ok1 := fields["account_id"].(string)
				customerASN, ok2 := fields["customer_asn"].(int)
				name, ok3 := fields["name"].(string)
				customerIPAddress, ok4 := fields["customer_ip_address"].(string)
				providerIPAddress, ok5 := fields["provider_ip_address"].(string)

				if !ok1 {
					return NewValidationError("IBM account_id", fields["account_id"], "must be a valid string")
				}
				if !ok2 {
					return NewValidationError("IBM customer_asn", fields["customer_asn"], "must be a valid integer")
				}
				if !ok3 {
					return NewValidationError("IBM name", fields["name"], "must be a valid string")
				}
				if !ok4 {
					return NewValidationError("IBM customer_ip_address", fields["customer_ip_address"], "must be a valid string")
				}
				if !ok5 {
					return NewValidationError("IBM provider_ip_address", fields["provider_ip_address"], "must be a valid string")
				}

				return ValidateIBMPartnerConfig(
					accountID,
					customerASN,
					name,
					customerIPAddress,
					providerIPAddress,
				)
			},
		},
		"vrouter": {
			configKey: "vrouter_config",
			validate: func(cfg map[string]interface{}) error {
				interfaces, _ := GetSliceMapStringInterfaceFromInterface(cfg["interfaces"])
				return ValidateVrouterPartnerConfig(interfaces)
			},
		},
	}

	partnerType, hasPartner := GetStringFromInterface(config["partner"])
	if !hasPartner || partnerType == "" {
		return NewValidationError("Partner type", "", "cannot be empty")
	}

	validatorEntry, ok := partnerValidators[partnerType]
	if !ok {
		// Build a list of valid partner types in the expected order for the error message
		orderedPartnerTypes := []string{"aws", "azure", "google", "oracle", "ibm", "vrouter"}
		// Join with commas and "or" before the last item
		var partnersMsg string
		if len(orderedPartnerTypes) > 1 {
			partnersMsg = strings.Join(orderedPartnerTypes[:len(orderedPartnerTypes)-1], ", ") + ", or " + orderedPartnerTypes[len(orderedPartnerTypes)-1]
		} else {
			partnersMsg = orderedPartnerTypes[0]
		}
		return NewValidationError("Partner type", partnerType,
			fmt.Sprintf("must be one of %s", partnersMsg))
	}

	// Count the number of partner configs provided
	configCount := 0
	var foundConfig map[string]interface{}

	for partnerKey, entry := range partnerValidators {
		cfg, has := GetMapStringInterfaceFromInterface(config[entry.configKey])
		if has {
			configCount++
			if partnerType == partnerKey {
				foundConfig = cfg
			} else {
				// Ensure consistent uppercase for "AWS" instead of "Aws"
				partnerKeyDisplay := partnerKey
				if partnerKey == "aws" {
					partnerKeyDisplay = "AWS"
				} else {
					// Use a simple alternative to strings.Title for ASCII keys
					if len(partnerKey) > 0 {
						partnerKeyDisplay = strings.ToUpper(partnerKey[:1]) + partnerKey[1:]
					}
				}
				return NewValidationError(partnerKeyDisplay+" config", cfg,
					fmt.Sprintf("cannot be provided when partner type is not %s", partnerKey))
			}
		}
	}

	// Check for deprecated partner_a_end_config and handle more gracefully
	_, hasPartnerAEnd := GetMapStringInterfaceFromInterface(config["partner_a_end_config"])
	if hasPartnerAEnd {
		configCount++
		fmt.Println("Warning: partner_a_end_config is deprecated, please use vrouter_config instead")
		// Make this a warning rather than an error if it's for the vrouter type
		if partnerType == "vrouter" {
			// Just continue with processing rather than returning an error
			return nil
		}
	}

	if configCount == 0 {
		return NewValidationError("Partner configuration", nil,
			fmt.Sprintf("no configuration provided for partner type '%s'", partnerType))
	} else if configCount > 1 && partnerType != "vrouter" {
		// Skip this check for vrouter type since we're handling partner_a_end_config as a warning
		return NewValidationError("Partner configuration", nil,
			"only one partner configuration can be provided")
	}

	if foundConfig == nil && partnerType != "vrouter" {
		// Skip this check for vrouter if we have the deprecated config
		return NewValidationError("Partner configuration", nil,
			fmt.Sprintf("missing configuration for partner type '%s'", partnerType))
	}

	if foundConfig != nil {
		return validatorEntry.validate(foundConfig)
	}

	return nil
}

// ValidateVrouterPartnerConfig validates partner-specific configuration for a VXC attached to a VRouter.
// Assuming it includes VLAN validation, update it similarly.
func ValidateVrouterPartnerConfig(interfaces []map[string]interface{}) error {
	// Interfaces are required for vRouter config
	if len(interfaces) == 0 {
		return NewValidationError("vRouter interfaces", nil, "at least one interface must be provided")
	}

	// Validate each interface
	for i, iface := range interfaces {
		// Validate VLAN if provided
		if vlan, ok := GetIntFromInterface(iface["vlan"]); ok && vlan != 0 {
			if err := ValidateVLAN(vlan); err != nil {
				return NewValidationError(fmt.Sprintf("vRouter interface [%d] VLAN", i), vlan,
					fmt.Sprintf("must be between %d-%d (%d is reserved)", MinVLAN, MaxVLAN, ReservedVLAN))
			}
		}

		// Validate IP addresses if provided
		if ipAddresses, ok := GetSliceInterfaceFromInterface(iface["ip_addresses"]); ok && len(ipAddresses) > 0 {
			for j, ip := range ipAddresses {
				ipStr, isStr := GetStringFromInterface(ip)
				if !isStr {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP address [%d]", i, j), ip,
						"must be a string in CIDR format")
				}
				if err := ValidateCIDR(ipStr, fmt.Sprintf("vRouter interface [%d] IP address [%d]", i, j)); err != nil {
					return err
				}
			}
		}

		// Validate NAT IP addresses if provided
		if natIPs, ok := GetSliceInterfaceFromInterface(iface["nat_ip_addresses"]); ok && len(natIPs) > 0 {
			for j, ip := range natIPs {
				ipStr, isStr := GetStringFromInterface(ip)
				if !isStr {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] NAT IP address [%d]", i, j), ip,
						"must be a string in CIDR format")
				}
				if err := ValidateCIDR(ipStr, fmt.Sprintf("vRouter interface [%d] NAT IP address [%d]", i, j)); err != nil {
					return err
				}
			}
		}

		// Validate IP routes if provided
		if ipRoutes, ok := GetSliceInterfaceFromInterface(iface["ip_routes"]); ok && len(ipRoutes) > 0 {
			for j, route := range ipRoutes {
				routeMap, isMap := GetMapStringInterfaceFromInterface(route)
				if !isMap {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d]", i, j), route,
						"must be a valid route configuration map")
				}
				if err := ValidateIPRoute(routeMap, i, j); err != nil {
					return err
				}
			}
		}

		// Validate BFD configuration if provided
		if bfd, ok := GetMapStringInterfaceFromInterface(iface["bfd"]); ok {
			if err := ValidateBFDConfig(bfd, i); err != nil {
				return err
			}
		}

		// Validate BGP connections if provided
		if bgpConns, ok := GetSliceInterfaceFromInterface(iface["bgp_connections"]); ok && len(bgpConns) > 0 {
			for j, conn := range bgpConns {
				connMap, isMap := GetMapStringInterfaceFromInterface(conn)
				if !isMap {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d]", i, j), conn,
						"must be a valid BGP connection configuration map")
				}
				if err := ValidateBGPConnection(connMap, i, j); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ValidateIPRoute validates a route configuration
func ValidateIPRoute(route map[string]interface{}, ifaceIndex, routeIndex int) error {
	// Prefix is required and must be a valid CIDR
	prefix, hasPrefix := GetStringFromInterface(route["prefix"])
	if !hasPrefix || prefix == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] prefix", ifaceIndex, routeIndex), route["prefix"],
			"cannot be empty and must be a string")
	}
	if err := ValidateCIDR(prefix, fmt.Sprintf("vRouter interface [%d] IP route [%d] prefix", ifaceIndex, routeIndex)); err != nil {
		return err
	}

	// Next hop is required and must be a valid IP address
	nextHop, hasNextHop := GetStringFromInterface(route["next_hop"])
	if !hasNextHop || nextHop == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), route["next_hop"],
			"cannot be empty and must be a string")
	}

	// Validate next hop as IP address (not CIDR)
	if strings.Contains(nextHop, "/") {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), nextHop,
			"must be a valid IPv4 address (not CIDR)")
	}
	if err := ValidateIPv4(nextHop, fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex)); err != nil {
		return err // Use the more specific error from ValidateIPv4
	}

	return nil
}

// ValidateBFDConfig validates a BFD configuration
func ValidateBFDConfig(bfd map[string]interface{}, ifaceIndex int) error {
	// TX interval - must be between MinBFDInterval-MaxBFDInterval ms
	if txInterval, ok := GetIntFromInterface(bfd["tx_interval"]); ok {
		if txInterval < MinBFDInterval || txInterval > MaxBFDInterval {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD TX interval", ifaceIndex), txInterval,
				fmt.Sprintf("must be between %d-%d milliseconds", MinBFDInterval, MaxBFDInterval))
		}
	} else if bfd["tx_interval"] != nil { // Check if key exists but type is wrong
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD TX interval", ifaceIndex), bfd["tx_interval"], "must be a valid integer")
	}

	// RX interval - must be between MinBFDInterval-MaxBFDInterval ms
	if rxInterval, ok := GetIntFromInterface(bfd["rx_interval"]); ok {
		if rxInterval < MinBFDInterval || rxInterval > MaxBFDInterval {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD RX interval", ifaceIndex), rxInterval,
				fmt.Sprintf("must be between %d-%d milliseconds", MinBFDInterval, MaxBFDInterval))
		}
	} else if bfd["rx_interval"] != nil {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD RX interval", ifaceIndex), bfd["rx_interval"], "must be a valid integer")
	}

	// Multiplier - must be between MinBFDMultiplier-MaxBFDMultiplier
	if multiplier, ok := GetIntFromInterface(bfd["multiplier"]); ok {
		if multiplier < MinBFDMultiplier || multiplier > MaxBFDMultiplier {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD multiplier", ifaceIndex), multiplier,
				fmt.Sprintf("must be between %d-%d", MinBFDMultiplier, MaxBFDMultiplier))
		}
	} else if bfd["multiplier"] != nil {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD multiplier", ifaceIndex), bfd["multiplier"], "must be a valid integer")
	}

	return nil
}

// ValidateBGPConnection validates a BGP connection configuration
func ValidateBGPConnection(conn map[string]interface{}, ifaceIndex, connIndex int) error {
	fieldPrefix := fmt.Sprintf("vRouter interface [%d] BGP connection [%d]", ifaceIndex, connIndex)

	// Peer ASN - required, type check (int or string parsable to int)
	peerASNVal, hasPeerASN := conn["peer_asn"]
	if !hasPeerASN {
		return NewValidationError(fmt.Sprintf("%s peer ASN", fieldPrefix), nil, "is required")
	}
	if _, ok := GetIntFromInterface(peerASNVal); !ok { // Allow int, float64, or string representation
		return NewValidationError(fmt.Sprintf("%s peer ASN", fieldPrefix), peerASNVal, "must be a valid integer ASN")
	}
	// Add ASN range validation if needed

	// Local ASN - optional, type check if present
	if localASNVal, hasLocalASN := conn["local_asn"]; hasLocalASN {
		if _, ok := GetIntFromInterface(localASNVal); !ok {
			return NewValidationError(fmt.Sprintf("%s local ASN", fieldPrefix), localASNVal, "must be a valid integer ASN")
		}
		// Add ASN range validation if needed
	}

	// Local IP address validation (required and must be a valid IP or CIDR)
	localIP, hasLocalIP := GetStringFromInterface(conn["local_ip_address"])
	if !hasLocalIP || localIP == "" {
		return NewValidationError(fmt.Sprintf("%s local IP address", fieldPrefix), conn["local_ip_address"], "cannot be empty and must be a string")
	}
	if strings.Contains(localIP, "/") {
		if err := ValidateCIDR(localIP, fmt.Sprintf("%s local IP address", fieldPrefix)); err != nil {
			return err
		}
	} else {
		if err := ValidateIPv4(localIP, fmt.Sprintf("%s local IP address", fieldPrefix)); err != nil {
			return err
		}
	}

	// Peer IP address validation (required and must be a valid IP or CIDR)
	peerIP, hasPeerIP := GetStringFromInterface(conn["peer_ip_address"])
	if !hasPeerIP || peerIP == "" {
		return NewValidationError(fmt.Sprintf("%s peer IP address", fieldPrefix), conn["peer_ip_address"], "cannot be empty and must be a string")
	}
	if strings.Contains(peerIP, "/") {
		if err := ValidateCIDR(peerIP, fmt.Sprintf("%s peer IP address", fieldPrefix)); err != nil {
			return err
		}
	} else {
		if err := ValidateIPv4(peerIP, fmt.Sprintf("%s peer IP address", fieldPrefix)); err != nil {
			return err
		}
	}

	// Password validation removed - let API handle it

	// Validate peer type if provided
	if peerType, ok := GetStringFromInterface(conn["peer_type"]); ok && peerType != "" {
		validTypes := []string{BGPPeerNonCloud, BGPPeerPrivCloud, BGPPeerPubCloud}
		isValid := false
		for _, vt := range validTypes {
			if peerType == vt {
				isValid = true
				break
			}
		}
		if !isValid {
			return NewValidationError(fmt.Sprintf("%s peer type", fieldPrefix), peerType,
				fmt.Sprintf("must be one of '%s', '%s', or '%s'", BGPPeerNonCloud, BGPPeerPrivCloud, BGPPeerPubCloud))
		}
	}

	// Validate MED (Multi-Exit Discriminator) values
	if medInVal, hasMedIn := conn["med_in"]; hasMedIn {
		if medIn, ok := GetIntFromInterface(medInVal); ok {
			if medIn < MinMED || medIn > MaxMED {
				return NewValidationError(fmt.Sprintf("%s MED in", fieldPrefix), medIn,
					fmt.Sprintf("must be between %d-%d", MinMED, MaxMED))
			}
		} else {
			return NewValidationError(fmt.Sprintf("%s MED in", fieldPrefix), medInVal, "must be a valid integer")
		}
	}

	if medOutVal, hasMedOut := conn["med_out"]; hasMedOut {
		if medOut, ok := GetIntFromInterface(medOutVal); ok {
			if medOut < MinMED || medOut > MaxMED {
				return NewValidationError(fmt.Sprintf("%s MED out", fieldPrefix), medOut,
					fmt.Sprintf("must be between %d-%d", MinMED, MaxMED))
			}
		} else {
			return NewValidationError(fmt.Sprintf("%s MED out", fieldPrefix), medOutVal, "must be a valid integer")
		}
	}

	// AS path prepend count validation
	if asPathPrependCountVal, hasASPath := conn["as_path_prepend_count"]; hasASPath {
		if asPathPrependCount, ok := GetIntFromInterface(asPathPrependCountVal); ok {
			if asPathPrependCount < MinASPathPrependCount || asPathPrependCount > MaxASPathPrependCount {
				return NewValidationError(fmt.Sprintf("%s AS path prepend count", fieldPrefix), asPathPrependCount,
					fmt.Sprintf("must be between %d-%d", MinASPathPrependCount, MaxASPathPrependCount))
			}
		} else {
			return NewValidationError(fmt.Sprintf("%s AS path prepend count", fieldPrefix), asPathPrependCountVal, "must be a valid integer")
		}
	}

	// Export policy validation
	if exportPolicy, ok := GetStringFromInterface(conn["export_policy"]); ok && exportPolicy != "" {
		if exportPolicy != BGPExportPolicyPermit && exportPolicy != BGPExportPolicyDeny {
			return NewValidationError(fmt.Sprintf("%s export policy", fieldPrefix), exportPolicy,
				fmt.Sprintf("must be '%s' or '%s'", BGPExportPolicyPermit, BGPExportPolicyDeny))
		}
	}

	return nil
}

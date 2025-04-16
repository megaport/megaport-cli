package validation

import (
	"fmt"
	"net"
	"strings"
)

// VXC Specific Validation
// This file contains validation functions for VXC-specific fields.
// These functions are used to validate requests and responses related to VXCs.

// Constants for validation
const (
	// VLAN ranges
	MinVLAN        = 2
	MaxVLAN        = 4093
	UntaggedVLAN   = -1
	AutoAssignVLAN = 0
	ReservedVLAN   = 1

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

// validateIPv4CIDR uses the net package to validate CIDR notation
func validateIPv4CIDR(cidr string, fieldName string) error {
	if cidr == "" {
		return NewValidationError(fieldName, cidr, "cannot be empty")
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return NewValidationError(fieldName, cidr, "must be a valid IPv4 CIDR notation")
	}

	// Ensure it's an IPv4 CIDR
	if ipNet.IP.To4() == nil {
		return NewValidationError(fieldName, cidr, "must be an IPv4 CIDR (not IPv6)")
	}

	return nil
}

// ValidateVXCEndVLAN validates a VXC endpoint VLAN
func ValidateVXCEndVLAN(vlan int, endName string) error {
	if vlan == UntaggedVLAN || vlan == AutoAssignVLAN || (vlan >= MinVLAN && vlan <= MaxVLAN) {
		return nil
	}
	return NewValidationError(fmt.Sprintf("%s VLAN", endName), vlan,
		fmt.Sprintf("must be %d (untagged), %d (auto-assigned), or between %d-%d (%d is reserved)",
			UntaggedVLAN, AutoAssignVLAN, MinVLAN, MaxVLAN, ReservedVLAN))
}

// ValidateVXCEndInnerVLAN validates a VXC endpoint inner VLAN (for QinQ)
func ValidateVXCEndInnerVLAN(vlan int, outerVLAN int, endName string) error {
	// Inner VLAN can't be set if outer VLAN is untagged (-1)
	if outerVLAN == -1 && vlan != 0 {
		return NewValidationError(fmt.Sprintf("%s inner VLAN", endName), vlan,
			"cannot be set when the outer VLAN is untagged (-1)")
	}

	// For non-zero inner VLANs, they should be in the valid range
	if vlan != 0 && (vlan < 2 || vlan > 4093) {
		return NewValidationError(fmt.Sprintf("%s inner VLAN", endName), vlan,
			"must be 0 (not set) or between 2-4093 (1 is reserved)")
	}

	return nil
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
		if err := ValidateIPv4CIDR(customerIPAddress); err != nil {
			return NewValidationError("AWS customer IP address", customerIPAddress, "must be a valid IPv4 CIDR")
		}
	}

	if amazonIPAddress != "" {
		if err := ValidateIPv4CIDR(amazonIPAddress); err != nil {
			return NewValidationError("AWS Amazon IP address", amazonIPAddress, "must be a valid IPv4 CIDR")
		}
	}

	// Validate AWS connection name if provided
	if name != "" {
		if len(name) > 255 {
			return NewValidationError("AWS connection name", name, "must be no longer than 255 characters")
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

// ValidateAzurePartnerConfig validates Azure partner configuration
func ValidateAzurePartnerConfig(serviceKey string, peers []map[string]interface{}) error {
	if serviceKey == "" {
		return NewValidationError("Azure service key", serviceKey, "cannot be empty")
	}

	// Validate peers if provided
	if len(peers) > 0 {
		for i, peer := range peers {
			// Validate peer type
			peerType, hasType := peer["type"].(string)
			if hasType {
				if peerType != "private" && peerType != "microsoft" {
					return NewValidationError(fmt.Sprintf("Azure peer [%d] type", i), peerType, "must be 'private' or 'microsoft'")
				}
			}

			// Validate peer_asn if provided
			if peerASN, ok := peer["peer_asn"].(string); ok && peerASN != "" {
				// Azure ASN is stored as string but should be parseable as integer
				// and be within valid ranges
				var asnValue int
				_, err := fmt.Sscanf(peerASN, "%d", &asnValue)
				if err != nil {
					return NewValidationError(fmt.Sprintf("Azure peer [%d] ASN", i), peerASN, "must be a valid ASN number")
				}
			}

			// Validate subnets if provided
			if primarySubnet, ok := peer["primary_subnet"].(string); ok && primarySubnet != "" {
				if err := ValidateIPv4CIDR(primarySubnet); err != nil {
					return NewValidationError(fmt.Sprintf("Azure peer [%d] primary subnet", i), primarySubnet,
						"must be a valid IPv4 CIDR")
				}
			}

			if secondarySubnet, ok := peer["secondary_subnet"].(string); ok && secondarySubnet != "" {
				if err := ValidateIPv4CIDR(secondarySubnet); err != nil {
					return NewValidationError(fmt.Sprintf("Azure peer [%d] secondary subnet", i), secondarySubnet,
						"must be a valid IPv4 CIDR")
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
		if err := validateIPv4CIDR(customerIPAddress, "IBM customer IP address"); err != nil {
			return err
		}
	}

	if providerIPAddress != "" {
		if err := validateIPv4CIDR(providerIPAddress, "IBM provider IP address"); err != nil {
			return err
		}
	}

	return nil
}

// ValidateIPv4CIDR validates that a string is a valid IPv4 CIDR
func ValidateIPv4CIDR(cidr string) error {
	return validateIPv4CIDR(cidr, "CIDR")
}

// ValidateVXCPartnerConfig validates that the partner configuration is valid
// It ensures that only one partner configuration type is provided, and that the
// configuration for that partner type is valid according to the schema.
func ValidateVXCPartnerConfig(config map[string]interface{}) error {
	// Extract the partner type
	partnerType, hasPartner := config["partner"].(string)
	if !hasPartner || partnerType == "" {
		return NewValidationError("Partner type", "", "cannot be empty")
	}

	// Check that partner type is one of the supported types
	validPartners := []string{"aws", "azure", "google", "oracle", "ibm", "vrouter"}
	isValidPartner := false
	for _, p := range validPartners {
		if partnerType == p {
			isValidPartner = true
			break
		}
	}
	if !isValidPartner {
		return NewValidationError("Partner type", partnerType,
			"must be one of aws, azure, google, oracle, ibm, or vrouter")
	}

	// Count the number of partner configs provided
	configCount := 0

	// Check for AWS config
	awsConfig, hasAWS := config["aws_config"].(map[string]interface{})
	if hasAWS && awsConfig != nil {
		configCount++

		// Only validate AWS config if partner type is aws
		if partnerType == "aws" {
			// Extract and validate AWS config fields
			connectType := ""
			if ct, ok := awsConfig["connect_type"].(string); ok {
				connectType = ct
			}

			ownerAccount := ""
			if oa, ok := awsConfig["owner_account"].(string); ok {
				ownerAccount = oa
			}

			asn := 0
			if a, ok := awsConfig["asn"].(int); ok {
				asn = a
			} else if a, ok := awsConfig["asn"].(float64); ok {
				asn = int(a)
			}

			amazonAsn := 0
			if a, ok := awsConfig["amazon_asn"].(int); ok {
				amazonAsn = a
			} else if a, ok := awsConfig["amazon_asn"].(float64); ok {
				amazonAsn = int(a)
			}

			authKey := ""
			if ak, ok := awsConfig["auth_key"].(string); ok {
				authKey = ak
			}

			customerIPAddress := ""
			if cip, ok := awsConfig["customer_ip_address"].(string); ok {
				customerIPAddress = cip
			}

			amazonIPAddress := ""
			if aip, ok := awsConfig["amazon_ip_address"].(string); ok {
				amazonIPAddress = aip
			}

			name := ""
			if n, ok := awsConfig["name"].(string); ok {
				name = n
			}

			awsType := ""
			if t, ok := awsConfig["type"].(string); ok {
				awsType = t
			}

			if err := ValidateAWSPartnerConfig(connectType, ownerAccount, asn, amazonAsn, authKey,
				customerIPAddress, amazonIPAddress, name, awsType); err != nil {
				return err
			}
		} else if hasAWS {
			// If AWS config is provided but partner type is not aws, return an error
			return NewValidationError("AWS config", awsConfig,
				"cannot be provided when partner type is not aws")
		}
	}

	// Check for Azure config
	azureConfig, hasAzure := config["azure_config"].(map[string]interface{})
	if hasAzure && azureConfig != nil {
		configCount++

		// Only validate Azure config if partner type is azure
		if partnerType == "azure" {
			// Extract and validate Azure config fields
			serviceKey := ""
			if sk, ok := azureConfig["service_key"].(string); ok {
				serviceKey = sk
			}

			var peers []map[string]interface{}
			if p, ok := azureConfig["peers"].([]map[string]interface{}); ok {
				peers = p
			} else if p, ok := azureConfig["peers"].([]interface{}); ok {
				for _, peer := range p {
					if peerMap, isPeerMap := peer.(map[string]interface{}); isPeerMap {
						peers = append(peers, peerMap)
					}
				}
			}

			if err := ValidateAzurePartnerConfig(serviceKey, peers); err != nil {
				return err
			}
		} else if hasAzure {
			// If Azure config is provided but partner type is not azure, return an error
			return NewValidationError("Azure config", azureConfig,
				"cannot be provided when partner type is not azure")
		}
	}

	// Check for Google config
	googleConfig, hasGoogle := config["google_config"].(map[string]interface{})
	if hasGoogle && googleConfig != nil {
		configCount++

		// Only validate Google config if partner type is google
		if partnerType == "google" {
			// Extract and validate Google config fields
			pairingKey := ""
			if pk, ok := googleConfig["pairing_key"].(string); ok {
				pairingKey = pk
			}

			if err := ValidateGooglePartnerConfig(pairingKey); err != nil {
				return err
			}
		} else if hasGoogle {
			// If Google config is provided but partner type is not google, return an error
			return NewValidationError("Google config", googleConfig,
				"cannot be provided when partner type is not google")
		}
	}

	// Check for Oracle config
	oracleConfig, hasOracle := config["oracle_config"].(map[string]interface{})
	if hasOracle && oracleConfig != nil {
		configCount++

		// Only validate Oracle config if partner type is oracle
		if partnerType == "oracle" {
			// Extract and validate Oracle config fields
			virtualCircuitID := ""
			if vcid, ok := oracleConfig["virtual_circuit_id"].(string); ok {
				virtualCircuitID = vcid
			}

			if err := ValidateOraclePartnerConfig(virtualCircuitID); err != nil {
				return err
			}
		} else if hasOracle {
			// If Oracle config is provided but partner type is not oracle, return an error
			return NewValidationError("Oracle config", oracleConfig,
				"cannot be provided when partner type is not oracle")
		}
	}

	// Check for IBM config
	ibmConfig, hasIBM := config["ibm_config"].(map[string]interface{})
	if hasIBM && ibmConfig != nil {
		configCount++

		// Only validate IBM config if partner type is ibm
		if partnerType == "ibm" {
			// Extract and validate IBM config fields
			accountID := ""
			if aid, ok := ibmConfig["account_id"].(string); ok {
				accountID = aid
			}

			customerASN := 0
			if casn, ok := ibmConfig["customer_asn"].(int); ok {
				customerASN = casn
			} else if casn, ok := ibmConfig["customer_asn"].(float64); ok {
				customerASN = int(casn)
			}

			name := ""
			if n, ok := ibmConfig["name"].(string); ok {
				name = n
			}

			customerIPAddress := ""
			if cip, ok := ibmConfig["customer_ip_address"].(string); ok {
				customerIPAddress = cip
			}

			providerIPAddress := ""
			if pip, ok := ibmConfig["provider_ip_address"].(string); ok {
				providerIPAddress = pip
			}

			if err := ValidateIBMPartnerConfig(accountID, customerASN, name, customerIPAddress, providerIPAddress); err != nil {
				return err
			}
		} else if hasIBM {
			// If IBM config is provided but partner type is not ibm, return an error
			return NewValidationError("IBM config", ibmConfig,
				"cannot be provided when partner type is not ibm")
		}
	}

	// Check for vRouter config
	vrouterConfig, hasVrouter := config["vrouter_config"].(map[string]interface{})
	if hasVrouter && vrouterConfig != nil {
		configCount++

		// Only validate vRouter config if partner type is vrouter
		if partnerType == "vrouter" {
			// vRouter validation is complex and would require additional functions
			// to validate the interfaces, BGP connections, etc.
			// For now, we'll just check if it exists

			// TODO: Add detailed validation for vRouter config
			var interfaces []map[string]interface{}
			if ifaces, ok := vrouterConfig["interfaces"].([]interface{}); ok {
				for _, iface := range ifaces {
					if ifaceMap, isIfaceMap := iface.(map[string]interface{}); isIfaceMap {
						interfaces = append(interfaces, ifaceMap)
					}
				}
			}

			if err := ValidateVrouterPartnerConfig(interfaces); err != nil {
				return err
			}
		} else if hasVrouter {
			// If vRouter config is provided but partner type is not vrouter, return an error
			return NewValidationError("vRouter config", vrouterConfig,
				"cannot be provided when partner type is not vrouter")
		}
	}

	// Check for deprecated partner_a_end_config
	_, hasPartnerAEnd := config["partner_a_end_config"].(map[string]interface{})
	if hasPartnerAEnd {
		configCount++

		// Display a warning about using deprecated config
		fmt.Println("Warning: partner_a_end_config is deprecated, please use vrouter_config instead")
	}

	// Ensure exactly one partner config is provided
	if configCount == 0 {
		return NewValidationError("Partner configuration", nil,
			fmt.Sprintf("no configuration provided for partner type '%s'", partnerType))
	} else if configCount > 1 {
		return NewValidationError("Partner configuration", nil,
			"only one partner configuration can be provided")
	}

	return nil
}

// ValidateVrouterPartnerConfig validates vRouter partner configuration
func ValidateVrouterPartnerConfig(interfaces []map[string]interface{}) error {
	// Interfaces are required for vRouter config
	if len(interfaces) == 0 {
		return NewValidationError("vRouter interfaces", nil, "at least one interface must be provided")
	}

	// Validate each interface
	for i, iface := range interfaces {
		// Validate VLAN if provided
		if vlan, ok := iface["vlan"].(int); ok && vlan != 0 {
			if vlan < 2 || vlan > 4093 {
				return NewValidationError(fmt.Sprintf("vRouter interface [%d] VLAN", i), vlan,
					"must be between 2-4093 (1 is reserved)")
			}
		}

		// Validate IP addresses if provided
		if ipAddresses, ok := iface["ip_addresses"].([]interface{}); ok && len(ipAddresses) > 0 {
			for j, ip := range ipAddresses {
				ipStr, isStr := ip.(string)
				if !isStr {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP address [%d]", i, j), ip,
						"must be a string in CIDR format")
				}

				if err := ValidateIPv4CIDR(ipStr); err != nil {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP address [%d]", i, j), ipStr,
						"must be a valid IPv4 CIDR")
				}
			}
		}

		// Validate NAT IP addresses if provided
		if natIPs, ok := iface["nat_ip_addresses"].([]interface{}); ok && len(natIPs) > 0 {
			for j, ip := range natIPs {
				ipStr, isStr := ip.(string)
				if !isStr {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] NAT IP address [%d]", i, j), ip,
						"must be a string in CIDR format")
				}

				if err := ValidateIPv4CIDR(ipStr); err != nil {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] NAT IP address [%d]", i, j), ipStr,
						"must be a valid IPv4 CIDR")
				}
			}
		}

		// Validate IP routes if provided
		if ipRoutes, ok := iface["ip_routes"].([]interface{}); ok && len(ipRoutes) > 0 {
			for j, route := range ipRoutes {
				routeMap, isMap := route.(map[string]interface{})
				if !isMap {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d]", i, j), route,
						"must be a valid route configuration")
				}

				if err := ValidateIPRoute(routeMap, i, j); err != nil {
					return err
				}
			}
		}

		// Validate BFD configuration if provided
		if bfd, ok := iface["bfd"].(map[string]interface{}); ok && bfd != nil {
			if err := ValidateBFDConfig(bfd, i); err != nil {
				return err
			}
		}

		// Validate BGP connections if provided
		if bgpConns, ok := iface["bgp_connections"].([]interface{}); ok && len(bgpConns) > 0 {
			for j, conn := range bgpConns {
				connMap, isMap := conn.(map[string]interface{})
				if !isMap {
					return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d]", i, j), conn,
						"must be a valid BGP connection configuration")
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
	prefix, hasPrefix := route["prefix"].(string)
	if !hasPrefix || prefix == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] prefix", ifaceIndex, routeIndex), prefix,
			"cannot be empty")
	}

	if err := ValidateIPv4CIDR(prefix); err != nil {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] prefix", ifaceIndex, routeIndex), prefix,
			"must be a valid IPv4 CIDR")
	}

	// Next hop is required and must be a valid IP address
	nextHop, hasNextHop := route["next_hop"].(string)
	if !hasNextHop || nextHop == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), nextHop,
			"cannot be empty")
	}

	// Validate next hop as IP address (not CIDR)
	if strings.Contains(nextHop, "/") {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), nextHop,
			"must be a valid IPv4 address (not CIDR)")
	}

	// Basic IPv4 validation for next hop
	octets := strings.Split(nextHop, ".")
	if len(octets) != 4 {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), nextHop,
			"must be a valid IPv4 address with 4 octets")
	}

	for _, octet := range octets {
		var val int
		if _, err := fmt.Sscanf(octet, "%d", &val); err != nil || val < 0 || val > 255 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), nextHop,
				"must have octet values between 0-255")
		}
	}

	return nil
}

// ValidateBFDConfig validates a BFD configuration
func ValidateBFDConfig(bfd map[string]interface{}, ifaceIndex int) error {
	// TX interval - must be between 300-30000 ms
	if txInterval, ok := bfd["tx_interval"].(int); ok {
		if txInterval < 300 || txInterval > 30000 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD TX interval", ifaceIndex), txInterval,
				"must be between 300-30000 milliseconds")
		}
	} else if txInterval, ok := bfd["tx_interval"].(float64); ok {
		if txInterval < 300 || txInterval > 30000 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD TX interval", ifaceIndex), txInterval,
				"must be between 300-30000 milliseconds")
		}
	}

	// RX interval - must be between 300-30000 ms
	if rxInterval, ok := bfd["rx_interval"].(int); ok {
		if rxInterval < 300 || rxInterval > 30000 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD RX interval", ifaceIndex), rxInterval,
				"must be between 300-30000 milliseconds")
		}
	} else if rxInterval, ok := bfd["rx_interval"].(float64); ok {
		if rxInterval < 300 || rxInterval > 30000 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD RX interval", ifaceIndex), rxInterval,
				"must be between 300-30000 milliseconds")
		}
	}

	// Multiplier - must be between 3-20
	if multiplier, ok := bfd["multiplier"].(int); ok {
		if multiplier < 3 || multiplier > 20 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD multiplier", ifaceIndex), multiplier,
				"must be between 3-20")
		}
	} else if multiplier, ok := bfd["multiplier"].(float64); ok {
		if multiplier < 3 || multiplier > 20 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD multiplier", ifaceIndex), multiplier,
				"must be between 3-20")
		}
	}

	return nil
}

// ValidateBGPConnection validates a BGP connection configuration
func ValidateBGPConnection(conn map[string]interface{}, ifaceIndex, connIndex int) error {
	// Peer ASN - no validation, let API handle it
	_, hasPeerASN := conn["peer_asn"]
	if !hasPeerASN {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] peer ASN", ifaceIndex, connIndex),
			nil, "is required")
	}

	// Local ASN - no validation, let API handle it

	// Local IP address validation (required and must be a valid IP)
	localIP, hasLocalIP := conn["local_ip_address"].(string)
	if !hasLocalIP || localIP == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] local IP address", ifaceIndex, connIndex),
			localIP, "cannot be empty")
	}

	// Basic IPv4 validation for local IP
	if strings.Contains(localIP, "/") {
		// If CIDR format, validate as CIDR
		if err := ValidateIPv4CIDR(localIP); err != nil {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] local IP address", ifaceIndex, connIndex),
				localIP, "must be a valid IPv4 address or CIDR")
		}
	} else {
		// Non-CIDR format, validate as simple IP
		octets := strings.Split(localIP, ".")
		if len(octets) != 4 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] local IP address", ifaceIndex, connIndex),
				localIP, "must be a valid IPv4 address with 4 octets")
		}

		for _, octet := range octets {
			var val int
			if _, err := fmt.Sscanf(octet, "%d", &val); err != nil || val < 0 || val > 255 {
				return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] local IP address", ifaceIndex, connIndex),
					localIP, "must have octet values between 0-255")
			}
		}
	}

	// Peer IP address validation (required and must be a valid IP)
	peerIP, hasPeerIP := conn["peer_ip_address"].(string)
	if !hasPeerIP || peerIP == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] peer IP address", ifaceIndex, connIndex),
			peerIP, "cannot be empty")
	}

	// Basic IPv4 validation for peer IP
	if strings.Contains(peerIP, "/") {
		// If CIDR format, validate as CIDR
		if err := ValidateIPv4CIDR(peerIP); err != nil {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] peer IP address", ifaceIndex, connIndex),
				peerIP, "must be a valid IPv4 address or CIDR")
		}
	} else {
		// Non-CIDR format, validate as simple IP
		octets := strings.Split(peerIP, ".")
		if len(octets) != 4 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] peer IP address", ifaceIndex, connIndex),
				peerIP, "must be a valid IPv4 address with 4 octets")
		}

		for _, octet := range octets {
			var val int
			if _, err := fmt.Sscanf(octet, "%d", &val); err != nil || val < 0 || val > 255 {
				return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] peer IP address", ifaceIndex, connIndex),
					peerIP, "must have octet values between 0-255")
			}
		}
	}

	// Password validation removed - let API handle it

	// Validate peer type if provided - update to match Terraform schema
	if peerType, ok := conn["peer_type"].(string); ok && peerType != "" {
		validTypes := []string{"NON_CLOUD", "PRIV_CLOUD", "PUB_CLOUD"}
		isValid := false
		for _, vt := range validTypes {
			if peerType == vt {
				isValid = true
				break
			}
		}
		if !isValid {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] peer type", ifaceIndex, connIndex),
				peerType, "must be one of 'NON_CLOUD', 'PRIV_CLOUD', or 'PUB_CLOUD'")
		}
	}

	// Validate MED (Multi-Exit Discriminator) values
	if medIn, ok := conn["med_in"].(int); ok {
		if medIn < 0 || medIn > 4294967295 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] MED in", ifaceIndex, connIndex),
				medIn, "must be between 0-4294967295")
		}
	} else if medIn, ok := conn["med_in"].(float64); ok {
		if medIn < 0 || medIn > 4294967295 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] MED in", ifaceIndex, connIndex),
				medIn, "must be between 0-4294967295")
		}
	}

	if medOut, ok := conn["med_out"].(int); ok {
		if medOut < 0 || medOut > 4294967295 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] MED out", ifaceIndex, connIndex),
				medOut, "must be between 0-4294967295")
		}
	} else if medOut, ok := conn["med_out"].(float64); ok {
		if medOut < 0 || medOut > 4294967295 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] MED out", ifaceIndex, connIndex),
				medOut, "must be between 0-4294967295")
		}
	}

	// AS path prepend count validation
	if asPathPrependCount, ok := conn["as_path_prepend_count"].(int); ok {
		if asPathPrependCount < 0 || asPathPrependCount > 10 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] AS path prepend count", ifaceIndex, connIndex),
				asPathPrependCount, "must be between 0-10")
		}
	} else if asPathPrependCount, ok := conn["as_path_prepend_count"].(float64); ok {
		if asPathPrependCount < 0 || asPathPrependCount > 10 {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] AS path prepend count", ifaceIndex, connIndex),
				asPathPrependCount, "must be between 0-10")
		}
	}

	// Export policy validation
	if exportPolicy, ok := conn["export_policy"].(string); ok && exportPolicy != "" {
		if exportPolicy != "permit" && exportPolicy != "deny" {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BGP connection [%d] export policy", ifaceIndex, connIndex),
				exportPolicy, "must be 'permit' or 'deny'")
		}
	}

	return nil
}

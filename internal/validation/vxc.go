package validation

import (
	"fmt"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

const (
	// MinASPathPrependCount is the minimum allowed AS path prepend count in BGP configurations.
	MinASPathPrependCount = 0
	// MaxASPathPrependCount is the maximum allowed AS path prepend count in BGP configurations.
	MaxASPathPrependCount = 10
	// MinBFDInterval is the minimum Bidirectional Forwarding Detection interval in milliseconds.
	MinBFDInterval = 300
	// MaxBFDInterval is the maximum Bidirectional Forwarding Detection interval in milliseconds.
	MaxBFDInterval = 30000
	// MinBFDMultiplier is the minimum Bidirectional Forwarding Detection multiplier value.
	MinBFDMultiplier = 3
	// MaxBFDMultiplier is the maximum Bidirectional Forwarding Detection multiplier value.
	MaxBFDMultiplier = 20
	// MinMED is the minimum Multi-Exit Discriminator value for BGP routing.
	MinMED = 0
	// MaxMED is the maximum Multi-Exit Discriminator value for BGP routing.
	MaxMED = 4294967295
	// BGPPeerNonCloud identifies a non-cloud BGP peer type.
	BGPPeerNonCloud = "NON_CLOUD"
	// BGPPeerPrivCloud identifies a private cloud BGP peer type.
	BGPPeerPrivCloud = "PRIV_CLOUD"
	// BGPPeerPubCloud identifies a public cloud BGP peer type.
	BGPPeerPubCloud = "PUB_CLOUD"
	// BGPExportPolicyPermit defines the permit policy for BGP route exports.
	BGPExportPolicyPermit = "permit"
	// BGPExportPolicyDeny defines the deny policy for BGP route exports.
	BGPExportPolicyDeny = "deny"
	// MaxIBMNameLength is the maximum allowed length of an IBM connection name.
	MaxIBMNameLength = 100
	// IBMAccountIDLength is the required length of an IBM account ID.
	IBMAccountIDLength = 32
	// AWSConnectTypeAWS denotes a standard AWS connection type.
	AWSConnectTypeAWS = "AWS"
	// AWSConnectTypeAWSHC denotes a high-capacity AWS connection type.
	AWSConnectTypeAWSHC = "AWSHC"
	// AWSConnectTypeTransit denotes a transit AWS connection type.
	AWSConnectTypeTransit = "transit"
	// AWSConnectTypePrivate denotes a private AWS connection type.
	AWSConnectTypePrivate = "private"
	// AWSConnectTypePublic denotes a public AWS connection type.
	AWSConnectTypePublic = "public"
)

func ValidateVXCEndVLAN(vlan int) error {
	return ValidateVLAN(vlan)
}

func ValidateVXCEndInnerVLAN(vlan int) error {
	return ValidateVLAN(vlan)
}

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
	if bEndUID == "" && !hasPartnerConfig {
		return NewValidationError("B-End UID", bEndUID, "cannot be empty when no partner configuration is provided")
	}
	return nil
}

func ValidateVXCRequestFromConfig(config *megaport.VXCOrderConfiguration) error {
	hasPartnerConfig := false
	if config.BEnd.PartnerConfig != nil {
		hasPartnerConfig = true
	}
	return ValidateVXCRequest(
		config.Name,
		config.Term,
		config.RateLimit,
		config.AEnd.ProductUID,
		config.BEnd.ProductUID,
		hasPartnerConfig,
	)
}

func ValidateAWSPartnerConfig(config *megaport.VXCPartnerConfigAWS) error {
	if config.ConnectType == "" {
		return NewValidationError("AWS connect type", config.ConnectType, "cannot be empty")
	}
	validTypes := []string{"AWS", "AWSHC", "private", "public"}
	isValidType := false
	for _, t := range validTypes {
		if config.ConnectType == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return NewValidationError("AWS connect type", config.ConnectType, "must be 'AWS', 'AWSHC', 'private', or 'public'")
	}
	if config.OwnerAccount == "" {
		return NewValidationError("AWS owner account", config.OwnerAccount, "cannot be empty")
	}
	if config.CustomerIPAddress != "" {
		if err := ValidateCIDR(config.CustomerIPAddress, "AWS customer IP address"); err != nil {
			return err
		}
	}
	if config.AmazonIPAddress != "" {
		if err := ValidateCIDR(config.AmazonIPAddress, "AWS Amazon IP address"); err != nil {
			return err
		}
	}
	if config.ConnectionName != "" && len(config.ConnectionName) > 255 {
		return NewValidationError("AWS connection name", config.ConnectionName, "cannot exceed 255 characters")
	}
	if config.ConnectType == "AWS" && config.Type != "" && config.Type != "private" && config.Type != "public" {
		return NewValidationError("AWS type", config.Type, "must be 'private' or 'public' for AWS connect type")
	}
	return nil
}

func ValidateAzurePartnerConfig(config *megaport.VXCPartnerConfigAzure) error {
	if config == nil {
		return NewValidationError("Azure partner config", nil, "cannot be nil")
	}
	if config.ServiceKey == "" {
		return NewValidationError("Azure service key", config.ServiceKey, "cannot be empty")
	}
	return nil
}

func ValidateGooglePartnerConfig(config *megaport.VXCPartnerConfigGoogle) error {
	if config.PairingKey == "" {
		return NewValidationError("Google pairing key", config.PairingKey, "cannot be empty")
	}
	return nil
}

func ValidateOraclePartnerConfig(config *megaport.VXCPartnerConfigOracle) error {
	if config.VirtualCircuitId == "" {
		return NewValidationError("Oracle virtual circuit ID", config.VirtualCircuitId, "cannot be empty")
	}
	return nil
}

func ValidateIBMPartnerConfig(config *megaport.VXCPartnerConfigIBM) error {
	if config.AccountID == "" {
		return NewValidationError("IBM account ID", config.AccountID, "cannot be empty")
	}
	if len(config.AccountID) != IBMAccountIDLength {
		return NewValidationError("IBM account ID", config.AccountID, fmt.Sprintf("must be exactly %d characters", IBMAccountIDLength))
	}
	for _, c := range config.AccountID {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') && !(c >= 'A' && c <= 'F') {
			return NewValidationError("IBM account ID", config.AccountID, "must contain only hexadecimal characters (0-9, a-f, A-F)")
		}
	}
	if config.Name != "" && len(config.Name) > MaxIBMNameLength {
		return NewValidationError("IBM connection name", config.Name, fmt.Sprintf("cannot exceed %d characters", MaxIBMNameLength))
	}
	if config.Name != "" && !isValidIBMName(config.Name) {
		return NewValidationError("IBM connection name", config.Name, "must only contain characters 0-9, a-z, A-Z, /, -, _, or ,")
	}
	if config.CustomerIPAddress != "" {
		if err := ValidateCIDR(config.CustomerIPAddress, "IBM customer IP address"); err != nil {
			return err
		}
	}
	if config.ProviderIPAddress != "" {
		if err := ValidateCIDR(config.ProviderIPAddress, "IBM provider IP address"); err != nil {
			return err
		}
	}
	return nil
}

func isValidIBMName(name string) bool {
	for _, c := range name {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && c != '/' && c != '-' && c != '_' && c != ',' {
			return false
		}
	}
	return true
}

func ValidateVrouterPartnerConfig(config *megaport.VXCOrderVrouterPartnerConfig) error {
	if config == nil {
		return NewValidationError("vRouter partner config", nil, "cannot be nil")
	}
	if len(config.Interfaces) == 0 {
		return NewValidationError("vRouter interfaces", nil, "at least one interface must be provided")
	}
	for i, iface := range config.Interfaces {
		if iface.VLAN != 0 {
			if iface.VLAN < AutoAssignVLAN || iface.VLAN > MaxVLAN || iface.VLAN == ReservedVLAN {
				return NewValidationError(fmt.Sprintf("vRouter interface [%d] VLAN", i), iface.VLAN, fmt.Sprintf("must be between %d-%d (%d is reserved)", AutoAssignVLAN, MaxVLAN, ReservedVLAN))
			}
		}
		if len(iface.IpAddresses) > 0 {
			for j, ip := range iface.IpAddresses {
				if err := ValidateCIDR(ip, fmt.Sprintf("vRouter interface [%d] IP address [%d]", i, j)); err != nil {
					return err
				}
			}
		}
		if len(iface.NatIpAddresses) > 0 {
			for j, ip := range iface.NatIpAddresses {
				if err := ValidateCIDR(ip, fmt.Sprintf("vRouter interface [%d] NAT IP address [%d]", i, j)); err != nil {
					return err
				}
			}
		}
		if len(iface.IpRoutes) > 0 {
			for j, route := range iface.IpRoutes {
				if err := ValidateIPRouteConfig(route, i, j); err != nil {
					return err
				}
			}
		}
		if err := ValidateBFDConfig(iface.Bfd, i); err != nil {
			return err
		}
		if len(iface.BgpConnections) > 0 {
			for j, conn := range iface.BgpConnections {
				if err := ValidateBGPConnectionConfig(conn, i, j); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func ValidateVXCPartnerConfig(config megaport.VXCPartnerConfiguration) error {
	switch v := config.(type) {
	case *megaport.VXCPartnerConfigAWS:
		return ValidateAWSPartnerConfig(v)
	case *megaport.VXCPartnerConfigAzure:
		return ValidateAzurePartnerConfig(v)
	case *megaport.VXCPartnerConfigGoogle:
		return ValidateGooglePartnerConfig(v)
	case *megaport.VXCPartnerConfigOracle:
		return ValidateOraclePartnerConfig(v)
	case *megaport.VXCPartnerConfigIBM:
		return ValidateIBMPartnerConfig(v)
	case *megaport.VXCOrderVrouterPartnerConfig:
		return ValidateVrouterPartnerConfig(v)
	default:
		return NewValidationError("Partner configuration type", fmt.Sprintf("%T", v), "is not supported")
	}
}

func ValidateIPRouteConfig(route megaport.IpRoute, ifaceIndex, routeIndex int) error {
	if route.Prefix == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] prefix", ifaceIndex, routeIndex), route.Prefix, "cannot be empty")
	}
	if err := ValidateCIDR(route.Prefix, fmt.Sprintf("vRouter interface [%d] IP route [%d] prefix", ifaceIndex, routeIndex)); err != nil {
		return err
	}
	if route.NextHop == "" {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), route.NextHop, "cannot be empty")
	}
	if strings.Contains(route.NextHop, "/") {
		return NewValidationError(fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex), route.NextHop, "must be a valid IPv4 address (not CIDR)")
	}
	if err := ValidateIPv4(route.NextHop, fmt.Sprintf("vRouter interface [%d] IP route [%d] next hop", ifaceIndex, routeIndex)); err != nil {
		return err
	}
	return nil
}

func ValidateBFDConfig(bfd megaport.BfdConfig, ifaceIndex int) error {
	if bfd.TxInterval != 0 {
		if bfd.TxInterval < MinBFDInterval || bfd.TxInterval > MaxBFDInterval {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD TX interval", ifaceIndex), bfd.TxInterval, fmt.Sprintf("must be between %d-%d milliseconds", MinBFDInterval, MaxBFDInterval))
		}
	}
	if bfd.RxInterval != 0 {
		if bfd.RxInterval < MinBFDInterval || bfd.RxInterval > MaxBFDInterval {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD RX interval", ifaceIndex), bfd.RxInterval, fmt.Sprintf("must be between %d-%d milliseconds", MinBFDInterval, MaxBFDInterval))
		}
	}
	if bfd.Multiplier != 0 {
		if bfd.Multiplier < MinBFDMultiplier || bfd.Multiplier > MaxBFDMultiplier {
			return NewValidationError(fmt.Sprintf("vRouter interface [%d] BFD multiplier", ifaceIndex), bfd.Multiplier, fmt.Sprintf("must be between %d-%d", MinBFDMultiplier, MaxBFDMultiplier))
		}
	}
	return nil
}

func ValidateBGPConnectionConfig(conn megaport.BgpConnectionConfig, ifaceIndex, connIndex int) error {
	fieldPrefix := fmt.Sprintf("vRouter interface [%d] BGP connection [%d]", ifaceIndex, connIndex)
	if conn.PeerAsn == 0 {
		return NewValidationError(fmt.Sprintf("%s peer ASN", fieldPrefix), nil, "is required")
	}
	if conn.LocalIpAddress == "" {
		return NewValidationError(fmt.Sprintf("%s local IP address", fieldPrefix), conn.LocalIpAddress, "cannot be empty")
	}
	if strings.Contains(conn.LocalIpAddress, "/") {
		if err := ValidateCIDR(conn.LocalIpAddress, fmt.Sprintf("%s local IP address", fieldPrefix)); err != nil {
			return err
		}
	} else {
		if err := ValidateIPv4(conn.LocalIpAddress, fmt.Sprintf("%s local IP address", fieldPrefix)); err != nil {
			return err
		}
	}
	if conn.PeerIpAddress == "" {
		return NewValidationError(fmt.Sprintf("%s peer IP address", fieldPrefix), conn.PeerIpAddress, "cannot be empty")
	}
	if strings.Contains(conn.PeerIpAddress, "/") {
		if err := ValidateCIDR(conn.PeerIpAddress, fmt.Sprintf("%s peer IP address", fieldPrefix)); err != nil {
			return err
		}
	} else {
		if err := ValidateIPv4(conn.PeerIpAddress, fmt.Sprintf("%s peer IP address", fieldPrefix)); err != nil {
			return err
		}
	}
	if conn.PeerType != "" {
		validTypes := []string{BGPPeerNonCloud, BGPPeerPrivCloud, BGPPeerPubCloud}
		isValid := false
		for _, vt := range validTypes {
			if conn.PeerType == vt {
				isValid = true
				break
			}
		}
		if !isValid {
			return NewValidationError(fmt.Sprintf("%s peer type", fieldPrefix), conn.PeerType, fmt.Sprintf("must be one of '%s', '%s', or '%s'", BGPPeerNonCloud, BGPPeerPrivCloud, BGPPeerPubCloud))
		}
	}
	if conn.MedIn != 0 {
		if conn.MedIn < MinMED || conn.MedIn > MaxMED {
			return NewValidationError(fmt.Sprintf("%s MED in", fieldPrefix), conn.MedIn, fmt.Sprintf("must be between %d-%d", MinMED, MaxMED))
		}
	}
	if conn.MedOut != 0 {
		if conn.MedOut < MinMED || conn.MedOut > MaxMED {
			return NewValidationError(fmt.Sprintf("%s MED out", fieldPrefix), conn.MedOut, fmt.Sprintf("must be between %d-%d", MinMED, MaxMED))
		}
	}
	if conn.AsPathPrependCount != 0 {
		if conn.AsPathPrependCount < MinASPathPrependCount || conn.AsPathPrependCount > MaxASPathPrependCount {
			return NewValidationError(fmt.Sprintf("%s AS path prepend count", fieldPrefix), conn.AsPathPrependCount, fmt.Sprintf("must be between %d-%d", MinASPathPrependCount, MaxASPathPrependCount))
		}
	}
	if conn.ExportPolicy != "" {
		if conn.ExportPolicy != BGPExportPolicyPermit && conn.ExportPolicy != BGPExportPolicyDeny {
			return NewValidationError(fmt.Sprintf("%s export policy", fieldPrefix), conn.ExportPolicy, "must be 'permit' or 'deny'")
		}
	}
	return nil
}

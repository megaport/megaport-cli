package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	megaport "github.com/megaport/megaportgo"
)

var buyVXCFunc = func(ctx context.Context, client *megaport.Client, req *megaport.BuyVXCRequest) (*megaport.BuyVXCResponse, error) {
	return client.VXCService.BuyVXC(ctx, req)
}

// VXCOutput represents the desired fields for JSON output.
type VXCOutput struct {
	output
	UID     string `json:"uid"`
	Name    string `json:"name"`
	AEndUID string `json:"a_end_uid"`
	BEndUID string `json:"b_end_uid"`
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
func printVXCs(vxcs []*megaport.VXC, format string) error {
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
	return printOutput(outputs, format)
}

func promptPartnerConfig(end string) (megaport.VXCPartnerConfiguration, error) {
	partner, err := prompt(fmt.Sprintf("Enter %s partner (AWS, Azure, Google, Oracle, IBM, VRouter, Transit) (optional): ", end))
	if err != nil {
		return nil, err
	}
	if partner == "" {
		return nil, nil
	}

	switch strings.ToLower(partner) {
	case "aws":
		return promptAWSConfig()
	case "azure":
		return promptAzureConfig()
	case "google":
		return promptGoogleConfig()
	case "oracle":
		return promptOracleConfig()
	case "ibm":
		return promptIBMConfig()
	case "vrouter":
		return promptVRouterConfig()
	case "transit":
		return promptTransitConfig()
	default:
		return nil, fmt.Errorf("unsupported partner: %s", partner)
	}
}

// promptVRouterConfig prompts the user for VRouter-specific configuration details.
func promptVRouterConfig() (*megaport.VXCOrderVrouterPartnerConfig, error) {
	var interfaces []megaport.PartnerConfigInterface

	for {
		addInterface, err := prompt("Add an interface? (yes/no): ")
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addInterface) != "yes" {
			break
		}

		vlanStr, err := prompt("Enter VLAN (required): ")
		if err != nil {
			return nil, err
		}
		vlan, err := strconv.Atoi(vlanStr)
		if err != nil {
			return nil, fmt.Errorf("invalid VLAN")
		}

		ipAddresses, err := promptIPAddresses()
		if err != nil {
			return nil, err
		}

		ipRoutes, err := promptIPRoutes()
		if err != nil {
			return nil, err
		}

		natIpAddresses, err := promptNatIPAddresses()
		if err != nil {
			return nil, err
		}

		bfdConfig, err := promptBfdConfig()
		if err != nil {
			return nil, err
		}

		bgpConnections, err := promptBgpConnections()
		if err != nil {
			return nil, err
		}

		interfaces = append(interfaces, megaport.PartnerConfigInterface{
			VLAN:           vlan,
			IpAddresses:    ipAddresses,
			IpRoutes:       ipRoutes,
			NatIpAddresses: natIpAddresses,
			Bfd:            bfdConfig,
			BgpConnections: bgpConnections,
		})
	}

	return &megaport.VXCOrderVrouterPartnerConfig{
		Interfaces: interfaces,
	}, nil
}

// Helper to prompt for IP addresses
func promptIPAddresses() ([]string, error) {
	var ipAddresses []string
	for {
		ipAddress, err := prompt("Enter an IP address (or leave blank to finish): ")
		if err != nil {
			return nil, err
		}
		if ipAddress == "" {
			break
		}
		ipAddresses = append(ipAddresses, ipAddress)
	}
	return ipAddresses, nil
}

// Helper to prompt for IP routes
func promptIPRoutes() ([]megaport.IpRoute, error) {
	var ipRoutes []megaport.IpRoute
	for {
		prefix, err := prompt("Enter IP route prefix (or leave blank to finish): ")
		if err != nil {
			return nil, err
		}
		if prefix == "" {
			break
		}
		description, err := prompt("Enter IP route description: ")
		if err != nil {
			return nil, err
		}
		nextHop, err := prompt("Enter IP route next hop: ")
		if err != nil {
			return nil, err
		}
		ipRoutes = append(ipRoutes, megaport.IpRoute{
			Prefix:      prefix,
			Description: description,
			NextHop:     nextHop,
		})
	}
	return ipRoutes, nil
}

// Helper to prompt for NAT IP addresses
func promptNatIPAddresses() ([]string, error) {
	var natIpAddresses []string
	for {
		natIpAddress, err := prompt("Enter a NAT IP address (or leave blank to finish): ")
		if err != nil {
			return nil, err
		}
		if natIpAddress == "" {
			break
		}
		natIpAddresses = append(natIpAddresses, natIpAddress)
	}
	return natIpAddresses, nil
}

// Helper to prompt for BFD configuration
func promptBfdConfig() (megaport.BfdConfig, error) {
	bfdEnabledStr, err := prompt("Enable BFD? (true/false): ")
	if err != nil {
		return megaport.BfdConfig{}, err
	}
	bfdEnabled, err := strconv.ParseBool(bfdEnabledStr)
	if err != nil {
		bfdEnabled = false
	}
	if !bfdEnabled {
		return megaport.BfdConfig{}, nil
	}

	txIntervalStr, err := prompt("Enter BFD TxInterval: ")
	if err != nil {
		return megaport.BfdConfig{}, err
	}
	txInterval, err := strconv.Atoi(txIntervalStr)
	if err != nil {
		return megaport.BfdConfig{}, fmt.Errorf("invalid TxInterval")
	}

	rxIntervalStr, err := prompt("Enter BFD RxInterval: ")
	if err != nil {
		return megaport.BfdConfig{}, err
	}
	rxInterval, err := strconv.Atoi(rxIntervalStr)
	if err != nil {
		return megaport.BfdConfig{}, fmt.Errorf("invalid RxInterval")
	}

	multiplierStr, err := prompt("Enter BFD Multiplier: ")
	if err != nil {
		return megaport.BfdConfig{}, err
	}
	multiplier, err := strconv.Atoi(multiplierStr)
	if err != nil {
		return megaport.BfdConfig{}, fmt.Errorf("invalid Multiplier")
	}

	return megaport.BfdConfig{
		TxInterval: txInterval,
		RxInterval: rxInterval,
		Multiplier: multiplier,
	}, nil
}

// Helper to prompt for BGP connections
func promptBgpConnections() ([]megaport.BgpConnectionConfig, error) {
	var bgpConnections []megaport.BgpConnectionConfig
	for {
		addConnection, err := prompt("Add a BGP connection? (yes/no): ")
		if err != nil {
			return nil, err
		}
		if strings.ToLower(addConnection) != "yes" {
			break
		}

		peerAsnStr, err := prompt("Enter Peer ASN: ")
		if err != nil {
			return nil, err
		}
		peerAsn, err := strconv.Atoi(peerAsnStr)
		if err != nil {
			return nil, fmt.Errorf("invalid Peer ASN")
		}

		localAsnStr, err := prompt("Enter Local ASN (optional): ")
		if err != nil {
			return nil, err
		}
		var localAsn *int
		if localAsnStr != "" {
			localAsnVal, err := strconv.Atoi(localAsnStr)
			if err != nil {
				return nil, fmt.Errorf("invalid Local ASN")
			}
			localAsn = &localAsnVal
		}

		localIpAddress, err := prompt("Enter Local IP Address: ")
		if err != nil {
			return nil, err
		}

		peerIpAddress, err := prompt("Enter Peer IP Address: ")
		if err != nil {
			return nil, err
		}

		password, err := prompt("Enter Password (optional): ")
		if err != nil {
			return nil, err
		}

		shutdownStr, err := prompt("Shutdown BGP connection? (true/false): ")
		if err != nil {
			return nil, err
		}
		shutdown, err := strconv.ParseBool(shutdownStr)
		if err != nil {
			shutdown = false
		}

		description, err := prompt("Enter Description (optional): ")
		if err != nil {
			return nil, err
		}

		medInStr, err := prompt("Enter MED In (optional): ")
		if err != nil {
			return nil, err
		}
		medIn, err := strconv.Atoi(medInStr)
		if err != nil {
			medIn = 0
		}

		medOutStr, err := prompt("Enter MED Out (optional): ")
		if err != nil {
			return nil, err
		}
		medOut, err := strconv.Atoi(medOutStr)
		if err != nil {
			medOut = 0
		}

		bfdEnabledStr, err := prompt("Enable BFD? (true/false): ")
		if err != nil {
			return nil, err
		}
		bfdEnabled, err := strconv.ParseBool(bfdEnabledStr)
		if err != nil {
			bfdEnabled = false
		}

		exportPolicy, err := prompt("Enter Export Policy (optional): ")
		if err != nil {
			return nil, err
		}

		permitExportToStr, err := prompt("Enter Permit Export To (comma-separated, optional): ")
		if err != nil {
			return nil, err
		}
		permitExportTo := strings.Split(permitExportToStr, ",")

		denyExportToStr, err := prompt("Enter Deny Export To (comma-separated, optional): ")
		if err != nil {
			return nil, err
		}
		denyExportTo := strings.Split(denyExportToStr, ",")

		importWhitelistStr, err := prompt("Enter Import Whitelist (optional): ")
		if err != nil {
			return nil, err
		}
		importWhitelist, err := strconv.Atoi(importWhitelistStr)
		if err != nil {
			importWhitelist = 0
		}

		importBlacklistStr, err := prompt("Enter Import Blacklist (optional): ")
		if err != nil {
			return nil, err
		}
		importBlacklist, err := strconv.Atoi(importBlacklistStr)
		if err != nil {
			importBlacklist = 0
		}

		exportWhitelistStr, err := prompt("Enter Export Whitelist (optional): ")
		if err != nil {
			return nil, err
		}
		exportWhitelist, err := strconv.Atoi(exportWhitelistStr)
		if err != nil {
			exportWhitelist = 0
		}

		exportBlacklistStr, err := prompt("Enter Export Blacklist (optional): ")
		if err != nil {
			return nil, err
		}
		exportBlacklist, err := strconv.Atoi(exportBlacklistStr)
		if err != nil {
			exportBlacklist = 0
		}

		asPathPrependCountStr, err := prompt("Enter AS Path Prepend Count (optional): ")
		if err != nil {
			return nil, err
		}
		asPathPrependCount, err := strconv.Atoi(asPathPrependCountStr)
		if err != nil {
			asPathPrependCount = 0
		}

		peerType, err := prompt("Enter Peer Type (NON_CLOUD, PRIV_CLOUD, PUB_CLOUD, optional): ")
		if err != nil {
			return nil, err
		}

		bgpConnections = append(bgpConnections, megaport.BgpConnectionConfig{
			PeerAsn:            peerAsn,
			LocalAsn:           localAsn,
			LocalIpAddress:     localIpAddress,
			PeerIpAddress:      peerIpAddress,
			Password:           password,
			Shutdown:           shutdown,
			Description:        description,
			MedIn:              medIn,
			MedOut:             medOut,
			BfdEnabled:         bfdEnabled,
			ExportPolicy:       exportPolicy,
			PermitExportTo:     permitExportTo,
			DenyExportTo:       denyExportTo,
			ImportWhitelist:    importWhitelist,
			ImportBlacklist:    importBlacklist,
			ExportWhitelist:    exportWhitelist,
			ExportBlacklist:    exportBlacklist,
			AsPathPrependCount: asPathPrependCount,
			PeerType:           peerType,
		})
	}
	return bgpConnections, nil
}

func promptTransitConfig() (*megaport.VXCPartnerConfigTransit, error) {
	connectType, err := prompt("Enter connect type (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VXCPartnerConfigTransit{
		ConnectType: connectType,
	}, nil
}

func promptAWSConfig() (*megaport.VXCPartnerConfigAWS, error) {
	connectType, err := prompt("Enter connect type (required - either AWS or AWSHC): ")
	if err != nil {
		return nil, err
	}

	ownerAccount, err := prompt("Enter owner account ID (required): ")
	if err != nil {
		return nil, err
	}

	asnStr, err := prompt("Enter ASN (optional): ")
	if err != nil {
		return nil, err
	}
	asn, err := strconv.Atoi(asnStr)
	if err != nil {
		asn = 0
	}

	amazonASNStr, err := prompt("Enter Amazon ASN (optional): ")
	if err != nil {
		return nil, err
	}
	amazonASN, err := strconv.Atoi(amazonASNStr)
	if err != nil {
		amazonASN = 0
	}

	authKey, err := prompt("Enter auth key (optional): ")
	if err != nil {
		return nil, err
	}

	prefixes, err := prompt("Enter prefixes (optional): ")
	if err != nil {
		return nil, err
	}

	customerIPAddress, err := prompt("Enter customer IP address (optional): ")
	if err != nil {
		return nil, err
	}

	amazonIPAddress, err := prompt("Enter Amazon IP address (optional): ")
	if err != nil {
		return nil, err
	}

	connectionName, err := prompt("Enter connection name (optional): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VXCPartnerConfigAWS{
		ConnectType:       connectType,
		OwnerAccount:      ownerAccount,
		ASN:               asn,
		AmazonASN:         amazonASN,
		AuthKey:           authKey,
		Prefixes:          prefixes,
		CustomerIPAddress: customerIPAddress,
		AmazonIPAddress:   amazonIPAddress,
		ConnectionName:    connectionName,
	}, nil
}

// promptAzureConfig prompts the user for Azure-specific configuration details.
func promptAzureConfig() (*megaport.VXCPartnerConfigAzure, error) {
	serviceKey, err := prompt("Enter service key (required): ")
	if err != nil {
		return nil, err
	}

	var peers []megaport.PartnerOrderAzurePeeringConfig
	for {
		addPeer, err := prompt("Add a peering config? (yes/no): ")
		if err != nil {
			return nil, err
		}
		if addPeer != "yes" {
			break
		}

		peerConfig, err := promptAzurePeeringConfig()
		if err != nil {
			return nil, err
		}
		peers = append(peers, peerConfig)
	}

	return &megaport.VXCPartnerConfigAzure{
		ConnectType: "AZURE",
		ServiceKey:  serviceKey,
		Peers:       peers,
	}, nil
}

// Helper to prompt for Azure Peering Config
func promptAzurePeeringConfig() (megaport.PartnerOrderAzurePeeringConfig, error) {
	peeringType, err := prompt("Enter peering type (required): ")
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	peerASN, err := prompt("Enter Peer ASN (optional): ")
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	primarySubnet, err := prompt("Enter Primary Subnet (optional): ")
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	secondarySubnet, err := prompt("Enter Secondary Subnet (optional): ")
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	prefixes, err := prompt("Enter Prefixes (optional): ")
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	sharedKey, err := prompt("Enter Shared Key (optional): ")
	if err != nil {
		return megaport.PartnerOrderAzurePeeringConfig{}, err
	}

	vlanStr, err := prompt("Enter VLAN (optional): ")
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

func promptGoogleConfig() (*megaport.VXCPartnerConfigGoogle, error) {
	pairingKey, err := prompt("Enter pairing key (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VXCPartnerConfigGoogle{
		ConnectType: "GOOGLE",
		PairingKey:  pairingKey,
	}, nil
}

func promptOracleConfig() (*megaport.VXCPartnerConfigOracle, error) {
	virtualCircuitId, err := prompt("Enter virtual circuit ID (required): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VXCPartnerConfigOracle{
		ConnectType:      "ORACLE",
		VirtualCircuitId: virtualCircuitId,
	}, nil
}

func promptIBMConfig() (*megaport.VXCPartnerConfigIBM, error) {
	accountID, err := prompt("Enter account ID (required): ")
	if err != nil {
		return nil, err
	}

	customerASNStr, err := prompt("Enter customer ASN (optional): ")
	if err != nil {
		return nil, err
	}
	customerASN, err := strconv.Atoi(customerASNStr)
	if err != nil {
		customerASN = 0
	}

	customerIPAddress, err := prompt("Enter customer IP address (optional): ")
	if err != nil {
		return nil, err
	}

	providerIPAddress, err := prompt("Enter provider IP address (optional): ")
	if err != nil {
		return nil, err
	}

	name, err := prompt("Enter name (optional): ")
	if err != nil {
		return nil, err
	}

	return &megaport.VXCPartnerConfigIBM{
		ConnectType:       "IBM",
		AccountID:         accountID,
		CustomerASN:       customerASN,
		CustomerIPAddress: customerIPAddress,
		ProviderIPAddress: providerIPAddress,
		Name:              name,
	}, nil
}

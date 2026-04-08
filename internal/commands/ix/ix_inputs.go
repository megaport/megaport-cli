package ix

import (
	"encoding/json"
	"fmt"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func buildIXRequestFromFlags(cmd *cobra.Command) (*megaport.BuyIXRequest, error) { //nolint:unparam
	// Flag read errors are intentionally ignored — flags are registered by the command builder.
	productUID, _ := cmd.Flags().GetString("product-uid")
	name, _ := cmd.Flags().GetString("name")
	networkServiceType, _ := cmd.Flags().GetString("network-service-type")
	asn, _ := cmd.Flags().GetInt("asn")
	macAddress, _ := cmd.Flags().GetString("mac-address")
	rateLimit, _ := cmd.Flags().GetInt("rate-limit")
	vlan, _ := cmd.Flags().GetInt("vlan")
	shutdown, _ := cmd.Flags().GetBool("shutdown")
	promoCode, _ := cmd.Flags().GetString("promo-code")

	req := &megaport.BuyIXRequest{
		ProductUID:         productUID,
		Name:               name,
		NetworkServiceType: networkServiceType,
		ASN:                asn,
		MACAddress:         macAddress,
		RateLimit:          rateLimit,
		VLAN:               vlan,
		Shutdown:           shutdown,
		PromoCode:          promoCode,
	}

	return req, nil
}

func buildIXRequestFromJSON(jsonStr, jsonFile string) (*megaport.BuyIXRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	req := &megaport.BuyIXRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return req, nil
}

func buildUpdateIXRequestFromFlags(cmd *cobra.Command) (*megaport.UpdateIXRequest, error) { //nolint:unparam
	req := &megaport.UpdateIXRequest{}

	if cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		req.Name = &name
	}

	if cmd.Flags().Changed("rate-limit") {
		rateLimit, _ := cmd.Flags().GetInt("rate-limit")
		req.RateLimit = &rateLimit
	}

	if cmd.Flags().Changed("cost-centre") {
		costCentre, _ := cmd.Flags().GetString("cost-centre")
		req.CostCentre = &costCentre
	}

	if cmd.Flags().Changed("vlan") {
		vlan, _ := cmd.Flags().GetInt("vlan")
		req.VLAN = &vlan
	}

	if cmd.Flags().Changed("mac-address") {
		macAddress, _ := cmd.Flags().GetString("mac-address")
		req.MACAddress = &macAddress
	}

	if cmd.Flags().Changed("asn") {
		asn, _ := cmd.Flags().GetInt("asn")
		req.ASN = &asn
	}

	if cmd.Flags().Changed("password") {
		password, _ := cmd.Flags().GetString("password")
		req.Password = &password
	}

	if cmd.Flags().Changed("public-graph") {
		publicGraph, _ := cmd.Flags().GetBool("public-graph")
		req.PublicGraph = &publicGraph
	}

	if cmd.Flags().Changed("reverse-dns") {
		reverseDns, _ := cmd.Flags().GetString("reverse-dns")
		req.ReverseDns = &reverseDns
	}

	if cmd.Flags().Changed("a-end-product-uid") {
		aEndProductUID, _ := cmd.Flags().GetString("a-end-product-uid")
		req.AEndProductUid = &aEndProductUID
	}

	if cmd.Flags().Changed("shutdown") {
		shutdown, _ := cmd.Flags().GetBool("shutdown")
		req.Shutdown = &shutdown
	}

	return req, nil
}

func buildUpdateIXRequestFromJSON(jsonStr, jsonFile string) (*megaport.UpdateIXRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	req := &megaport.UpdateIXRequest{}
	if err := json.Unmarshal(jsonData, req); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return req, nil
}

package nat_gateway

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/megaport/megaport-cli/internal/validation"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

func processJSONCreateNATGatewayInput(jsonStr, jsonFile string) (*megaport.CreateNATGatewayRequest, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, err
	}

	// Unmarshal into a flat helper to handle camelCase JSON keys.
	var raw struct {
		Name                  string            `json:"name"`
		LocationID            int               `json:"locationId"`
		Speed                 int               `json:"speed"`
		Term                  int               `json:"term"`
		SessionCount          int               `json:"sessionCount"`
		DiversityZone         string            `json:"diversityZone"`
		ASN                   int               `json:"asn"`
		BGPShutdownDefault    bool              `json:"bgpShutdownDefault"`
		AutoRenewTerm         bool              `json:"autoRenewTerm"`
		PromoCode             string            `json:"promoCode"`
		ServiceLevelReference string            `json:"serviceLevelReference"`
		ResourceTags          map[string]string `json:"resourceTags"`
	}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	req := &megaport.CreateNATGatewayRequest{
		ProductName:           raw.Name,
		LocationID:            raw.LocationID,
		Speed:                 raw.Speed,
		Term:                  raw.Term,
		AutoRenewTerm:         raw.AutoRenewTerm,
		PromoCode:             raw.PromoCode,
		ServiceLevelReference: raw.ServiceLevelReference,
		Config: megaport.NATGatewayNetworkConfig{
			ASN:                raw.ASN,
			BGPShutdownDefault: raw.BGPShutdownDefault,
			DiversityZone:      raw.DiversityZone,
			SessionCount:       raw.SessionCount,
		},
	}

	if raw.ResourceTags != nil {
		tags := make([]megaport.ResourceTag, 0, len(raw.ResourceTags))
		for k, v := range raw.ResourceTags {
			tags = append(tags, megaport.ResourceTag{Key: k, Value: v})
		}
		req.ResourceTags = tags
	}

	if err := validation.ValidateCreateNATGatewayRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func processFlagCreateNATGatewayInput(cmd *cobra.Command) (*megaport.CreateNATGatewayRequest, error) {
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	speed, _ := cmd.Flags().GetInt("speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	sessionCount, _ := cmd.Flags().GetInt("session-count")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	serviceLevelRef, _ := cmd.Flags().GetString("service-level-reference")
	autoRenew, _ := cmd.Flags().GetBool("auto-renew")

	resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")
	resourceTagsFile, _ := cmd.Flags().GetString("resource-tags-file")
	var resourceTags []megaport.ResourceTag
	if resourceTagsStr != "" || resourceTagsFile != "" {
		tagData, err := utils.ReadJSONInput(resourceTagsStr, resourceTagsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read resource tags: %w", err)
		}
		var tagsMap map[string]string
		if err := json.Unmarshal(tagData, &tagsMap); err != nil {
			return nil, fmt.Errorf("failed to parse resource tags JSON: %w", err)
		}
		for k, v := range tagsMap {
			resourceTags = append(resourceTags, megaport.ResourceTag{Key: k, Value: v})
		}
	}

	req := &megaport.CreateNATGatewayRequest{
		ProductName:           name,
		LocationID:            locationID,
		Speed:                 speed,
		Term:                  term,
		AutoRenewTerm:         autoRenew,
		PromoCode:             promoCode,
		ServiceLevelReference: serviceLevelRef,
		ResourceTags:          resourceTags,
		Config: megaport.NATGatewayNetworkConfig{
			DiversityZone: diversityZone,
			SessionCount:  sessionCount,
		},
	}

	if err := validation.ValidateCreateNATGatewayRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

// updateExplicitFields tracks which zero-valued fields were explicitly provided
// in an update request. This lets mergeUpdateDefaults distinguish "user wants
// zero/false" from "user omitted the field and we should inherit from original".
type updateExplicitFields struct {
	AutoRenewTerm      bool // was autoRenewTerm present in input?
	BGPShutdownDefault bool // was bgpShutdownDefault present in input?
	ASN                bool // was asn present in JSON input? (no flag path for ASN)
}

// processJSONUpdateNATGatewayInput parses a JSON update request and returns the
// request along with explicit-presence flags for fields whose zero value is
// ambiguous. Absent fields use pointer types (*bool, *int) so nil means "not
// provided" and mergeUpdateDefaults will inherit from the original resource.
func processJSONUpdateNATGatewayInput(jsonStr, jsonFile, uid string) (*megaport.UpdateNATGatewayRequest, updateExplicitFields, error) {
	jsonData, err := utils.ReadJSONInput(jsonStr, jsonFile)
	if err != nil {
		return nil, updateExplicitFields{}, err
	}

	var raw struct {
		Name                  string            `json:"name"`
		LocationID            int               `json:"locationId"`
		Speed                 int               `json:"speed"`
		Term                  int               `json:"term"`
		SessionCount          int               `json:"sessionCount"`
		DiversityZone         string            `json:"diversityZone"`
		ASN                   *int              `json:"asn"`
		BGPShutdownDefault    *bool             `json:"bgpShutdownDefault"`
		AutoRenewTerm         *bool             `json:"autoRenewTerm"`
		PromoCode             string            `json:"promoCode"`
		ServiceLevelReference string            `json:"serviceLevelReference"`
		ResourceTags          map[string]string `json:"resourceTags"`
	}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, updateExplicitFields{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	explicit := updateExplicitFields{
		AutoRenewTerm:      raw.AutoRenewTerm != nil,
		BGPShutdownDefault: raw.BGPShutdownDefault != nil,
		ASN:                raw.ASN != nil,
	}

	var autoRenew, bgpShutdown bool
	if raw.AutoRenewTerm != nil {
		autoRenew = *raw.AutoRenewTerm
	}
	if raw.BGPShutdownDefault != nil {
		bgpShutdown = *raw.BGPShutdownDefault
	}
	var asn int
	if raw.ASN != nil {
		asn = *raw.ASN
	}

	req := &megaport.UpdateNATGatewayRequest{
		ProductUID:            uid,
		ProductName:           raw.Name,
		LocationID:            raw.LocationID,
		Speed:                 raw.Speed,
		Term:                  raw.Term,
		AutoRenewTerm:         autoRenew,
		PromoCode:             raw.PromoCode,
		ServiceLevelReference: raw.ServiceLevelReference,
		Config: megaport.NATGatewayNetworkConfig{
			ASN:                asn,
			BGPShutdownDefault: bgpShutdown,
			DiversityZone:      raw.DiversityZone,
			SessionCount:       raw.SessionCount,
		},
	}

	if raw.ResourceTags != nil {
		tags := make([]megaport.ResourceTag, 0, len(raw.ResourceTags))
		for k, v := range raw.ResourceTags {
			tags = append(tags, megaport.ResourceTag{Key: k, Value: v})
		}
		req.ResourceTags = tags
	}

	return req, explicit, nil
}

func processFlagUpdateNATGatewayInput(cmd *cobra.Command, uid string) (*megaport.UpdateNATGatewayRequest, error) {
	if uid == "" {
		return nil, fmt.Errorf("product UID is required")
	}
	name, _ := cmd.Flags().GetString("name")
	term, _ := cmd.Flags().GetInt("term")
	speed, _ := cmd.Flags().GetInt("speed")
	locationID, _ := cmd.Flags().GetInt("location-id")
	sessionCount, _ := cmd.Flags().GetInt("session-count")
	diversityZone, _ := cmd.Flags().GetString("diversity-zone")
	promoCode, _ := cmd.Flags().GetString("promo-code")
	serviceLevelRef, _ := cmd.Flags().GetString("service-level-reference")
	autoRenew, _ := cmd.Flags().GetBool("auto-renew")

	resourceTagsStr, _ := cmd.Flags().GetString("resource-tags")
	resourceTagsFile, _ := cmd.Flags().GetString("resource-tags-file")
	var resourceTags []megaport.ResourceTag
	if resourceTagsStr != "" || resourceTagsFile != "" {
		tagData, err := utils.ReadJSONInput(resourceTagsStr, resourceTagsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read resource tags: %w", err)
		}
		var tagsMap map[string]string
		if err := json.Unmarshal(tagData, &tagsMap); err != nil {
			return nil, fmt.Errorf("failed to parse resource tags JSON: %w", err)
		}
		for k, v := range tagsMap {
			resourceTags = append(resourceTags, megaport.ResourceTag{Key: k, Value: v})
		}
	}

	req := &megaport.UpdateNATGatewayRequest{
		ProductUID:            uid,
		ProductName:           name,
		LocationID:            locationID,
		Speed:                 speed,
		Term:                  term,
		AutoRenewTerm:         autoRenew,
		PromoCode:             promoCode,
		ServiceLevelReference: serviceLevelRef,
		ResourceTags:          resourceTags,
		Config: megaport.NATGatewayNetworkConfig{
			DiversityZone: diversityZone,
			SessionCount:  sessionCount,
		},
	}

	return req, nil
}

// parseTelemetryTypes splits a comma-separated types string into a slice.
func parseTelemetryTypes(typesStr string) []string {
	if typesStr == "" {
		return nil
	}
	parts := strings.Split(typesStr, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToUpper(p))
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

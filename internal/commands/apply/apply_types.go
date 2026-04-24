package apply

import "github.com/megaport/megaport-cli/internal/base/output"

// InfraConfig is the top-level structure for a megaport apply config file.
type InfraConfig struct {
	Ports []PortConfig `yaml:"ports" json:"ports"`
	MCRs  []MCRConfig  `yaml:"mcrs"  json:"mcrs"`
	MVEs  []MVEConfig  `yaml:"mves"  json:"mves"`
	VXCs  []VXCConfig  `yaml:"vxcs"  json:"vxcs"`
}

// PortConfig describes a port to provision.
type PortConfig struct {
	Name                  string            `yaml:"name" json:"name"`
	LocationID            int               `yaml:"location_id" json:"location_id"`
	Speed                 int               `yaml:"speed" json:"speed"`
	Term                  int               `yaml:"term" json:"term"`
	MarketplaceVisibility bool              `yaml:"marketplace_visibility" json:"marketplace_visibility"`
	DiversityZone         string            `yaml:"diversity_zone" json:"diversity_zone"`
	CostCentre            string            `yaml:"cost_centre" json:"cost_centre"`
	ResourceTags          map[string]string `yaml:"resource_tags" json:"resource_tags"`
}

// MCRConfig describes an MCR to provision.
type MCRConfig struct {
	Name          string            `yaml:"name" json:"name"`
	LocationID    int               `yaml:"location_id" json:"location_id"`
	Speed         int               `yaml:"speed" json:"speed"`
	Term          int               `yaml:"term" json:"term"`
	ASN           int               `yaml:"asn" json:"asn"`
	DiversityZone string            `yaml:"diversity_zone" json:"diversity_zone"`
	CostCentre    string            `yaml:"cost_centre" json:"cost_centre"`
	ResourceTags  map[string]string `yaml:"resource_tags" json:"resource_tags"`
}

// MVEConfig describes an MVE to provision.
// VendorConfig holds vendor-specific fields (e.g. vendor, imageId, productSize).
type MVEConfig struct {
	Name          string                 `yaml:"name" json:"name"`
	LocationID    int                    `yaml:"location_id" json:"location_id"`
	Term          int                    `yaml:"term" json:"term"`
	VendorConfig  map[string]interface{} `yaml:"vendor_config" json:"vendor_config"`
	DiversityZone string                 `yaml:"diversity_zone" json:"diversity_zone"`
	CostCentre    string                 `yaml:"cost_centre" json:"cost_centre"`
	ResourceTags  map[string]string      `yaml:"resource_tags" json:"resource_tags"`
}

// VXCEndpointConfig describes one end of a VXC connection.
type VXCEndpointConfig struct {
	ProductUID string `yaml:"product_uid" json:"product_uid"`
	VLAN       int    `yaml:"vlan" json:"vlan"`
}

// VXCConfig describes a VXC to provision.
type VXCConfig struct {
	Name         string            `yaml:"name" json:"name"`
	RateLimit    int               `yaml:"rate_limit" json:"rate_limit"`
	Term         int               `yaml:"term" json:"term"`
	AEnd         VXCEndpointConfig `yaml:"a_end" json:"a_end"`
	BEnd         VXCEndpointConfig `yaml:"b_end" json:"b_end"`
	CostCentre   string            `yaml:"cost_centre" json:"cost_centre"`
	ResourceTags map[string]string `yaml:"resource_tags" json:"resource_tags"`
}

// ApplyResult records the outcome of provisioning a single resource.
type ApplyResult struct {
	output.Output `json:"-" header:"-"`
	Type          string `json:"type"   header:"Type"`
	Name          string `json:"name"   header:"Name"`
	UID           string `json:"uid"    header:"UID"`
	Status        string `json:"status" header:"Status"`
}

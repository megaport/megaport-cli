package mcr

import (
	"context"
	"os"
	"testing"

	"github.com/megaport/megaport-cli/internal/commands/config"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessJSONMCRInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid JSON string",
			jsonStr: `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"name":"file-mcr","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
		},
		{
			name:          "empty tag key rejected",
			jsonStr:       `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"resourceTags":{"":"x"}}`,
			expectedError: "tag key must not be empty",
		},
		{
			name:          "empty tag key rejected via file",
			writeFile:     `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"resourceTags":{"":"x"}}`,
			expectedError: "tag key must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "mcr-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONMCRInput(tt.jsonStr, jsonFile)
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessJSONMCRInput_ValidResourceTags(t *testing.T) {
	req, err := processJSONMCRInput(`{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"resourceTags":{"env":"prod"}}`, "")
	require.NoError(t, err)
	require.NotNil(t, req)
	assert.Equal(t, "prod", req.ResourceTags["env"])
}

func TestProcessJSONUpdateMCRInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid update",
			jsonStr: `{"name":"updated-mcr"}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "no fields updated",
			jsonStr:       `{}`,
			expectedError: "at least one field must be updated",
		},
		{
			name:          "invalid contract term",
			jsonStr:       `{"contractTermMonths":99}`,
			expectedError: "Invalid contract term",
		},
		{
			name:          "empty name provided",
			jsonStr:       `{"name":""}`,
			expectedError: "name cannot be empty",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"name":"file-update"}`,
		},
		{
			name:    "mcrAsn only",
			jsonStr: `{"mcrAsn":65000}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "mcr-update-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONUpdateMCRInput(tt.jsonStr, jsonFile, "")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessJSONUpdateMCRInput_MCRAsnSet(t *testing.T) {
	req, err := processJSONUpdateMCRInput(`{"mcrAsn":65010}`, "", "")
	assert.NoError(t, err)
	require.NotNil(t, req)
	require.NotNil(t, req.MCRAsn, "MCRAsn should be set when mcrAsn is provided in JSON")
	assert.Equal(t, 65010, *req.MCRAsn)
}

func TestProcessJSONUpdateMCRInput_MCRAsnOutOfRangeRejected(t *testing.T) {
	_, err := processJSONUpdateMCRInput(`{"mcrAsn":0}`, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCR ASN")
}

func TestProcessJSONUpdateMCRInput_MCRAsnUnsetWhenKeyAbsent(t *testing.T) {
	req, err := processJSONUpdateMCRInput(`{"name":"only-name"}`, "", "")
	assert.NoError(t, err)
	require.NotNil(t, req)
	assert.Nil(t, req.MCRAsn, "MCRAsn should remain nil when mcrAsn key is absent")
}

func TestProcessJSONUpdateMCRInput_MCRAsnNullRejected(t *testing.T) {
	_, err := processJSONUpdateMCRInput(`{"mcrAsn":null}`, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid MCR ASN: null value")
}

func TestProcessFlagUpdateMCRInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name:          "no flags changed",
			flags:         map[string]string{},
			expectedError: "at least one field must be updated",
		},
		{
			name:  "name only",
			flags: map[string]string{"name": "new-name"},
		},
		{
			name:  "cost-centre only",
			flags: map[string]string{"cost-centre": "IT-2024"},
		},
		{
			name:  "marketplace-visibility only",
			flags: map[string]string{"marketplace-visibility": "true"},
		},
		{
			name:  "term only",
			flags: map[string]string{"term": "24"},
		},
		{
			name:          "invalid term",
			flags:         map[string]string{"term": "99"},
			expectedError: "Invalid contract term",
		},
		{
			name:          "empty name",
			flags:         map[string]string{"name": ""},
			expectedError: "name cannot be empty",
		},
		{
			name:  "multiple fields",
			flags: map[string]string{"name": "new", "cost-centre": "IT", "term": "12"},
		},
		{
			name:  "mcr-asn only",
			flags: map[string]string{"mcr-asn": "65000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("mcr-asn", 0, "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagUpdateMCRInput(cmd, "mcr-123", "")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "mcr-123", req.MCRID)
			}
		})
	}
}

func TestProcessFlagUpdateMCRInput_MCRAsnSet(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("mcr-asn", 0, "")

	require.NoError(t, cmd.Flags().Set("mcr-asn", "65001"))

	req, err := processFlagUpdateMCRInput(cmd, "mcr-123", "")
	assert.NoError(t, err)
	require.NotNil(t, req)
	require.NotNil(t, req.MCRAsn, "MCRAsn should be set when --mcr-asn flag is provided")
	assert.Equal(t, 65001, *req.MCRAsn)
}

func TestProcessFlagUpdateMCRInput_MCRAsnOutOfRangeRejected(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("mcr-asn", 0, "")

	require.NoError(t, cmd.Flags().Set("mcr-asn", "0"))

	_, err := processFlagUpdateMCRInput(cmd, "mcr-123", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCR ASN")
}

func TestProcessFlagUpdateMCRInput_MCRAsnUnsetWhenFlagAbsent(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().Int("term", 0, "")
	cmd.Flags().Int("mcr-asn", 0, "")

	require.NoError(t, cmd.Flags().Set("name", "stays"))

	req, err := processFlagUpdateMCRInput(cmd, "mcr-123", "")
	assert.NoError(t, err)
	require.NotNil(t, req)
	assert.Nil(t, req.MCRAsn, "MCRAsn should remain nil when --mcr-asn flag is not changed")
}

// The SDK sends costCentre without omitempty, so an empty value on update wipes
// it. These tests pin the preserve-unless-specified behavior for each input mode.
func TestProcessFlagUpdateMCRInput_CostCentrePreserved(t *testing.T) {
	newCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("name", "", "")
		cmd.Flags().String("cost-centre", "", "")
		cmd.Flags().Bool("marketplace-visibility", false, "")
		cmd.Flags().Int("term", 0, "")
		cmd.Flags().Int("mcr-asn", 0, "")
		return cmd
	}

	t.Run("name-only update preserves current cost centre", func(t *testing.T) {
		cmd := newCmd()
		require.NoError(t, cmd.Flags().Set("name", "new-name"))
		req, err := processFlagUpdateMCRInput(cmd, "mcr-123", "IT Dept")
		require.NoError(t, err)
		assert.Equal(t, "IT Dept", req.CostCentre)
	})

	t.Run("explicit cost centre overrides current", func(t *testing.T) {
		cmd := newCmd()
		require.NoError(t, cmd.Flags().Set("cost-centre", "Finance"))
		req, err := processFlagUpdateMCRInput(cmd, "mcr-123", "IT Dept")
		require.NoError(t, err)
		assert.Equal(t, "Finance", req.CostCentre)
	})

	t.Run("explicit empty cost centre clears it", func(t *testing.T) {
		cmd := newCmd()
		require.NoError(t, cmd.Flags().Set("name", "new-name"))
		require.NoError(t, cmd.Flags().Set("cost-centre", ""))
		req, err := processFlagUpdateMCRInput(cmd, "mcr-123", "IT Dept")
		require.NoError(t, err)
		assert.Equal(t, "", req.CostCentre)
	})
}

func TestProcessJSONUpdateMCRInput_CostCentrePreserved(t *testing.T) {
	t.Run("name-only update preserves current cost centre", func(t *testing.T) {
		req, err := processJSONUpdateMCRInput(`{"name":"new-name"}`, "", "IT Dept")
		require.NoError(t, err)
		assert.Equal(t, "IT Dept", req.CostCentre)
	})

	t.Run("explicit cost centre overrides current", func(t *testing.T) {
		req, err := processJSONUpdateMCRInput(`{"name":"new-name","costCentre":"Finance"}`, "", "IT Dept")
		require.NoError(t, err)
		assert.Equal(t, "Finance", req.CostCentre)
	})

	t.Run("explicit empty cost centre clears it", func(t *testing.T) {
		req, err := processJSONUpdateMCRInput(`{"name":"new-name","costCentre":""}`, "", "IT Dept")
		require.NoError(t, err)
		assert.Equal(t, "", req.CostCentre)
	})
}

func TestProcessJSONPrefixFilterListInput(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		jsonFile      string
		writeFile     string
		expectedError string
	}{
		{
			name:    "valid",
			jsonStr: `{"description":"PFL","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8"}]}`,
		},
		{
			name:          "invalid JSON",
			jsonStr:       `{invalid}`,
			expectedError: "failed to parse JSON",
		},
		{
			name:          "file not found",
			jsonFile:      "/nonexistent/path.json",
			expectedError: "failed to read JSON file",
		},
		{
			name:      "valid JSON file",
			writeFile: `{"description":"File PFL","addressFamily":"IPv4","entries":[{"action":"permit","prefix":"10.0.0.0/8"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonFile := tt.jsonFile
			if tt.writeFile != "" {
				tmpFile, err := os.CreateTemp("", "pfl-test-*.json")
				require.NoError(t, err)
				defer os.Remove(tmpFile.Name())
				_, err = tmpFile.WriteString(tt.writeFile)
				require.NoError(t, err)
				tmpFile.Close()
				jsonFile = tmpFile.Name()
			}

			req, err := processJSONPrefixFilterListInput(tt.jsonStr, jsonFile, "mcr-123")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "mcr-123", req.MCRID)
			}
		})
	}
}

func TestProcessFlagPrefixFilterListInput(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "valid",
			flags: map[string]string{
				"description":    "My PFL",
				"address-family": "IPv4",
				"entries":        `[{"action":"permit","prefix":"10.0.0.0/8"}]`,
			},
		},
		{
			name: "invalid entries JSON",
			flags: map[string]string{
				"description":    "My PFL",
				"address-family": "IPv4",
				"entries":        `{invalid}`,
			},
			expectedError: "failed to parse entries JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("address-family", "", "")
			cmd.Flags().String("entries", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagPrefixFilterListInput(cmd, "mcr-123")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

func TestProcessFlagUpdatePrefixFilterListInput(t *testing.T) {
	// Setup mock for getMCRPrefixFilterListFunc
	originalGetPFL := getMCRPrefixFilterListFunc
	originalLogin := config.GetLoginFunc()
	defer func() {
		getMCRPrefixFilterListFunc = originalGetPFL
		config.SetLoginFunc(originalLogin)
	}()

	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return &megaport.Client{}, nil
	})
	getMCRPrefixFilterListFunc = func(ctx context.Context, client *megaport.Client, mcrUID string, pflID int) (*megaport.MCRPrefixFilterList, error) {
		return &megaport.MCRPrefixFilterList{
			ID:            1,
			Description:   "Current PFL",
			AddressFamily: "IPv4",
			Entries: []*megaport.MCRPrefixListEntry{
				{Action: "permit", Prefix: "10.0.0.0/8"},
			},
		}, nil
	}

	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name:          "no flags changed",
			flags:         map[string]string{},
			expectedError: "at least one field",
		},
		{
			name:  "description only",
			flags: map[string]string{"description": "Updated PFL"},
		},
		{
			name:  "entries only",
			flags: map[string]string{"entries": `[{"action":"deny","prefix":"192.168.0.0/16"}]`},
		},
		{
			name:  "both description and entries",
			flags: map[string]string{"description": "New PFL", "entries": `[{"action":"permit","prefix":"172.16.0.0/12"}]`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("description", "", "")
			cmd.Flags().String("address-family", "", "")
			cmd.Flags().String("entries", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			pfl, err := processFlagUpdatePrefixFilterListInput(cmd, "mcr-123", 1)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pfl)
				assert.Equal(t, 1, pfl.ID)
			}
		})
	}
}

func TestProcessJSONMCRInput_WithTunnelCount(t *testing.T) {
	tests := []struct {
		name          string
		jsonStr       string
		expectAddOns  bool
		expectedCount int
		expectedError string
	}{
		{
			name:          "tunnelCount 10 populates AddOns",
			jsonStr:       `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"tunnelCount":10}`,
			expectAddOns:  true,
			expectedCount: 10,
		},
		{
			name:          "tunnelCount 0 includes add-on with API default",
			jsonStr:       `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"tunnelCount":0}`,
			expectAddOns:  true,
			expectedCount: 0,
		},
		{
			name:         "no tunnelCount field",
			jsonStr:      `{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true}`,
			expectAddOns: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := processJSONMCRInput(tt.jsonStr, "")
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, req)
			if tt.expectAddOns {
				assert.Len(t, req.AddOns, 1)
				addon, ok := req.AddOns[0].(*megaport.MCRAddOnIPsecConfig)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCount, addon.TunnelCount)
			} else {
				assert.Empty(t, req.AddOns)
			}
		})
	}
}

func TestProcessFlagMCRInput_IPSec(t *testing.T) {
	newCmd := func() *cobra.Command {
		cmd := &cobra.Command{Use: "test"}
		cmd.Flags().String("name", "test-mcr", "")
		cmd.Flags().Int("term", 12, "")
		cmd.Flags().Int("port-speed", 5000, "")
		cmd.Flags().Int("location-id", 1, "")
		cmd.Flags().Int("mcr-asn", 0, "")
		cmd.Flags().Bool("marketplace-visibility", false, "")
		cmd.Flags().String("cost-centre", "", "")
		cmd.Flags().String("promo-code", "", "")
		cmd.Flags().String("diversity-zone", "", "")
		cmd.Flags().String("resource-tags", "", "")
		cmd.Flags().Int("ipsec-tunnel-count", 0, "")
		return cmd
	}

	t.Run("ipsec-tunnel-count flag set", func(t *testing.T) {
		cmd := newCmd()
		require.NoError(t, cmd.Flags().Set("ipsec-tunnel-count", "20"))

		req, err := processFlagMCRInput(cmd)
		assert.NoError(t, err)
		assert.Len(t, req.AddOns, 1)
		addon, ok := req.AddOns[0].(*megaport.MCRAddOnIPsecConfig)
		assert.True(t, ok)
		assert.Equal(t, 20, addon.TunnelCount)
	})

	t.Run("ipsec-tunnel-count flag not set", func(t *testing.T) {
		cmd := newCmd()
		req, err := processFlagMCRInput(cmd)
		assert.NoError(t, err)
		assert.Empty(t, req.AddOns)
	})

	t.Run("ipsec-tunnel-count zero includes add-on with API default", func(t *testing.T) {
		cmd := newCmd()
		require.NoError(t, cmd.Flags().Set("ipsec-tunnel-count", "0"))
		req, err := processFlagMCRInput(cmd)
		assert.NoError(t, err)
		assert.Len(t, req.AddOns, 1)
		addon, ok := req.AddOns[0].(*megaport.MCRAddOnIPsecConfig)
		assert.True(t, ok)
		assert.Equal(t, 0, addon.TunnelCount)
	})

	t.Run("invalid ipsec-tunnel-count rejected", func(t *testing.T) {
		cmd := newCmd()
		require.NoError(t, cmd.Flags().Set("ipsec-tunnel-count", "5"))
		_, err := processFlagMCRInput(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid IPSec tunnel count")
	})
}

func TestProcessJSONMCRInput_InvalidTunnelCount(t *testing.T) {
	_, err := processJSONMCRInput(`{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"tunnelCount":5}`, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IPSec tunnel count")
}

func TestProcessJSONMCRInput_NegativeTunnelCount(t *testing.T) {
	_, err := processJSONMCRInput(`{"name":"test","term":12,"portSpeed":5000,"locationId":1,"marketplaceVisibility":true,"tunnelCount":-1}`, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tunnelCount must be")
}

func TestProcessFlagMCRInput_NegativeIPSecTunnelCount(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "test-mcr", "")
	cmd.Flags().Int("term", 12, "")
	cmd.Flags().Int("port-speed", 5000, "")
	cmd.Flags().Int("location-id", 1, "")
	cmd.Flags().Int("mcr-asn", 0, "")
	cmd.Flags().Bool("marketplace-visibility", false, "")
	cmd.Flags().String("cost-centre", "", "")
	cmd.Flags().String("promo-code", "", "")
	cmd.Flags().String("diversity-zone", "", "")
	cmd.Flags().String("resource-tags", "", "")
	cmd.Flags().Int("ipsec-tunnel-count", 0, "")
	require.NoError(t, cmd.Flags().Set("ipsec-tunnel-count", "-1"))

	_, err := processFlagMCRInput(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ipsec-tunnel-count must be")
}

func TestProcessFlagMCRInput_ResourceTags(t *testing.T) {
	tests := []struct {
		name          string
		flags         map[string]string
		expectedError string
	}{
		{
			name: "with valid resource tags",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
				"location-id":            "1",
				"marketplace-visibility": "true",
				"resource-tags":          `{"env":"prod"}`,
			},
		},
		{
			name: "with invalid resource tags JSON",
			flags: map[string]string{
				"name":                   "test-mcr",
				"term":                   "12",
				"port-speed":             "5000",
				"location-id":            "1",
				"marketplace-visibility": "true",
				"resource-tags":          `{invalid}`,
			},
			expectedError: "failed to parse resource tags JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("term", 0, "")
			cmd.Flags().Int("port-speed", 0, "")
			cmd.Flags().Int("location-id", 0, "")
			cmd.Flags().Int("mcr-asn", 0, "")
			cmd.Flags().Bool("marketplace-visibility", false, "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().String("promo-code", "", "")
			cmd.Flags().String("diversity-zone", "", "")
			cmd.Flags().String("resource-tags", "", "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := processFlagMCRInput(cmd)
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
			}
		})
	}
}

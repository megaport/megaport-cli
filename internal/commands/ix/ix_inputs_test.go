package ix

import (
	"os"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIXRequestFromJSON_ValidAllFields(t *testing.T) {
	jsonStr := `{"productUid":"port-123","productName":"test-ix","networkServiceType":"Los Angeles IX","asn":65000,"macAddress":"00:11:22:33:44:55","rateLimit":1000,"vlan":100,"shutdown":true,"promoCode":"PROMO1"}`
	req, err := buildIXRequestFromJSON(jsonStr, "")
	require.NoError(t, err)

	assert.Equal(t, "port-123", req.ProductUID)
	assert.Equal(t, "test-ix", req.Name)
	assert.Equal(t, "Los Angeles IX", req.NetworkServiceType)
	assert.Equal(t, 65000, req.ASN)
	assert.Equal(t, "00:11:22:33:44:55", req.MACAddress)
	assert.Equal(t, 1000, req.RateLimit)
	assert.Equal(t, 100, req.VLAN)
	assert.True(t, req.Shutdown)
	assert.Equal(t, "PROMO1", req.PromoCode)
}

func TestBuildIXRequestFromJSON_InvalidJSON(t *testing.T) {
	_, err := buildIXRequestFromJSON(`{invalid}`, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestBuildIXRequestFromJSON_EmptyJSON(t *testing.T) {
	req, err := buildIXRequestFromJSON(`{}`, "")
	require.NoError(t, err)
	assert.NotNil(t, req)
	assert.Equal(t, "", req.ProductUID)
	assert.Equal(t, "", req.Name)
	assert.Equal(t, 0, req.ASN)
}

func TestBuildIXRequestFromJSON_File(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ix-test-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := `{"productUid":"port-456","productName":"file-ix","networkServiceType":"Sydney IX","asn":65001,"macAddress":"AA:BB:CC:DD:EE:FF","rateLimit":500,"vlan":200}`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	req, err := buildIXRequestFromJSON("", tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, "port-456", req.ProductUID)
	assert.Equal(t, "file-ix", req.Name)
	assert.Equal(t, 65001, req.ASN)
}

func TestBuildIXRequestFromFlags_AllFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("product-uid", "", "")
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("network-service-type", "", "")
	cmd.Flags().Int("asn", 0, "")
	cmd.Flags().String("mac-address", "", "")
	cmd.Flags().Int("rate-limit", 0, "")
	cmd.Flags().Int("vlan", 0, "")
	cmd.Flags().Bool("shutdown", false, "")
	cmd.Flags().String("promo-code", "", "")

	require.NoError(t, cmd.Flags().Set("product-uid", "port-789"))
	require.NoError(t, cmd.Flags().Set("name", "flag-ix"))
	require.NoError(t, cmd.Flags().Set("network-service-type", "Sydney IX"))
	require.NoError(t, cmd.Flags().Set("asn", "65100"))
	require.NoError(t, cmd.Flags().Set("mac-address", "AA:BB:CC:DD:EE:FF"))
	require.NoError(t, cmd.Flags().Set("rate-limit", "2000"))
	require.NoError(t, cmd.Flags().Set("vlan", "300"))
	require.NoError(t, cmd.Flags().Set("shutdown", "true"))
	require.NoError(t, cmd.Flags().Set("promo-code", "DEAL"))

	req, err := buildIXRequestFromFlags(cmd)
	require.NoError(t, err)

	assert.Equal(t, "port-789", req.ProductUID)
	assert.Equal(t, "flag-ix", req.Name)
	assert.Equal(t, "Sydney IX", req.NetworkServiceType)
	assert.Equal(t, 65100, req.ASN)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", req.MACAddress)
	assert.Equal(t, 2000, req.RateLimit)
	assert.Equal(t, 300, req.VLAN)
	assert.True(t, req.Shutdown)
	assert.Equal(t, "DEAL", req.PromoCode)
}

func TestBuildUpdateIXRequestFromJSON_Valid(t *testing.T) {
	jsonStr := `{"name":"updated-ix","rateLimit":5000}`
	req, err := buildUpdateIXRequestFromJSON(jsonStr, "")
	require.NoError(t, err)
	assert.NotNil(t, req.Name)
	assert.Equal(t, "updated-ix", *req.Name)
	assert.NotNil(t, req.RateLimit)
	assert.Equal(t, 5000, *req.RateLimit)
}

func TestBuildUpdateIXRequestFromJSON_InvalidJSON(t *testing.T) {
	_, err := buildUpdateIXRequestFromJSON(`{invalid}`, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestBuildUpdateIXRequestFromFlags_ChangedProducePointers(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]string
		validate func(t *testing.T, req *megaport.UpdateIXRequest)
	}{
		{
			name:  "name only - unchanged fields are nil",
			flags: map[string]string{"name": "new-ix-name"},
			validate: func(t *testing.T, req *megaport.UpdateIXRequest) {
				assert.NotNil(t, req.Name)
				assert.Equal(t, "new-ix-name", *req.Name)
				assert.Nil(t, req.RateLimit)
				assert.Nil(t, req.CostCentre)
				assert.Nil(t, req.VLAN)
				assert.Nil(t, req.MACAddress)
				assert.Nil(t, req.ASN)
				assert.Nil(t, req.Password)
				assert.Nil(t, req.PublicGraph)
				assert.Nil(t, req.ReverseDns)
				assert.Nil(t, req.AEndProductUid)
				assert.Nil(t, req.Shutdown)
			},
		},
		{
			name:  "multiple flags changed",
			flags: map[string]string{"name": "new", "cost-centre": "IT", "vlan": "200", "asn": "65200"},
			validate: func(t *testing.T, req *megaport.UpdateIXRequest) {
				assert.NotNil(t, req.Name)
				assert.Equal(t, "new", *req.Name)
				assert.NotNil(t, req.CostCentre)
				assert.Equal(t, "IT", *req.CostCentre)
				assert.NotNil(t, req.VLAN)
				assert.Equal(t, 200, *req.VLAN)
				assert.NotNil(t, req.ASN)
				assert.Equal(t, 65200, *req.ASN)
				assert.Nil(t, req.RateLimit)
				assert.Nil(t, req.Password)
			},
		},
		{
			name:  "no flags changed - all nil",
			flags: map[string]string{},
			validate: func(t *testing.T, req *megaport.UpdateIXRequest) {
				assert.Nil(t, req.Name)
				assert.Nil(t, req.RateLimit)
				assert.Nil(t, req.CostCentre)
				assert.Nil(t, req.VLAN)
				assert.Nil(t, req.MACAddress)
				assert.Nil(t, req.ASN)
				assert.Nil(t, req.Password)
				assert.Nil(t, req.PublicGraph)
				assert.Nil(t, req.ReverseDns)
				assert.Nil(t, req.AEndProductUid)
				assert.Nil(t, req.Shutdown)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("name", "", "")
			cmd.Flags().Int("rate-limit", 0, "")
			cmd.Flags().String("cost-centre", "", "")
			cmd.Flags().Int("vlan", 0, "")
			cmd.Flags().String("mac-address", "", "")
			cmd.Flags().Int("asn", 0, "")
			cmd.Flags().String("password", "", "")
			cmd.Flags().Bool("public-graph", false, "")
			cmd.Flags().String("reverse-dns", "", "")
			cmd.Flags().String("a-end-product-uid", "", "")
			cmd.Flags().Bool("shutdown", false, "")

			for k, v := range tt.flags {
				require.NoError(t, cmd.Flags().Set(k, v))
			}

			req, err := buildUpdateIXRequestFromFlags(cmd)
			assert.NoError(t, err)
			assert.NotNil(t, req)
			tt.validate(t, req)
		})
	}
}

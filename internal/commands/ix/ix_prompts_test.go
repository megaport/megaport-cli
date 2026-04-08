package ix

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func mockPromptSequence(responses []string) func(string, string, bool) (string, error) {
	idx := 0
	return func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		val := responses[idx]
		idx++
		return val, nil
	}
}

func TestPromptBuildIXRequest_Success(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// productUID, name, networkServiceType, asn, macAddress, rateLimit, vlan, promoCode
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"port-123", "Test IX", "Los Angeles IX", "65000", "00:11:22:33:44:55", "1000", "100", "PROMO1",
	}))

	req, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.NoError(t, err)
	assert.Equal(t, "port-123", req.ProductUID)
	assert.Equal(t, "Test IX", req.Name)
	assert.Equal(t, "Los Angeles IX", req.NetworkServiceType)
	assert.Equal(t, 65000, req.ASN)
	assert.Equal(t, "00:11:22:33:44:55", req.MACAddress)
	assert.Equal(t, 1000, req.RateLimit)
	assert.Equal(t, 100, req.VLAN)
	assert.Equal(t, "PROMO1", req.PromoCode)
}

func TestPromptBuildIXRequest_EmptyProductUID(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "product UID is required")
}

func TestPromptBuildIXRequest_EmptyName(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"port-123", ""}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestPromptBuildIXRequest_EmptyNetworkServiceType(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"port-123", "Test IX", ""}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network service type is required")
}

func TestPromptBuildIXRequest_InvalidASN(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"port-123", "Test IX", "Los Angeles IX", "abc"}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ASN")
}

func TestPromptBuildIXRequest_EmptyMACAddress(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"port-123", "Test IX", "Los Angeles IX", "65000", ""}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MAC address is required")
}

func TestPromptBuildIXRequest_InvalidRateLimit(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"port-123", "Test IX", "Los Angeles IX", "65000", "00:11:22:33:44:55", "abc"}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rate limit")
}

func TestPromptBuildIXRequest_InvalidVLAN(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"port-123", "Test IX", "Los Angeles IX", "65000", "00:11:22:33:44:55", "1000", "abc"}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid VLAN")
}

func TestPromptBuildUpdateIXRequest_SuccessNameOnly(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	// name, rateLimit, costCentre, vlan, macAddress, asn, password, reverseDns
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"Updated IX", "", "", "", "", "", "", "",
	}))

	req, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.NoError(t, err)
	assert.NotNil(t, req.Name)
	assert.Equal(t, "Updated IX", *req.Name)
	assert.Nil(t, req.RateLimit)
	assert.Nil(t, req.CostCentre)
	assert.Nil(t, req.VLAN)
	assert.Nil(t, req.MACAddress)
	assert.Nil(t, req.ASN)
	assert.Nil(t, req.Password)
	assert.Nil(t, req.ReverseDns)
}

func TestPromptBuildUpdateIXRequest_NoFields(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "", "", "", "", "", "", "",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be updated")
}

func TestPromptBuildUpdateIXRequest_InvalidRateLimit(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "abc",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid rate limit")
}

func TestPromptBuildUpdateIXRequest_InvalidVLAN(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "", "", "abc",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid VLAN")
}

func TestPromptBuildUpdateIXRequest_InvalidASN(t *testing.T) {
	original := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(original) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "", "", "", "", "abc",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ASN")
}

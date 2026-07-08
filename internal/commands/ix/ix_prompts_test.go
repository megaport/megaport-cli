package ix

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func mockIXPromptSequence(responses []string) func(string, string, bool) (string, error) {
	idx := 0
	return func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		val := responses[idx]
		idx++
		if val == "ERROR" {
			return "", fmt.Errorf("simulated prompt error")
		}
		return val, nil
	}
}

func TestBuildIXRequestFromPrompt_ErrorOnNamePrompt(t *testing.T) {
	orig := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(orig)

	// First prompt (product UID) succeeds; second (name) errors
	utils.SetResourcePrompt(mockIXPromptSequence([]string{"port-uid-1", "ERROR"}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated prompt error")
}

func TestBuildIXRequestFromPrompt_ErrorOnNetworkServiceTypePrompt(t *testing.T) {
	orig := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(orig)

	utils.SetResourcePrompt(mockIXPromptSequence([]string{"port-uid-1", "My IX", "ERROR"}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated prompt error")
}

func TestBuildIXRequestFromPrompt_ErrorOnMACPrompt(t *testing.T) {
	orig := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(orig)

	// productUID, name, networkServiceType, asn, then ERROR on mac
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"port-uid-1", "My IX", "Los Angeles IX", "65000", "ERROR",
	}))

	_, err := buildIXRequestFromPrompt(context.Background(), true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_ErrorOnRateLimitPrompt(t *testing.T) {
	orig := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(orig)

	// name prompt returns "", then rate-limit prompt errors
	utils.SetResourcePrompt(mockIXPromptSequence([]string{"", "ERROR"}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_ErrorOnCostCentrePrompt(t *testing.T) {
	orig := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(orig)

	// name="", rate-limit="", cost-centre=ERROR
	utils.SetResourcePrompt(mockIXPromptSequence([]string{"", "", "ERROR"}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_ErrorOnPasswordPrompt(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", fmt.Errorf("simulated password prompt error") })
	// name="Updated", rate-limit="", cost-centre="", vlan="", mac="", asn="" skipped, then password errors
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"Updated", "", "", "", "", "",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_ErrorOnPublicGraphPrompt(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", nil })
	// name, rate-limit, cost-centre, vlan, mac, asn skipped, public-graph=ERROR
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"", "", "", "", "", "", "ERROR",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_ErrorOnAEndProductUIDPrompt(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", nil })
	// name, rate-limit, cost-centre, vlan, mac, asn, public-graph="", reverse-dns="" skipped, a-end-product-uid=ERROR
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"", "", "", "", "", "", "", "", "ERROR",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_ErrorOnShutdownPrompt(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", nil })
	// all prior prompts skipped, shutdown=ERROR
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"", "", "", "", "", "", "", "", "", "ERROR",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

func TestBuildUpdateIXRequestFromPrompt_InvalidPublicGraphValue(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", nil })
	// name, rate-limit, cost-centre, vlan, mac-address, asn skipped, public-graph="maybe"
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"", "", "", "", "", "", "maybe",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response for public-graph")
}

func TestBuildUpdateIXRequestFromPrompt_InvalidShutdownValue(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", nil })
	// name, rate-limit, cost-centre, vlan, mac-address, asn, public-graph, reverse-dns, a-end-product-uid skipped, shutdown="maybe"
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"", "", "", "", "", "", "", "", "", "maybe",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response for shutdown")
}

func TestBuildUpdateIXRequestFromPrompt_AEndProductUIDTrimmed(t *testing.T) {
	origResource := utils.GetResourcePrompt()
	origPassword := utils.GetPasswordPrompt()
	defer func() {
		utils.SetResourcePrompt(origResource)
		utils.SetPasswordPrompt(origPassword)
	}()

	utils.SetPasswordPrompt(func(_ string, _ bool) (string, error) { return "", nil })
	// name, rate-limit, cost-centre, vlan, mac-address, asn, public-graph, reverse-dns skipped, a-end-product-uid="  port-new-uid  ", shutdown skipped
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"", "", "", "", "", "", "", "", "  port-new-uid  ", "",
	}))

	req, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.NoError(t, err)
	assert.NotNil(t, req.AEndProductUid)
	assert.Equal(t, "port-new-uid", *req.AEndProductUid)
}

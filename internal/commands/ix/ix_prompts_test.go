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
	orig := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(orig)

	// name="Updated", rate-limit="", cost-centre="", vlan="", mac="", asn="", password=ERROR
	utils.SetResourcePrompt(mockIXPromptSequence([]string{
		"Updated", "", "", "", "", "", "ERROR",
	}))

	_, err := buildUpdateIXRequestFromPrompt("ix-123", true)
	assert.Error(t, err)
}

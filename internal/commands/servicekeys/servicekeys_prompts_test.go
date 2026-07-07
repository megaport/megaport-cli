package servicekeys

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

func mockServiceKeyPromptSequence(responses []string) func(string, string, bool) (string, error) {
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

func TestPromptForCreateServiceKeyDetails_ProductUIDPromptError(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(originalPrompt)

	utils.SetResourcePrompt(mockServiceKeyPromptSequence(nil))

	_, err := promptForCreateServiceKeyDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected prompt call")
}

func TestPromptForCreateServiceKeyDetails_InvalidProductID(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(originalPrompt)

	// productUID empty enters the product ID prompt, which fails to parse.
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{"", "not-a-number"}))

	_, err := promptForCreateServiceKeyDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid whole number")
}

func TestPromptForCreateServiceKeyDetails_ValidProductID(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	originalConfirm := utils.GetConfirmPrompt()
	defer func() {
		utils.SetResourcePrompt(originalPrompt)
		utils.SetConfirmPrompt(originalConfirm)
	}()

	// productUID empty, productID "77"; skip max speed/description/dates/vlan.
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{"", "77", "", "", "", "", ""}))
	utils.SetConfirmPrompt(func(_ string, _ bool) bool { return false })

	req, err := promptForCreateServiceKeyDetails(true)
	assert.NoError(t, err)
	if assert.NotNil(t, req) {
		assert.Empty(t, req.ProductUID)
		assert.Equal(t, 77, req.ProductID)
	}
}

func TestPromptForCreateServiceKeyDetails_InvalidMaxSpeed(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	originalConfirm := utils.GetConfirmPrompt()
	defer func() {
		utils.SetResourcePrompt(originalPrompt)
		utils.SetConfirmPrompt(originalConfirm)
	}()

	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{"prod-uid", "not-a-number"}))
	utils.SetConfirmPrompt(func(_ string, _ bool) bool { return false })

	_, err := promptForCreateServiceKeyDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid whole number")
}

func TestPromptForCreateServiceKeyDetails_InvalidDateRange(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	originalConfirm := utils.GetConfirmPrompt()
	defer func() {
		utils.SetResourcePrompt(originalPrompt)
		utils.SetConfirmPrompt(originalConfirm)
	}()

	// productUID, maxSpeed(skip), description(skip), startDate, endDate (end before start).
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{
		"prod-uid", "", "", "2025-12-31", "2025-01-01",
	}))
	utils.SetConfirmPrompt(func(_ string, _ bool) bool { return false })

	_, err := promptForCreateServiceKeyDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end date must be after start date")
}

func TestPromptForCreateServiceKeyDetails_InvalidVLAN(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	originalConfirm := utils.GetConfirmPrompt()
	defer func() {
		utils.SetResourcePrompt(originalPrompt)
		utils.SetConfirmPrompt(originalConfirm)
	}()

	// productUID, maxSpeed(skip), description(skip), startDate(skip), endDate(skip), vlan.
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{
		"prod-uid", "", "", "", "", "not-a-number",
	}))
	utils.SetConfirmPrompt(func(_ string, _ bool) bool { return false })

	_, err := promptForCreateServiceKeyDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid whole number")
}

func TestPromptForUpdateServiceKeyDetails_ProductUIDPromptError(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(originalPrompt)

	utils.SetResourcePrompt(mockServiceKeyPromptSequence(nil))

	current := &megaport.ServiceKey{ProductUID: "current-prod-uid"}
	_, err := promptForUpdateServiceKeyDetails("key-123", current, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected prompt call")
}

func TestPromptForUpdateServiceKeyDetails_InvalidProductID(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(originalPrompt)

	// productUID empty enters the product ID prompt, which fails to parse.
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{"", "not-a-number"}))

	current := &megaport.ServiceKey{ProductUID: "current-prod-uid"}
	_, err := promptForUpdateServiceKeyDetails("key-123", current, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid whole number")
}

func TestPromptForUpdateServiceKeyDetails_SingleUseAnsPromptError(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(originalPrompt)

	// productUID skip, single-use answer prompt fails.
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{"new-prod-uid"}))

	current := &megaport.ServiceKey{ProductUID: "current-prod-uid"}
	_, err := promptForUpdateServiceKeyDetails("key-123", current, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected prompt call")
}

func TestPromptForUpdateServiceKeyDetails_ActiveAnsPromptError(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer utils.SetResourcePrompt(originalPrompt)

	// productUID, single-use answer skip, active answer prompt fails.
	utils.SetResourcePrompt(mockServiceKeyPromptSequence([]string{"new-prod-uid", "no"}))

	current := &megaport.ServiceKey{ProductUID: "current-prod-uid"}
	_, err := promptForUpdateServiceKeyDetails("key-123", current, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected prompt call")
}

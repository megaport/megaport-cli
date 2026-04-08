package managed_account

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	"github.com/stretchr/testify/assert"
)

func mockPromptSequence(responses []string) func(string, string, bool) (string, error) {
	idx := 0
	return func(resourceType, msg string, noColor bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		val := responses[idx]
		idx++
		return val, nil
	}
}

func TestPromptBuildManagedAccountRequest_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"TestAccount", "REF-001"}))

	req, err := buildManagedAccountRequestFromPrompt(true)
	assert.NoError(t, err)
	assert.Equal(t, "TestAccount", req.AccountName)
	assert.Equal(t, "REF-001", req.AccountRef)
}

func TestPromptBuildManagedAccountRequest_EmptyName(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, err := buildManagedAccountRequestFromPrompt(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account name is required")
}

func TestPromptBuildManagedAccountRequest_EmptyRef(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"TestAccount", ""}))

	_, err := buildManagedAccountRequestFromPrompt(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "account reference is required")
}

func TestPromptBuildUpdateManagedAccountRequest_OneField(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"NewName", ""}))

	req, err := buildUpdateManagedAccountRequestFromPrompt(true)
	assert.NoError(t, err)
	assert.Equal(t, "NewName", req.AccountName)
	assert.Equal(t, "", req.AccountRef)
}

func TestPromptBuildUpdateManagedAccountRequest_BothFields(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"NewName", "NewRef"}))

	req, err := buildUpdateManagedAccountRequestFromPrompt(true)
	assert.NoError(t, err)
	assert.Equal(t, "NewName", req.AccountName)
	assert.Equal(t, "NewRef", req.AccountRef)
}

func TestPromptBuildUpdateManagedAccountRequest_NoFields(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"", ""}))

	_, err := buildUpdateManagedAccountRequestFromPrompt(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be updated")
}

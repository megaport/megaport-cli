package users

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

func TestPromptForCreateUserDetails_Success(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	// firstName, lastName, email, position, phone
	utils.ResourcePrompt = mockPromptSequence([]string{
		"John", "Doe", "john@example.com", "Technical Admin", "+61412345678",
	})

	req, err := promptForCreateUserDetails(true)
	assert.NoError(t, err)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
	assert.Equal(t, "john@example.com", req.Email)
	assert.Equal(t, "Technical Admin", string(req.Position))
	assert.Equal(t, "+61412345678", req.Phone)
	assert.True(t, req.Active)
}

func TestPromptForCreateUserDetails_EmptyFirstName(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	utils.ResourcePrompt = mockPromptSequence([]string{""})

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "first name is required")
}

func TestPromptForCreateUserDetails_EmptyLastName(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	utils.ResourcePrompt = mockPromptSequence([]string{"John", ""})

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "last name is required")
}

func TestPromptForCreateUserDetails_EmptyEmail(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	utils.ResourcePrompt = mockPromptSequence([]string{"John", "Doe", ""})

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestPromptForCreateUserDetails_EmptyPosition(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	utils.ResourcePrompt = mockPromptSequence([]string{"John", "Doe", "john@example.com", ""})

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "position is required")
}

func TestPromptForCreateUserDetails_InvalidPosition(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	utils.ResourcePrompt = mockPromptSequence([]string{"John", "Doe", "john@example.com", "Super Admin"})

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid position")
}

func TestPromptForUpdateUserDetails_Success(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	// firstName, lastName, email, position, phone
	utils.ResourcePrompt = mockPromptSequence([]string{
		"Jane", "Smith", "jane@example.com", "Finance", "+61400000000",
	})

	req, err := promptForUpdateUserDetails(true)
	assert.NoError(t, err)
	assert.NotNil(t, req.FirstName)
	assert.Equal(t, "Jane", *req.FirstName)
	assert.NotNil(t, req.LastName)
	assert.Equal(t, "Smith", *req.LastName)
	assert.NotNil(t, req.Email)
	assert.Equal(t, "jane@example.com", *req.Email)
	assert.NotNil(t, req.Position)
	assert.Equal(t, "Finance", *req.Position)
	assert.NotNil(t, req.Phone)
	assert.Equal(t, "+61400000000", *req.Phone)
}

func TestPromptForUpdateUserDetails_PartialUpdate(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	// Only update firstName, skip the rest
	utils.ResourcePrompt = mockPromptSequence([]string{
		"NewFirst", "", "", "", "",
	})

	req, err := promptForUpdateUserDetails(true)
	assert.NoError(t, err)
	assert.NotNil(t, req.FirstName)
	assert.Equal(t, "NewFirst", *req.FirstName)
	assert.Nil(t, req.LastName)
	assert.Nil(t, req.Email)
	assert.Nil(t, req.Position)
	assert.Nil(t, req.Phone)
}

func TestPromptForUpdateUserDetails_NoChanges(t *testing.T) {
	originalPrompt := utils.ResourcePrompt
	defer func() { utils.ResourcePrompt = originalPrompt }()

	utils.ResourcePrompt = mockPromptSequence([]string{"", "", "", "", ""})

	_, err := promptForUpdateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be updated")
}

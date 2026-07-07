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
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// firstName, lastName, email, position, phone
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"John", "Doe", "john@example.com", "Technical Admin", "+61412345678",
	}))

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
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{""}))

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "first name is required")
}

func TestPromptForCreateUserDetails_EmptyLastName(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"John", ""}))

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "last name is required")
}

func TestPromptForCreateUserDetails_EmptyEmail(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"John", "Doe", ""}))

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestPromptForCreateUserDetails_EmptyPosition(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"John", "Doe", "john@example.com", ""}))

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "position is required")
}

func TestPromptForCreateUserDetails_InvalidPosition(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"John", "Doe", "john@example.com", "Super Admin"}))

	_, err := promptForCreateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid position")
}

func TestPromptForUpdateUserDetails_Success(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// firstName, lastName, email, position, phone, active, notification-enabled
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"Jane", "Smith", "jane@example.com", "Finance", "+61400000000", "yes", "no",
	}))

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
	assert.NotNil(t, req.Active)
	assert.True(t, *req.Active)
	assert.NotNil(t, req.NotificationEnabled)
	assert.False(t, *req.NotificationEnabled)
}

func TestPromptForUpdateUserDetails_PartialUpdate(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// Only update firstName, skip the rest
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"NewFirst", "", "", "", "", "", "",
	}))

	req, err := promptForUpdateUserDetails(true)
	assert.NoError(t, err)
	assert.NotNil(t, req.FirstName)
	assert.Equal(t, "NewFirst", *req.FirstName)
	assert.Nil(t, req.LastName)
	assert.Nil(t, req.Email)
	assert.Nil(t, req.Position)
	assert.Nil(t, req.Phone)
	assert.Nil(t, req.Active)
	assert.Nil(t, req.NotificationEnabled)
}

func TestPromptForUpdateUserDetails_ActiveAndNotificationOnly(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// Skip everything except active and notification-enabled
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "", "", "", "", "no", "yes",
	}))

	req, err := promptForUpdateUserDetails(true)
	assert.NoError(t, err)
	assert.Nil(t, req.FirstName)
	assert.Nil(t, req.LastName)
	assert.Nil(t, req.Email)
	assert.Nil(t, req.Position)
	assert.Nil(t, req.Phone)
	assert.NotNil(t, req.Active)
	assert.False(t, *req.Active)
	assert.NotNil(t, req.NotificationEnabled)
	assert.True(t, *req.NotificationEnabled)
}

func TestPromptForUpdateUserDetails_ErrorOnActivePrompt(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	idx := 0
	responses := []string{"", "", "", "", ""}
	utils.SetResourcePrompt(func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("simulated prompt error")
		}
		val := responses[idx]
		idx++
		return val, nil
	})

	_, err := promptForUpdateUserDetails(true)
	assert.Error(t, err)
}

func TestPromptForUpdateUserDetails_ErrorOnNotificationEnabledPrompt(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	idx := 0
	responses := []string{"", "", "", "", "", ""}
	utils.SetResourcePrompt(func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("simulated prompt error")
		}
		val := responses[idx]
		idx++
		return val, nil
	})

	_, err := promptForUpdateUserDetails(true)
	assert.Error(t, err)
}

func TestPromptForUpdateUserDetails_InvalidActiveValue(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// firstName, lastName, email, position, phone skipped, active="maybe"
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "", "", "", "", "maybe",
	}))

	_, err := promptForUpdateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response for active")
}

func TestPromptForUpdateUserDetails_InvalidNotificationEnabledValue(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	// firstName, lastName, email, position, phone, active skipped, notification-enabled="maybe"
	utils.SetResourcePrompt(mockPromptSequence([]string{
		"", "", "", "", "", "", "maybe",
	}))

	_, err := promptForUpdateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid response for notification-enabled")
}

func TestPromptForUpdateUserDetails_NoChanges(t *testing.T) {
	originalPrompt := utils.GetResourcePrompt()
	defer func() { utils.SetResourcePrompt(originalPrompt) }()

	utils.SetResourcePrompt(mockPromptSequence([]string{"", "", "", "", "", "", ""}))

	_, err := promptForUpdateUserDetails(true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one field must be updated")
}

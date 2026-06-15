package managed_account

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/utils"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

// setManagedAccountPrompts installs a prompt stub that replays responses in
// order. A response of "ERROR" makes that prompt return an error.
func setManagedAccountPrompts(t *testing.T, responses []string) {
	t.Helper()
	original := utils.GetResourcePrompt()
	t.Cleanup(func() { utils.SetResourcePrompt(original) })

	idx := 0
	utils.SetResourcePrompt(func(_, _ string, _ bool) (string, error) {
		if idx >= len(responses) {
			return "", fmt.Errorf("unexpected prompt call #%d", idx)
		}
		resp := responses[idx]
		idx++
		if resp == "ERROR" {
			return "", fmt.Errorf("prompt failed")
		}
		return resp, nil
	})
}

func TestBuildManagedAccountRequestFromPrompt(t *testing.T) {
	tests := []struct {
		name          string
		prompts       []string
		expectedError string
		validate      func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name:    "all prompts answered successfully",
			prompts: []string{"Test Account", "REF-001"},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Test Account", req.AccountName)
				assert.Equal(t, "REF-001", req.AccountRef)
			},
		},
		{
			name:          "empty account name",
			prompts:       []string{""},
			expectedError: "account name is required",
		},
		{
			name:          "empty account ref",
			prompts:       []string{"Test Account", ""},
			expectedError: "account reference is required",
		},
		{
			name:          "error on first prompt (name)",
			prompts:       []string{"ERROR"},
			expectedError: "prompt failed",
		},
		{
			name:          "error on second prompt (ref)",
			prompts:       []string{"Test Account", "ERROR"},
			expectedError: "prompt failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setManagedAccountPrompts(t, tt.prompts)

			req, err := buildManagedAccountRequestFromPrompt(true)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				tt.validate(t, req)
			}
		})
	}
}

func TestBuildUpdateManagedAccountRequestFromPrompt(t *testing.T) {
	tests := []struct {
		name          string
		prompts       []string
		expectedError string
		validate      func(t *testing.T, req *megaport.ManagedAccountRequest)
	}{
		{
			name:    "update name only",
			prompts: []string{"Updated Account", ""},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "", req.AccountRef)
			},
		},
		{
			name:    "update ref only",
			prompts: []string{"", "NEW-REF"},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name:    "update both fields",
			prompts: []string{"Updated Account", "NEW-REF"},
			validate: func(t *testing.T, req *megaport.ManagedAccountRequest) {
				assert.Equal(t, "Updated Account", req.AccountName)
				assert.Equal(t, "NEW-REF", req.AccountRef)
			},
		},
		{
			name:          "no fields updated",
			prompts:       []string{"", ""},
			expectedError: "at least one field must be updated",
		},
		{
			name:          "error on first prompt (name)",
			prompts:       []string{"ERROR"},
			expectedError: "prompt failed",
		},
		{
			name:          "error on second prompt (ref)",
			prompts:       []string{"Updated Account", "ERROR"},
			expectedError: "prompt failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setManagedAccountPrompts(t, tt.prompts)

			req, err := buildUpdateManagedAccountRequestFromPrompt(true)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				tt.validate(t, req)
			}
		})
	}
}

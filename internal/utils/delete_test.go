package utils

import (
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/exitcodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirmDelete(t *testing.T) {
	originalFn := GetConfirmPrompt()
	defer SetConfirmPrompt(originalFn)

	tests := []struct {
		name          string
		resourceType  ResourceType
		resourceID    string
		force         bool
		noColor       bool
		confirmResult bool
		wantConfirmed bool
		wantErr       bool
		wantExitCode  int
	}{
		{
			name:          "force skips prompt",
			resourceType:  ResourceTypePort,
			resourceID:    "abc-123",
			force:         true,
			noColor:       true,
			wantConfirmed: true,
		},
		{
			name:          "user confirms deletion",
			resourceType:  ResourceTypeVXC,
			resourceID:    "vxc-456",
			force:         false,
			noColor:       true,
			confirmResult: true,
			wantConfirmed: true,
		},
		{
			name:          "user declines deletion",
			resourceType:  ResourceTypeMCR,
			resourceID:    "mcr-789",
			force:         false,
			noColor:       true,
			confirmResult: false,
			wantConfirmed: false,
			wantErr:       true,
			wantExitCode:  exitcodes.Cancelled,
		},
		{
			name:          "noColor forwarded to prompt",
			resourceType:  ResourceTypeMVE,
			resourceID:    "mve-012",
			force:         false,
			noColor:       false,
			confirmResult: true,
			wantConfirmed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			promptCalled := false
			expectedMsg := fmt.Sprintf("Are you sure you want to delete %s %s?", tt.resourceType, tt.resourceID)
			SetConfirmPrompt(func(message string, noColor bool) bool {
				promptCalled = true
				assert.Equal(t, expectedMsg, message)
				assert.Equal(t, tt.noColor, noColor, "noColor should be forwarded to ConfirmPrompt")
				return tt.confirmResult
			})

			confirmed, err := ConfirmDelete(tt.resourceType, tt.resourceID, tt.force, tt.noColor)

			assert.Equal(t, tt.wantConfirmed, confirmed)

			if tt.force {
				assert.False(t, promptCalled, "prompt should not be called when force=true")
			} else {
				assert.True(t, promptCalled, "prompt should be called when force=false")
			}

			if tt.wantErr {
				require.Error(t, err)
				var cliErr *exitcodes.CLIError
				require.ErrorAs(t, err, &cliErr)
				assert.Equal(t, tt.wantExitCode, cliErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

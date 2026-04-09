package mcr

import (
	"context"
	"fmt"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	"github.com/megaport/megaport-cli/internal/commands/config"
	"github.com/megaport/megaport-cli/internal/testutil"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAddMCRIPSecAddOn_Flags(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name           string
		mcrUID         string
		tunnelCount    int
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
		expectCalled   bool
		expectCount    int
	}{
		{
			name:           "add IPSec addon tunnel-count 10",
			mcrUID:         "mcr-abc",
			tunnelCount:    10,
			setupMock:      func(m *MockMCRService) {},
			expectCalled:   true,
			expectCount:    10,
			expectedOutput: "IPSec add-on added successfully",
		},
		{
			name:           "add IPSec addon tunnel-count 20",
			mcrUID:         "mcr-abc",
			tunnelCount:    20,
			setupMock:      func(m *MockMCRService) {},
			expectCalled:   true,
			expectCount:    20,
			expectedOutput: "IPSec add-on added successfully",
		},
		{
			name:        "service returns error",
			mcrUID:      "mcr-abc",
			tunnelCount: 10,
			setupMock: func(m *MockMCRService) {
				m.UpdateMCRWithAddOnErr = fmt.Errorf("service unavailable")
			},
			expectedError: "service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockMCRService{}
			tt.setupMock(mockSvc)

			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockSvc
				return client, nil
			})

			cmd := &cobra.Command{Use: "add-ipsec-addon [mcrUID]"}
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().Int("tunnel-count", 0, "")
			_ = cmd.Flags().Set("tunnel-count", fmt.Sprintf("%d", tt.tunnelCount))

			var err error
			out := output.CaptureOutput(func() {
				err = AddMCRIPSecAddOn(cmd, []string{tt.mcrUID}, false)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, out, tt.expectedOutput)
				if tt.expectCalled {
					assert.Equal(t, tt.mcrUID, mockSvc.CapturedUpdateMCRWithAddOnMCRID)
					addon, ok := mockSvc.CapturedUpdateMCRWithAddOnReq.AddOn.(*megaport.MCRAddOnIPsecConfig)
					assert.True(t, ok)
					assert.Equal(t, tt.expectCount, addon.TunnelCount)
					assert.Equal(t, megaport.AddOnTypeIPsec, addon.AddOnType)
				}
			}
		})
	}
}

func TestAddMCRIPSecAddOn_JSON(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	mockSvc := &MockMCRService{}
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.MCRService = mockSvc
		return client, nil
	})

	cmd := &cobra.Command{Use: "add-ipsec-addon [mcrUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", `{"tunnelCount":20}`, "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")

	var err error
	out := output.CaptureOutput(func() {
		err = AddMCRIPSecAddOn(cmd, []string{"mcr-json"}, false)
	})

	assert.NoError(t, err)
	assert.Contains(t, out, "IPSec add-on added successfully")
	assert.Equal(t, "mcr-json", mockSvc.CapturedUpdateMCRWithAddOnMCRID)
	addon, ok := mockSvc.CapturedUpdateMCRWithAddOnReq.AddOn.(*megaport.MCRAddOnIPsecConfig)
	assert.True(t, ok)
	assert.Equal(t, 20, addon.TunnelCount)
}

func TestAddMCRIPSecAddOn_NoInput(t *testing.T) {
	cmd := &cobra.Command{Use: "add-ipsec-addon [mcrUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")

	err := AddMCRIPSecAddOn(cmd, []string{"mcr-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}

func TestUpdateMCRIPSecAddOn_Flags(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	tests := []struct {
		name           string
		mcrUID         string
		addOnUID       string
		tunnelCount    int
		setupMock      func(*MockMCRService)
		expectedError  string
		expectedOutput string
		expectCount    int
	}{
		{
			name:           "update tunnel count to 30",
			mcrUID:         "mcr-abc",
			addOnUID:       "addon-123",
			tunnelCount:    30,
			setupMock:      func(m *MockMCRService) {},
			expectedOutput: "IPSec add-on updated successfully - tunnel count: 30",
			expectCount:    30,
		},
		{
			name:           "disable IPSec with tunnel count 0",
			mcrUID:         "mcr-abc",
			addOnUID:       "addon-123",
			tunnelCount:    0,
			setupMock:      func(m *MockMCRService) {},
			expectedOutput: "IPSec add-on disabled successfully",
			expectCount:    0,
		},
		{
			name:        "service returns error",
			mcrUID:      "mcr-abc",
			addOnUID:    "addon-123",
			tunnelCount: 20,
			setupMock: func(m *MockMCRService) {
				m.UpdateMCRIPsecAddOnErr = fmt.Errorf("update failed")
			},
			expectedError: "update failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &MockMCRService{}
			tt.setupMock(mockSvc)

			config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
				client := &megaport.Client{}
				client.MCRService = mockSvc
				return client, nil
			})

			cmd := &cobra.Command{Use: "update-ipsec-addon [mcrUID] [addOnUID]"}
			cmd.Flags().Bool("interactive", false, "")
			cmd.Flags().String("json", "", "")
			cmd.Flags().String("json-file", "", "")
			cmd.Flags().Int("tunnel-count", 0, "")
			_ = cmd.Flags().Set("tunnel-count", fmt.Sprintf("%d", tt.tunnelCount))

			var err error
			out := output.CaptureOutput(func() {
				err = UpdateMCRIPSecAddOn(cmd, []string{tt.mcrUID, tt.addOnUID}, false)
			})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Contains(t, out, tt.expectedOutput)
				assert.Equal(t, tt.mcrUID, mockSvc.CapturedUpdateMCRIPsecAddOnMCRID)
				assert.Equal(t, tt.addOnUID, mockSvc.CapturedUpdateMCRIPsecAddOnUID)
				assert.Equal(t, tt.expectCount, mockSvc.CapturedUpdateMCRIPsecTunnelCount)
			}
		})
	}
}

func TestUpdateMCRIPSecAddOn_JSON(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	mockSvc := &MockMCRService{}
	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		client := &megaport.Client{}
		client.MCRService = mockSvc
		return client, nil
	})

	cmd := &cobra.Command{Use: "update-ipsec-addon [mcrUID] [addOnUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", `{"tunnelCount":30}`, "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")

	var err error
	out := output.CaptureOutput(func() {
		err = UpdateMCRIPSecAddOn(cmd, []string{"mcr-json", "addon-json"}, false)
	})

	assert.NoError(t, err)
	assert.Contains(t, out, "tunnel count: 30")
	assert.Equal(t, "mcr-json", mockSvc.CapturedUpdateMCRIPsecAddOnMCRID)
	assert.Equal(t, "addon-json", mockSvc.CapturedUpdateMCRIPsecAddOnUID)
	assert.Equal(t, 30, mockSvc.CapturedUpdateMCRIPsecTunnelCount)
}

func TestUpdateMCRIPSecAddOn_NoInput(t *testing.T) {
	cmd := &cobra.Command{Use: "update-ipsec-addon [mcrUID] [addOnUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")

	err := UpdateMCRIPSecAddOn(cmd, []string{"mcr-abc", "addon-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no input provided")
}

func TestAddMCRIPSecAddOn_InvalidTunnelCount(t *testing.T) {
	cmd := &cobra.Command{Use: "add-ipsec-addon [mcrUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")
	_ = cmd.Flags().Set("tunnel-count", "5") // not 10, 20, or 30

	err := AddMCRIPSecAddOn(cmd, []string{"mcr-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IPSec tunnel count")
}

func TestUpdateMCRIPSecAddOn_InvalidTunnelCount(t *testing.T) {
	cmd := &cobra.Command{Use: "update-ipsec-addon [mcrUID] [addOnUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")
	_ = cmd.Flags().Set("tunnel-count", "15") // not 0, 10, 20, or 30

	err := UpdateMCRIPSecAddOn(cmd, []string{"mcr-abc", "addon-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid IPSec tunnel count")
}

func TestAddMCRIPSecAddOn_LoginError(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("auth failed")
	})

	cmd := &cobra.Command{Use: "add-ipsec-addon [mcrUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")
	_ = cmd.Flags().Set("tunnel-count", "10")

	err := AddMCRIPSecAddOn(cmd, []string{"mcr-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestUpdateMCRIPSecAddOn_LoginError(t *testing.T) {
	cleanup := testutil.SetupLogin(func(c *megaport.Client) {})
	defer cleanup()

	config.SetLoginFunc(func(ctx context.Context) (*megaport.Client, error) {
		return nil, fmt.Errorf("auth failed")
	})

	cmd := &cobra.Command{Use: "update-ipsec-addon [mcrUID] [addOnUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", "", "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")
	_ = cmd.Flags().Set("tunnel-count", "10")

	err := UpdateMCRIPSecAddOn(cmd, []string{"mcr-abc", "addon-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestParseIPSecTunnelCountFromJSON(t *testing.T) {
	tests := []struct {
		name        string
		jsonStr     string
		expected    int
		wantErr     bool
		errContains string
	}{
		{"valid tunnelCount", `{"tunnelCount":20}`, 20, false, ""},
		{"zero tunnelCount", `{"tunnelCount":0}`, 0, false, ""},
		{"missing field returns zero", `{}`, 0, false, ""},
		{"invalid JSON", `{bad}`, 0, true, "failed to parse JSON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := parseIPSecTunnelCountFromJSON(tt.jsonStr, "")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, count)
			}
		})
	}
}

func TestAddMCRIPSecAddOn_BadJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "add-ipsec-addon [mcrUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", `{bad json}`, "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")

	err := AddMCRIPSecAddOn(cmd, []string{"mcr-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestUpdateMCRIPSecAddOn_BadJSON(t *testing.T) {
	cmd := &cobra.Command{Use: "update-ipsec-addon [mcrUID] [addOnUID]"}
	cmd.Flags().Bool("interactive", false, "")
	cmd.Flags().String("json", `{bad json}`, "")
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().Int("tunnel-count", 0, "")

	err := UpdateMCRIPSecAddOn(cmd, []string{"mcr-abc", "addon-abc"}, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

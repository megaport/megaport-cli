package vxc

import (
	"context"
	"testing"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testVXCs = []*megaport.VXC{
	{
		UID:  "vxc-1",
		Name: "MyVXCOne",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-1",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-1",
		},
		RateLimit:          100,
		ProvisioningStatus: "CONFIGURED",
	},
	{
		UID:  "vxc-2",
		Name: "AnotherVXC",
		AEndConfiguration: megaport.VXCEndConfiguration{
			UID: "a-end-2",
		},
		BEndConfiguration: megaport.VXCEndConfiguration{
			UID: "b-end-2",
		},
		RateLimit:          200,
		ProvisioningStatus: "LIVE",
	},
}

func TestPrintVXCs_Table(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "table", true)
		assert.NoError(t, err)
	})

	// Check for headers and content
	assert.Contains(t, output, "UID")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "A END UID")
	assert.Contains(t, output, "B END UID")
	assert.Contains(t, output, "A END VLAN")
	assert.Contains(t, output, "B END VLAN")
	assert.Contains(t, output, "RATE LIMIT")
	assert.Contains(t, output, "STATUS")

	// Check for actual data
	assert.Contains(t, output, "vxc-1")
	assert.Contains(t, output, "MyVXCOne")
	assert.Contains(t, output, "a-end-1")
	assert.Contains(t, output, "b-end-1")
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "CONFIGURED")

	assert.Contains(t, output, "vxc-2")
	assert.Contains(t, output, "AnotherVXC")
	assert.Contains(t, output, "a-end-2")
	assert.Contains(t, output, "b-end-2")
	assert.Contains(t, output, "200")
	assert.Contains(t, output, "LIVE")

	// Check for box drawing characters
	assert.Contains(t, output, "┌")
	assert.Contains(t, output, "┐")
	assert.Contains(t, output, "└")
	assert.Contains(t, output, "┘")
	assert.Contains(t, output, "├")
	assert.Contains(t, output, "┤")
	assert.Contains(t, output, "│")
	assert.Contains(t, output, "─")
}

func TestPrintVXCs_JSON(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "json", true)
		assert.NoError(t, err)
	})

	expected := `[
  {
    "uid": "vxc-1",
    "name": "MyVXCOne",
    "a_end_uid": "a-end-1",
    "b_end_uid": "b-end-1",
    "a_end_vlan": 0,
    "b_end_vlan": 0,
    "rate_limit": 100,
    "status": "CONFIGURED"
  },
  {
    "uid": "vxc-2",
    "name": "AnotherVXC",
    "a_end_uid": "a-end-2",
    "b_end_uid": "b-end-2",
    "a_end_vlan": 0,
    "b_end_vlan": 0,
    "rate_limit": 200,
    "status": "LIVE"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestPrintVXCs_CSV(t *testing.T) {
	output := output.CaptureOutput(func() {
		err := printVXCs(testVXCs, "csv", true)
		assert.NoError(t, err)
	})

	expected := `uid,name,a_end_uid,b_end_uid,a_end_vlan,b_end_vlan,rate_limit,status
vxc-1,MyVXCOne,a-end-1,b-end-1,0,0,100,CONFIGURED
vxc-2,AnotherVXC,a-end-2,b-end-2,0,0,200,LIVE
`
	assert.Equal(t, expected, output)
}

func TestPrintVXCs_Invalid(t *testing.T) {
	var err error
	output := output.CaptureOutput(func() {
		err = printVXCs(testVXCs, "invalid", true)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Empty(t, output)
}

func TestPrintVXCs_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		vxcs         []*megaport.VXC
		format       string
		shouldError  bool
		validateFunc func(*testing.T, string) // New function to validate output
		expected     string                   // Keep for JSON and CSV validation
	}{
		{
			name:        "nil slice",
			vxcs:        nil,
			format:      "table",
			shouldError: false,
			validateFunc: func(t *testing.T, output string) {
				// Check for headers and box drawing characters in empty table
				assert.Contains(t, output, "UID")
				assert.Contains(t, output, "NAME")
				assert.Contains(t, output, "A END UID")
				assert.Contains(t, output, "B END UID")
				assert.Contains(t, output, "A END VLAN")
				assert.Contains(t, output, "B END VLAN")
				assert.Contains(t, output, "RATE LIMIT")
				assert.Contains(t, output, "STATUS")

				// Check for box drawing characters
				assert.Contains(t, output, "┌")
				assert.Contains(t, output, "┐")
				assert.Contains(t, output, "└")
				assert.Contains(t, output, "┘")
				assert.Contains(t, output, "│")
				assert.Contains(t, output, "─")
			},
		},
		{
			name:        "empty slice",
			vxcs:        []*megaport.VXC{},
			format:      "json",
			shouldError: false,
			expected:    "[]",
		},
		{
			name: "nil vxc in slice",
			vxcs: []*megaport.VXC{
				nil,
				{
					UID:  "vxc-1",
					Name: "TestVXC",
				},
			},
			format:      "table",
			shouldError: true,
			expected:    "invalid VXC: nil value",
		},
		{
			name: "nil end configurations",
			vxcs: []*megaport.VXC{
				{
					UID:                "vxc-1",
					Name:               "TestVXC",
					RateLimit:          50,
					ProvisioningStatus: "PENDING",
				},
			},
			format:      "csv",
			shouldError: false,
			expected:    "uid,name,a_end_uid,b_end_uid,a_end_vlan,b_end_vlan,rate_limit,status\nvxc-1,TestVXC,,,0,0,50,PENDING\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			capturedOutput := output.CaptureOutput(func() {
				err = printVXCs(tt.vxcs, tt.format, true)
			})

			if tt.shouldError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expected)
				assert.Empty(t, capturedOutput)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, capturedOutput)
				} else if tt.expected != "" {
					switch tt.format {
					case "json":
						assert.JSONEq(t, tt.expected, capturedOutput)
					case "csv":
						assert.Equal(t, tt.expected, capturedOutput)
					}
				}
			}
		})
	}
}

func TestFilterVXCs(t *testing.T) {
	activeVXCs := []*megaport.VXC{
		{
			UID:  "vxc-1",
			Name: "TestVXC-1",
			AEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-aaa",
			},
			BEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-bbb",
			},
			RateLimit:          100,
			ProvisioningStatus: "LIVE",
		},
		{
			UID:  "vxc-2",
			Name: "TestVXC-2",
			AEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-ccc",
			},
			BEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-ddd",
			},
			RateLimit:          500,
			ProvisioningStatus: "CONFIGURED",
		},
		{
			UID:  "vxc-3",
			Name: "Production-VXC",
			AEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-aaa",
			},
			BEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-eee",
			},
			RateLimit:          1000,
			ProvisioningStatus: "LIVE",
		},
		{
			UID:  "vxc-4",
			Name: "Staging-VXC",
			AEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-fff",
			},
			BEndConfiguration: megaport.VXCEndConfiguration{
				UID: "port-bbb",
			},
			RateLimit:          500,
			ProvisioningStatus: "LIVE",
		},
	}

	tests := []struct {
		name         string
		vxcs         []*megaport.VXC
		vxcName      string
		expected     int
		expectedUIDs []string
	}{
		{
			name:         "no filters",
			vxcs:         activeVXCs,
			expected:     4,
			expectedUIDs: []string{"vxc-1", "vxc-2", "vxc-3", "vxc-4"},
		},
		{
			name:         "filter by name (case insensitive)",
			vxcs:         activeVXCs,
			vxcName:      "test",
			expected:     2,
			expectedUIDs: []string{"vxc-1", "vxc-2"},
		},
		{
			name:         "filter by name (partial match)",
			vxcs:         activeVXCs,
			vxcName:      "Production",
			expected:     1,
			expectedUIDs: []string{"vxc-3"},
		},
		{
			name:         "filter by exact name match",
			vxcs:         activeVXCs,
			vxcName:      "TestVXC-1",
			expected:     1,
			expectedUIDs: []string{"vxc-1"},
		},
		{
			name:         "no matching VXCs",
			vxcs:         activeVXCs,
			vxcName:      "nonexistent",
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "nil slice",
			vxcs:         nil,
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "empty slice",
			vxcs:         []*megaport.VXC{},
			expected:     0,
			expectedUIDs: []string{},
		},
		{
			name:         "slice with nil VXC",
			vxcs:         []*megaport.VXC{nil, activeVXCs[0], nil, activeVXCs[1]},
			expected:     2,
			expectedUIDs: []string{"vxc-1", "vxc-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterVXCs(tt.vxcs, tt.vxcName)

			assert.Equal(t, tt.expected, len(filtered), "Filtered VXC count should match expected")

			if len(tt.expectedUIDs) > 0 {
				actualUIDs := make([]string, len(filtered))
				for i, vxc := range filtered {
					actualUIDs[i] = vxc.UID
				}
				assert.ElementsMatch(t, tt.expectedUIDs, actualUIDs, "Filtered VXC UIDs should match expected")
			}
		})
	}
}

func TestDisplayVXCChanges(t *testing.T) {
	tests := []struct {
		name             string
		original         *megaport.VXC
		updated          *megaport.VXC
		expectedContains []string
		expectEmpty      bool
	}{
		{
			name:        "nil original",
			original:    nil,
			updated:     &megaport.VXC{},
			expectEmpty: true,
		},
		{
			name:        "nil updated",
			original:    &megaport.VXC{},
			updated:     nil,
			expectEmpty: true,
		},
		{
			name:             "no changes",
			original:         &megaport.VXC{Name: "Same", RateLimit: 100},
			updated:          &megaport.VXC{Name: "Same", RateLimit: 100},
			expectedContains: []string{"No changes detected"},
		},
		{
			name:             "name change",
			original:         &megaport.VXC{Name: "Old"},
			updated:          &megaport.VXC{Name: "New"},
			expectedContains: []string{"Name:", "Old", "New"},
		},
		{
			name:             "rate limit change",
			original:         &megaport.VXC{RateLimit: 100},
			updated:          &megaport.VXC{RateLimit: 500},
			expectedContains: []string{"Rate Limit:", "100 Mbps", "500 Mbps"},
		},
		{
			name:             "cost centre change from empty",
			original:         &megaport.VXC{CostCentre: ""},
			updated:          &megaport.VXC{CostCentre: "CC-123"},
			expectedContains: []string{"Cost Centre:", "(none)", "CC-123"},
		},
		{
			name:             "term change",
			original:         &megaport.VXC{ContractTermMonths: 12},
			updated:          &megaport.VXC{ContractTermMonths: 24},
			expectedContains: []string{"Contract Term:", "12 months", "24 months"},
		},
		{
			name: "a-end vlan change",
			original: &megaport.VXC{
				AEndConfiguration: megaport.VXCEndConfiguration{VLAN: 100},
			},
			updated: &megaport.VXC{
				AEndConfiguration: megaport.VXCEndConfiguration{VLAN: 200},
			},
			expectedContains: []string{"A-End VLAN:", "100", "200"},
		},
		{
			name: "b-end vlan change",
			original: &megaport.VXC{
				BEndConfiguration: megaport.VXCEndConfiguration{VLAN: 300},
			},
			updated: &megaport.VXC{
				BEndConfiguration: megaport.VXCEndConfiguration{VLAN: 400},
			},
			expectedContains: []string{"B-End VLAN:"},
		},
		{
			name:             "locked change",
			original:         &megaport.VXC{Locked: false},
			updated:          &megaport.VXC{Locked: true},
			expectedContains: []string{"Locked:", "No", "Yes"},
		},
		{
			name: "multiple changes",
			original: &megaport.VXC{
				Name:      "OldName",
				RateLimit: 100,
				Locked:    false,
			},
			updated: &megaport.VXC{
				Name:      "NewName",
				RateLimit: 500,
				Locked:    true,
			},
			expectedContains: []string{"Name:", "Rate Limit:", "Locked:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := output.CaptureOutput(func() {
				displayVXCChanges(tt.original, tt.updated, true)
			})

			if tt.expectEmpty {
				assert.Empty(t, captured)
			} else {
				for _, expected := range tt.expectedContains {
					assert.Contains(t, captured, expected)
				}
			}
		})
	}
}

func TestToVXCOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		vxc           *megaport.VXC
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, vxcOutput)
	}{
		{
			name:          "nil vxc",
			vxc:           nil,
			shouldError:   true,
			errorContains: "invalid VXC: nil value",
		},
		{
			name: "zero values",
			vxc:  &megaport.VXC{},
			validateFunc: func(t *testing.T, output vxcOutput) {
				assert.Empty(t, output.UID)
				assert.Empty(t, output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
				assert.Equal(t, 0, output.AEndVLAN)
				assert.Equal(t, 0, output.BEndVLAN)
				assert.Equal(t, 0, output.RateLimit)
				assert.Empty(t, output.Status)
			},
		},
		{
			name: "nil end configurations",
			vxc: &megaport.VXC{
				UID:                "vxc-1",
				Name:               "TestVXC",
				RateLimit:          75,
				ProvisioningStatus: "CONFIGURED",
			},
			validateFunc: func(t *testing.T, output vxcOutput) {
				assert.Equal(t, "vxc-1", output.UID)
				assert.Equal(t, "TestVXC", output.Name)
				assert.Empty(t, output.AEndUID)
				assert.Empty(t, output.BEndUID)
				assert.Equal(t, 0, output.AEndVLAN)
				assert.Equal(t, 0, output.BEndVLAN)
				assert.Equal(t, 75, output.RateLimit)
				assert.Equal(t, "CONFIGURED", output.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := toVXCOutput(tt.vxc)

			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, output)
				}
			}
		})
	}
}

func TestVXCUpdateTagsHasGenerateSkeleton(t *testing.T) {
	root := &cobra.Command{Use: "megaport-cli"}
	AddCommandsTo(root)
	updateTagsCmd, _, err := root.Find([]string{"vxc", "update-tags"})
	require.NoError(t, err)
	require.NotNil(t, updateTagsCmd)
	assert.NotNil(t, updateTagsCmd.Flags().Lookup("generate-skeleton"))
}

func TestVXCListHasTagFlag(t *testing.T) {
	root := &cobra.Command{Use: "megaport-cli"}
	AddCommandsTo(root)
	listCmd, _, err := root.Find([]string{"vxc", "list"})
	require.NoError(t, err)
	require.NotNil(t, listCmd)
	assert.NotNil(t, listCmd.Flags().Lookup("tag"), "list command should have --tag flag")
}

func TestListVXCResourceTagsFunc(t *testing.T) {
	want := map[string]string{"env": "prod"}
	mockSvc := &MockVXCService{ListVXCResourceTagsResult: want}
	client := &megaport.Client{}
	client.VXCService = mockSvc

	got, err := listVXCResourceTagsFunc(context.Background(), client, "vxc-uid-1")
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

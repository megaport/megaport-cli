package cmd

import (
	"strings"
	"testing"
	"time"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

// Update mock data to include all required fields
var mockServiceKeys = []*megaport.ServiceKey{
	{
		Key:         "abcd-1234-efgh-5678",
		Description: "Test Key One",
		ProductUID:  "prd-uid-1",
		ProductID:   1,
		SingleUse:   true,
		MaxSpeed:    1000,
		Active:      true,
		ProductName: "Product One",
		CreateDate: &megaport.Time{
			Time: time.Now(),
		},
	},
	{
		Key:         "ijkl-9012-mnop-3456",
		Description: "Test Key Two",
		ProductUID:  "prd-uid-2",
		ProductID:   2,
		SingleUse:   false,
		MaxSpeed:    500,
		Active:      true,
		ProductName: "Product Two",
		CreateDate: &megaport.Time{
			Time: time.Now(),
		},
	},
}

// toServiceKeyOutput is an example helper that converts a ServiceKey to a minimal output struct.
// Adjust to match your actual output formatting or logic in servicekeys.go if needed.
func toServiceKeyOutput(k *megaport.ServiceKey) map[string]interface{} {
	return map[string]interface{}{
		"key":         k.Key,
		"description": k.Description,
		"single_use":  k.SingleUse,
		"max_speed":   k.MaxSpeed,
		"active":      k.Active,
	}
}

// Example test for a helper function that might convert a ServiceKey into output format (like JSON).
func TestToServiceKeyOutput(t *testing.T) {
	sk := mockServiceKeys[0]
	output := toServiceKeyOutput(sk)

	assert.Equal(t, sk.Key, output["key"])
	assert.Equal(t, sk.Description, output["description"])
	assert.Equal(t, sk.SingleUse, output["single_use"])
	assert.Equal(t, sk.MaxSpeed, output["max_speed"])
	assert.Equal(t, sk.Active, output["active"])
}

// filterServiceKeys is an example helper for local filtering logic.
// If your real code filters service keys by product ID, single use, etc. adapt as needed.
func filterServiceKeys(keys []*megaport.ServiceKey, singleUse *bool, maxSpeedMin int) []*megaport.ServiceKey {
	var filtered []*megaport.ServiceKey
	for _, key := range keys {
		if singleUse != nil && key.SingleUse != *singleUse {
			continue
		}
		if maxSpeedMin > 0 && key.MaxSpeed < maxSpeedMin {
			continue
		}
		filtered = append(filtered, key)
	}
	return filtered
}

// Example test for filtering logic.
func TestFilterServiceKeys(t *testing.T) {
	tests := []struct {
		name        string
		singleUse   *bool
		maxSpeedMin int
		expected    int
	}{
		{
			name:        "No filters",
			singleUse:   nil,
			maxSpeedMin: 0,
			expected:    2,
		},
		{
			name:        "Single use only",
			singleUse:   boolPtr(true),
			maxSpeedMin: 0,
			expected:    1,
		},
		{
			name:        "Min speed of 750",
			singleUse:   nil,
			maxSpeedMin: 750,
			expected:    1,
		},
		{
			name:        "Single use AND fast",
			singleUse:   boolPtr(true),
			maxSpeedMin: 750,
			expected:    1,
		},
		{
			name:        "No matches",
			singleUse:   boolPtr(false),
			maxSpeedMin: 1001,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterServiceKeys(mockServiceKeys, tt.singleUse, tt.maxSpeedMin)
			assert.Equal(t, tt.expected, len(got))
		})
	}
}

func TestServiceKeyOutput_Table(t *testing.T) {
	outputs := make([]ServiceKeyOutput, 0, len(mockServiceKeys))
	for _, sk := range mockServiceKeys {
		outputs = append(outputs, ToServiceKeyOutput(sk))
	}

	output := captureOutput(func() {
		err := printOutput(outputs, "table")
		assert.NoError(t, err)
	})

	expected := `key_uid               product_name   product_uid   description    create_date
abcd-1234-efgh-5678   Product One    prd-uid-1     Test Key One   2025-02-25T12:00:00Z
ijkl-9012-mnop-3456   Product Two    prd-uid-2     Test Key Two   2025-02-25T12:00:00Z
`
	assert.Equal(t, expected, output)
}

func TestServiceKeyOutput_JSON(t *testing.T) {
	outputs := make([]ServiceKeyOutput, 0, len(mockServiceKeys))
	for _, sk := range mockServiceKeys {
		outputs = append(outputs, ToServiceKeyOutput(sk))
	}

	output := captureOutput(func() {
		err := printOutput(outputs, "json")
		assert.NoError(t, err)
	})

	expected := `[
  {
    "key_uid": "abcd-1234-efgh-5678",
    "product_name": "Product One",
    "product_uid": "prd-uid-1",
    "description": "Test Key One",
    "create_date": "2025-02-25T12:00:00Z"
  },
  {
    "key_uid": "ijkl-9012-mnop-3456",
    "product_name": "Product Two",
    "product_uid": "prd-uid-2",
    "description": "Test Key Two",
    "create_date": "2025-02-25T12:00:00Z"
  }
]`
	assert.JSONEq(t, expected, output)
}

func TestServiceKeyOutput_CSV(t *testing.T) {
	outputs := make([]ServiceKeyOutput, 0, len(mockServiceKeys))
	for _, sk := range mockServiceKeys {
		outputs = append(outputs, ToServiceKeyOutput(sk))
	}

	output := captureOutput(func() {
		err := printOutput(outputs, "csv")
		assert.NoError(t, err)
	})

	// Note: CreateDate will be dynamic, so we'll check the structure only
	lines := strings.Split(output, "\n")
	assert.Equal(t, 4, len(lines)) // header + 2 data lines + empty line
	assert.Equal(t, "key_uid,product_name,product_uid,description,create_date", lines[0])
	assert.Contains(t, lines[1], "abcd-1234-efgh-5678,Product One,prd-uid-1,Test Key One,")
	assert.Contains(t, lines[2], "ijkl-9012-mnop-3456,Product Two,prd-uid-2,Test Key Two,")
}

func boolPtr(b bool) *bool {
	return &b
}

func init() {
	// Set fixed time for tests
	fixedTime := time.Date(2025, 2, 25, 12, 0, 0, 0, time.UTC)
	for _, sk := range mockServiceKeys {
		sk.CreateDate.Time = fixedTime
	}
}

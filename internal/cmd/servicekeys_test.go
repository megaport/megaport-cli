package cmd

import (
	"encoding/json"
	"testing"

	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

// mockServiceKeys is our fake data for testing
var mockServiceKeys = []*megaport.ServiceKey{
	{
		Key:         "abcd-1234-efgh-5678",
		Description: "Test Key One",
		ProductUID:  "prd-uid-1",
		ProductID:   1,
		SingleUse:   true,
		MaxSpeed:    1000,
		Active:      true,
	},
	{
		Key:         "ijkl-9012-mnop-3456",
		Description: "Test Key Two",
		ProductUID:  "prd-uid-2",
		ProductID:   2,
		SingleUse:   false,
		MaxSpeed:    500,
		Active:      true,
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

// Example test showing how you might verify JSON output of your service keys logic
// without making any real API calls or requiring login.
func TestServiceKeysJSONOutput(t *testing.T) {
	// Suppose you have a function that formats your service keys as JSON.
	// We'll just demonstrate a mock approach here.
	data, err := json.Marshal(mockServiceKeys)
	assert.NoError(t, err)

	var decoded []*megaport.ServiceKey
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, len(mockServiceKeys), len(decoded))
	assert.Equal(t, mockServiceKeys[0].Key, decoded[0].Key)
	assert.Equal(t, mockServiceKeys[1].Key, decoded[1].Key)
}

func boolPtr(b bool) *bool {
	return &b
}

package servicekeys

import (
	"strings"
	"testing"
	"time"

	"github.com/megaport/megaport-cli/internal/base/output"
	megaport "github.com/megaport/megaportgo"
	"github.com/stretchr/testify/assert"
)

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

func toServiceKeyOutput(k *megaport.ServiceKey) map[string]interface{} {
	return map[string]interface{}{
		"key":         k.Key,
		"description": k.Description,
		"single_use":  k.SingleUse,
		"max_speed":   k.MaxSpeed,
		"active":      k.Active,
	}
}

func TestToServiceKeyOutput(t *testing.T) {
	sk := mockServiceKeys[0]
	output := toServiceKeyOutput(sk)

	assert.Equal(t, sk.Key, output["key"])
	assert.Equal(t, sk.Description, output["description"])
	assert.Equal(t, sk.SingleUse, output["single_use"])
	assert.Equal(t, sk.MaxSpeed, output["max_speed"])
	assert.Equal(t, sk.Active, output["active"])
}

func filterServiceKeys(keys []*megaport.ServiceKey, singleUse *bool, maxSpeedMin int) []*megaport.ServiceKey {
	if keys == nil {
		return nil
	}

	var filtered []*megaport.ServiceKey
	for _, key := range keys {
		if key == nil {
			continue
		}

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
		skOutput, err := ToServiceKeyOutput(sk)
		assert.NoError(t, err)
		outputs = append(outputs, skOutput)
	}

	output := output.CaptureOutput(func() {
		err := output.PrintOutput(outputs, "table", true)
		assert.NoError(t, err)
	})

	expected := ` KEY_UID             │ PRODUCT_NAME │ PRODUCT_UID │ DESCRIPTION  │ CREATE_DATE          
─────────────────────┼──────────────┼─────────────┼──────────────┼──────────────────────
 abcd-1234-efgh-5678 │ Product One  │ prd-uid-1   │ Test Key One │ 2025-02-25T12:00:00Z 
 ijkl-9012-mnop-3456 │ Product Two  │ prd-uid-2   │ Test Key Two │ 2025-02-25T12:00:00Z 
`
	assert.Equal(t, expected, output)
}

func TestServiceKeyOutput_JSON(t *testing.T) {
	outputs := make([]ServiceKeyOutput, 0, len(mockServiceKeys))
	for _, sk := range mockServiceKeys {
		skOutput, err := ToServiceKeyOutput(sk)
		assert.NoError(t, err)
		outputs = append(outputs, skOutput)
	}

	output := output.CaptureOutput(func() {
		err := output.PrintOutput(outputs, "json", false)
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
		skOutput, err := ToServiceKeyOutput(sk)
		assert.NoError(t, err)
		outputs = append(outputs, skOutput)
	}

	output := output.CaptureOutput(func() {
		err := output.PrintOutput(outputs, "csv", false)
		assert.NoError(t, err)
	})

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
	fixedTime := time.Date(2025, 2, 25, 12, 0, 0, 0, time.UTC)
	for _, sk := range mockServiceKeys {
		sk.CreateDate.Time = fixedTime
	}
}

func TestFilterServiceKeys_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		keys        []*megaport.ServiceKey
		singleUse   *bool
		maxSpeedMin int
		expected    int
	}{
		{
			name:        "nil slice",
			keys:        nil,
			singleUse:   nil,
			maxSpeedMin: 0,
			expected:    0,
		},
		{
			name:        "empty slice",
			keys:        []*megaport.ServiceKey{},
			singleUse:   boolPtr(true),
			maxSpeedMin: 1000,
			expected:    0,
		},
		{
			name: "nil key in slice",
			keys: []*megaport.ServiceKey{
				nil,
				mockServiceKeys[0],
			},
			singleUse:   nil,
			maxSpeedMin: 0,
			expected:    1,
		},
		{
			name: "zero values",
			keys: []*megaport.ServiceKey{
				{
					Key:         "",
					Description: "",
					ProductUID:  "",
					MaxSpeed:    0,
					SingleUse:   false,
				},
			},
			singleUse:   nil,
			maxSpeedMin: 0,
			expected:    1,
		},
		{
			name:        "negative max speed",
			keys:        mockServiceKeys,
			singleUse:   nil,
			maxSpeedMin: -1000,
			expected:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterServiceKeys(tt.keys, tt.singleUse, tt.maxSpeedMin)
			assert.Equal(t, tt.expected, len(result))

			for _, key := range result {
				assert.NotNil(t, key, "Filtered results should not contain nil keys")
			}
		})
	}
}

func TestToServiceKeyOutput_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		key           *megaport.ServiceKey
		shouldError   bool
		errorContains string
		validateFunc  func(*testing.T, ServiceKeyOutput)
	}{
		{
			name:          "nil service key",
			key:           nil,
			shouldError:   true,
			errorContains: "nil service key",
		},
		{
			name: "zero values",
			key:  &megaport.ServiceKey{},
			validateFunc: func(t *testing.T, output ServiceKeyOutput) {
				assert.Empty(t, output.KeyUID)
				assert.Empty(t, output.ProductName)
				assert.Empty(t, output.ProductUID)
				assert.Empty(t, output.Description)
				assert.Empty(t, output.CreateDate)
			},
		},
		{
			name: "nil create date",
			key: &megaport.ServiceKey{
				Key:         "test-key",
				ProductName: "Test Product",
				ProductUID:  "prod-123",
				Description: "Test Description",
				CreateDate:  nil,
			},
			validateFunc: func(t *testing.T, output ServiceKeyOutput) {
				assert.Equal(t, "test-key", output.KeyUID)
				assert.Equal(t, "Test Product", output.ProductName)
				assert.Equal(t, "prod-123", output.ProductUID)
				assert.Equal(t, "Test Description", output.Description)
				assert.Empty(t, output.CreateDate)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ToServiceKeyOutput(tt.key)

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

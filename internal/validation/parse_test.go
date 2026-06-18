package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		want      int
		wantErr   bool
		errSubstr []string
	}{
		{name: "valid positive", field: "location ID", value: "123", want: 123},
		{name: "valid negative", field: "ASN", value: "-5", want: -5},
		{name: "valid zero", field: "VLAN ID", value: "0", want: 0},
		{
			name:    "non-numeric",
			field:   "location ID",
			value:   "bad-location",
			wantErr: true,
			// Friendly message: names the field and value, no strconv internals.
			errSubstr: []string{"invalid location ID", "bad-location"},
		},
		{
			name:      "empty string",
			field:     "employee ID",
			value:     "",
			wantErr:   true,
			errSubstr: []string{"invalid employee ID"},
		},
		{
			name:      "float string",
			field:     "term",
			value:     "1.5",
			wantErr:   true,
			errSubstr: []string{"invalid term"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.field, tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				for _, s := range tt.errSubstr {
					assert.Contains(t, err.Error(), s)
				}
				// Must not leak strconv internals to the user.
				assert.NotContains(t, err.Error(), "strconv")
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// classifyError keys off a lowercase "invalid" + "ID" substring, so the helper
// must keep that wording for ID fields to preserve the Usage exit code.
func TestParseIntPreservesUsageClassification(t *testing.T) {
	_, err := ParseInt("location ID", "abc")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid"))
	assert.True(t, strings.Contains(err.Error(), "ID"))
}

//go:build !wasm

package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type csvTagStruct struct {
	A      string `csv:"col_a"`        // explicit csv header
	B      string `json:"b_json"`      // csv falls back to json tag
	C      string `csv:"-"`            // skipped: csv tag "-"
	D      string `json:"-"`           // skipped: no csv, json "-"
	E      string `json:"e,omitempty"` // json options stripped to "e"
	F      string // skipped: no csv/json tag
	G      func() `csv:"g"` // skipped: not output-compatible
	hidden string //nolint:unused // exercises the unexported-field skip
}

func TestExtractCSVFieldInfo(t *testing.T) {
	t.Run("value struct honors csv/json tags and skips the rest", func(t *testing.T) {
		headers, jsonNames, indices, err := extractCSVFieldInfo([]csvTagStruct{{}})
		assert.NoError(t, err)
		assert.Equal(t, []string{"col_a", "b_json", "e"}, headers)
		assert.Equal(t, []string{"a", "b_json", "e"}, jsonNames)
		assert.Len(t, indices, 3)
	})

	t.Run("empty slice still derives fields from the element type", func(t *testing.T) {
		headers, _, _, err := extractCSVFieldInfo([]csvTagStruct{})
		assert.NoError(t, err)
		assert.Equal(t, []string{"col_a", "b_json", "e"}, headers)
	})

	t.Run("non-nil pointer element", func(t *testing.T) {
		headers, _, _, err := extractCSVFieldInfo([]*csvTagStruct{{}})
		assert.NoError(t, err)
		assert.Equal(t, []string{"col_a", "b_json", "e"}, headers)
	})

	t.Run("nil pointer to struct uses the element type", func(t *testing.T) {
		headers, _, _, err := extractCSVFieldInfo([]*csvTagStruct{nil})
		assert.NoError(t, err)
		assert.Equal(t, []string{"col_a", "b_json", "e"}, headers)
	})

	t.Run("nil pointer to non-struct yields nothing", func(t *testing.T) {
		headers, jsonNames, indices, err := extractCSVFieldInfo([]*int{nil})
		assert.NoError(t, err)
		assert.Nil(t, headers)
		assert.Nil(t, jsonNames)
		assert.Nil(t, indices)
	})

	t.Run("non-struct type yields nothing", func(t *testing.T) {
		headers, _, _, err := extractCSVFieldInfo([]int{5})
		assert.NoError(t, err)
		assert.Nil(t, headers)
	})

	t.Run("invalid sample (nil interface) yields nothing", func(t *testing.T) {
		headers, _, _, err := extractCSVFieldInfo([]any{})
		assert.NoError(t, err)
		assert.Nil(t, headers)
	})
}

func TestIsNilOrInvalid(t *testing.T) {
	assert.True(t, isNilOrInvalid(nil), "nil interface is invalid")
	assert.True(t, isNilOrInvalid((*int)(nil)), "typed nil pointer")
	assert.False(t, isNilOrInvalid(5), "non-nil value")
	assert.False(t, isNilOrInvalid(&csvTagStruct{}), "non-nil pointer")
}

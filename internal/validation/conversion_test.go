package validation

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIntFromInterface(t *testing.T) {
	// Derived from math.MaxInt (rather than a hardcoded 2^63) so the boundary
	// cases stay in range regardless of platform int width.
	nearMaxInt := float64(math.MaxInt) - 4096
	justAboveMaxInt := float64(math.MaxInt) + 4096

	tests := []struct {
		name  string
		value interface{}
		want  int
		ok    bool
	}{
		{"int", 42, 42, true},
		{"negative int", -7, -7, true},
		{"zero int", 0, 0, true},
		{"float64 whole", float64(10), 10, true},
		{"float64 fractional rejected", float64(3.9), 0, false},
		{"float64 above int range rejected", float64(math.MaxInt) * 2, 0, false},
		{"float64 just above max int rejected", justAboveMaxInt, 0, false},
		{"float64 just below max int accepted", nearMaxInt, int(nearMaxInt), true},
		{"numeric string", "123", 123, true},
		{"negative numeric string", "-5", -5, true},
		{"empty string", "", 0, false},
		{"non-numeric string", "abc", 0, false},
		{"float string rejected by Atoi", "3.14", 0, false},
		{"overflow string", "999999999999999999999999", 0, false},
		{"nil", nil, 0, false},
		{"bool wrong type", true, 0, false},
		{"int64 wrong type", int64(5), 0, false},
		{"slice wrong type", []interface{}{1}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetIntFromInterface(tt.value)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetStringFromInterface(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
		ok    bool
	}{
		{"string", "hello", "hello", true},
		{"empty string", "", "", true},
		{"int wrong type", 5, "", false},
		{"float wrong type", 1.5, "", false},
		{"bool wrong type", true, "", false},
		{"nil", nil, "", false},
		{"slice wrong type", []interface{}{"x"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetStringFromInterface(tt.value)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetFloatFromInterface(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  float64
		ok    bool
	}{
		{"float64", 1.5, 1.5, true},
		{"int", 7, 7.0, true},
		{"zero", 0, 0.0, true},
		{"numeric string", "2.75", 2.75, true},
		{"integer string", "10", 10.0, true},
		{"scientific string", "1e3", 1000.0, true},
		{"empty string", "", 0, false},
		{"non-numeric string", "abc", 0, false},
		{"bool wrong type", false, 0, false},
		{"nil", nil, 0, false},
		{"int64 wrong type", int64(5), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetFloatFromInterface(tt.value)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetBoolFromInterface(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  bool
		ok    bool
	}{
		{"bool true", true, true, true},
		{"bool false", false, false, true},
		{"string true", "true", true, true},
		{"string false", "false", false, true},
		{"string True", "True", true, true},
		{"string FALSE", "FALSE", false, true},
		{"string TRUE", "TRUE", true, true},
		{"string mixed case", "TrUe", true, true},
		{"empty string", "", false, false},
		{"string yes", "yes", false, false},
		{"string 1", "1", false, false},
		{"string 0", "0", false, false},
		{"int wrong type", 1, false, false},
		{"nil", nil, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := GetBoolFromInterface(tt.value)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetMapStringInterfaceFromInterface(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		in := map[string]interface{}{"a": 1, "b": "two"}
		got, ok := GetMapStringInterfaceFromInterface(in)
		assert.True(t, ok)
		assert.Equal(t, in, got)
	})
	t.Run("nested map preserved", func(t *testing.T) {
		in := map[string]interface{}{"outer": map[string]interface{}{"inner": 1}}
		got, ok := GetMapStringInterfaceFromInterface(in)
		assert.True(t, ok)
		assert.Equal(t, in, got)
	})
	t.Run("empty map", func(t *testing.T) {
		got, ok := GetMapStringInterfaceFromInterface(map[string]interface{}{})
		assert.True(t, ok)
		assert.Empty(t, got)
	})
	for _, tc := range []struct {
		name  string
		value interface{}
	}{
		{"nil", nil},
		{"string", "x"},
		{"int", 1},
		{"slice", []interface{}{1}},
		{"map wrong value type", map[string]int{"a": 1}},
	} {
		t.Run(tc.name+" rejected", func(t *testing.T) {
			got, ok := GetMapStringInterfaceFromInterface(tc.value)
			assert.False(t, ok)
			assert.Nil(t, got)
		})
	}
}

func TestGetSliceMapStringInterfaceFromInterface(t *testing.T) {
	t.Run("direct slice of maps", func(t *testing.T) {
		in := []map[string]interface{}{{"a": 1}, {"b": 2}}
		got, ok := GetSliceMapStringInterfaceFromInterface(in)
		assert.True(t, ok)
		assert.Equal(t, in, got)
	})
	t.Run("slice of interface holding maps", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{"a": 1},
			map[string]interface{}{"b": 2},
		}
		got, ok := GetSliceMapStringInterfaceFromInterface(in)
		assert.True(t, ok)
		assert.Len(t, got, 2)
		assert.Equal(t, 1, got[0]["a"])
	})
	t.Run("empty interface slice", func(t *testing.T) {
		got, ok := GetSliceMapStringInterfaceFromInterface([]interface{}{})
		assert.True(t, ok)
		assert.Empty(t, got)
	})
	t.Run("heterogeneous slice rejected", func(t *testing.T) {
		in := []interface{}{
			map[string]interface{}{"a": 1},
			"not a map",
		}
		got, ok := GetSliceMapStringInterfaceFromInterface(in)
		assert.False(t, ok)
		assert.Nil(t, got)
	})
	for _, tc := range []struct {
		name  string
		value interface{}
	}{
		{"nil", nil},
		{"string", "x"},
		{"single map", map[string]interface{}{"a": 1}},
		{"slice of strings", []string{"a", "b"}},
	} {
		t.Run(tc.name+" rejected", func(t *testing.T) {
			got, ok := GetSliceMapStringInterfaceFromInterface(tc.value)
			assert.False(t, ok)
			assert.Nil(t, got)
		})
	}
}

func TestGetSliceInterfaceFromInterface(t *testing.T) {
	t.Run("slice of interface", func(t *testing.T) {
		in := []interface{}{1, "two", 3.0}
		got, ok := GetSliceInterfaceFromInterface(in)
		assert.True(t, ok)
		assert.Equal(t, in, got)
	})
	t.Run("empty slice", func(t *testing.T) {
		got, ok := GetSliceInterfaceFromInterface([]interface{}{})
		assert.True(t, ok)
		assert.Empty(t, got)
	})
	for _, tc := range []struct {
		name  string
		value interface{}
	}{
		{"nil", nil},
		{"string", "x"},
		{"typed slice", []string{"a"}},
		{"map", map[string]interface{}{"a": 1}},
	} {
		t.Run(tc.name+" rejected", func(t *testing.T) {
			got, ok := GetSliceInterfaceFromInterface(tc.value)
			assert.False(t, ok)
			assert.Nil(t, got)
		})
	}
}

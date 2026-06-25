package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONString(t *testing.T) {
	m := map[string]interface{}{"name": "abc", "num": 1.0}

	t.Run("present and correct type", func(t *testing.T) {
		v, present, err := JSONString(m, "name")
		require.NoError(t, err)
		assert.True(t, present)
		assert.Equal(t, "abc", v)
	})

	t.Run("absent key is optional", func(t *testing.T) {
		v, present, err := JSONString(m, "missing")
		require.NoError(t, err)
		assert.False(t, present)
		assert.Equal(t, "", v)
	})

	t.Run("present but wrong type errors", func(t *testing.T) {
		_, present, err := JSONString(m, "num")
		require.Error(t, err)
		assert.True(t, present)
		assert.Contains(t, err.Error(), "num must be a string")
	})
}

func TestJSONNumber(t *testing.T) {
	m := map[string]interface{}{"n": 10.0, "s": "x"}

	t.Run("present and correct type", func(t *testing.T) {
		v, present, err := JSONNumber(m, "n")
		require.NoError(t, err)
		assert.True(t, present)
		assert.Equal(t, 10.0, v)
	})

	t.Run("absent key is optional", func(t *testing.T) {
		_, present, err := JSONNumber(m, "missing")
		require.NoError(t, err)
		assert.False(t, present)
	})

	t.Run("present but wrong type errors", func(t *testing.T) {
		_, present, err := JSONNumber(m, "s")
		require.Error(t, err)
		assert.True(t, present)
		assert.Contains(t, err.Error(), "s must be a number")
	})
}

func TestJSONBool(t *testing.T) {
	m := map[string]interface{}{"b": true, "s": "x"}

	t.Run("present and correct type", func(t *testing.T) {
		v, present, err := JSONBool(m, "b")
		require.NoError(t, err)
		assert.True(t, present)
		assert.True(t, v)
	})

	t.Run("absent key is optional", func(t *testing.T) {
		_, present, err := JSONBool(m, "missing")
		require.NoError(t, err)
		assert.False(t, present)
	})

	t.Run("present but wrong type errors", func(t *testing.T) {
		_, present, err := JSONBool(m, "s")
		require.Error(t, err)
		assert.True(t, present)
		assert.Contains(t, err.Error(), "s must be a boolean")
	})
}

func TestJSONObject(t *testing.T) {
	m := map[string]interface{}{"o": map[string]interface{}{"k": "v"}, "s": "x"}

	t.Run("present and correct type", func(t *testing.T) {
		v, present, err := JSONObject(m, "o")
		require.NoError(t, err)
		assert.True(t, present)
		assert.Equal(t, "v", v["k"])
	})

	t.Run("absent key is optional", func(t *testing.T) {
		_, present, err := JSONObject(m, "missing")
		require.NoError(t, err)
		assert.False(t, present)
	})

	t.Run("present but wrong type errors", func(t *testing.T) {
		_, present, err := JSONObject(m, "s")
		require.Error(t, err)
		assert.True(t, present)
		assert.Contains(t, err.Error(), "s must be an object")
	})
}

func TestJSONArray(t *testing.T) {
	m := map[string]interface{}{"a": []interface{}{1.0, 2.0}, "s": "x"}

	t.Run("present and correct type", func(t *testing.T) {
		v, present, err := JSONArray(m, "a")
		require.NoError(t, err)
		assert.True(t, present)
		assert.Len(t, v, 2)
	})

	t.Run("absent key is optional", func(t *testing.T) {
		_, present, err := JSONArray(m, "missing")
		require.NoError(t, err)
		assert.False(t, present)
	})

	t.Run("present but wrong type errors", func(t *testing.T) {
		_, present, err := JSONArray(m, "s")
		require.Error(t, err)
		assert.True(t, present)
		assert.Contains(t, err.Error(), "s must be an array")
	})
}

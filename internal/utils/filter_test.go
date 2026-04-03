package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := Filter[int](nil, func(n int) bool { return true })
		assert.Nil(t, result)
	})

	t.Run("empty slice returns nil", func(t *testing.T) {
		result := Filter([]int{}, func(n int) bool { return true })
		assert.Nil(t, result)
	})

	t.Run("all elements match returns full slice", func(t *testing.T) {
		input := []int{1, 2, 3}
		result := Filter(input, func(n int) bool { return true })
		assert.Equal(t, input, result)
	})

	t.Run("no elements match returns nil", func(t *testing.T) {
		result := Filter([]int{1, 2, 3}, func(n int) bool { return false })
		assert.Nil(t, result)
	})

	t.Run("mixed - only matching elements returned", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		result := Filter(input, func(n int) bool { return n%2 == 0 })
		assert.Equal(t, []int{2, 4}, result)
	})

	t.Run("works with pointer element types", func(t *testing.T) {
		a, b, c := 1, 2, 3
		input := []*int{&a, &b, &c}
		result := Filter(input, func(p *int) bool { return *p > 1 })
		assert.Equal(t, []*int{&b, &c}, result)
	})

	t.Run("works with string type", func(t *testing.T) {
		input := []string{"apple", "banana", "apricot", "cherry"}
		result := Filter(input, func(s string) bool { return s[0] == 'a' })
		assert.Equal(t, []string{"apple", "apricot"}, result)
	})
}

package utils

// Filter returns a new slice containing only the elements of items for which
// predicate returns true. Returns nil (not an empty slice) when no elements
// match, consistent with the existing filter functions across this codebase.
// Callers need not special-case a nil result: len(nil) == 0 in Go.
func Filter[T any](items []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

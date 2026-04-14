package utils

import "strings"

// MatchesTagFilters returns true if the tags map satisfies all filter specs.
// Each filter is either "key=value" (exact match) or "key" (key-exists match).
// An empty filters slice always returns true.
func MatchesTagFilters(tags map[string]string, filters []string) bool {
	for _, f := range filters {
		if key, value, hasValue := strings.Cut(f, "="); hasValue {
			if v, ok := tags[key]; !ok || v != value {
				return false
			}
		} else {
			if _, ok := tags[f]; !ok {
				return false
			}
		}
	}
	return true
}

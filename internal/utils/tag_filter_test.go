package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchesTagFilters(t *testing.T) {
	tests := []struct {
		name    string
		tags    map[string]string
		filters []string
		want    bool
	}{
		{
			name:    "no filters always matches",
			tags:    map[string]string{"env": "prod"},
			filters: nil,
			want:    true,
		},
		{
			name:    "empty filters always matches",
			tags:    map[string]string{"env": "prod"},
			filters: []string{},
			want:    true,
		},
		{
			name:    "exact match hit",
			tags:    map[string]string{"env": "prod"},
			filters: []string{"env=prod"},
			want:    true,
		},
		{
			name:    "exact match miss - wrong value",
			tags:    map[string]string{"env": "staging"},
			filters: []string{"env=prod"},
			want:    false,
		},
		{
			name:    "exact match miss - key absent",
			tags:    map[string]string{"team": "net"},
			filters: []string{"env=prod"},
			want:    false,
		},
		{
			name:    "key-exists match hit",
			tags:    map[string]string{"env": "anything"},
			filters: []string{"env"},
			want:    true,
		},
		{
			name:    "key-exists match miss",
			tags:    map[string]string{"team": "net"},
			filters: []string{"env"},
			want:    false,
		},
		{
			name:    "AND logic - all match",
			tags:    map[string]string{"env": "prod", "team": "net"},
			filters: []string{"env=prod", "team=net"},
			want:    true,
		},
		{
			name:    "AND logic - one misses",
			tags:    map[string]string{"env": "prod", "team": "ops"},
			filters: []string{"env=prod", "team=net"},
			want:    false,
		},
		{
			name:    "mixed exact and key-exists",
			tags:    map[string]string{"env": "prod", "owner": "alice"},
			filters: []string{"env=prod", "owner"},
			want:    true,
		},
		{
			name:    "nil tags map with key filter",
			tags:    nil,
			filters: []string{"env"},
			want:    false,
		},
		{
			name:    "nil tags map with no filters",
			tags:    nil,
			filters: nil,
			want:    true,
		},
		{
			name:    "value containing equals sign",
			tags:    map[string]string{"key": "a=b"},
			filters: []string{"key=a=b"},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesTagFilters(tt.tags, tt.filters)
			assert.Equal(t, tt.want, got)
		})
	}
}

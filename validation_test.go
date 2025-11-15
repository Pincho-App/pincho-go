package wirepusher

import (
	"reflect"
	"testing"
)

func TestNormalizeTags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "lowercase conversion",
			input:    []string{"Production", "RELEASE", "Deploy"},
			expected: []string{"production", "release", "deploy"},
		},
		{
			name:     "whitespace trimming",
			input:    []string{"  production  ", "release", "  deploy"},
			expected: []string{"production", "release", "deploy"},
		},
		{
			name:     "duplicate removal case-insensitive",
			input:    []string{"production", "Production", "PRODUCTION", "release"},
			expected: []string{"production", "release"},
		},
		{
			name:     "empty string filtering",
			input:    []string{"production", "", "  ", "release"},
			expected: []string{"production", "release"},
		},
		{
			name:     "invalid characters filtered",
			input:    []string{"production", "invalid tag", "release@123", "deploy"},
			expected: []string{"production", "deploy"},
		},
		{
			name:     "valid special characters",
			input:    []string{"prod-1", "release_2", "deploy-tag_3"},
			expected: []string{"prod-1", "release_2", "deploy-tag_3"},
		},
		{
			name:     "all invalid tags",
			input:    []string{"invalid tag", "another bad", "  "},
			expected: nil,
		},
		{
			name:     "mixed valid and invalid",
			input:    []string{"Production", "  Release  ", "production", "Bad Tag", "deploy"},
			expected: []string{"production", "release", "deploy"},
		},
		{
			name:     "numbers allowed",
			input:    []string{"tag1", "tag2", "123", "abc123"},
			expected: []string{"tag1", "tag2", "123", "abc123"},
		},
		{
			name:     "preserves order",
			input:    []string{"zebra", "apple", "mango"},
			expected: []string{"zebra", "apple", "mango"},
		},
		{
			name:     "complex example from Python tests",
			input:    []string{"Production", "  Release  ", "production", "Deploy", "invalid tag"},
			expected: []string{"production", "release", "deploy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeTags(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("NormalizeTags(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

package wirepusher

import (
	"regexp"
	"strings"
)

var tagPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

// NormalizeTags normalizes tags by converting to lowercase, trimming whitespace,
// validating characters, and removing duplicates.
//
// Tags are normalized in the following way:
//   - Converted to lowercase
//   - Whitespace trimmed
//   - Only alphanumeric, hyphens, and underscores allowed
//   - Duplicates removed (case-insensitive)
//   - Empty tags filtered out
//
// Returns nil if the input is nil or if all tags are invalid.
//
// Example:
//
//	tags := []string{"Production", "  Release  ", "production", "Deploy"}
//	normalized := NormalizeTags(tags)
//	// Returns: []string{"production", "release", "deploy"}
func NormalizeTags(tags []string) []string {
	if tags == nil || len(tags) == 0 {
		return nil
	}

	var normalized []string
	seen := make(map[string]bool)

	for _, tag := range tags {
		// Lowercase and trim
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))

		// Skip empty tags
		if normalizedTag == "" {
			continue
		}

		// Skip duplicates (case-insensitive)
		if seen[normalizedTag] {
			continue
		}

		// Validate characters (alphanumeric, hyphens, underscores only)
		if !tagPattern.MatchString(normalizedTag) {
			continue
		}

		normalized = append(normalized, normalizedTag)
		seen[normalizedTag] = true
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

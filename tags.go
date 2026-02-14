package main

import (
	"regexp"
	"strings"
)

var tagRegex = regexp.MustCompile(`#[a-zA-Z0-9_-]+`)

// ExtractTags extracts hashtags from text and returns them in lowercase
func ExtractTags(text string) []string {
	matches := tagRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return []string{}
	}

	// Remove duplicates and convert to lowercase
	seen := make(map[string]bool)
	tags := []string{}

	for _, match := range matches {
		// Remove the # prefix and convert to lowercase
		tag := strings.ToLower(strings.TrimPrefix(match, "#"))
		if !seen[tag] {
			seen[tag] = true
			tags = append(tags, tag)
		}
	}

	return tags
}

// HighlightTags returns text with tag positions for highlighting
// Returns the text and a slice of [start, end] positions for each tag
func HighlightTags(text string) (string, [][2]int) {
	matches := tagRegex.FindAllStringIndex(text, -1)
	positions := make([][2]int, len(matches))
	for i, match := range matches {
		positions[i] = [2]int{match[0], match[1]}
	}
	return text, positions
}

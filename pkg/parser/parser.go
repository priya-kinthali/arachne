package parser

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ExtractTitle extracts title from HTML or JSON responses
func ExtractTitle(content, contentType string) string {
	// Check if it's JSON based on content type or content
	if strings.Contains(contentType, "application/json") ||
		(strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[")) {
		return ExtractJSONTitle(content)
	}

	// Otherwise treat as HTML
	return ExtractHTMLTitle(content)
}

// ExtractHTMLTitle extracts title from HTML
func ExtractHTMLTitle(html string) string {
	// Look for <title> tag
	titleStart := strings.Index(strings.ToLower(html), "<title>")
	if titleStart == -1 {
		return "No HTML title found"
	}

	titleStart += 7 // length of "<title>"
	titleEnd := strings.Index(html[titleStart:], "</title>")
	if titleEnd == -1 {
		return "Malformed HTML title"
	}

	title := html[titleStart : titleStart+titleEnd]
	title = strings.TrimSpace(title)

	if title == "" {
		return "Empty HTML title"
	}

	return title
}

// ExtractJSONTitle extracts meaningful title from JSON responses
func ExtractJSONTitle(jsonStr string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "Invalid JSON"
	}

	// Look for common title fields in JSON
	titleFields := []string{"title", "name", "login", "message", "description"}
	for _, field := range titleFields {
		if value, exists := data[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				return str
			}
		}
	}

	// If no title field, return first meaningful string value in sorted order
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := data[key]
		if str, ok := value.(string); ok && len(str) < 100 && str != "" {
			return fmt.Sprintf("%s: %s", key, str)
		}
	}

	return "JSON response (no title field)"
}

package util

import (
	"html"
	"strings"
)

// Sanitize trims whitespace and escapes HTML special characters.
func Sanitize(s string) string {
	return html.EscapeString(strings.TrimSpace(s))
}

package util

import "strings"

func SanitizeString(input string) string {
	// Replace any non-printable characters with an empty string
	sanitized := ""
	for _, r := range input {
		if r >= 32 && r <= 126 { // ASCII printable characters
			sanitized += string(r)
		}
	}
	sanitized = strings.TrimSpace(sanitized)
	sanitized = strings.ReplaceAll(sanitized, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	return sanitized
}

package logging

import (
	"strings"
)

func SanitizeString(input string) string {
	// Replace any non-printable characters with an empty string
	sanitised := strings.ReplaceAll(input, `"`, "")
	sanitised = strings.TrimSpace(sanitised)
	sanitised = strings.ReplaceAll(sanitised, "\n", "")
	sanitised = strings.ReplaceAll(sanitised, "\r", "")
	return sanitised
}

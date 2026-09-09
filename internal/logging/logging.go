package logging

import (
	"strconv"
	"strings"
)

func SanitizeString(input string) string {
	// Replace any non-printable characters with an empty string
	sanitised := strconv.QuoteToASCII(input)
	sanitised = strings.TrimSpace(sanitised)
	sanitised = strings.ReplaceAll(sanitised, "\n", "")
	sanitised = strings.ReplaceAll(sanitised, "\r", "")
	return sanitised
}

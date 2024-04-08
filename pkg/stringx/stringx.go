package stringx

import "unicode/utf8"

// TruncateString truncates a string to ensure it does not exceed maxByteLength bytes.
// It respects UTF-8 encoding by not cutting a character in the middle.
func TruncateString(str string, maxByteLength int) string {
	if len(str) <= maxByteLength {
		return str
	}

	truncated := str[:maxByteLength]
	for len(truncated) > 0 {
		if !utf8.ValidString(truncated) {
			// Remove the last byte until the string is valid UTF-8
			truncated = truncated[:len(truncated)-1]
			continue
		}
		break
	}

	const suffixLen = 3
	if runes := []rune(truncated); len(runes) > suffixLen {
		runes = runes[:len(runes)-suffixLen]
		truncated = string(runes) + "..."
	}
	return truncated
}

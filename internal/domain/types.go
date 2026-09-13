package domain

import "strings"

// NormalizeLookupName produces a case-insensitive, whitespace-normalized
// key for entity and project name indexes.
func NormalizeLookupName(value string) string {
	if value == "" {
		return ""
	}
	b := make([]byte, 0, len(value))
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c >= 'A' && c <= 'Z' {
			b = append(b, c+'a'-'A')
		} else if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			if len(b) > 0 && b[len(b)-1] != ' ' {
				b = append(b, ' ')
			}
		} else {
			b = append(b, c)
		}
	}
	return strings.TrimSpace(string(b))
}
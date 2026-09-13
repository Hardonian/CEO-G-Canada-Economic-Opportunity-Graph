package identity

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

// StableID returns a deterministic UUID for a canonical resource key. The
// namespace string is deliberately part of the input so identifiers remain
// stable across ordinary renames while distinct source systems cannot collide.
func StableID(resourceType, namespace, externalID string) string {
	key := strings.Join([]string{
		strings.ToLower(strings.TrimSpace(resourceType)),
		strings.ToLower(strings.TrimSpace(namespace)),
		strings.ToLower(strings.TrimSpace(externalID)),
	}, ":")
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("https://canadaopportunitygraph.ca/"+key)).String()
}

// Slug creates a conservative ASCII URL slug. Stable IDs, not slugs, are the
// canonical identity and therefore survive future name changes.
func Slug(value string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) && r <= unicode.MaxASCII || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return strings.Trim(nonSlug.ReplaceAllString(b.String(), "-"), "-")
}

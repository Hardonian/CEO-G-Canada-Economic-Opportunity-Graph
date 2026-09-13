package identity

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

var latinSlugFold = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ä", "a", "ã", "a", "å", "a",
	"æ", "ae", "ç", "c", "è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i", "ñ", "n", "ò", "o",
	"ó", "o", "ô", "o", "ö", "o", "õ", "o", "œ", "oe", "ù", "u",
	"ú", "u", "û", "u", "ü", "u", "ý", "y", "ÿ", "y",
)

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
	value = latinSlugFold.Replace(strings.ToLower(strings.TrimSpace(value)))
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) && r <= unicode.MaxASCII || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return strings.Trim(nonSlug.ReplaceAllString(b.String(), "-"), "-")
}

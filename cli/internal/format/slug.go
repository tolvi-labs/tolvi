package format

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const slugMaxLen = 80

// SlugFromTitle converts a free-form title into a slug matching the
// format spec's slug rule: `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`.
//
// Rules applied in order:
//  1. Unicode NFKD normalize, then drop combining marks (ASCII-fold accents)
//  2. Lowercase
//  3. Replace any run of non-[a-z0-9] chars with a single "-"
//  4. Trim leading/trailing "-"
//  5. Truncate to 80 chars
//  6. Trim trailing "-" again if truncation left one
//
// Empty input yields empty output (caller's responsibility to handle).
func SlugFromTitle(title string) string {
	if title == "" {
		return ""
	}

	// 1. ASCII-fold via NFKD + drop combining marks.
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	folded, _, err := transform.String(t, title)
	if err != nil {
		folded = title
	}

	// 2. Lowercase.
	folded = strings.ToLower(folded)

	// 3. Replace non-alphanumeric runs with "-".
	var b strings.Builder
	b.Grow(len(folded))
	prevDash := false
	for _, r := range folded {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
		} else {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	s := b.String()

	// 4. Trim leading/trailing "-".
	s = strings.Trim(s, "-")

	// 5. Truncate.
	if len(s) > slugMaxLen {
		s = s[:slugMaxLen]
	}

	// 6. Trim trailing "-" post-truncation.
	s = strings.TrimRight(s, "-")

	return s
}

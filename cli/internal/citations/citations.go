// Package citations extracts [[slug]] references from LLM output and
// verifies them against a known set of vault slugs.
package citations

import "regexp"

// citationRe matches the slug shape from spec/schemas/decision.json
// (and friends): `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`. Same shape used
// by the server (server/src/ask/citations.ts).
var citationRe = regexp.MustCompile(`\[\[([a-z0-9](?:[a-z0-9-]*[a-z0-9])?)\]\]`)

// Extract returns the deduplicated, in-order list of [[slug]] refs
// found in text. Order is first-occurrence order in the input.
func Extract(text string) []string {
	matches := citationRe.FindAllStringSubmatch(text, -1)
	seen := map[string]bool{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		slug := m[1]
		if seen[slug] {
			continue
		}
		seen[slug] = true
		out = append(out, slug)
	}
	return out
}

// Verify partitions slugs into (matched, unmatched) given a set of
// known slugs (typically built from the loaded vault).
func Verify(slugs []string, known map[string]bool) (matched, unmatched []string) {
	for _, s := range slugs {
		if known[s] {
			matched = append(matched, s)
		} else {
			unmatched = append(unmatched, s)
		}
	}
	return
}

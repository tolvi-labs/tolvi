package llm

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// DefaultSurfacedStatuses matches the server's DEFAULT_SURFACED_STATUSES
// (see server/src/search/ranking.ts). Statuses surfaced when no
// --include-status flag is set.
var DefaultSurfacedStatuses = []string{"active", "in-progress", "historical"}

// AssembleOpts controls prompt assembly.
type AssembleOpts struct {
	Workspace   string
	Docs        []vault.Doc
	GeneratedAt time.Time

	// Status filter; empty means use DefaultSurfacedStatuses.
	IncludeStatuses []string

	// Type filter; empty means include all 3 types.
	ExcludeTypes []string
}

// AssemblePrompt produces the full system-prompt body (instructions
// + <vault> XML block). The output is deterministic for a given input
// (modulo GeneratedAt truncation to the minute) so it can be reasoned
// about for prompt-cache behavior.
func AssemblePrompt(opts AssembleOpts) string {
	statusSet := map[string]bool{}
	if len(opts.IncludeStatuses) == 0 {
		for _, s := range DefaultSurfacedStatuses {
			statusSet[s] = true
		}
	} else {
		for _, s := range opts.IncludeStatuses {
			statusSet[s] = true
		}
	}
	excludeTypeSet := map[string]bool{}
	for _, t := range opts.ExcludeTypes {
		excludeTypeSet[t] = true
	}

	var b strings.Builder
	b.WriteString(instructions)
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "<vault workspace=%q generated_at=%q>\n",
		opts.Workspace,
		opts.GeneratedAt.UTC().Truncate(time.Minute).Format(time.RFC3339),
	)
	for _, d := range opts.Docs {
		status := d.Frontmatter.String("status")
		if status == "" {
			status = "active"
		}
		if !statusSet[status] {
			continue
		}
		if excludeTypeSet[d.Type] {
			continue
		}

		attrs := []string{
			fmt.Sprintf(`slug=%q`, d.Slug),
			fmt.Sprintf(`type=%q`, d.Type),
			fmt.Sprintf(`status=%q`, status),
		}
		if repo := d.Frontmatter.String("repo"); repo != "" {
			attrs = append(attrs, fmt.Sprintf(`repo=%q`, repo))
		}
		if d.Date != "" {
			attrs = append(attrs, fmt.Sprintf(`date=%q`, d.Date))
		}
		attrs = append(attrs, fmt.Sprintf(`path=%q`, d.Path))

		fmt.Fprintf(&b, "  <doc %s>\n", strings.Join(attrs, " "))
		body := string(d.Body)
		// Defensive: escape literal </doc> in body if any (rare).
		if strings.Contains(body, "</doc>") {
			body = strings.ReplaceAll(body, "</doc>", html.EscapeString("</doc>"))
		}
		b.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			b.WriteString("\n")
		}
		b.WriteString("  </doc>\n\n")
	}
	b.WriteString("</vault>\n")
	return b.String()
}

const instructions = `You are an assistant that answers questions about a software project's engineering knowledge vault. The vault contains three doc types:

- DECISION: a non-obvious choice the team has made
- SESSION: a dated work-session log
- PATTERN: a reusable technique that outlives any single decision

Below is the entire vault. Use it to answer the user's question.

REQUIREMENTS:
- Cite every claim with [[slug]] of the doc it came from. Cite by slug only.
- Never invent slugs not present in the vault below.
- If the vault doesn't answer the question, say so honestly.
- Use [[wiki-link]] syntax exactly: double square brackets around the slug.
- Prefer durable docs (decisions, patterns) over session logs when both apply.
- Match the project's voice: factual, concise.`

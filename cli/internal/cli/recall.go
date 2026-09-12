package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

const (
	RecallDefaultSessionCount  = 3
	RecallDefaultDecisionCount = 10
	RecallDefaultMaxBytes      = 8000
)

// recallSkipStatuses mirrors the /recall skill's status filter: skip
// superseded, deprecated, and draft. active, in-progress, and historical
// are all surfaced.
var recallSkipStatuses = map[string]bool{
	"superseded": true,
	"deprecated": true,
	"draft":      true,
}

// RecallOpts controls the recall command.
type RecallOpts struct {
	VaultPath       string
	SessionCount    int    // 0 → RecallDefaultSessionCount
	DecisionCount   int    // 0 → RecallDefaultDecisionCount
	MaxBytes        int    // 0 → unlimited; caps hook-json additionalContext length
	IncludePatterns bool   // false by default; patterns add bulk without session-resumption value
	Format          string // "" | "human" | "hook-json"
	Stdout          io.Writer
}

type recallSession struct {
	Date     string
	Heading  string // text of last "## [HH:MM] Session — ..." in the file
	Content  []byte // full file content (included in hook-json)
	priority int    // higher wins when two sources carry the same date
}

type recallDecision struct {
	Slug     string
	Title    string
	Status   string
	Oneliner string // TL;DR first line, or Why section opening, or first body line
}

// RunRecall loads recent sessions and active decisions from the vault and
// writes a structured recall summary. Format "hook-json" emits the
// Claude Code SessionStart hook JSON blob; "" or "human" emits the
// RECALL SUMMARY text block.
func RunRecall(opts RecallOpts) error {
	if opts.SessionCount <= 0 {
		opts.SessionCount = RecallDefaultSessionCount
	}
	if opts.DecisionCount <= 0 {
		opts.DecisionCount = RecallDefaultDecisionCount
	}

	sessionSources, decisionSources := recallSources(opts.VaultPath)
	sessions, err := recallLoadSessions(sessionSources, opts.SessionCount)
	if err != nil {
		return err
	}
	decisions, filteredOut, err := recallLoadDecisions(decisionSources, opts.DecisionCount)
	if err != nil {
		return err
	}

	switch opts.Format {
	case "hook-json":
		return recallWriteHookJSON(opts.Stdout, opts.VaultPath, sessions, decisions, filteredOut, opts.MaxBytes)
	default:
		return recallWriteHuman(opts.Stdout, sessions, decisions, filteredOut)
	}
}

// recallSessionSource is one directory recall reads session notes from. The
// private root of a routed vault is shared by every repo in the org, so suffix
// restricts the read to this workspace's own notes; an empty suffix takes every
// *.md, which is what a repo's own vault wants.
type recallSessionSource struct {
	dir      string
	suffix   string
	priority int
}

// recallDecisionSource is one directory recall reads decisions from. repo, when
// set, keeps only decisions whose frontmatter `repo` matches it — the private
// vault holds every repo's decisions, a repo's own vault only its own.
type recallDecisionSource struct {
	dir  string
	repo string
}

// recallSources resolves which vault roots recall reads. A vault with a local
// routing config keeps its session notes in the private vault and splits its
// decisions across both roots, so recall has to read both or it reports a
// stale local note as the latest session.
//
// Meta is read softly: recall is a read-only path that runs as a SessionStart
// hook, so a broken routing config degrades to local-only here rather than
// failing the session. The loud failure belongs to sync and commit, where
// writing into the wrong vault is what actually does harm.
func recallSources(vaultPath string) ([]recallSessionSource, []recallDecisionSource) {
	sessions := []recallSessionSource{{dir: filepath.Join(vaultPath, "sessions")}}
	decisions := []recallDecisionSource{{dir: filepath.Join(vaultPath, "decisions")}}

	meta, err := vault.ReadMeta(vaultPath)
	if err != nil {
		return sessions, decisions
	}
	privRoot, routed, err := vault.ResolveDocDestination(vaultPath, meta, "session", "")
	if err != nil || !routed {
		return sessions, decisions
	}
	sessions = append(sessions, recallSessionSource{
		dir:      filepath.Join(privRoot, "sessions"),
		suffix:   "-" + meta.Workspace + ".md",
		priority: 1,
	})
	decisions = append(decisions, recallDecisionSource{
		dir:  filepath.Join(privRoot, "decisions"),
		repo: meta.Workspace,
	})
	return sessions, decisions
}

// recallLoadSessions reads the N most recent session notes across sources,
// newest first. Filenames are YYYY-MM-DD[-workspace].md, so the leading date
// sorts chronologically.
func recallLoadSessions(sources []recallSessionSource, count int) ([]recallSession, error) {
	var out []recallSession
	for _, src := range sources {
		entries, err := os.ReadDir(src.dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read sessions: %w", err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".md") {
				continue
			}
			if src.suffix != "" && !strings.HasSuffix(name, src.suffix) {
				continue
			}
			content, err := os.ReadFile(filepath.Join(src.dir, name))
			if err != nil {
				continue // soft error — skip unreadable files without aborting
			}
			out = append(out, recallSession{
				Date:     recallSessionDate(name),
				Heading:  recallExtractSessionHeading(content),
				Content:  content,
				priority: src.priority,
			})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Date != out[j].Date {
			return out[i].Date > out[j].Date
		}
		return out[i].priority > out[j].priority
	})
	if count < len(out) {
		out = out[:count]
	}
	return out, nil
}

// recallSessionDate strips the .md extension and any -workspace suffix a
// routed session note carries, leaving the leading YYYY-MM-DD.
func recallSessionDate(name string) string {
	base := strings.TrimSuffix(name, ".md")
	const dateLen = len("YYYY-MM-DD")
	if len(base) > dateLen && base[dateLen] == '-' {
		return base[:dateLen]
	}
	return base
}

// recallLoadDecisions reads decisions across sources and returns those that
// pass the status filter, up to maxCount, newest first. The second return is
// the count hidden by the status filter; decisions belonging to another repo
// are not counted, since they were never candidates for this vault.
func recallLoadDecisions(sources []recallDecisionSource, maxCount int) ([]recallDecision, int, error) {
	type candidate struct {
		name string
		dir  string
		repo string
	}
	var candidates []candidate
	for _, src := range sources {
		entries, err := os.ReadDir(src.dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, 0, fmt.Errorf("read decisions: %w", err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				candidates = append(candidates, candidate{name: e.Name(), dir: src.dir, repo: src.repo})
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].name > candidates[j].name })

	var out []recallDecision
	filteredOut := 0

	for _, c := range candidates {
		data, err := os.ReadFile(filepath.Join(c.dir, c.name))
		if err != nil {
			continue
		}
		fm, body, err := format.ParseFrontmatter(data)
		if err != nil {
			continue
		}
		if c.repo != "" && fm.String("repo") != c.repo {
			continue // another repo's decision in the shared private vault
		}
		status := fm.String("status")
		if status == "" {
			status = "active"
		}
		if recallSkipStatuses[status] {
			filteredOut++
			continue
		}

		// Slug: strip the YYYY-MM-DD- prefix from the filename.
		slug := strings.TrimSuffix(c.name, ".md")
		const datePrefixLen = len("YYYY-MM-DD-") // 11
		if len(slug) > datePrefixLen {
			slug = slug[datePrefixLen:]
		}

		out = append(out, recallDecision{
			Slug:     slug,
			Title:    recallExtractTitle(body),
			Status:   status,
			Oneliner: recallExtractOneliner(body),
		})
		if len(out) >= maxCount {
			break
		}
	}
	return out, filteredOut, nil
}

var sessionHeadingRe = regexp.MustCompile(`(?m)^## \[\d{2}:\d{2}\] Session — (.+)$`)

// recallExtractSessionHeading returns the summary text from the last
// "## [HH:MM] Session — <summary>" heading in the file. Multiple blocks
// may be appended to a single session file; the last one is most recent.
func recallExtractSessionHeading(content []byte) string {
	matches := sessionHeadingRe.FindAllSubmatch(content, -1)
	if len(matches) == 0 {
		return ""
	}
	return string(matches[len(matches)-1][1])
}

var h1Re = regexp.MustCompile(`(?m)^# (.+)$`)

// recallExtractTitle returns the text of the first level-1 heading.
func recallExtractTitle(body []byte) string {
	m := h1Re.FindSubmatch(body)
	if m == nil {
		return ""
	}
	return string(m[1])
}

var tldrSectionRe = regexp.MustCompile(`(?ms)^## TL;DR\s*\n(.+?)(?:\n## |\z)`)
var whySectionRe = regexp.MustCompile(`(?ms)^## Why\s*\n(.+?)(?:\n## |\z)`)

// recallExtractOneliner returns the most useful one-line summary of a
// decision body. Priority: TL;DR first line → Why section first line →
// first non-empty body line.
func recallExtractOneliner(body []byte) string {
	for _, re := range []*regexp.Regexp{tldrSectionRe, whySectionRe} {
		m := re.FindSubmatch(body)
		if m != nil {
			if line := recallFirstNonEmptyLine(m[1]); line != "" {
				return recallTruncate(line, 120)
			}
		}
	}
	return recallTruncate(recallFirstNonEmptyLine(body), 120)
}

func recallFirstNonEmptyLine(b []byte) string {
	for _, line := range bytes.Split(b, []byte("\n")) {
		s := strings.TrimSpace(string(line))
		if s != "" && !strings.HasPrefix(s, "#") {
			return s
		}
	}
	return ""
}

func recallTruncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	// "..." is 3 bytes; trim to max-3 so the total stays within max.
	if max < 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

const recallSeparator = "──────────────────────────────────────────"

func recallWriteHuman(w io.Writer, sessions []recallSession, decisions []recallDecision, filteredOut int) error {
	fmt.Fprintln(w, "RECALL SUMMARY")
	fmt.Fprintln(w, recallSeparator)

	switch len(sessions) {
	case 0:
		fmt.Fprintln(w, "Last session:     (none)")
		fmt.Fprintln(w, "Prior session:    (none)")
	case 1:
		fmt.Fprintf(w, "Last session:     %s — %s\n", sessions[0].Date, sessions[0].Heading)
		fmt.Fprintln(w, "Prior session:    (none)")
	default:
		fmt.Fprintf(w, "Last session:     %s — %s\n", sessions[0].Date, sessions[0].Heading)
		fmt.Fprintf(w, "Prior session:    %s — %s\n", sessions[1].Date, sessions[1].Heading)
	}

	if len(decisions) == 0 {
		fmt.Fprintln(w, "Decisions:        none")
	} else {
		fmt.Fprintf(w, "Decisions:        %d relevant\n", len(decisions))
		for _, d := range decisions {
			if d.Status != "active" {
				fmt.Fprintf(w, "  %s — %s  (status: %s)\n", d.Slug, d.Title, d.Status)
			} else {
				fmt.Fprintf(w, "  %s — %s\n", d.Slug, d.Title)
			}
		}
	}

	if filteredOut > 0 {
		fmt.Fprintf(w, "Filtered out:     %d hidden by status filter\n", filteredOut)
	} else {
		fmt.Fprintln(w, "Filtered out:     none")
	}
	fmt.Fprintln(w, "Patterns:         (not loaded at recall — query on demand with tolvi ask)")
	fmt.Fprintln(w, recallSeparator)
	return nil
}

type recallHookOutput struct {
	HookSpecificOutput recallHookSpecific `json:"hookSpecificOutput"`
}

type recallHookSpecific struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

// recallWriteHookJSON emits the Claude Code SessionStart hook JSON blob.
// The additionalContext includes a compact summary followed by full
// session content — trimmed to maxBytes when set (0 = unlimited).
func recallWriteHookJSON(w io.Writer, vaultPath string, sessions []recallSession, decisions []recallDecision, filteredOut int, maxBytes int) error {
	var sb strings.Builder

	fmt.Fprintf(&sb, "TOLVI VAULT RECALL (auto-injected at session start)\nVault: %s\n\n", vaultPath)

	// Compact summary — always included first so it survives any truncation.
	sb.WriteString("=== SUMMARY ===\n")
	switch len(sessions) {
	case 0:
		sb.WriteString("Last session: (none)\n")
	case 1:
		fmt.Fprintf(&sb, "Last session: %s — %s\n", sessions[0].Date, sessions[0].Heading)
	default:
		fmt.Fprintf(&sb, "Last session: %s — %s\n", sessions[0].Date, sessions[0].Heading)
		fmt.Fprintf(&sb, "Prior session: %s — %s\n", sessions[1].Date, sessions[1].Heading)
	}
	if len(decisions) > 0 {
		fmt.Fprintf(&sb, "Active decisions: %d\n", len(decisions))
	}
	sb.WriteString("\n")

	// Decision summaries (compact — no full body).
	if len(decisions) > 0 {
		sb.WriteString("=== ACTIVE DECISIONS ===\n")
		for _, d := range decisions {
			line := d.Slug
			if d.Title != "" {
				line += " — " + d.Title
			}
			if d.Status != "active" {
				line += " (" + d.Status + ")"
			}
			sb.WriteString(line + "\n")
			if d.Oneliner != "" {
				sb.WriteString("  " + d.Oneliner + "\n")
			}
		}
		sb.WriteString("\n")
	}

	// Full session file content — most context value, placed last so
	// maxBytes truncation removes it before the summary.
	if len(sessions) > 0 {
		sb.WriteString("=== RECENT SESSIONS ===\n")
		for _, s := range sessions {
			fmt.Fprintf(&sb, "--- %s ---\n", s.Date)
			sb.Write(s.Content)
			sb.WriteString("\n")
		}
	}

	content := sb.String()
	if maxBytes > 0 && len(content) > maxBytes {
		content = content[:maxBytes] + "\n[...truncated — run /recall for full context]"
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(recallHookOutput{
		HookSpecificOutput: recallHookSpecific{
			HookEventName:     "SessionStart",
			AdditionalContext: content,
		},
	})
}

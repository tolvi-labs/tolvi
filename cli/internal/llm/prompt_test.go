package llm

import (
	"strings"
	"testing"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

func mkDoc(docType, slug, date, path string, body string) vault.Doc {
	return vault.Doc{
		Type: docType, Slug: slug, Date: date, Path: path,
		Body: []byte(body),
		Frontmatter: map[string]any{
			"status": "active",
			"repo":   "my-project",
		},
	}
}

func TestAssemblePrompt_ContainsAllDocs(t *testing.T) {
	docs := []vault.Doc{
		mkDoc("decision", "postgres", "2026-04-12", "decisions/2026-04-12-postgres.md", "# Postgres\n\nbecause pgvector."),
		mkDoc("pattern", "idempotent", "", "patterns/idempotent.md", "# Pattern\n\nbody."),
		mkDoc("session", "2026-04-12", "2026-04-12", "sessions/2026-04-12.md", "# Session\n\nlog."),
	}
	out := AssemblePrompt(AssembleOpts{
		Workspace:   "my-project",
		Docs:        docs,
		GeneratedAt: time.Date(2026, 5, 14, 17, 23, 45, 0, time.UTC),
	})

	for _, want := range []string{
		`<doc slug="postgres"`,
		`type="decision"`,
		`status="active"`,
		`<doc slug="idempotent"`,
		`type="pattern"`,
		`<doc slug="2026-04-12"`,
		`type="session"`,
		`because pgvector`,
		`<vault workspace="my-project"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestAssemblePrompt_GeneratedAtMinuteTruncate(t *testing.T) {
	t1 := time.Date(2026, 5, 14, 17, 23, 15, 0, time.UTC)
	t2 := time.Date(2026, 5, 14, 17, 23, 45, 0, time.UTC)

	out1 := AssemblePrompt(AssembleOpts{Workspace: "w", Docs: []vault.Doc{}, GeneratedAt: t1})
	out2 := AssemblePrompt(AssembleOpts{Workspace: "w", Docs: []vault.Doc{}, GeneratedAt: t2})
	if out1 != out2 {
		t.Errorf("same-minute outputs differ — minute-truncation broken")
	}

	t3 := time.Date(2026, 5, 14, 17, 24, 0, 0, time.UTC)
	out3 := AssemblePrompt(AssembleOpts{Workspace: "w", Docs: []vault.Doc{}, GeneratedAt: t3})
	if out1 == out3 {
		t.Errorf("different-minute outputs match — minute-truncation broken")
	}
}

func TestAssemblePrompt_StatusFilter_DefaultExcludesSuperseded(t *testing.T) {
	docs := []vault.Doc{
		{Type: "decision", Slug: "active-one", Path: "decisions/x.md", Body: []byte("active"),
			Frontmatter: map[string]any{"status": "active"}},
		{Type: "decision", Slug: "superseded-one", Path: "decisions/y.md", Body: []byte("superseded"),
			Frontmatter: map[string]any{"status": "superseded"}},
	}
	out := AssemblePrompt(AssembleOpts{
		Workspace: "w", Docs: docs, GeneratedAt: time.Now(),
	})
	if !strings.Contains(out, "active-one") {
		t.Errorf("active doc missing")
	}
	if strings.Contains(out, "superseded-one") {
		t.Errorf("superseded doc should be excluded by default")
	}
}

func TestAssemblePrompt_StatusFilter_All(t *testing.T) {
	docs := []vault.Doc{
		{Type: "decision", Slug: "a", Path: "x.md", Body: []byte("a"),
			Frontmatter: map[string]any{"status": "active"}},
		{Type: "decision", Slug: "s", Path: "y.md", Body: []byte("s"),
			Frontmatter: map[string]any{"status": "superseded"}},
	}
	out := AssemblePrompt(AssembleOpts{
		Workspace: "w", Docs: docs, GeneratedAt: time.Now(),
		IncludeStatuses: []string{"active", "in-progress", "superseded", "deprecated", "draft", "historical"},
	})
	if !strings.Contains(out, `slug="a"`) || !strings.Contains(out, `slug="s"`) {
		t.Errorf("--include-status all should include both: %s", out)
	}
}

func TestAssemblePrompt_TypeFilter(t *testing.T) {
	docs := []vault.Doc{
		{Type: "decision", Slug: "d", Path: "x.md", Body: []byte("d"), Frontmatter: map[string]any{"status": "active"}},
		{Type: "session", Slug: "s", Path: "y.md", Body: []byte("s"), Frontmatter: map[string]any{"status": "active"}},
	}
	out := AssemblePrompt(AssembleOpts{
		Workspace: "w", Docs: docs, GeneratedAt: time.Now(),
		ExcludeTypes: []string{"session"},
	})
	if !strings.Contains(out, `slug="d"`) {
		t.Errorf("decision should remain: %s", out)
	}
	if strings.Contains(out, `slug="s"`) {
		t.Errorf("session should be excluded: %s", out)
	}
}

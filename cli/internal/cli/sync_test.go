package cli

import (
	"testing"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
)

func TestAssembleFrontmatter_Decision(t *testing.T) {
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   "decision",
		Title:     "Why we chose Postgres",
		Slug:      "why-we-chose-postgres",
		Workspace: "tolvi",
		Date:      time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		Status:    "active",
	})
	if fm.String("status") != "active" {
		t.Errorf("status = %q", fm.String("status"))
	}
	if fm.String("repo") != "tolvi" {
		t.Errorf("repo = %q", fm.String("repo"))
	}
	if fm.String("date") != "2026-05-14" {
		t.Errorf("date = %q", fm.String("date"))
	}
	tags, _ := fm.Tags()
	if len(tags) != 1 || tags[0] != "decision" {
		t.Errorf("tags = %v", tags)
	}
}

func TestAssembleFrontmatter_Pattern_NoDateOrRepo(t *testing.T) {
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   "pattern",
		Title:     "Idempotent migrations",
		Slug:      "idempotent-migrations",
		Workspace: "tolvi",
		Date:      time.Now(),
		Status:    "active",
	})
	if _, ok := fm["date"]; ok {
		t.Errorf("pattern frontmatter should not include date: %v", fm)
	}
	if _, ok := fm["repo"]; ok {
		t.Errorf("pattern frontmatter should not include repo: %v", fm)
	}
}

func TestPathForDoc(t *testing.T) {
	tests := []struct {
		docType, slug, date string
		want                string
	}{
		{"decision", "postgres", "2026-04-12", "decisions/2026-04-12-postgres.md"},
		{"session", "", "2026-04-12", "sessions/2026-04-12.md"},
		{"pattern", "idempotent-migrations", "", "patterns/idempotent-migrations.md"},
	}
	for _, tt := range tests {
		got := PathForDoc(tt.docType, tt.slug, tt.date)
		if got != tt.want {
			t.Errorf("PathForDoc(%q, %q, %q) = %q, want %q",
				tt.docType, tt.slug, tt.date, got, tt.want)
		}
	}
}

// Silence unused-import if format is removed later.
var _ = format.Frontmatter{}

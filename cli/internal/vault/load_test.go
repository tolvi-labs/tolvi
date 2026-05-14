package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// seedDoc writes a Markdown file with the given frontmatter + body.
func seedDoc(t *testing.T, path, frontmatter, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\n" + frontmatter + "---\n\n" + body
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadAll_HappyPath(t *testing.T) {
	vault := t.TempDir()
	for _, d := range []string{"decisions", "sessions", "patterns"} {
		_ = os.MkdirAll(filepath.Join(vault, d), 0o755)
	}
	if err := WriteMeta(vault, Meta{Workspace: "test", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1}); err != nil {
		t.Fatal(err)
	}

	seedDoc(t, filepath.Join(vault, "decisions", "2026-04-12-postgres.md"),
		"tags: [decision]\ndate: 2026-04-12\nrepo: test\nstatus: active\n",
		"# postgres\n\nbody\n")
	seedDoc(t, filepath.Join(vault, "sessions", "2026-04-12.md"),
		"tags: [session]\ndate: 2026-04-12\nstatus: active\n",
		"## [09:00] Session — work\n\nstuff\n")
	seedDoc(t, filepath.Join(vault, "patterns", "idempotent.md"),
		"tags: [pattern]\nstatus: active\n",
		"# Pattern\n\nbody\n")

	docs, errs, err := LoadAll(vault)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(errs) != 0 {
		t.Errorf("unexpected per-doc errors: %v", errs)
	}
	if len(docs) != 3 {
		t.Fatalf("len(docs) = %d, want 3", len(docs))
	}
}

func TestLoadAll_SortOrder(t *testing.T) {
	vault := t.TempDir()
	for _, d := range []string{"decisions", "sessions", "patterns"} {
		_ = os.MkdirAll(filepath.Join(vault, d), 0o755)
	}
	_ = WriteMeta(vault, Meta{Workspace: "test", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1})

	seedDoc(t, filepath.Join(vault, "sessions", "2026-04-12.md"),
		"tags: [session]\ndate: 2026-04-12\nstatus: active\n", "body\n")
	seedDoc(t, filepath.Join(vault, "decisions", "2026-04-12-foo.md"),
		"tags: [decision]\ndate: 2026-04-12\nrepo: t\nstatus: active\n", "body\n")
	seedDoc(t, filepath.Join(vault, "patterns", "p.md"),
		"tags: [pattern]\nstatus: active\n", "body\n")

	docs, _, err := LoadAll(vault)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(docs) != 3 {
		t.Fatalf("len(docs) = %d, want 3", len(docs))
	}
	// Order: decisions → patterns → sessions
	gotOrder := []string{docs[0].Type, docs[1].Type, docs[2].Type}
	want := []string{"decision", "pattern", "session"}
	if !sliceEq(gotOrder, want) {
		t.Errorf("type order = %v, want %v", gotOrder, want)
	}
}

func TestLoadAll_PerFileErrorSkips(t *testing.T) {
	vault := t.TempDir()
	_ = os.MkdirAll(filepath.Join(vault, "decisions"), 0o755)
	_ = WriteMeta(vault, Meta{Workspace: "test", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1})

	// One valid, one malformed.
	seedDoc(t, filepath.Join(vault, "decisions", "2026-04-12-good.md"),
		"tags: [decision]\ndate: 2026-04-12\nrepo: t\nstatus: active\n", "body\n")
	if err := os.WriteFile(filepath.Join(vault, "decisions", "2026-04-12-bad.md"),
		[]byte("not frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	docs, errs, err := LoadAll(vault)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("len(docs) = %d, want 1", len(docs))
	}
	if len(errs) != 1 {
		t.Fatalf("len(errs) = %d, want 1", len(errs))
	}
}

func TestSlugFromFilename(t *testing.T) {
	tests := []struct {
		path     string
		docType  string
		wantSlug string
		wantDate string
	}{
		{"decisions/2026-04-12-postgres.md", "decision", "postgres", "2026-04-12"},
		{"sessions/2026-04-12.md", "session", "2026-04-12", "2026-04-12"},
		{"patterns/idempotent-migrations.md", "pattern", "idempotent-migrations", ""},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			slug, date := SlugAndDateFromPath(tt.path, tt.docType)
			if slug != tt.wantSlug {
				t.Errorf("slug = %q, want %q", slug, tt.wantSlug)
			}
			if date != tt.wantDate {
				t.Errorf("date = %q, want %q", date, tt.wantDate)
			}
		})
	}
}

func sliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

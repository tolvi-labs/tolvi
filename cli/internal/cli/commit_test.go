package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasSessionNote(t *testing.T) {
	vault := t.TempDir()
	sessions := filepath.Join(vault, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}

	const date = "2026-06-07"

	// No file → false.
	if hasSessionNote(vault, date) {
		t.Error("expected false when session file is absent")
	}

	// File exists but no "## " heading → false (frontmatter-only skeleton).
	skeleton := "---\ntags: [session]\ndate: 2026-06-07\nstatus: active\n---\n"
	if err := os.WriteFile(filepath.Join(sessions, date+".md"), []byte(skeleton), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasSessionNote(vault, date) {
		t.Error("expected false when file has no '## ' session block")
	}

	// File with a session block → true.
	withBlock := skeleton + "\n## [10:00] Session — did the thing\n\n### What happened\n- stuff\n"
	if err := os.WriteFile(filepath.Join(sessions, date+".md"), []byte(withBlock), 0o644); err != nil {
		t.Fatal(err)
	}
	if !hasSessionNote(vault, date) {
		t.Error("expected true when file has a '## ' session block")
	}

	// A different date is unaffected.
	if hasSessionNote(vault, "2026-06-06") {
		t.Error("expected false for a date with no file")
	}
}

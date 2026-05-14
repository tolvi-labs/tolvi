package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunSync_Session_NewFile(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	var out bytes.Buffer
	err := RunSync(SyncOpts{
		VaultPath: vaultDir,
		DocType:   "session",
		Date:      time.Date(2026, 5, 14, 9, 30, 0, 0, time.UTC),
		BodyFlag:  "## [09:30] Session — first work\n\nbody\n",
		Stdout:    &out,
	})
	if err != nil {
		t.Fatalf("RunSync: %v", err)
	}
	created := filepath.Join(vaultDir, "sessions", "2026-05-14.md")
	data, _ := os.ReadFile(created)
	if !bytes.Contains(data, []byte("## [09:30]")) {
		t.Errorf("session block missing: %s", data)
	}
}

func TestRunSync_Session_AppendsToExisting(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	day := time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC)

	// First session of the day.
	var out bytes.Buffer
	if err := RunSync(SyncOpts{
		VaultPath: vaultDir, DocType: "session", Date: day,
		BodyFlag: "## [09:00] Session — morning\n\nbody-A\n",
		Stdout:   &out,
	}); err != nil {
		t.Fatalf("first sync: %v", err)
	}

	// Second session, same day. Should APPEND, not refuse.
	out.Reset()
	if err := RunSync(SyncOpts{
		VaultPath: vaultDir, DocType: "session", Date: day,
		BodyFlag: "## [14:00] Session — afternoon\n\nbody-B\n",
		Stdout:   &out,
	}); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	merged, _ := os.ReadFile(filepath.Join(vaultDir, "sessions", "2026-05-14.md"))
	if !bytes.Contains(merged, []byte("## [09:00]")) {
		t.Errorf("first block missing after append: %s", merged)
	}
	if !bytes.Contains(merged, []byte("## [14:00]")) {
		t.Errorf("second block missing after append: %s", merged)
	}
	// The file should still have exactly ONE frontmatter section.
	if c := bytes.Count(merged, []byte("\n---\n")); c != 1 {
		t.Errorf("frontmatter delimiter count = %d, want 1 (one closing ---): %s", c, merged)
	}
}

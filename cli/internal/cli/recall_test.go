package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// writeVaultFile is a helper to create vault docs for recall tests.
func writeVaultFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ── extraction helpers ────────────────────────────────────────────────────

func TestRecallExtractSessionHeading(t *testing.T) {
	content := []byte("---\nstatus: active\n---\n\n## [09:00] Session — morning standup\n\nbody\n\n## [14:30] Session — afternoon review\n\nmore body\n")
	got := recallExtractSessionHeading(content)
	if got != "afternoon review" {
		t.Errorf("got %q, want %q", got, "afternoon review")
	}
}

func TestRecallExtractSessionHeading_Empty(t *testing.T) {
	if got := recallExtractSessionHeading([]byte("no headings here")); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestRecallExtractTitle(t *testing.T) {
	body := []byte("# My Decision Title\n\n## Why\nsome reason\n")
	if got := recallExtractTitle(body); got != "My Decision Title" {
		t.Errorf("got %q", got)
	}
}

func TestRecallExtractOneliner_TLDRWins(t *testing.T) {
	body := []byte("# Title\n\n## TL;DR\nAdopt the vault index pattern.\n\n## Why\nBecause reasons.\n")
	if got := recallExtractOneliner(body); got != "Adopt the vault index pattern." {
		t.Errorf("got %q", got)
	}
}

func TestRecallExtractOneliner_WhyFallback(t *testing.T) {
	body := []byte("# Title\n\n## Why\nThe problem is X.\n\n## How\nDo Y.\n")
	if got := recallExtractOneliner(body); got != "The problem is X." {
		t.Errorf("got %q", got)
	}
}

func TestRecallExtractOneliner_Truncates(t *testing.T) {
	long := strings.Repeat("a", 200)
	body := []byte("# T\n\n## Why\n" + long + "\n")
	got := recallExtractOneliner(body)
	if len(got) > 120 {
		t.Errorf("expected ≤120 chars, got %d", len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
}

// ── RunRecall human format ───────────────────────────────────────────────

func TestRunRecall_Human_EmptyVault(t *testing.T) {
	vault := mkVaultForTest(t)
	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: vault, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "RECALL SUMMARY") {
		t.Errorf("missing header: %s", got)
	}
	if !strings.Contains(got, "Last session:     (none)") {
		t.Errorf("expected no-session line: %s", got)
	}
	if !strings.Contains(got, "Decisions:        none") {
		t.Errorf("expected no-decisions line: %s", got)
	}
}

func TestRunRecall_Human_WithContent(t *testing.T) {
	vault := mkVaultForTest(t)

	writeVaultFile(t, vault, "sessions/2026-05-27.md",
		"---\nstatus: active\ndate: 2026-05-27\ntags: [session]\n---\n\n## [20:00] Session — launch kickoff\n\nbody\n")
	writeVaultFile(t, vault, "sessions/2026-05-24.md",
		"---\nstatus: active\ndate: 2026-05-24\ntags: [session]\n---\n\n## [10:00] Session — SDK shipped\n\nbody\n")
	writeVaultFile(t, vault, "decisions/2026-05-27-v0-launch-scope.md",
		"---\nstatus: active\ndate: 2026-05-27\nrepo: tolvi\ntags: [decision]\n---\n\n# v0.1.0 launch scope\n\n## TL;DR\nChannels locked for soft launch.\n\n## Why\nPhase 6 targets 5-15 devs.\n")
	// This one should be filtered out.
	writeVaultFile(t, vault, "decisions/2026-05-10-old-thing.md",
		"---\nstatus: superseded\ndate: 2026-05-10\nrepo: tolvi\ntags: [decision]\n---\n\n# Old decision\n")

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: vault, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}
	got := out.String()

	if !strings.Contains(got, "2026-05-27 — launch kickoff") {
		t.Errorf("last session missing: %s", got)
	}
	if !strings.Contains(got, "2026-05-24 — SDK shipped") {
		t.Errorf("prior session missing: %s", got)
	}
	if !strings.Contains(got, "v0-launch-scope") {
		t.Errorf("decision slug missing: %s", got)
	}
	if !strings.Contains(got, "Filtered out:     1") {
		t.Errorf("expected 1 filtered out: %s", got)
	}
	if !strings.Contains(got, "query on demand") {
		t.Errorf("patterns note missing: %s", got)
	}
}

func TestRunRecall_Human_SessionCountRespected(t *testing.T) {
	vault := mkVaultForTest(t)
	for _, date := range []string{"2026-05-01", "2026-05-02", "2026-05-03", "2026-05-04"} {
		writeVaultFile(t, vault, "sessions/"+date+".md",
			"---\nstatus: active\ndate: "+date+"\ntags: [session]\n---\n\n## [09:00] Session — work on "+date+"\n")
	}

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: vault, SessionCount: 2, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}
	got := out.String()
	if strings.Contains(got, "2026-05-01") {
		t.Errorf("oldest session should be excluded: %s", got)
	}
	if !strings.Contains(got, "2026-05-04") {
		t.Errorf("newest session should be included: %s", got)
	}
}

// ── RunRecall hook-json format ────────────────────────────────────────────

func TestRunRecall_HookJSON_ValidStructure(t *testing.T) {
	vault := mkVaultForTest(t)
	writeVaultFile(t, vault, "sessions/2026-05-27.md",
		"---\nstatus: active\ndate: 2026-05-27\ntags: [session]\n---\n\n## [20:00] Session — hook test\n\nbody\n")

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: vault, Format: "hook-json", Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out.String())
	}
	hso, ok := parsed["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("missing hookSpecificOutput: %v", parsed)
	}
	if hso["hookEventName"] != "SessionStart" {
		t.Errorf("hookEventName = %v", hso["hookEventName"])
	}
	ctx, _ := hso["additionalContext"].(string)
	if !strings.Contains(ctx, "TOLVI VAULT RECALL") {
		t.Errorf("additionalContext missing header: %s", ctx)
	}
	if !strings.Contains(ctx, "hook test") {
		t.Errorf("session heading missing from context: %s", ctx)
	}
}

func TestRunRecall_HookJSON_MaxBytesApplied(t *testing.T) {
	vault := mkVaultForTest(t)
	// Write a large session to trigger truncation.
	body := strings.Repeat("x", 5000)
	writeVaultFile(t, vault, "sessions/2026-05-27.md",
		"---\nstatus: active\ndate: 2026-05-27\ntags: [session]\n---\n\n## [09:00] Session — big\n\n"+body+"\n")

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: vault, Format: "hook-json", MaxBytes: 500, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(out.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	hso := parsed["hookSpecificOutput"].(map[string]any)
	ctx := hso["additionalContext"].(string)
	if len(ctx) > 600 { // 500 + truncation message + some slack
		t.Errorf("context too long (%d bytes), expected ≤600", len(ctx))
	}
	if !strings.Contains(ctx, "truncated") {
		t.Errorf("expected truncation notice: %s", ctx)
	}
}

// ── routed (public/private) vaults ────────────────────────────────────────
//
// On a repo with vault/.vault-routing.local.json, sync writes session notes
// to the private vault, so recall has to read them from there or it reports a
// stale local note as the latest session. The private vault is shared by every
// repo in the org, so recall must also filter to this workspace's own content.

// mkRoutedVaultForTest returns a public vault carrying a routing config that
// points at a fresh private vault, plus that private vault's path.
func mkRoutedVaultForTest(t *testing.T) (string, string) {
	t.Helper()
	pub := mkVaultForTest(t)
	priv := filepath.Join(t.TempDir(), "private-vault")
	for _, sub := range []string{"decisions", "sessions"} {
		if err := os.MkdirAll(filepath.Join(priv, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeVaultFile(t, pub, vault.RoutingConfigFileName, `{"private_vault": "`+priv+`"}`)
	return pub, priv
}

func sessionDoc(heading string) string {
	return "---\nstatus: active\ntags: [session]\n---\n\n## [10:00] Session — " + heading + "\n\nbody\n"
}

func decisionDoc(repo, status, title string) string {
	return "---\nstatus: " + status + "\nrepo: " + repo + "\ntags: [decision]\n---\n\n# " + title + "\n\n## TL;DR\nOne line.\n"
}

func TestRunRecall_RoutedVault_PrefersPrivateSession(t *testing.T) {
	pub, priv := mkRoutedVaultForTest(t)
	writeVaultFile(t, pub, "sessions/2026-06-07.md", sessionDoc("stale local note"))
	writeVaultFile(t, priv, "sessions/2026-09-12-test.md", sessionDoc("routed private note"))

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: pub, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Last session:     2026-09-12 — routed private note") {
		t.Errorf("private session not surfaced as latest: %s", got)
	}
	if !strings.Contains(got, "2026-06-07 — stale local note") {
		t.Errorf("legacy local session should still appear as prior: %s", got)
	}
}

func TestRunRecall_RoutedVault_IgnoresOtherWorkspaces(t *testing.T) {
	pub, priv := mkRoutedVaultForTest(t)
	writeVaultFile(t, priv, "sessions/2026-09-12-test.md", sessionDoc("our note"))
	// Newer, but belongs to a sibling repo sharing the private vault.
	writeVaultFile(t, priv, "sessions/2026-09-13-canary.md", sessionDoc("sibling repo note"))
	// Newer still, and the private vault's own host repo.
	writeVaultFile(t, priv, "sessions/2026-09-14.md", sessionDoc("host repo note"))

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: pub, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Last session:     2026-09-12 — our note") {
		t.Errorf("own workspace session not latest: %s", got)
	}
	if strings.Contains(got, "sibling repo note") || strings.Contains(got, "host repo note") {
		t.Errorf("leaked another workspace's session: %s", got)
	}
}

func TestRunRecall_RoutedVault_MergesDecisionsForThisRepo(t *testing.T) {
	pub, priv := mkRoutedVaultForTest(t)
	writeVaultFile(t, pub, "decisions/2026-09-01-public-thing.md", decisionDoc("test", "active", "Public thing"))
	writeVaultFile(t, priv, "decisions/2026-09-12-private-thing.md", decisionDoc("test", "active", "Private thing"))
	writeVaultFile(t, priv, "decisions/2026-09-11-dead-thing.md", decisionDoc("test", "superseded", "Dead thing"))
	// Another repo's private decision — must not appear.
	writeVaultFile(t, priv, "decisions/2026-09-13-other-repo-thing.md", decisionDoc("canary", "active", "Other repo thing"))

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: pub, Stdout: &out}); err != nil {
		t.Fatalf("RunRecall: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "private-thing") {
		t.Errorf("private decision missing: %s", got)
	}
	if !strings.Contains(got, "public-thing") {
		t.Errorf("public decision missing: %s", got)
	}
	if strings.Contains(got, "other-repo-thing") {
		t.Errorf("leaked another repo's decision: %s", got)
	}
	if !strings.Contains(got, "Filtered out:     1") {
		t.Errorf("status filter should count the superseded private decision: %s", got)
	}
}

func TestRunRecall_BrokenRoutingConfig_FallsBackToLocal(t *testing.T) {
	pub := mkVaultForTest(t)
	writeVaultFile(t, pub, vault.RoutingConfigFileName, `{"private_vault":`)
	writeVaultFile(t, pub, "sessions/2026-06-07.md", sessionDoc("local note"))

	var out bytes.Buffer
	if err := RunRecall(RecallOpts{VaultPath: pub, Stdout: &out}); err != nil {
		t.Fatalf("recall must not fail the session start on a broken config: %v", err)
	}
	if got := out.String(); !strings.Contains(got, "2026-06-07 — local note") {
		t.Errorf("expected local-only fallback: %s", got)
	}
}

//go:build live

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestLive_AskRoundTrip exercises the real Anthropic API.
// Skipped unless ANTHROPIC_API_KEY is set and the `live` build tag is on.
//
// Run manually:
//
//	ANTHROPIC_API_KEY=sk-ant-... go test -tags=live ./cmd/tolvi/...
//
// In CI: workflow_dispatch only on .github/workflows/cli.yml.
func TestLive_AskRoundTrip(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set")
	}
	bin := buildToTmp(t)
	work := t.TempDir()

	// Setup: init + sync one decision the LLM can find.
	initCmd := exec.Command(bin, "init", "--workspace", "live-smoke")
	initCmd.Dir = work
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	syncCmd := exec.Command(bin, "sync", "decision", "Cats are best",
		"--body", "# Cats\n\nCats are the best pets. We agreed.\n")
	syncCmd.Dir = work
	if out, err := syncCmd.CombinedOutput(); err != nil {
		t.Fatalf("sync: %v\n%s", err, out)
	}

	askCmd := exec.Command(bin, "ask", "what pets do we like?", "--no-stream")
	askCmd.Dir = work
	askCmd.Env = os.Environ() // inherits ANTHROPIC_API_KEY
	out, err := askCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ask: %v\n%s", err, out)
	}
	if !strings.Contains(strings.ToLower(string(out)), "cat") {
		t.Errorf("LLM didn't mention cats — context might not be feeding correctly: %s", out)
	}
	if !strings.Contains(string(out), "Sources:") {
		t.Errorf("no Sources footer: %s", out)
	}
}

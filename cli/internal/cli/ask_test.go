package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/llm"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// fakeLLM lets the test drive Ask without hitting the network.
type fakeLLM struct {
	response llm.StreamResult
}

func (f *fakeLLM) StreamCompletion(ctx context.Context, sys, msg string, max int64, onText func(string)) (llm.StreamResult, error) {
	if onText != nil {
		onText(f.response.Text)
	}
	return f.response, nil
}

func TestRunAsk_HappyPath(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	doc := vault.Doc{
		Type: "decision", Slug: "postgres", Date: "2026-04-12",
		Path:        "decisions/2026-04-12-postgres.md",
		Body:        []byte("# Postgres\n\nbecause pgvector.\n"),
		Frontmatter: map[string]any{"status": "active", "repo": "test"},
	}

	var out bytes.Buffer
	err := RunAsk(AskOpts{
		VaultPath: vaultDir,
		Query:     "why postgres",
		Docs:      []vault.Doc{doc},
		Now:       time.Date(2026, 5, 14, 17, 23, 0, 0, time.UTC),
		LLM: &fakeLLM{response: llm.StreamResult{
			Text:         "We chose Postgres for pgvector. [[postgres]]",
			Model:        "claude-sonnet-4-7",
			InputTokens:  100,
			OutputTokens: 20,
		}},
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("RunAsk: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "[[postgres]]") {
		t.Errorf("answer missing in stdout: %s", s)
	}
	if !strings.Contains(s, "Sources:") {
		t.Errorf("Sources footer missing: %s", s)
	}
	if !strings.Contains(s, "decisions/2026-04-12-postgres.md") {
		t.Errorf("path missing from Sources footer: %s", s)
	}
}

func TestRunAsk_UnverifiedCitationFlagged(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	doc := vault.Doc{
		Type: "decision", Slug: "postgres",
		Path: "decisions/x.md", Body: []byte("body"),
		Frontmatter: map[string]any{"status": "active"},
	}

	var out bytes.Buffer
	err := RunAsk(AskOpts{
		VaultPath: vaultDir, Query: "x",
		Docs: []vault.Doc{doc},
		Now:  time.Now(),
		LLM: &fakeLLM{response: llm.StreamResult{
			Text: "See [[postgres]] and also [[hallucinated]].",
		}},
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("RunAsk: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "⚠ unverified") || !strings.Contains(s, "hallucinated") {
		t.Errorf("unverified citation not flagged in footer: %s", s)
	}
}

func TestRunAsk_GuardError(t *testing.T) {
	vaultDir := mkVaultForTest(t)
	hugeBody := strings.Repeat("x", 800_000)
	doc := vault.Doc{
		Type: "decision", Slug: "big",
		Path: "decisions/big.md", Body: []byte(hugeBody),
		Frontmatter: map[string]any{"status": "active"},
	}

	var out bytes.Buffer
	err := RunAsk(AskOpts{
		VaultPath: vaultDir, Query: "x",
		Docs:   []vault.Doc{doc},
		Now:    time.Now(),
		LLM:    &fakeLLM{response: llm.StreamResult{Text: "should not be called"}},
		Stdout: &out,
	})
	if err == nil {
		t.Fatal("expected guard error on huge vault")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error message lacks 'too large': %v", err)
	}
}

// TestRunAsk_ReadsRepoAndProductNotTheWholeChain pins the CAG boundary. Reading
// the whole chain would pull the org root's corpus into every paid query, which
// is the premise cag-for-local-cli rests on. Repo plus product is the default,
// and the answer says so when the corpus was narrower than the chain.
func TestRunAsk_ReadsRepoAndProductNotTheWholeChain(t *testing.T) {
	repo := mkVaultForTest(t)
	product := filepath.Join(t.TempDir(), "product-root")
	org := filepath.Join(t.TempDir(), "org-root")
	for _, base := range []string{product, org} {
		for _, sub := range []string{"decisions", "sessions", "patterns"} {
			if err := os.MkdirAll(filepath.Join(base, sub), 0o755); err != nil {
				t.Fatal(err)
			}
		}
	}
	declareRawRoots(t, `{"roots":[
		{"role":"product","product":"stack","path":"`+product+`"},
		{"role":"org","workspace":"test","path":"`+org+`"}
	]}`)
	if err := vault.WriteMeta(repo, vault.Meta{
		Workspace: "test", Repo: "test", Product: "stack",
		EmbeddingModel: "nomic-embed-text", SchemaVersion: vault.SupportedSchemaVersion,
	}); err != nil {
		t.Fatal(err)
	}

	writeVaultFile(t, repo, "decisions/2026-09-01-repo-doc.md", decisionDoc("test", "active", "Repo doc"))
	writeVaultFile(t, product, "decisions/2026-09-02-product-doc.md", decisionDoc("test", "active", "Product doc"))
	writeVaultFile(t, org, "decisions/2026-09-03-org-doc.md", decisionDoc("test", "active", "Org doc"))

	roots, err := askReadRoots(repo)
	if err != nil {
		t.Fatalf("askReadRoots: %v", err)
	}
	got := map[string]bool{}
	for _, r := range roots {
		got[r.Path] = true
	}
	if !got[repo] || !got[product] {
		t.Errorf("ask should read repo and product roots, got %v", roots)
	}
	if got[org] {
		t.Errorf("ask must not read the org root by default, got %v", roots)
	}
}

package cli

import (
	"bytes"
	"context"
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
		Path: "decisions/2026-04-12-postgres.md",
		Body: []byte("# Postgres\n\nbecause pgvector.\n"),
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
		Docs: []vault.Doc{doc},
		Now:  time.Now(),
		LLM:  &fakeLLM{response: llm.StreamResult{Text: "should not be called"}},
		Stdout: &out,
	})
	if err == nil {
		t.Fatal("expected guard error on huge vault")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("error message lacks 'too large': %v", err)
	}
}

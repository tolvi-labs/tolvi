package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/citations"
	"github.com/tolvi-labs/tolvi/cli/internal/llm"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// LLMClient is the narrow interface used by RunAsk. Concrete impl is
// *llm.Client; tests inject a fake.
type LLMClient interface {
	StreamCompletion(ctx context.Context, sys, msg string, maxTokens int64, onText func(string)) (llm.StreamResult, error)
}

// AskOpts holds parsed args + injected dependencies for `tolvi ask`.
type AskOpts struct {
	VaultPath string
	Query     string

	// Pre-loaded docs (injected for tests; production loads via vault.LoadAll).
	Docs []vault.Doc

	// Filter overrides.
	IncludeStatuses []string
	ExcludeTypes    []string

	// LLM injection point.
	LLM LLMClient

	// I/O.
	Stdout   io.Writer
	Stderr   io.Writer
	NoStream bool
	JSON     bool

	// Knobs.
	Model     string
	MaxTokens int64
	Now       time.Time
}

// RunAsk is the end-to-end ask flow.
func RunAsk(opts AskOpts) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	if opts.LLM == nil {
		return fmt.Errorf("internal: AskOpts.LLM is required")
	}
	if opts.Query == "" {
		return fmt.Errorf("query is required")
	}

	meta, err := vault.ReadMeta(opts.VaultPath)
	if err != nil {
		return err
	}

	docs := opts.Docs
	if docs == nil {
		loaded, errs, err := vault.LoadAll(opts.VaultPath)
		if err != nil {
			return err
		}
		for _, e := range errs {
			fmt.Fprintf(opts.Stderr, "warn: skipped invalid doc: %v\n", e)
		}
		docs = loaded
	}

	systemPrompt := llm.AssemblePrompt(llm.AssembleOpts{
		Workspace:       meta.Workspace,
		Docs:            docs,
		GeneratedAt:     opts.Now,
		IncludeStatuses: opts.IncludeStatuses,
		ExcludeTypes:    opts.ExcludeTypes,
	})

	tokens := llm.EstimateTokens(systemPrompt)
	switch llm.EvaluateGuard(tokens) {
	case llm.GuardWarn:
		fmt.Fprintf(opts.Stderr,
			"warn: vault is large (~%d tokens). Approaching context limits. Consider archiving old sessions or splitting vaults.\n",
			tokens)
	case llm.GuardError:
		return fmt.Errorf("vault too large for v1 (~%d tokens). Filter with --include-status active (already default), --exclude-type session, or migrate to the server arm",
			tokens)
	}

	// Build set of known slugs for citation verification.
	known := map[string]bool{}
	docBySlug := map[string]vault.Doc{}
	for _, d := range docs {
		known[d.Slug] = true
		docBySlug[d.Slug] = d
	}

	// Stream.
	var streamSink func(string)
	if !opts.JSON && !opts.NoStream {
		streamSink = func(chunk string) {
			fmt.Fprint(opts.Stdout, chunk)
		}
	}
	result, err := opts.LLM.StreamCompletion(
		context.Background(),
		systemPrompt, opts.Query,
		opts.MaxTokens, streamSink,
	)
	if err != nil {
		return fmt.Errorf("llm error: %w", err)
	}

	// Citations.
	extracted := citations.Extract(result.Text)
	matched, unmatched := citations.Verify(extracted, known)

	if opts.JSON {
		return printAskJSON(opts.Stdout, opts.Stderr, result, matched, unmatched, docBySlug)
	}

	if opts.NoStream {
		fmt.Fprintln(opts.Stdout, result.Text)
	} else {
		fmt.Fprintln(opts.Stdout)
	}
	printSourcesFooter(opts.Stdout, matched, unmatched, docBySlug)
	return nil
}

func printSourcesFooter(w io.Writer, matched, unmatched []string, docBySlug map[string]vault.Doc) {
	if len(matched) == 0 && len(unmatched) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Sources:")
	for _, slug := range matched {
		d := docBySlug[slug]
		status := d.Frontmatter.String("status")
		fmt.Fprintf(w, "  [[%s]]  %s  (%s)\n", slug, d.Path, status)
	}
	for _, slug := range unmatched {
		fmt.Fprintf(w, "  ⚠ unverified [[%s]]\n", slug)
	}
}

package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// SyncOpts holds the parsed args + flags for `tolvi sync`.
type SyncOpts struct {
	VaultPath string
	DocType   string // decision | session | pattern
	Title     string

	// Frontmatter overrides
	Slug   string // empty → derive from Title
	Status string // empty → "active"
	Date   time.Time

	// Body capture (precedence: BodyFlag > StdinReader > editor)
	BodyFlag    string
	StdinReader io.Reader
	NoEdit      bool // write skeleton-only (no body capture at all)
	RunEditor   func(string) error

	// I/O
	Stdout    io.Writer
	PrintPath bool // when true, print only the resulting path to Stdout
}

// RunSync is the full `tolvi sync` flow.
//
// For sessions, the same-day file is treated as an append-target (handled
// in Task 15); this version refuses to overwrite as a safety net.
func RunSync(opts SyncOpts) error {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.DocType != "decision" && opts.DocType != "session" && opts.DocType != "pattern" {
		return fmt.Errorf("type must be decision|session|pattern, got %q", opts.DocType)
	}
	if opts.VaultPath == "" {
		return fmt.Errorf("internal: VaultPath required")
	}
	if opts.DocType == "pattern" && strings.TrimSpace(opts.Title) == "" {
		return fmt.Errorf("title is required for pattern")
	}

	meta, err := vault.ReadMeta(opts.VaultPath)
	if err != nil {
		return fmt.Errorf("read vault meta: %w", err)
	}

	if opts.Date.IsZero() {
		opts.Date = time.Now()
	}

	// Resolve slug.
	slug := opts.Slug
	if slug == "" {
		slug = format.SlugFromTitle(opts.Title)
	}
	if opts.DocType != "session" && slug == "" {
		return fmt.Errorf("could not derive a slug from title %q — pass --slug explicitly", opts.Title)
	}
	if opts.DocType == "session" {
		slug = opts.Date.Format("2006-01-02")
	}

	relPath := PathForDoc(opts.DocType, slug, opts.Date.Format("2006-01-02"))
	absPath := filepath.Join(opts.VaultPath, relPath)

	if _, err := os.Stat(absPath); err == nil {
		if opts.DocType != "session" {
			return fmt.Errorf("%s already exists — refusing to overwrite", absPath)
		}
		// Session same-day append flow: capture a new block (no
		// frontmatter on the block), then re-render the existing file
		// with the new block appended at the end of the body.
		return appendSessionBlock(opts, absPath)
	}

	// Assemble + validate frontmatter.
	fm := AssembleFrontmatter(AssembleOpts{
		DocType:   opts.DocType,
		Title:     opts.Title,
		Slug:      slug,
		Workspace: meta.Workspace,
		Date:      opts.Date,
		Status:    opts.Status,
	})
	if err := ValidateFrontmatter(opts.DocType, fm); err != nil {
		return err
	}

	// Body.
	var body []byte
	if opts.NoEdit {
		body = []byte("\n")
	} else {
		captured, err := CaptureBody(CaptureOpts{
			InitialContent: skeletonForType(opts.DocType, opts.Title, fm),
			BodyFlag:       opts.BodyFlag,
			StdinReader:    opts.StdinReader,
			RunEditor:      opts.RunEditor,
		})
		if err != nil {
			return err
		}
		body = stripFrontmatterIfPresent(captured)
	}

	// Render.
	full, err := format.RenderDocument(fm, body)
	if err != nil {
		return fmt.Errorf("render document: %w", err)
	}

	// Atomic write.
	if err := atomicWriteFile(absPath, full); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	if opts.PrintPath {
		fmt.Fprintln(opts.Stdout, absPath)
	} else {
		fmt.Fprintf(opts.Stdout, "✓ Wrote %s\n", relPath)
	}
	return nil
}

// skeletonForType returns the initial editor content (frontmatter rendered
// at the top + a blank body area).
func skeletonForType(docType, title string, fm format.Frontmatter) []byte {
	body := []byte(fmt.Sprintf("# %s\n\n<!-- write your %s body here -->\n", title, docType))
	out, err := format.RenderDocument(fm, body)
	if err != nil {
		return body
	}
	return out
}

// stripFrontmatterIfPresent removes any frontmatter the user left in the
// captured content (the CLI renders frontmatter separately).
func stripFrontmatterIfPresent(content []byte) []byte {
	fm, body, err := format.ParseFrontmatter(content)
	if err != nil {
		return content
	}
	_ = fm
	return body
}

// atomicWriteFile writes data to path via a temp file + rename so a
// crashed write never leaves a half-written .md.
func atomicWriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tolvi-sync-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

// appendSessionBlock handles the case where `tolvi sync session` is
// invoked on a day that already has a sessions/<date>.md file. The
// existing frontmatter is preserved; a fresh session-block template
// is captured via the same body-capture pipeline; the new block is
// appended to the body with a blank line separator.
func appendSessionBlock(opts SyncOpts, path string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read existing session file: %w", err)
	}
	fm, body, err := format.ParseFrontmatter(existing)
	if err != nil {
		return fmt.Errorf("parse existing session file: %w", err)
	}

	blockTemplate := []byte(fmt.Sprintf(
		"## [%s] Session — \n\n**Tickets:** none\n\n### What happened\n\n### Files touched\n\n### Left open\n",
		opts.Date.Format("15:04"),
	))

	var newBlock []byte
	if opts.NoEdit {
		newBlock = blockTemplate
	} else {
		captured, err := CaptureBody(CaptureOpts{
			InitialContent: blockTemplate,
			BodyFlag:       opts.BodyFlag,
			StdinReader:    opts.StdinReader,
			RunEditor:      opts.RunEditor,
		})
		if err != nil {
			return err
		}
		newBlock = captured
	}

	// Ensure separation between existing body and new block.
	var merged []byte
	merged = append(merged, body...)
	if len(body) > 0 && body[len(body)-1] != '\n' {
		merged = append(merged, '\n')
	}
	merged = append(merged, '\n')
	merged = append(merged, newBlock...)

	rendered, err := format.RenderDocument(fm, merged)
	if err != nil {
		return fmt.Errorf("render appended session: %w", err)
	}
	if err := atomicWriteFile(path, rendered); err != nil {
		return fmt.Errorf("write appended session: %w", err)
	}
	if opts.PrintPath {
		fmt.Fprintln(opts.Stdout, path)
	} else {
		rel, _ := filepath.Rel(opts.VaultPath, path)
		fmt.Fprintf(opts.Stdout, "✓ Appended session block to %s\n", rel)
	}
	return nil
}

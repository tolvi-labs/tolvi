package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
)

// AssembleOpts builds the inputs to AssembleFrontmatter.
type AssembleOpts struct {
	DocType   string    // "decision" | "session" | "pattern"
	Title     string    // free-form title (used for slug generation only; not stored as a field)
	Slug      string    // pre-computed slug (caller-derived from Title or --slug flag)
	Workspace string    // from vault.Meta.Workspace
	Date      time.Time // typically time.Now()
	Status    string    // defaults to "active"
}

// AssembleFrontmatter builds a Frontmatter map per the doc-type rules of
// tolvi-format-v1. Fields included depend on type:
//
//	decision: tags, date, repo, status
//	session:  tags, date, status
//	pattern:  tags, status   (no date, no repo — patterns are timeless)
func AssembleFrontmatter(opts AssembleOpts) format.Frontmatter {
	if opts.Status == "" {
		opts.Status = "active"
	}
	fm := format.Frontmatter{
		"tags":   []any{opts.DocType},
		"status": opts.Status,
	}
	switch opts.DocType {
	case "decision":
		fm["date"] = opts.Date.Format("2006-01-02")
		fm["repo"] = opts.Workspace
	case "session":
		fm["date"] = opts.Date.Format("2006-01-02")
	case "pattern":
		// nothing extra
	}
	return fm
}

// PathForDoc returns the vault-relative path for a new doc per the
// format-spec layout rule (decisions: date-slug, sessions: date only,
// patterns: slug only).
func PathForDoc(docType, slug, date string) string {
	switch docType {
	case "decision":
		return filepath.ToSlash(filepath.Join("decisions", fmt.Sprintf("%s-%s.md", date, slug)))
	case "session":
		return filepath.ToSlash(filepath.Join("sessions", fmt.Sprintf("%s.md", date)))
	case "pattern":
		return filepath.ToSlash(filepath.Join("patterns", fmt.Sprintf("%s.md", slug)))
	}
	return ""
}

// ValidateFrontmatter checks that the given frontmatter satisfies the
// embedded JSON Schema for the given doc type. Returns nil on success;
// returns a descriptive error listing each violation otherwise.
func ValidateFrontmatter(docType string, fm format.Frontmatter) error {
	validator, err := format.ValidatorForDocType(docType)
	if err != nil {
		return err
	}
	// jsonschema works on map[string]any with primitive value types; our
	// Frontmatter already satisfies that shape.
	if err := validator.Validate(map[string]any(fm)); err != nil {
		return fmt.Errorf("frontmatter does not match %s schema: %w", docType, err)
	}
	return nil
}

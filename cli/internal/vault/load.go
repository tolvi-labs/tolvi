package vault

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
)

// Doc is a single loaded vault document.
type Doc struct {
	Type        string             // "decision" | "session" | "pattern"
	Slug        string             // extracted from filename
	Date        string             // YYYY-MM-DD for decisions/sessions; "" for patterns
	Path        string             // path relative to vault root, forward-slash
	AbsPath     string             // absolute path on disk
	Frontmatter format.Frontmatter // parsed frontmatter
	Body        []byte             // doc body (everything after closing ---)
}

// LoadAll reads every .md file under <vaultPath>/{decisions,sessions,patterns}/.
// Per-file parse errors are collected (returned in the second slice) and
// don't stop the load. A non-nil top-level error means something stopped
// the load entirely (e.g., the vault directory doesn't exist).
//
// Output ordering: decisions → patterns → sessions, then by date
// descending within each type (newest first), then by slug asc as a
// deterministic tie-breaker.
func LoadAll(vaultPath string) ([]Doc, []error, error) {
	if _, err := os.Stat(vaultPath); err != nil {
		return nil, nil, fmt.Errorf("vault path: %w", err)
	}
	var docs []Doc
	var errs []error

	for _, sub := range []struct {
		dir     string
		docType string
	}{
		{"decisions", "decision"},
		{"patterns", "pattern"},
		{"sessions", "session"},
	} {
		subDir := filepath.Join(vaultPath, sub.dir)
		entries, err := os.ReadDir(subDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			errs = append(errs, fmt.Errorf("read %s: %w", subDir, err))
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			absPath := filepath.Join(subDir, e.Name())
			doc, err := loadOne(vaultPath, absPath, sub.docType)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", absPath, err))
				continue
			}
			docs = append(docs, doc)
		}
	}

	sort.SliceStable(docs, func(i, j int) bool {
		ti, tj := typeOrder(docs[i].Type), typeOrder(docs[j].Type)
		if ti != tj {
			return ti < tj
		}
		if docs[i].Date != docs[j].Date {
			return docs[i].Date > docs[j].Date
		}
		return docs[i].Slug < docs[j].Slug
	})
	return docs, errs, nil
}

func loadOne(vaultPath, absPath, docType string) (Doc, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return Doc{}, fmt.Errorf("read: %w", err)
	}
	fm, body, err := format.ParseFrontmatter(data)
	if err != nil {
		return Doc{}, fmt.Errorf("frontmatter: %w", err)
	}
	rel, err := filepath.Rel(vaultPath, absPath)
	if err != nil {
		rel = absPath
	}
	rel = filepath.ToSlash(rel)
	slug, date := SlugAndDateFromPath(rel, docType)
	return Doc{
		Type:        docType,
		Slug:        slug,
		Date:        date,
		Path:        rel,
		AbsPath:     absPath,
		Frontmatter: fm,
		Body:        body,
	}, nil
}

func typeOrder(t string) int {
	switch t {
	case "decision":
		return 0
	case "pattern":
		return 1
	case "session":
		return 2
	}
	return 3
}

// SlugAndDateFromPath extracts (slug, date) from a vault-relative path
// like "decisions/2026-04-12-postgres.md" → ("postgres", "2026-04-12").
//
//   - decisions: <date>-<slug>.md → (slug, date)
//   - sessions:  <date>.md         → (date, date) — the date IS the slug
//   - patterns:  <slug>.md         → (slug, "")
//
// Returns ("", "") when the filename doesn't match the expected shape.
var datePrefixRe = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})(?:-(.+))?\.md$`)

func SlugAndDateFromPath(relPath, docType string) (slug, date string) {
	base := filepath.Base(relPath)
	switch docType {
	case "decision":
		m := datePrefixRe.FindStringSubmatch(base)
		if m == nil || m[2] == "" {
			return "", ""
		}
		return m[2], m[1]
	case "session":
		m := datePrefixRe.FindStringSubmatch(base)
		if m == nil {
			return "", ""
		}
		return m[1], m[1]
	case "pattern":
		name := strings.TrimSuffix(base, ".md")
		return name, ""
	}
	return "", ""
}

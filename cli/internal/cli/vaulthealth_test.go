package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// healthVault writes docs into a vault and returns its path.
func healthVault(t *testing.T, files map[string]string) string {
	t.Helper()
	v := filepath.Join(t.TempDir(), "vault")
	for _, sub := range []string{"decisions", "sessions", "patterns"} {
		if err := os.MkdirAll(filepath.Join(v, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	meta := `{"workspace":"t","embedding_model":"nomic-embed-text","schema_version":1}`
	if err := os.WriteFile(filepath.Join(v, ".vault-meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	for rel, body := range files {
		full := filepath.Join(v, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return v
}

const cleanDoc = `---
tags: [decision, test]
date: 2026-09-12
status: active
---

# A clean decision

## Why
Because.
`

func findingIDs(fs []HealthFinding) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.CheckID)
	}
	return out
}

func hasFinding(fs []HealthFinding, id string) bool {
	for _, f := range fs {
		if f.CheckID == id {
			return true
		}
	}
	return false
}

func TestVaultHealth_CleanVaultScoresAPlus(t *testing.T) {
	v := healthVault(t, map[string]string{"decisions/2026-09-12-clean.md": cleanDoc})
	rep, err := RunVaultHealth(v)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Findings) != 0 {
		t.Fatalf("expected no findings, got %v", findingIDs(rep.Findings))
	}
	if rep.OverallGrade != "A+" {
		t.Errorf("got grade %q, want A+", rep.OverallGrade)
	}
	if rep.FilesScanned != 1 {
		t.Errorf("scanned %d files, want 1", rep.FilesScanned)
	}
}

func TestVaultHealth_EmptyTagsIsHigh(t *testing.T) {
	v := healthVault(t, map[string]string{"decisions/2026-09-12-x.md": "---\ndate: 2026-09-12\nstatus: active\n---\n\n# No tags\n"})
	rep, _ := RunVaultHealth(v)
	if !hasFinding(rep.Findings, "empty-tags") {
		t.Fatalf("expected empty-tags, got %v", findingIDs(rep.Findings))
	}
}

// An absent status means `active` by convention, so it is not a defect; only
// a present-but-unrecognized value is.
func TestVaultHealth_StatusEnumOnlyFlagsUnknownValues(t *testing.T) {
	v := healthVault(t, map[string]string{
		"decisions/2026-09-12-a.md": "---\ntags: [d]\nstatus: wibble\n---\n\n# Bad status\n",
		"decisions/2026-09-12-b.md": "---\ntags: [d]\n---\n\n# Absent status\n",
	})
	rep, _ := RunVaultHealth(v)
	var flagged []string
	for _, f := range rep.Findings {
		if f.CheckID == "status-enum" {
			flagged = append(flagged, f.File)
		}
	}
	if len(flagged) != 1 || !strings.Contains(flagged[0], "-a.md") {
		t.Fatalf("expected only the unknown-status file flagged, got %v", flagged)
	}
}

func TestVaultHealth_DuplicateTitlesNameEachOther(t *testing.T) {
	doc := "---\ntags: [d]\nstatus: active\n---\n\n# Choose Postgres\n"
	v := healthVault(t, map[string]string{
		"decisions/2026-09-12-a.md": doc,
		"decisions/2026-09-11-b.md": doc,
	})
	rep, _ := RunVaultHealth(v)
	var dupes []HealthFinding
	for _, f := range rep.Findings {
		if f.CheckID == "duplicate-title" {
			dupes = append(dupes, f)
		}
	}
	if len(dupes) != 2 {
		t.Fatalf("expected both files flagged, got %d", len(dupes))
	}
	for _, f := range dupes {
		if strings.Contains(f.Reason, f.File) {
			t.Errorf("a file should not be listed as its own duplicate: %s", f.Reason)
		}
	}
}

// Placeholders are scoped to frontmatter and headings on purpose: body prose
// may legitimately document the `sessions/YYYY-MM-DD.md` naming convention.
func TestVaultHealth_PlaceholdersIgnoreBodyProse(t *testing.T) {
	v := healthVault(t, map[string]string{
		"decisions/2026-09-12-real.md":  "---\ntags: [d]\ndate: 2026-09-12\n---\n\n# Naming\n\nSessions are named YYYY-MM-DD.md by convention.\n",
		"decisions/2026-09-12-blank.md": "---\ntags: [d]\ndate: YYYY-MM-DD\n---\n\n# Unfilled\n",
	})
	rep, _ := RunVaultHealth(v)
	var flagged []string
	for _, f := range rep.Findings {
		if f.CheckID == "template-placeholder" {
			flagged = append(flagged, f.File)
		}
	}
	if len(flagged) != 1 || !strings.Contains(flagged[0], "blank") {
		t.Fatalf("only the unfilled frontmatter should be flagged, got %v", flagged)
	}
}

func TestVaultHealth_AngleBracketPlaceholderDoesNotMatchHTMLTags(t *testing.T) {
	v := healthVault(t, map[string]string{
		"decisions/2026-09-12-html.md": "---\ntags: [d]\nrepo: real-repo\n---\n\n# Has html\n\nLine<br>break and a <div> too.\n",
		"decisions/2026-09-12-slug.md": "---\ntags: [d]\nrepo: <repo-slug>\n---\n\n# Unfilled slug\n",
	})
	rep, _ := RunVaultHealth(v)
	for _, f := range rep.Findings {
		if f.CheckID == "template-placeholder" && strings.Contains(f.File, "html") {
			t.Errorf("<br>/<div> must not count as a placeholder: %s", f.Reason)
		}
	}
	if !hasFinding(rep.Findings, "template-placeholder") {
		t.Error("expected the <repo-slug> frontmatter to be flagged")
	}
}

func TestVaultHealth_UnicodeEscapingIsLowSeverity(t *testing.T) {
	v := healthVault(t, map[string]string{
		"decisions/2026-09-12-u.md": "---\ntags: [d]\nstatus: active\n---\n\n# Escaped\n\nIt says caf\\u00e9 here.\n",
	})
	rep, _ := RunVaultHealth(v)
	for _, f := range rep.Findings {
		if f.CheckID == "unicode-escaping" {
			if f.Severity != "low" {
				t.Errorf("severity = %q, want low", f.Severity)
			}
			return
		}
	}
	t.Fatalf("expected unicode-escaping, got %v", findingIDs(rep.Findings))
}

func TestVaultHealth_DefectsDragTheGradeDown(t *testing.T) {
	files := map[string]string{}
	for i, name := range []string{"a", "b", "c", "d"} {
		_ = i
		files["decisions/2026-09-1"+name[:1]+"-"+name+".md"] = "---\nstatus: active\n---\n\n# Untagged " + name + "\n"
	}
	v := healthVault(t, files)
	rep, _ := RunVaultHealth(v)
	if rep.OverallGrade == "A+" {
		t.Fatal("a vault where every note is untagged must not grade A+")
	}
	if g := rep.Grades["tags"]; g.Letter != "F" {
		t.Errorf("tags dimension = %q, want F", g.Letter)
	}
}

func TestVaultHealth_EmptyVaultIsNotAFailure(t *testing.T) {
	v := healthVault(t, map[string]string{})
	rep, err := RunVaultHealth(v)
	if err != nil {
		t.Fatal(err)
	}
	if rep.OverallGrade != "A+" {
		t.Errorf("an empty vault should not be graded down, got %q", rep.OverallGrade)
	}
}

func TestRenderVaultHealth_NamesEveryFindingsFile(t *testing.T) {
	v := healthVault(t, map[string]string{"decisions/2026-09-12-x.md": "---\nstatus: nope\n---\n\n# Bad\n"})
	rep, _ := RunVaultHealth(v)
	var out bytes.Buffer
	if err := RenderVaultHealth(&out, rep); err != nil {
		t.Fatal(err)
	}
	report := out.String()
	for _, f := range rep.Findings {
		if !strings.Contains(report, f.File) {
			t.Errorf("report omits %s", f.File)
		}
	}
	if !strings.Contains(report, rep.OverallGrade) {
		t.Error("report omits the overall grade")
	}
}

// Parity regression. Six files with this exact defect profile average to
// exactly 80.0, the B- boundary. Float accumulation once made this C+, and
// Go's randomised map iteration made it vary between runs. Both the Python
// implementation and this one must say B-.
func TestVaultHealth_BoundaryGradeMatchesPythonAndIsStable(t *testing.T) {
	files := map[string]string{
		"decisions/2026-01-01-untagged.md":  "---\ndate: 2026-01-01\nstatus: active\n---\n\n# Untagged note\n",
		"decisions/2026-01-02-badstatus.md": "---\ntags: [d]\nstatus: wibble\n---\n\n# Bad status\n",
		"decisions/2026-01-03-unfilled.md":  "---\ntags: [d]\ndate: YYYY-MM-DD\nrepo: <repo-slug>\n---\n\n# Unfilled template\n",
		"decisions/2026-01-04-dupe-a.md":    "---\ntags: [d]\nstatus: active\n---\n\n# Choose Postgres\n",
		"decisions/2026-01-05-dupe-b.md":    "---\ntags: [d]\nstatus: active\n---\n\n# Choose Postgres\n",
		"decisions/2026-01-06-escaped.md":   "---\ntags: [d]\nstatus: active\n---\n\n# Escapes\n\nIt says caf\\u00e9 here.\n",
	}
	v := healthVault(t, files)
	want := map[string]string{"tags": "B", "lifecycle": "B", "dedup": "D", "completeness": "B", "unicode": "B"}

	// Repeat: map-iteration order differs per run, so a single pass can pass by luck.
	for i := 0; i < 20; i++ {
		rep, err := RunVaultHealth(v)
		if err != nil {
			t.Fatal(err)
		}
		if rep.OverallGrade != "B-" {
			t.Fatalf("run %d: overall = %q (%.15f), want B-", i, rep.OverallGrade, rep.OverallPct)
		}
		for dim, letter := range want {
			if got := rep.Grades[dim].Letter; got != letter {
				t.Fatalf("run %d: %s = %q, want %q", i, dim, got, letter)
			}
		}
	}
}

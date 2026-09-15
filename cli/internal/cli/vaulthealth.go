package cli

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/tolvi-labs/tolvi/cli/internal/vault"
)

// Vault health: a Go port of skills/vault-health/scripts/vault_health.py.
//
// Ported rather than shelled out to on purpose. `tolvi doctor` runs this
// automatically, and shelling out would give doctor a hard dependency on
// Python, uv, and PyYAML — reintroducing exactly the silent-degradation
// failure doctor exists to detect. The Python skill stays as the standalone
// graded report; this is the embedded copy. Both are pinned by
// .github/scripts/vault-health-parity-check.sh.

// HealthFinding is one defect in one file.
type HealthFinding struct {
	CheckID  string
	Severity string // "high" | "medium" | "low"
	File     string
	Reason   string
}

// HealthGrade is one dimension's score.
type HealthGrade struct {
	Pct    float64
	Letter string
}

// HealthReport is the result of a full scan.
type HealthReport struct {
	VaultPath    string
	FilesScanned int
	Findings     []HealthFinding
	Grades       map[string]HealthGrade
	OverallGrade string
	OverallPct   float64
}

// knownStatuses is the tolvi-format status vocabulary, frozen since v1. An absent status
// means `active` by convention, so only a present-but-unrecognized value is
// a defect.
var knownStatuses = map[string]bool{
	"active": true, "in-progress": true, "draft": true,
	"superseded": true, "deprecated": true,
}

// healthDimensions maps a reported dimension to the checks that feed it.
var healthDimensions = map[string][]string{
	"tags":         {"empty-tags"},
	"lifecycle":    {"status-enum"},
	"dedup":        {"duplicate-title"},
	"completeness": {"template-placeholder"},
	"unicode":      {"unicode-escaping"},
}

var gradeThresholds = []struct {
	min    float64
	letter string
}{
	{97, "A+"}, {93, "A"}, {90, "A-"}, {87, "B+"}, {83, "B"}, {80, "B-"},
	{77, "C+"}, {73, "C"}, {70, "C-"}, {60, "D"},
}

var (
	headingRe    = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)
	slugRe       = regexp.MustCompile(`[^a-z0-9]+`)
	dateHolderRe = regexp.MustCompile(`YYYY-MM-DD`)
	// Requires an internal hyphen so it cannot match HTML tags like <br> or <div>.
	slugHolderRe = regexp.MustCompile(`<[a-z]+(?:-[a-z]+)+>`)
	timeHolderRe = regexp.MustCompile(`(?m)^#{1,6}\s.*\[HH:MM\]`)
	unicodeEscRe = regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
)

// gradeEpsilon absorbs float accumulation error. A dimension set averaging
// exactly 80.0 must grade B-, but summing 83.333... four times can land on
// 79.999999999999986, which would silently drop a whole grade.
const gradeEpsilon = 1e-9

func gradeLetter(pct float64) string {
	for _, t := range gradeThresholds {
		if pct >= t.min-gradeEpsilon {
			return t.letter
		}
	}
	return "F"
}

func healthSlugify(title string) string {
	s := slugRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(title)), "-")
	return strings.Trim(s, "-")
}

// RunVaultHealth scans a vault and grades it.
func RunVaultHealth(vaultPath string) (HealthReport, error) {
	docs, parseErrs, err := vault.LoadAll(vaultPath)
	if err != nil {
		return HealthReport{}, err
	}

	rep := HealthReport{VaultPath: vaultPath, FilesScanned: len(docs) + len(parseErrs)}
	var findings []HealthFinding

	for _, e := range parseErrs {
		findings = append(findings, HealthFinding{
			CheckID: "unparseable-frontmatter", Severity: "high",
			File: parseErrFile(e), Reason: e.Error(),
		})
	}

	clusters := map[string][]string{}
	for _, d := range docs {
		fmRaw := frontmatterText(d)

		if !hasTags(d.Frontmatter) {
			findings = append(findings, HealthFinding{
				CheckID: "empty-tags", Severity: "high", File: d.Path,
				Reason: "tags field is empty or missing",
			})
		}

		if raw, ok := d.Frontmatter["status"]; ok {
			s := fmt.Sprintf("%v", raw)
			if !knownStatuses[strings.TrimSpace(s)] {
				findings = append(findings, HealthFinding{
					CheckID: "status-enum", Severity: "medium", File: d.Path,
					Reason: fmt.Sprintf("unrecognized status value: %q", s),
				})
			}
		}

		if m := headingRe.FindSubmatch(d.Body); m != nil {
			if slug := healthSlugify(string(m[1])); slug != "" {
				clusters[slug] = append(clusters[slug], d.Path)
			}
		}

		var hits []string
		if m := dateHolderRe.FindString(fmRaw); m != "" {
			hits = append(hits, "date ("+m+")")
		}
		if m := slugHolderRe.FindString(fmRaw); m != "" {
			hits = append(hits, "slug ("+m+")")
		}
		if timeHolderRe.Match(d.Body) {
			hits = append(hits, "time ([HH:MM])")
		}
		if len(hits) > 0 {
			findings = append(findings, HealthFinding{
				CheckID: "template-placeholder", Severity: "high", File: d.Path,
				Reason: "unfilled template placeholder: " + strings.Join(hits, ", "),
			})
		}

		if m := unicodeEscRe.FindAllString(fmRaw+string(d.Body), -1); len(m) > 0 {
			findings = append(findings, HealthFinding{
				CheckID: "unicode-escaping", Severity: "low", File: d.Path,
				Reason: fmt.Sprintf("%d escaped unicode sequence(s) found (e.g. %s)", len(m), m[0]),
			})
		}
	}

	for slug, members := range clusters {
		if len(members) < 2 {
			continue
		}
		sort.Strings(members)
		for _, f := range members {
			var others []string
			for _, o := range members {
				if o != f {
					others = append(others, o)
				}
			}
			findings = append(findings, HealthFinding{
				CheckID: "duplicate-title", Severity: "high", File: f,
				Reason: fmt.Sprintf("title slug %q also used by: %s", slug, strings.Join(others, ", ")),
			})
		}
	}

	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].CheckID != findings[j].CheckID {
			return findings[i].CheckID < findings[j].CheckID
		}
		return findings[i].File < findings[j].File
	})
	rep.Findings = findings
	rep.Grades, rep.OverallPct, rep.OverallGrade = computeHealthGrades(len(docs), findings)
	return rep, nil
}

func computeHealthGrades(evaluable int, findings []HealthFinding) (map[string]HealthGrade, float64, string) {
	grades := make(map[string]HealthGrade, len(healthDimensions))
	// Sum in a fixed order: Go randomises map iteration, and float addition is
	// not associative, so an unsorted sum makes the overall grade vary between
	// runs on the same vault.
	dims := make([]string, 0, len(healthDimensions))
	for dim := range healthDimensions {
		dims = append(dims, dim)
	}
	sort.Strings(dims)

	var sum float64
	for _, dim := range dims {
		ids := healthDimensions[dim]
		affected := map[string]bool{}
		for _, f := range findings {
			for _, id := range ids {
				if f.CheckID == id {
					affected[f.File] = true
				}
			}
		}
		pct := 100.0
		if evaluable > 0 {
			pct = 100.0 * float64(evaluable-len(affected)) / float64(evaluable)
		}
		grades[dim] = HealthGrade{Pct: pct, Letter: gradeLetter(pct)}
		sum += pct
	}
	overall := 100.0
	if len(grades) > 0 {
		overall = sum / float64(len(grades))
	}
	return grades, overall, gradeLetter(overall)
}

func hasTags(fm map[string]any) bool {
	raw, ok := fm["tags"]
	if !ok || raw == nil {
		return false
	}
	switch v := raw.(type) {
	case []any:
		return len(v) > 0
	case string:
		return strings.TrimSpace(v) != ""
	}
	return true
}

// frontmatterText reconstructs the frontmatter as text for placeholder and
// escape scanning. Values are what the checks look at, so key order is
// irrelevant; sorting only keeps the scan deterministic.
func frontmatterText(d vault.Doc) string {
	keys := make([]string, 0, len(d.Frontmatter))
	for k := range d.Frontmatter {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s: %v\n", k, d.Frontmatter[k])
	}
	return b.String()
}

func parseErrFile(err error) string {
	s := err.Error()
	if i := strings.Index(s, ": "); i > 0 {
		return s[:i]
	}
	return s
}

const maxShownPerCheck = 10

// RenderVaultHealth writes the human report.
func RenderVaultHealth(w io.Writer, rep HealthReport) error {
	var b strings.Builder
	fmt.Fprintf(&b, "\nVAULT HEALTH — %s\n", rep.VaultPath)
	b.WriteString(strings.Repeat("─", 40) + "\n")
	fmt.Fprintf(&b, "Files scanned: %d\n", rep.FilesScanned)
	fmt.Fprintf(&b, "Overall grade: %s\n", rep.OverallGrade)

	dims := make([]string, 0, len(rep.Grades))
	for d := range rep.Grades {
		dims = append(dims, d)
	}
	sort.Strings(dims)
	parts := make([]string, 0, len(dims))
	for _, d := range dims {
		parts = append(parts, d+" "+rep.Grades[d].Letter)
	}
	b.WriteString(strings.Join(parts, " · ") + "\n")

	if len(rep.Findings) == 0 {
		b.WriteString("\nNo defects found.\n")
		_, err := io.WriteString(w, b.String())
		return err
	}

	byCheck := map[string][]HealthFinding{}
	var order []string
	for _, f := range rep.Findings {
		if _, seen := byCheck[f.CheckID]; !seen {
			order = append(order, f.CheckID)
		}
		byCheck[f.CheckID] = append(byCheck[f.CheckID], f)
	}
	for _, id := range order {
		group := byCheck[id]
		fmt.Fprintf(&b, "\n%s (%s) — %d file(s)\n", id, group[0].Severity, len(group))
		for i, f := range group {
			if i == maxShownPerCheck {
				fmt.Fprintf(&b, "  … and %d more\n", len(group)-maxShownPerCheck)
				break
			}
			fmt.Fprintf(&b, "  %s — %s\n", f.File, f.Reason)
		}
	}
	b.WriteString("\n")
	_, err := io.WriteString(w, b.String())
	return err
}

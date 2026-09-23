package cli

import (
	"encoding/json"
	"io"

	"github.com/tolvi-labs/tolvi/cli/internal/packs"
)

// The --json forms of doctor and doctor vault-health. The shapes are published
// as spec/schemas/doctor.json and spec/schemas/vault-health.json, so these
// structs carry explicit tags rather than inheriting Go field names, and the
// domain types they are built from stay free to change.
//
// Every payload names the CLI that produced it. A consumer that needs a newer
// contract than it is reading can say so instead of parsing output it cannot
// trust.

type doctorJSON struct {
	TolviVersion string          `json:"tolvi_version"`
	OK           bool            `json:"ok"`
	Checks       []doctorCheckJS `json:"checks"`
}

type doctorCheckJS struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
	Fix    string `json:"fix"`
}

type vaultHealthJSON struct {
	TolviVersion string                 `json:"tolvi_version"`
	VaultPath    string                 `json:"vault_path"`
	FilesScanned int                    `json:"files_scanned"`
	OverallGrade string                 `json:"overall_grade"`
	OverallPct   float64                `json:"overall_pct"`
	HighSeverity int                    `json:"high_severity"`
	Grades       map[string]gradeJS     `json:"grades"`
	Findings     []vaultHealthFindingJS `json:"findings"`
}

type gradeJS struct {
	Pct    float64 `json:"pct"`
	Letter string  `json:"letter"`
}

type vaultHealthFindingJS struct {
	CheckID  string `json:"check_id"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Reason   string `json:"reason"`
}

// PrintDoctorJSON writes the setup checks as the published doctor shape.
func PrintDoctorJSON(w io.Writer, version string, checks []Check) error {
	out := doctorJSON{
		TolviVersion: version,
		OK:           DoctorFailures(checks) == 0,
		Checks:       make([]doctorCheckJS, 0, len(checks)),
	}
	for _, c := range checks {
		out.Checks = append(out.Checks, doctorCheckJS{
			ID: c.ID, Name: c.Name, OK: c.OK, Detail: c.Detail, Fix: c.Fix,
		})
	}
	return encode(w, out)
}

// PrintVaultHealthJSON writes a scan as the published vault-health shape.
func PrintVaultHealthJSON(w io.Writer, version string, rep HealthReport) error {
	out := vaultHealthJSON{
		TolviVersion: version,
		VaultPath:    rep.VaultPath,
		FilesScanned: rep.FilesScanned,
		OverallGrade: rep.OverallGrade,
		OverallPct:   rep.OverallPct,
		HighSeverity: HealthHighSeverity(rep),
		Grades:       make(map[string]gradeJS, len(rep.Grades)),
		// Never nil: a clean vault encodes as [] so consumers can iterate
		// unconditionally.
		Findings: make([]vaultHealthFindingJS, 0, len(rep.Findings)),
	}
	for dim, g := range rep.Grades {
		out.Grades[dim] = gradeJS{Pct: g.Pct, Letter: g.Letter}
	}
	for _, f := range rep.Findings {
		out.Findings = append(out.Findings, vaultHealthFindingJS{
			CheckID: f.CheckID, Severity: f.Severity, File: f.File, Reason: f.Reason,
		})
	}
	return encode(w, out)
}

func encode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// repos-list is the third published output shape. It lives here beside the
// other emitters so the wire structs for every --json command are in one file.

type reposListJSON struct {
	TolviVersion string       `json:"tolvi_version"`
	Repos        []repoRowJSN `json:"repos"`
}

type repoRowJSN struct {
	Path        string `json:"path"`
	Workspace   string `json:"workspace"`
	Repo        string `json:"repo,omitempty"`
	Product     string `json:"product,omitempty"`
	Status      string `json:"status"`
	VaultPath   string `json:"vault_path,omitempty"`
	SessionNote string `json:"session_note,omitempty"`
	Registered  string `json:"registered"`
}

func printReposListJSON(w io.Writer, version string, rows []repoSummary) error {
	out := reposListJSON{TolviVersion: version, Repos: make([]repoRowJSN, 0, len(rows))}
	for _, r := range rows {
		out.Repos = append(out.Repos, repoRowJSN{
			Path: r.Path, Workspace: r.Workspace, Repo: r.Repo, Product: r.Product,
			Status: r.Status, VaultPath: r.VaultPath, SessionNote: r.SessionNote,
			Registered: r.Registered,
		})
	}
	return encode(w, out)
}

// packs-list is the fourth published output shape.

type packsListJSON struct {
	TolviVersion string       `json:"tolvi_version"`
	Packs        []packJSONed `json:"packs"`
}

type packJSONed struct {
	Name      string           `json:"name"`
	Status    string           `json:"status"`
	Summary   string           `json:"summary"`
	Verticals []string         `json:"verticals"`
	Templates []packTemplateJS `json:"templates"`
}

type packTemplateJS struct {
	File string `json:"file"`
	Use  string `json:"use"`
}

func printPacksListJSON(w io.Writer, version string, all []packs.Pack) error {
	out := packsListJSON{TolviVersion: version, Packs: make([]packJSONed, 0, len(all))}
	for _, p := range all {
		row := packJSONed{
			Name: p.Name, Status: p.Status, Summary: p.Summary,
			Verticals: p.Verticals,
			Templates: make([]packTemplateJS, 0, len(p.Templates)),
		}
		if row.Verticals == nil {
			row.Verticals = []string{}
		}
		for _, t := range p.Templates {
			row.Templates = append(row.Templates, packTemplateJS{File: t.File, Use: t.Use})
		}
		out.Packs = append(out.Packs, row)
	}
	return encode(w, out)
}

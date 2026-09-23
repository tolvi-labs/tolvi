package cli

import (
	"encoding/json"
	"io"
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

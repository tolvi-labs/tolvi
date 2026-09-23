package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/tolvi-labs/tolvi/cli/internal/format"
)

// The JSON is a published contract, so these tests assert the wire shape and
// validate it against the schema shipped in spec/schemas/. A field rename that
// only updates the struct fails here rather than downstream.

// validateAgainst compiles an embedded schema and validates real command
// output against it, which is what keeps the schema and the emitter honest.
func validateAgainst(t *testing.T, schema []byte, name string, out []byte) {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(name, bytes.NewReader(schema)); err != nil {
		t.Fatalf("add %s: %v", name, err)
	}
	compiled, err := compiler.Compile(name)
	if err != nil {
		t.Fatalf("compile %s: %v", name, err)
	}
	var doc any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if err := compiled.Validate(doc); err != nil {
		t.Errorf("output does not satisfy %s: %v\n%s", name, err, out)
	}
}

func TestPrintDoctorJSON_ShapeAndStableIDs(t *testing.T) {
	var report bytes.Buffer
	checks, err := RunDoctor(healthyOpts(t, &report))
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := PrintDoctorJSON(&out, "v0.2.0", checks); err != nil {
		t.Fatal(err)
	}

	var got struct {
		TolviVersion string `json:"tolvi_version"`
		OK           bool   `json:"ok"`
		Checks       []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			OK     bool   `json:"ok"`
			Detail string `json:"detail"`
			Fix    string `json:"fix"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if got.TolviVersion != "v0.2.0" {
		t.Errorf("tolvi_version = %q, want v0.2.0", got.TolviVersion)
	}
	if !got.OK {
		t.Errorf("ok = false for a healthy setup: %s", out.String())
	}

	var ids []string
	for _, c := range got.Checks {
		ids = append(ids, c.ID)
		if c.Name == "" {
			t.Errorf("check %q has no name", c.ID)
		}
	}
	want := []string{"path", "vault", "api_key", "claude_permissions"}
	if len(ids) != len(want) {
		t.Fatalf("ids = %v, want %v", ids, want)
	}
	for i, id := range want {
		if ids[i] != id {
			t.Errorf("ids[%d] = %q, want %q", i, ids[i], id)
		}
	}
}

func TestPrintDoctorJSON_FailingCheckCarriesItsFixAndFlipsOK(t *testing.T) {
	var report bytes.Buffer
	opts := healthyOpts(t, &report)
	opts.Env = func(string) string { return "" } // no ANTHROPIC_API_KEY
	checks, err := RunDoctor(opts)
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := PrintDoctorJSON(&out, "dev", checks); err != nil {
		t.Fatal(err)
	}
	var got struct {
		OK     bool `json:"ok"`
		Checks []struct {
			ID  string `json:"id"`
			OK  bool   `json:"ok"`
			Fix string `json:"fix"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.OK {
		t.Error("ok = true although a check failed")
	}
	for _, c := range got.Checks {
		if c.ID != "api_key" {
			continue
		}
		if c.OK {
			t.Error("api_key check reported ok")
		}
		if c.Fix == "" {
			t.Error("a failing check carries no fix")
		}
		return
	}
	t.Error("no api_key check in the output")
}

func TestPrintDoctorJSON_ValidatesAgainstThePublishedSchema(t *testing.T) {
	var report bytes.Buffer
	checks, err := RunDoctor(healthyOpts(t, &report))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := PrintDoctorJSON(&out, "v0.2.0", checks); err != nil {
		t.Fatal(err)
	}
	validateAgainst(t, format.DoctorSchema, "doctor.json", out.Bytes())
}

func TestPrintVaultHealthJSON_ShapeAndSchema(t *testing.T) {
	rep, err := RunVaultHealth(healthVault(t, map[string]string{
		"decisions/2026-09-12-clean.md":    cleanDoc,
		"decisions/2026-09-13-untagged.md": "---\ndate: 2026-09-13\nstatus: active\n---\n\n# An untagged decision\n",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := PrintVaultHealthJSON(&out, "v0.2.0", rep); err != nil {
		t.Fatal(err)
	}

	var got struct {
		TolviVersion string `json:"tolvi_version"`
		VaultPath    string `json:"vault_path"`
		FilesScanned int    `json:"files_scanned"`
		OverallGrade string `json:"overall_grade"`
		OverallPct   float64
		HighSeverity int `json:"high_severity"`
		Grades       map[string]struct {
			Pct    float64 `json:"pct"`
			Letter string  `json:"letter"`
		} `json:"grades"`
		Findings []struct {
			CheckID  string `json:"check_id"`
			Severity string `json:"severity"`
			File     string `json:"file"`
			Reason   string `json:"reason"`
		} `json:"findings"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out.String())
	}
	if got.TolviVersion != "v0.2.0" {
		t.Errorf("tolvi_version = %q", got.TolviVersion)
	}
	if got.VaultPath == "" || got.FilesScanned == 0 || got.OverallGrade == "" {
		t.Errorf("empty scan summary: %s", out.String())
	}
	for _, dim := range []string{"tags", "lifecycle", "dedup", "completeness", "unicode"} {
		if _, ok := got.Grades[dim]; !ok {
			t.Errorf("grades has no %q dimension", dim)
		}
	}
	if got.HighSeverity != HealthHighSeverity(rep) {
		t.Errorf("high_severity = %d, want %d", got.HighSeverity, HealthHighSeverity(rep))
	}

	validateAgainst(t, format.VaultHealthSchema, "vault-health.json", out.Bytes())
}

func TestPrintVaultHealthJSON_FindingsAreAnArrayNotNull(t *testing.T) {
	rep := HealthReport{VaultPath: "/v", FilesScanned: 1, OverallGrade: "A+", Grades: map[string]HealthGrade{}}
	var out bytes.Buffer
	if err := PrintVaultHealthJSON(&out, "dev", rep); err != nil {
		t.Fatal(err)
	}
	// A consumer iterating findings must not have to special-case null.
	if !bytes.Contains(out.Bytes(), []byte(`"findings": []`)) {
		t.Errorf("empty findings did not encode as []: %s", out.String())
	}
	validateAgainst(t, format.VaultHealthSchema, "vault-health.json", out.Bytes())
}

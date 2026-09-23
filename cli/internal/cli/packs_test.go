package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
)

func TestPacksListJSON_ShapeAndSchema(t *testing.T) {
	var out bytes.Buffer
	if err := RunPacksList(PacksOpts{JSON: true, Stdout: &out, Version: "v0.2.0"}); err != nil {
		t.Fatal(err)
	}
	var got struct {
		TolviVersion string `json:"tolvi_version"`
		Packs        []struct {
			Name      string   `json:"name"`
			Status    string   `json:"status"`
			Verticals []string `json:"verticals"`
			Templates []struct {
				File string `json:"file"`
				Use  string `json:"use"`
			} `json:"templates"`
		} `json:"packs"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if got.TolviVersion != "v0.2.0" {
		t.Errorf("tolvi_version = %q", got.TolviVersion)
	}
	if len(got.Packs) != 6 {
		t.Fatalf("packs = %d, want the six vendored packs", len(got.Packs))
	}
	if got.Packs[0].Name != "cpa" {
		t.Errorf("packs are not sorted: first is %q", got.Packs[0].Name)
	}
	for _, p := range got.Packs {
		if p.Status != "available" || len(p.Templates) == 0 {
			t.Errorf("%s: status %q with %d templates", p.Name, p.Status, len(p.Templates))
		}
	}
	validateAgainst(t, format.PacksListSchema, "packs-list.json", out.Bytes())
}

func TestPacksListText_NamesEveryPackAndHowToUseOne(t *testing.T) {
	var out bytes.Buffer
	if err := RunPacksList(PacksOpts{Stdout: &out, Version: "dev"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"engineer", "security", "tolvi init --pack"} {
		if !bytes.Contains(out.Bytes(), []byte(want)) {
			t.Errorf("output does not mention %q:\n%s", want, out.String())
		}
	}
}

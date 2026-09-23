package packs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestList_ReturnsEveryVendoredPackSorted(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, p := range got {
		names = append(names, p.Name)
	}
	want := []string{"cpa", "data", "engineer", "entrepreneur", "security", "writer"}
	if len(names) != len(want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Errorf("names[%d] = %q, want %q", i, names[i], want[i])
		}
	}
}

func TestList_CarriesTheManifestFieldsAConsumerRenders(t *testing.T) {
	got, err := List()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range got {
		if p.Status == "" || p.Summary == "" {
			t.Errorf("%s: status or summary is empty", p.Name)
		}
		if len(p.Templates) == 0 {
			t.Errorf("%s: no templates", p.Name)
		}
		for _, tpl := range p.Templates {
			if tpl.File == "" || tpl.Use == "" {
				t.Errorf("%s: template %+v is missing a field", p.Name, tpl)
			}
		}
	}
}

func TestInstall_WritesEveryTemplateOfThePack(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "vault", "templates")

	written, err := Install("engineer", dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) == 0 {
		t.Fatal("Install wrote nothing")
	}
	for _, name := range written {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Errorf("%s not on disk: %v", name, err)
		}
	}
	// The manifest is the contract; it must not be written into the vault.
	if _, err := os.Stat(filepath.Join(dest, "pack.json")); err == nil {
		t.Error("pack.json was copied into the vault's templates")
	}
}

func TestInstall_UnknownPackNamesTheOnesThereAre(t *testing.T) {
	_, err := Install("nope", t.TempDir())
	if err == nil {
		t.Fatal("installing an unknown pack succeeded")
	}
	if !contains(err.Error(), "engineer") {
		t.Errorf("error does not list the available packs: %v", err)
	}
}

// Templates a user has edited are theirs. Re-running init should not quietly
// replace them.
func TestInstall_SkipsATemplateThatIsAlreadyThere(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "templates")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	mine := "# my own decision template\n"
	if err := os.WriteFile(filepath.Join(dest, "decision.md"), []byte(mine), 0o644); err != nil {
		t.Fatal(err)
	}

	written, err := Install("engineer", dest)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(dest, "decision.md"))
	if string(got) != mine {
		t.Error("an existing template was overwritten")
	}
	for _, name := range written {
		if name == "decision.md" {
			t.Error("decision.md was reported as written although it was skipped")
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

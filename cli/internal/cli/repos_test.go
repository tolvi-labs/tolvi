package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tolvi-labs/tolvi/cli/internal/format"
	"github.com/tolvi-labs/tolvi/cli/internal/registry"
)

func reposRepo(t *testing.T, root, name, workspace string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Join(dir, "vault", "decisions"), 0o755); err != nil {
		t.Fatal(err)
	}
	meta := `{"workspace":"` + workspace + `","repo":"` + name + `","embedding_model":"nomic-embed-text","schema_version":2}`
	if err := os.WriteFile(filepath.Join(dir, "vault", ".vault-meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func reposOpts(t *testing.T, out *bytes.Buffer) (ReposOpts, string) {
	t.Helper()
	dir := t.TempDir()
	reg := filepath.Join(dir, "cfg", "repos.json")
	// The resolver reads the machine's roots.json, so the test gets its own
	// config dir: without this the summary routes against whatever this
	// developer happens to have declared.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "cfg"))
	return ReposOpts{
		RegistryPath: reg,
		Stdout:       out,
		Today:        "2026-09-23",
		Version:      "v0.2.0",
	}, dir
}

func TestReposList_EmptyRegistrySaysSoAndIsNotAnError(t *testing.T) {
	var out bytes.Buffer
	opts, _ := reposOpts(t, &out)
	if err := RunReposList(opts); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		t.Error("an empty registry printed nothing at all")
	}
}

func TestReposListJSON_SummaryComesFromTheResolver(t *testing.T) {
	var out bytes.Buffer
	opts, dir := reposOpts(t, &out)
	repo := reposRepo(t, dir, "tolvi", "tolvi-labs")
	if err := registry.Add(opts.RegistryPath, repo); err != nil {
		t.Fatal(err)
	}
	opts.JSON = true
	if err := RunReposList(opts); err != nil {
		t.Fatal(err)
	}

	var got struct {
		TolviVersion string `json:"tolvi_version"`
		Repos        []struct {
			Path        string `json:"path"`
			Workspace   string `json:"workspace"`
			Repo        string `json:"repo"`
			Status      string `json:"status"`
			VaultPath   string `json:"vault_path"`
			SessionNote string `json:"session_note"`
			Registered  string `json:"registered"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("not valid JSON: %v\n%s", err, out.String())
	}
	if got.TolviVersion != "v0.2.0" {
		t.Errorf("tolvi_version = %q", got.TolviVersion)
	}
	if len(got.Repos) != 1 {
		t.Fatalf("repos = %v, want one", got.Repos)
	}
	r := got.Repos[0]
	if r.Path != repo || r.Workspace != "tolvi-labs" || r.Repo != "tolvi" {
		t.Errorf("entry = %+v, want the repo's identity", r)
	}
	if r.Status != "ok" {
		t.Errorf("status = %q, want ok", r.Status)
	}
	if want := filepath.Join(repo, "vault"); r.VaultPath != want {
		t.Errorf("vault_path = %q, want %q", r.VaultPath, want)
	}
	// The whole point of computing this here is that a consumer never
	// re-derives routing. With no roots.json the note belongs in the repo.
	if want := filepath.Join(repo, "vault", "sessions", "2026-09-23.md"); r.SessionNote != want {
		t.Errorf("session_note = %q, want %q", r.SessionNote, want)
	}
	if r.Registered == "" {
		t.Error("registered is empty")
	}

	validateAgainst(t, format.ReposListSchema, "repos-list.json", out.Bytes())
}

func TestReposListJSON_MovedRepoIsMarkedStaleAndKeepsItsEntry(t *testing.T) {
	var out bytes.Buffer
	opts, dir := reposOpts(t, &out)
	repo := reposRepo(t, dir, "tolvi", "tolvi-labs")
	if err := registry.Add(opts.RegistryPath, repo); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(repo); err != nil {
		t.Fatal(err)
	}

	opts.JSON = true
	if err := RunReposList(opts); err != nil {
		t.Fatal(err)
	}
	var got struct {
		Repos []struct {
			Path        string `json:"path"`
			Status      string `json:"status"`
			SessionNote string `json:"session_note"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Repos) != 1 {
		t.Fatalf("a stale entry was dropped rather than reported: %v", got.Repos)
	}
	if got.Repos[0].Status != "stale" {
		t.Errorf("status = %q, want stale", got.Repos[0].Status)
	}
	if got.Repos[0].SessionNote != "" {
		t.Errorf("session_note = %q, want empty for a repo that is not there", got.Repos[0].SessionNote)
	}
	validateAgainst(t, format.ReposListSchema, "repos-list.json", out.Bytes())
}

func TestReposListJSON_EmptyRegistryEncodesReposAsArray(t *testing.T) {
	var out bytes.Buffer
	opts, _ := reposOpts(t, &out)
	opts.JSON = true
	if err := RunReposList(opts); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"repos": []`)) {
		t.Errorf("empty registry did not encode repos as []: %s", out.String())
	}
	validateAgainst(t, format.ReposListSchema, "repos-list.json", out.Bytes())
}

func TestReposScan_RebuildsFromADirectoryAndDropsWhatIsGone(t *testing.T) {
	var out bytes.Buffer
	opts, dir := reposOpts(t, &out)
	gone := reposRepo(t, dir, "gone", "tolvi-labs")
	if err := registry.Add(opts.RegistryPath, gone); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(gone); err != nil {
		t.Fatal(err)
	}
	live := reposRepo(t, dir, "here", "tolvi-labs")

	opts.Dirs = []string{dir}
	if err := RunReposScan(opts); err != nil {
		t.Fatal(err)
	}
	f, err := registry.Load(opts.RegistryPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Repos) != 1 || f.Repos[0].Path != live {
		t.Errorf("registry = %v, want only %s", f.Repos, live)
	}
}

func TestReposForget_ReportsWhetherAnythingWasRemoved(t *testing.T) {
	var out bytes.Buffer
	opts, dir := reposOpts(t, &out)
	repo := reposRepo(t, dir, "tolvi", "tolvi-labs")
	if err := registry.Add(opts.RegistryPath, repo); err != nil {
		t.Fatal(err)
	}

	opts.Dirs = []string{repo}
	if err := RunReposForget(opts); err != nil {
		t.Fatal(err)
	}
	f, _ := registry.Load(opts.RegistryPath)
	if len(f.Repos) != 0 {
		t.Errorf("registry = %v, want empty", f.Repos)
	}

	out.Reset()
	if err := RunReposForget(opts); err != nil {
		t.Fatalf("forgetting an unregistered path errored: %v", err)
	}
	if !bytes.Contains(bytes.ToLower(out.Bytes()), []byte("not registered")) {
		t.Errorf("second forget did not say the path was not registered: %s", out.String())
	}
}

// The registry is a cache over vaults, so failing to update it must never fail
// the command that provisioned the vault.
func TestRegisterAfterInit_FailureWarnsAndNamesTheRepair(t *testing.T) {
	var out bytes.Buffer
	dir := t.TempDir()
	repo := reposRepo(t, dir, "tolvi", "tolvi-labs")

	// A registry path that cannot be written: the parent is a file.
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	RegisterAfterInit(&out, filepath.Join(blocker, "repos.json"), repo)

	got := out.String()
	if got == "" {
		t.Fatal("a failed registration said nothing")
	}
	if !bytes.Contains([]byte(got), []byte("tolvi repos scan")) {
		t.Errorf("warning does not name the repair: %q", got)
	}
}

func TestRegisterAfterInit_SuccessIsSilent(t *testing.T) {
	var out bytes.Buffer
	dir := t.TempDir()
	repo := reposRepo(t, dir, "tolvi", "tolvi-labs")
	reg := filepath.Join(dir, "cfg", "repos.json")

	RegisterAfterInit(&out, reg, repo)
	if out.Len() != 0 {
		t.Errorf("a successful registration printed %q, want nothing", out.String())
	}
	f, _ := registry.Load(reg)
	if len(f.Repos) != 1 {
		t.Errorf("init did not register the repo: %v", f.Repos)
	}
}

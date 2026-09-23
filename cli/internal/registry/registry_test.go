package registry

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// repoAt provisions a repo whose vault carries a meta, which is what makes a
// path registrable at all.
func repoAt(t *testing.T, root, name, workspace string) string {
	t.Helper()
	dir := filepath.Join(root, name)
	v := filepath.Join(dir, "vault", "decisions")
	if err := os.MkdirAll(v, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := `{"workspace":"` + workspace + `","repo":"` + name + `","embedding_model":"nomic-embed-text","schema_version":2}`
	if err := os.WriteFile(filepath.Join(dir, "vault", ".vault-meta.json"), []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoad_AbsentFileIsEmptyNotAnError(t *testing.T) {
	f, err := Load(filepath.Join(t.TempDir(), "repos.json"))
	if err != nil {
		t.Fatalf("absent registry should load empty: %v", err)
	}
	if len(f.Repos) != 0 {
		t.Errorf("Repos = %v, want empty", f.Repos)
	}
}

func TestAddThenLoad_RoundTrips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg", "repos.json")
	repo := repoAt(t, dir, "tolvi", "tolvi-labs")

	if err := Add(path, repo); err != nil {
		t.Fatal(err)
	}
	f, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Repos) != 1 {
		t.Fatalf("Repos = %v, want one entry", f.Repos)
	}
	e := f.Repos[0]
	if e.Path != repo || e.Workspace != "tolvi-labs" || e.Repo != "tolvi" {
		t.Errorf("entry = %+v, want the repo's own identity", e)
	}
	if _, err := time.Parse(time.RFC3339, e.Registered); err != nil {
		t.Errorf("registered = %q, want RFC3339: %v", e.Registered, err)
	}
}

func TestAdd_IsIdempotentAndKeepsTheFirstRegistration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.json")
	repo := repoAt(t, dir, "tolvi", "tolvi-labs")

	if err := Add(path, repo); err != nil {
		t.Fatal(err)
	}
	first, _ := Load(path)
	if err := Add(path, repo); err != nil {
		t.Fatal(err)
	}
	again, _ := Load(path)

	if len(again.Repos) != 1 {
		t.Fatalf("re-registering duplicated the entry: %v", again.Repos)
	}
	if again.Repos[0].Registered != first.Repos[0].Registered {
		t.Error("re-registering rewrote the original registration time")
	}
}

func TestAdd_RefusesAPathWithNoVault(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "not-a-repo")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Add(filepath.Join(dir, "repos.json"), plain); err == nil {
		t.Error("registered a directory with no vault")
	}
}

func TestForget_RemovesOnlyThatEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.json")
	a := repoAt(t, dir, "tolvi", "tolvi-labs")
	b := repoAt(t, dir, "forge", "tolvi-labs")
	mustAdd(t, path, a, b)

	removed, err := Forget(path, a)
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Error("Forget reported no entry removed")
	}
	f, _ := Load(path)
	if len(f.Repos) != 1 || f.Repos[0].Path != b {
		t.Errorf("Repos = %v, want only %s", f.Repos, b)
	}
}

func TestForget_UnknownPathIsNotAnError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.json")
	mustAdd(t, path, repoAt(t, dir, "tolvi", "tolvi-labs"))

	removed, err := Forget(path, filepath.Join(dir, "never-registered"))
	if err != nil {
		t.Fatalf("forgetting an unknown path errored: %v", err)
	}
	if removed {
		t.Error("Forget claimed to remove an entry that was not there")
	}
}

// A moved repo is marked, never dropped: the registry is a cache over vaults,
// and silently deleting an entry loses the only record that it was ever there.
func TestVerify_MarksAMovedRepoStaleRatherThanDeletingIt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.json")
	repo := repoAt(t, dir, "tolvi", "tolvi-labs")
	mustAdd(t, path, repo)

	if err := os.RemoveAll(repo); err != nil {
		t.Fatal(err)
	}
	f, _ := Load(path)
	if got := Verify(f.Repos[0]); got != StatusStale {
		t.Errorf("status = %q, want %q", got, StatusStale)
	}
	if len(f.Repos) != 1 {
		t.Error("loading dropped the entry instead of keeping it for a forget")
	}
}

func TestVerify_MarksAPathThatLostItsMetaStale(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.json")
	repo := repoAt(t, dir, "tolvi", "tolvi-labs")
	mustAdd(t, path, repo)

	if err := os.Remove(filepath.Join(repo, "vault", ".vault-meta.json")); err != nil {
		t.Fatal(err)
	}
	f, _ := Load(path)
	if got := Verify(f.Repos[0]); got != StatusStale {
		t.Errorf("status = %q, want %q", got, StatusStale)
	}
}

func TestScan_FindsEveryVaultUnderADirectory(t *testing.T) {
	dir := t.TempDir()
	a := repoAt(t, dir, "tolvi", "tolvi-labs")
	b := repoAt(t, filepath.Join(dir, "nested"), "forge", "tolvi-labs")
	if err := os.MkdirAll(filepath.Join(dir, "plain", "src"), 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, e := range found {
		paths = append(paths, e.Path)
	}
	if len(paths) != 2 {
		t.Fatalf("Scan found %v, want exactly %s and %s", paths, a, b)
	}
}

func TestScan_SkipsDotDirectoriesAndNodeModules(t *testing.T) {
	dir := t.TempDir()
	repoAt(t, filepath.Join(dir, "node_modules"), "vendored", "tolvi-labs")
	repoAt(t, filepath.Join(dir, ".git"), "hidden", "tolvi-labs")
	want := repoAt(t, dir, "real", "tolvi-labs")

	found, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Path != want {
		t.Errorf("Scan = %v, want only %s", found, want)
	}
}

func TestScan_ResultIsSortedSoOutputIsStable(t *testing.T) {
	dir := t.TempDir()
	repoAt(t, dir, "zeta", "tolvi-labs")
	repoAt(t, dir, "alpha", "tolvi-labs")

	found, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 2 || filepath.Base(found[0].Path) != "alpha" {
		t.Errorf("Scan = %v, want alpha first", found)
	}
}

func TestReplace_RebuildsFromNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "repos.json")
	stale := repoAt(t, dir, "gone", "tolvi-labs")
	mustAdd(t, path, stale)
	if err := os.RemoveAll(stale); err != nil {
		t.Fatal(err)
	}
	live := repoAt(t, dir, "here", "tolvi-labs")

	found, err := Scan([]string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if err := Replace(path, found); err != nil {
		t.Fatal(err)
	}
	f, _ := Load(path)
	if len(f.Repos) != 1 || f.Repos[0].Path != live {
		t.Errorf("Repos = %v, want only the live repo", f.Repos)
	}
}

func mustAdd(t *testing.T, path string, repos ...string) {
	t.Helper()
	for _, r := range repos {
		if err := Add(path, r); err != nil {
			t.Fatal(err)
		}
	}
}

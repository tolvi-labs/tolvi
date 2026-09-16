package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// writeRoots drops a roots.json into a temp dir and returns its path.
func writeRoots(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "roots.json")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write roots.json: %v", err)
	}
	return path
}

// tolviLabsRoots is the shape step 5 of the topology plan writes: org roots for
// three workspaces and one product root.
const tolviLabsRoots = `{
  "roots": [
    {"role": "org", "workspace": "tolvi-labs", "path": "/vaults/acme-org/vault"},
    {"role": "org", "workspace": "acme", "path": "/vaults/acme-shared/vault"},
    {"role": "org", "workspace": "isolated-org", "path": "/vaults/isolated-org/vault"},
    {"role": "product", "product": "bravo-os", "path": "/vaults/bravo-vault/vault"}
  ]
}`

func loadTolviLabs(t *testing.T) Roots {
	t.Helper()
	r, err := LoadRoots(writeRoots(t, tolviLabsRoots))
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	return r
}

// --- LoadRoots ---

func TestLoadRootsMissingFileIsSingleRootMode(t *testing.T) {
	r, err := LoadRoots(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("absent roots.json must not error (it is the contributor state): %v", err)
	}
	if r.Declared() {
		t.Fatal("absent roots.json must leave Roots undeclared, so writes fall back to the repo vault")
	}
}

func TestLoadRootsMalformedIsAnError(t *testing.T) {
	if _, err := LoadRoots(writeRoots(t, "{not json")); err == nil {
		t.Fatal("expected an error for a malformed roots.json — ignoring it publishes internal notes")
	}
}

func TestLoadRootsRejectsUnknownRole(t *testing.T) {
	_, err := LoadRoots(writeRoots(t, `{"roots":[{"role":"wardrobe","path":"/x"}]}`))
	if err == nil {
		t.Fatal("expected an error for an unknown role")
	}
}

func TestLoadRootsRejectsOrgRootWithoutWorkspace(t *testing.T) {
	_, err := LoadRoots(writeRoots(t, `{"roots":[{"role":"org","path":"/x"}]}`))
	if err == nil {
		t.Fatal("an org root without a workspace can never be selected; expected an error")
	}
}

// --- Chain ---

func TestChainOrdersRootsNearestFirst(t *testing.T) {
	c := loadTolviLabs(t).Chain(Identity{Workspace: "bravo-os", Repo: "bravo-web", Product: "bravo-os"}, "/repo/vault")

	// bravo-web declares no matching org root, so the chain is repo + product.
	want := []RootRole{RoleRepo, RoleProduct}
	got := make([]RootRole, 0, len(c.Roots()))
	for _, r := range c.Roots() {
		got = append(got, r.Role)
	}
	if len(got) != len(want) {
		t.Fatalf("chain roles = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("chain roles = %v, want %v", got, want)
		}
	}
}

func TestChainSelectsOrgRootByWorkspace(t *testing.T) {
	c := loadTolviLabs(t).Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")

	org, ok := c.Role(RoleOrg)
	if !ok {
		t.Fatal("expected an org root in the chain")
	}
	if org.Path != "/vaults/acme-org/vault" {
		t.Errorf("org root = %q, want the tolvi-labs root", org.Path)
	}
}

func TestChainExcludesOtherWorkspacesOrgRoot(t *testing.T) {
	// Isolation is structural: a repo in one workspace must never resolve
	// another workspace's org root, with no flag to forget.
	c := loadTolviLabs(t).Chain(Identity{Workspace: "isolated-org", Repo: "isolated-web"}, "/repo/vault")

	for _, r := range c.Roots() {
		if r.Path == "/vaults/acme-shared/vault" {
			t.Fatalf("chain contains another workspace's root (%s); isolation must be structural", r.Path)
		}
	}
}

func TestChainOmitsProductWhenIdentityDeclaresNone(t *testing.T) {
	c := loadTolviLabs(t).Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")

	if _, ok := c.Role(RoleProduct); ok {
		t.Fatal("a repo that declares no product must not get a product root")
	}
}

// --- Target: the resolution table ---

func TestTargetResolutionTable(t *testing.T) {
	roots := loadTolviLabs(t)
	// tolvi: workspace tolvi-labs, no product. Chain = repo, org.
	c := roots.Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")

	cases := []struct {
		name       string
		docType    string
		visibility string
		want       string
	}{
		{"public decision stays in the repo vault", "decision", "", "/repo/vault"},
		{"private decision routes to the org root", "decision", "private", "/vaults/acme-org/vault"},
		{"public pattern stays in the repo vault", "pattern", "", "/repo/vault"},
		{"private pattern routes to the org root", "pattern", "private", "/vaults/acme-org/vault"},
		{"a session always routes private", "session", "", "/vaults/acme-org/vault"},
		{"an explicitly private session routes private too", "session", "private", "/vaults/acme-org/vault"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.Target(tc.docType, tc.visibility)
			if err != nil {
				t.Fatalf("Target(%q, %q): %v", tc.docType, tc.visibility, err)
			}
			if got.Path != tc.want {
				t.Errorf("Target(%q, %q) = %q, want %q", tc.docType, tc.visibility, got.Path, tc.want)
			}
		})
	}
}

func TestTargetPrefersProductOverOrgForPrivateDocs(t *testing.T) {
	// "The nearest declared private root above its scope" — product sits
	// between repo and org, so a repo that declares one routes there.
	c := loadTolviLabs(t).Chain(Identity{Workspace: "bravo-os", Repo: "bravo-web", Product: "bravo-os"}, "/repo/vault")

	got, err := c.Target("decision", "private")
	if err != nil {
		t.Fatalf("Target: %v", err)
	}
	if got.Path != "/vaults/bravo-vault/vault" {
		t.Errorf("private decision routed to %q, want the product root", got.Path)
	}
}

func TestTargetSingleRootModeIgnoresVisibility(t *testing.T) {
	// No roots.json at all is the external contributor. Their session note
	// belongs with their PR, in the repo's own vault.
	r, err := LoadRoots(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	c := r.Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")

	for _, docType := range []string{"session", "decision", "pattern"} {
		got, err := c.Target(docType, "private")
		if err != nil {
			t.Fatalf("Target(%q) in single-root mode: %v", docType, err)
		}
		if got.Path != "/repo/vault" {
			t.Errorf("Target(%q) = %q, want the repo vault in single-root mode", docType, got.Path)
		}
	}
}

func TestTargetRefusesWhenDeclaredConfigLacksTheNeededRoot(t *testing.T) {
	// A roots.json that exists but has no private root for this workspace is a
	// misconfiguration, not a contributor. Falling back to the repo vault is
	// exactly how an internal note gets published, so the write is refused.
	r, err := LoadRoots(writeRoots(t, `{"roots":[{"role":"org","workspace":"other-org","path":"/vaults/other/vault"}]}`))
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	c := r.Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")

	if _, err := c.Target("session", ""); err == nil {
		t.Fatal("expected a refusal: declared config with no private root must not fall back to the public repo vault")
	}
}

// --- ReadRoots ---

func TestReadRootsDedupesSharedPaths(t *testing.T) {
	// Two roles may be declared over one directory. Reading it twice would
	// double every doc it holds.
	r, err := LoadRoots(writeRoots(t, `{
  "roots": [
    {"role": "org", "workspace": "tolvi-labs", "path": "/vaults/shared/vault"},
    {"role": "product", "product": "stack", "path": "/vaults/shared/vault"}
  ]
}`))
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	c := r.Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi", Product: "stack"}, "/repo/vault")

	seen := map[string]int{}
	for _, rt := range c.ReadRoots() {
		seen[rt.Path]++
	}
	if seen["/vaults/shared/vault"] != 1 {
		t.Errorf("shared path appears %d times in ReadRoots, want 1", seen["/vaults/shared/vault"])
	}
}

func TestLoadRootsRejectsTheDroppedPersonalRole(t *testing.T) {
	// The fourth role was inferred from a directory that turned out to be an
	// org root, not a personal one. Three roles describe the real topology;
	// declaring the old one should fail loudly rather than be ignored.
	if _, err := LoadRoots(writeRoots(t, `{"roots":[{"role":"personal","path":"/x"}]}`)); err == nil {
		t.Fatal("expected an error: the personal role no longer exists")
	}
}

func TestLoadRootsRejectsADeclaredRepoRoot(t *testing.T) {
	// The repo root is always the vault discovered from the working directory.
	// Declaring one would let roots.json point a repo at someone else's vault.
	if _, err := LoadRoots(writeRoots(t, `{"roots":[{"role":"repo","path":"/x"}]}`)); err == nil {
		t.Fatal("expected an error: the repo root is discovered, never declared")
	}
}

// TestTargetIsAlwaysReadable is the invariant the whole topology exists to
// enforce: every destination write can choose must be somewhere read looks.
// Neither routing bug fixed on 2026-09-12 could have survived this test.
func TestTargetIsAlwaysReadable(t *testing.T) {
	roots := loadTolviLabs(t)

	identities := []Identity{
		{Workspace: "tolvi-labs", Repo: "tolvi"},
		{Workspace: "tolvi-labs", Repo: "acme-site"},
		{Workspace: "acme", Repo: "bravo-web"},
		{Workspace: "bravo-os", Repo: "bravo-web", Product: "bravo-os"},
		{Workspace: "isolated-org", Repo: "isolated-web"},
		{Workspace: "unconfigured-org", Repo: "some-repo"},
	}

	for _, id := range identities {
		c := roots.Chain(id, "/repo/vault")
		readable := map[string]bool{}
		for _, r := range c.ReadRoots() {
			readable[r.Path] = true
		}

		for _, docType := range []string{"decision", "session", "pattern"} {
			for _, vis := range []string{"", "public", "private"} {
				target, err := c.Target(docType, vis)
				if err != nil {
					continue // a refusal writes nothing, so there is nothing to read back
				}
				if !readable[target.Path] {
					t.Errorf("%s/%s: Target(%q, %q) = %q, which ReadRoots never reads",
						id.Workspace, id.Repo, docType, vis, target.Path)
				}
			}
		}
	}
}

// --- Accepts ---

func TestAcceptsFiltersSharedRootsByRepo(t *testing.T) {
	c := loadTolviLabs(t).Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")
	org, ok := c.Role(RoleOrg)
	if !ok {
		t.Fatal("expected an org root")
	}

	if !c.Accepts(org, DocScope{Repo: "tolvi"}) {
		t.Error("org root must yield this repo's own docs")
	}
	if c.Accepts(org, DocScope{Repo: "acme-site"}) {
		t.Error("org root must not yield a sibling repo's docs")
	}
}

func TestAcceptsTakesEverythingFromTheRepoRoot(t *testing.T) {
	c := loadTolviLabs(t).Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")
	repo, ok := c.Role(RoleRepo)
	if !ok {
		t.Fatal("expected a repo root")
	}

	// The repo's own vault is unambiguously its own; an untagged doc there
	// belongs to it, and requiring a repo: tag would hide contributor docs.
	if !c.Accepts(repo, DocScope{}) {
		t.Error("repo root must yield its own untagged docs")
	}
}

func TestAcceptsRejectsContainerScopedDocsFromSharedRoots(t *testing.T) {
	// Accepted risk, chosen twice: a doc with no repo: tag in a shared root is
	// container-scoped and surfaces in no repo's recall.
	c := loadTolviLabs(t).Chain(Identity{Workspace: "tolvi-labs", Repo: "tolvi"}, "/repo/vault")
	org, _ := c.Role(RoleOrg)

	if c.Accepts(org, DocScope{}) {
		t.Error("an untagged doc in a shared root is container-scoped and must not surface in a repo's recall")
	}
}

package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// helper: create a vault at <dir>/vault/
func mkVault(t *testing.T, dir string) string {
	t.Helper()
	vaultDir := filepath.Join(dir, "vault")
	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteMeta(vaultDir, Meta{Workspace: "test", EmbeddingModel: "nomic-embed-text", SchemaVersion: 1}); err != nil {
		t.Fatal(err)
	}
	return vaultDir
}

func TestDiscover_StartingInVaultRoot(t *testing.T) {
	root := t.TempDir()
	vaultDir := mkVault(t, root)
	got, err := Discover(DiscoverOpts{StartDir: root, HomeDir: t.TempDir() /* unrelated */})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != vaultDir {
		t.Errorf("Discover = %q, want %q", got, vaultDir)
	}
}

func TestDiscover_WalkUpFromSubdir(t *testing.T) {
	root := t.TempDir()
	vaultDir := mkVault(t, root)
	deep := filepath.Join(root, "src", "pkg", "deep")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Discover(DiscoverOpts{StartDir: deep, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != vaultDir {
		t.Errorf("Discover = %q, want %q", got, vaultDir)
	}
}

func TestDiscover_StartingInsideVaultSubdir(t *testing.T) {
	root := t.TempDir()
	vaultDir := mkVault(t, root)
	decisions := filepath.Join(vaultDir, "decisions")
	if err := os.MkdirAll(decisions, 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := Discover(DiscoverOpts{StartDir: decisions, HomeDir: t.TempDir()})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != vaultDir {
		t.Errorf("Discover = %q, want %q", got, vaultDir)
	}
}

func TestDiscover_NoVaultFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Discover(DiscoverOpts{StartDir: dir, HomeDir: t.TempDir()})
	if err == nil {
		t.Fatal("expected error when no vault found")
	}
}

func TestDiscover_WalkStopsAtHome(t *testing.T) {
	home := t.TempDir()
	// Put a vault ABOVE home — should not be found.
	above := filepath.Dir(home)
	mkVault(t, above)
	// Start inside home with no vault inside.
	inHome := filepath.Join(home, "project")
	if err := os.MkdirAll(inHome, 0o755); err != nil {
		t.Fatal(err)
	}
	_, err := Discover(DiscoverOpts{StartDir: inHome, HomeDir: home})
	if err == nil {
		t.Fatal("expected walk to stop at $HOME and not find vault above it")
	}
}

func TestDiscover_ExplicitOverride(t *testing.T) {
	overrideRoot := t.TempDir()
	vaultDir := mkVault(t, overrideRoot)
	// Start somewhere completely different; --vault override wins.
	startDir := t.TempDir()
	got, err := Discover(DiscoverOpts{
		StartDir:     startDir,
		HomeDir:      t.TempDir(),
		ExplicitPath: vaultDir,
	})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != vaultDir {
		t.Errorf("Discover = %q, want %q", got, vaultDir)
	}
}

func TestDiscover_ExplicitOverride_InvalidPath(t *testing.T) {
	_, err := Discover(DiscoverOpts{
		StartDir:     t.TempDir(),
		HomeDir:      t.TempDir(),
		ExplicitPath: "/nonexistent/path",
	})
	if err == nil {
		t.Fatal("expected error on invalid --vault path")
	}
}

func TestDiscover_DefaultVaultFallback(t *testing.T) {
	defaultRoot := t.TempDir()
	vaultDir := mkVault(t, defaultRoot)
	startDir := t.TempDir()
	got, err := Discover(DiscoverOpts{
		StartDir:     startDir,
		HomeDir:      t.TempDir(),
		DefaultVault: vaultDir,
	})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if got != vaultDir {
		t.Errorf("Discover = %q, want %q", got, vaultDir)
	}
}

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Integration tests run the real binary, and most runs set no environment of
// their own. `tolvi init` registers its vault in the machine's repos.json, so
// without isolation every test run added a temp vault to the developer's real
// ~/.config/tolvi/repos.json. The check runs before init so that a failure here
// cannot itself write to the real registry.
func TestIntegration_InitRegistersIntoTheIsolatedConfig(t *testing.T) {
	cfg := os.Getenv("XDG_CONFIG_HOME")
	tmp, _ := filepath.EvalSymlinks(os.TempDir())
	resolved, _ := filepath.EvalSymlinks(cfg)
	if cfg == "" || !strings.HasPrefix(resolved, tmp) {
		t.Fatalf("XDG_CONFIG_HOME is %q, not a temp dir: spawned binaries would write the developer's real config", cfg)
	}

	bin := buildToTmp(t)
	work := t.TempDir()
	init := exec.Command(bin, "init", "--workspace", "isolation-it")
	init.Dir = work
	if out, err := init.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}

	data, err := os.ReadFile(filepath.Join(cfg, "tolvi", "repos.json"))
	if err != nil {
		t.Fatalf("init did not register into the isolated config: %v", err)
	}
	var reg struct {
		Repos []struct {
			Path string `json:"path"`
		} `json:"repos"`
	}
	if err := json.Unmarshal(data, &reg); err != nil {
		t.Fatal(err)
	}
	for _, r := range reg.Repos {
		if filepath.Base(r.Path) == filepath.Base(work) {
			return
		}
	}
	t.Errorf("isolated repos.json does not list %s: %s", work, data)
}

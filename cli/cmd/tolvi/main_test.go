package main

import (
	"fmt"
	"os"
	"testing"
)

// TestMain points XDG_CONFIG_HOME at a temp dir for the whole test process, so
// every binary a test spawns inherits isolation by default. Most integration
// runs set no environment of their own, and `tolvi init` registers into the
// machine's repos.json, which left temp vaults in the developer's real
// ~/.config/tolvi/repos.json. HOME is left alone because the tests' git commits
// read the developer's global git identity.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tolvi-test-config-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Setenv("XDG_CONFIG_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

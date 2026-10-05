package main

import (
	"runtime/debug"
	"testing"
)

func buildInfo(v string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Path: "github.com/tolvi-labs/tolvi/cli", Version: v}}, true
	}
}

func TestResolveVersion_ReleaseLdflagWins(t *testing.T) {
	if got := resolveVersion("v0.3.0", buildInfo("v0.2.0")); got != "v0.3.0" {
		t.Fatalf("got %q, want the ldflag version", got)
	}
}

func TestResolveVersion_GoInstallReportsTheModuleVersion(t *testing.T) {
	if got := resolveVersion("dev", buildInfo("v0.3.0")); got != "v0.3.0" {
		t.Fatalf("got %q, want the module version go install recorded", got)
	}
}

func TestResolveVersion_UnversionedBuildsStayDev(t *testing.T) {
	for _, mod := range []string{"", "(devel)"} {
		if got := resolveVersion("dev", buildInfo(mod)); got != "dev" {
			t.Fatalf("module version %q: got %q, want dev", mod, got)
		}
	}
	none := func() (*debug.BuildInfo, bool) { return nil, false }
	if got := resolveVersion("dev", none); got != "dev" {
		t.Fatalf("no build info: got %q, want dev", got)
	}
}

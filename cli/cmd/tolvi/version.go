package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is baked at release time via -ldflags "-X main.version=v0.1.0". A build from
// `go install ...@vX.Y.Z` has no ldflags, so init falls back to the module version Go
// records in the binary; only a build with neither reports "dev".
var version = "dev"

func init() {
	version = resolveVersion(version, debug.ReadBuildInfo)
}

// resolveVersion prefers the release ldflag, then the module version from the build info.
// "(devel)" is what Go records for a build with no module version, so it stays "dev".
func resolveVersion(ldflag string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	if ldflag != "dev" {
		return ldflag
	}
	if info, ok := readBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return ldflag
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Tolvi CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version)
		return nil
	},
}

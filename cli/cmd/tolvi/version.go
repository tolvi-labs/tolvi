package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is baked at release time via -ldflags "-X main.version=v0.1.0".
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Tolvi CLI version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(version)
		return nil
	},
}

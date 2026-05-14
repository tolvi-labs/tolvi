package cli

// Exit codes per docs/superpowers/specs/2026-05-14-phase-3-cli-design.md §6.
const (
	ExitOK         = 0
	ExitInternal   = 1
	ExitUserInput  = 2
	ExitVaultState = 3
	ExitConfig     = 4
	ExitNetwork    = 5
)

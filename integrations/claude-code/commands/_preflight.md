<!--
Shared preflight contract. Every command in this directory that prefers the
`tolvi` CLI includes a copy of the block below, between the PREFLIGHT markers.
Edit it here and re-run `.github/scripts/preflight-sync-check.sh` to find copies
that have drifted.
-->

<!-- PREFLIGHT:BEGIN -->
## Preflight — say it once if the CLI is missing

Before the steps below, check whether the `tolvi` CLI is available:

    command -v tolvi

**If it is present**, use it. It is one invocation, it discovers the vault itself, and it does not raise a permission prompt per file.

**If it is absent**, fall back to reading the vault directly (the steps below work either way), and tell the user exactly once per conversation:

> `!` tolvi CLI not on PATH. Reading the vault directly, which works but skips
> semantic retrieval and costs one shell call per step. Fix: `tolvi doctor`,
> or install with `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest`
> and add `$(go env GOPATH)/bin` to your PATH.

**Once per conversation means once.** If you have already reported this in the current conversation, do not repeat it — later commands in the same session stay quiet. You know what you have already said; no marker file is needed. A user who has chosen not to install the CLI should not be told four times in one session, because a warning repeated that often stops being read.

Never silently degrade. The fallback path is legitimate and produces real answers, but the user has to learn once that they are on it, or a broken install looks identical to a working one.
<!-- PREFLIGHT:END -->

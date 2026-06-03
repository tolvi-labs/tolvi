#!/usr/bin/env bash
# commit-sync-nudge.sh — Claude Code PostToolUse(git commit) hook for Tolvi.
#
# Fires after any Bash call whose command matches "git commit*".
# Injects additionalContext so Claude proactively offers to run
# /sync-session when the current task is complete — not mid-task.
#
# Silently exits 0 when not in a Tolvi-vaulted repo or if the
# tolvi binary is missing.
#
# Installed by: integrations/claude-code/install.sh --with-hooks
# See: https://github.com/tolvi-labs/tolvi

command -v tolvi >/dev/null 2>&1 || exit 0

# Walk up from cwd to find vault/.vault-meta.json — mirrors
# vault.Discover so the nudge only fires in Tolvi repos.
dir="$PWD"
in_tolvi_repo=false
while [[ "$dir" != "/" && "$dir" != "$HOME" ]]; do
  if [[ -f "$dir/vault/.vault-meta.json" ]]; then
    in_tolvi_repo=true
    break
  fi
  dir="$(dirname "$dir")"
done
[[ "$in_tolvi_repo" == "true" ]] || exit 0

printf '%s\n' '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"A git commit was just made. When the current task is complete (not right now), offer to run /sync-session to capture this work in the vault. One brief mention is enough; do not interrupt the current task."}}'

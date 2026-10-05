---
tags: [decision, tolvi]
date: 2026-10-05
repo: tolvi
status: active
supersedes: [[2026-06-05-pretooluse-enforcement-for-vault-sync]]
ticket: none
user_impact: high
product_area: Claude Code integration
---

# The commit hooks check their own command, and answer Claude Code in hookSpecificOutput

**Date:** 2026-10-05
**Repo:** tolvi

## Why
A user found that the commit hook fired on every shell command, not only on commits, and Claude Code's documentation says the output format the hook used is not supported for this kind of hook at all. So the hook could interfere with unrelated commands while its intended block or warning may never have reached Claude.

## How
- Applies to both `tolvi/skills/tolvi/hooks/tolvi-sync` (with its embedded copy in `cli/internal/integrations/files/hooks/`, kept identical by `integration-copy-parity-check.sh`) and `tolvi-solo/hooks/tolvi-sync`.
- **The hook checks its own command.** It reads the PreToolUse payload from stdin and exits 0 with no output unless `tool_input.command` matches `(^|[;&|(]\s*)git(\s+-[cC]\s+\S+)*\s+commit(\s|$)`. The settings entry keeps `"if": "Bash(git commit*)"`, which current Claude Code documents and honors, but a reporter's older version ignored it and the hook then staged `vault/` and blocked a `python3` command. The filter in the settings is an optimization; the check in the hook is the guarantee.
- **Output is `hookSpecificOutput`.** The docs state the top-level `decision` field is not supported for PreToolUse. A block is `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"..."}}`; a note to Claude is `{"hookSpecificOutput":{"hookEventName":"PreToolUse","additionalContext":"..."}}`.
- **The hook never emits `"allow"`.** `permissionDecision: "allow"` skips the user's permission prompt, and `git commit` is deliberately left as a conscious per-call approval (the installers allowlist only the read-only `tolvi recall` and `tolvi ask`). Exit 0 with only context leaves the normal permission flow in place.
- **The two products differ on blocking.** tolvi's hook is warn-only (a missing note adds context and the commit proceeds), because a fail-closed hook turns any hook bug into an inability to commit. tolvi-solo's gate denies a commit with no session note for today. This decision changes neither stance, only how each one is expressed.
- Unchanged from the superseded decision: the hook auto-stages `vault/` on a commit, and exits silently when no `vault/.vault-meta.json` is found above `$PWD`.

## Outcome
Both hooks stay silent on non-commit commands, tolvi-solo's gate blocks a note-less commit with the documented `deny`, and neither hook bypasses the commit permission prompt, verified by feeding both hooks sample payloads in a scratch repo.

**Supersedes:** [[2026-06-05-pretooluse-enforcement-for-vault-sync]]: its output format (top-level `decision`) is unsupported for PreToolUse, and it relied on the `if` filter alone to scope the hook to commits.

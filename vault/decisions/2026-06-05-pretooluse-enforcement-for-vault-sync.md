---
tags: [decision, tolvi]
date: 2026-06-05
repo: tolvi
status: active
ticket: none
user_impact: low
product_area: Claude Code integration
---

# PreToolUse enforcement for vault sync before every git commit

**Date:** 2026-06-05
**Repo:** tolvi

## Why

The old `commit-sync-nudge` PostToolUse hook fired after the commit and only suggested syncing — it was easy to ignore, and vault updates often ended up in a separate follow-up commit or not at all. The vault is only useful if it stays in sync with the codebase; enforcing sync before the commit guarantees vault and code always move together.

## How

- Hook renamed `commit-sync-nudge.sh` → `tolvi-sync` (no `.sh` extension for cleaner naming).
- Changed from `PostToolUse` (fires after commit) to `PreToolUse` (fires before commit, can block).
- `tolvi-sync` behavior: (1) auto-stages any modified `vault/` files so they land in the same commit; (2) blocks the commit with a clear instruction if no session note (`vault/sessions/YYYY-MM-DD.md`) exists for today; (3) allows the commit once a session note is present, confirming vault is staged.
- Output format: `{"decision":"block","reason":"..."}` to block; `{"decision":"allow","additionalContext":"..."}` to allow. Both handled by Claude Code's PreToolUse hook system.
- Exits silently (exit 0) if no `vault/.vault-meta.json` is found walking up from `$PWD` — never interferes in repos without a vault.
- Applied identically to `tolvi-solo/hooks/tolvi-sync` so both products have the same enforcement behavior.
- Rejected: keeping PostToolUse as a stronger nudge — nudges are still ignorable; the PreToolUse block is the only mechanism that actually guarantees inclusion.
- Rejected: a git pre-commit hook (`.git/hooks/pre-commit`) — would fire for all commits including non-Claude ones and can't prompt interactively; the Claude Code PreToolUse hook is targeted to exactly the sessions where Claude is committing.

## Outcome

Every `git commit` made through Claude Code in a vaulted repo now auto-stages vault changes and blocks until a session note for today exists — vault and code are always committed together.

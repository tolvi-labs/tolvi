---
tags: [decision, tolvi]
date: 2026-06-03
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: Claude Code integration
---

# Claude Code hooks use `additionalContext`, not `systemMessage`

**Date:** 2026-06-03
**Repo:** tolvi

## Why

Claude Code hooks can inject output in two ways: `additionalContext` (injected into Claude's context — Claude sees it and can act on it) or `systemMessage` (displayed to the user only — Claude does not see it). For both the session-recall hook and the commit-sync-nudge hook, the goal is to influence Claude's behavior, not just surface a notification to the user.

## How

- Both hooks output the `hookSpecificOutput.additionalContext` field
- `session-recall.sh` → `hookEventName: "SessionStart"`, `additionalContext`: full TOLVI VAULT RECALL block — Claude receives recent sessions and decisions as context before the first user message, enabling proactive resumption without the user having to recap
- `commit-sync-nudge.sh` → `hookEventName: "PostToolUse"`, `additionalContext`: a one-sentence nudge asking Claude to offer `/sync-session` when the current task is complete, explicitly not mid-task — Claude decides the right moment rather than interrupting
- `systemMessage` was rejected for both: a recall summary the user sees but Claude doesn't is useless; a sync nudge Claude doesn't see would never result in an offer

## Outcome

Both hooks target Claude's context window, enabling autonomous proactive behavior without requiring any user-side trigger or display.

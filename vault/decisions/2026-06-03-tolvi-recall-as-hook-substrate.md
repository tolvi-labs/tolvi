---
tags: [decision, tolvi]
date: 2026-06-03
repo: tolvi
status: active
ticket: none
user_impact: high
product_area: CLI / Claude Code integration
---

# `tolvi recall` as the Claude Code hook substrate

**Date:** 2026-06-03
**Repo:** tolvi

## Why

The Claude Code SessionStart hook needed a way to surface vault context before the first user message. A raw shell script that read vault files directly would have duplicated logic already belonging in the CLI, and would be hard to configure or test. Building `tolvi recall` as a proper CLI subcommand keeps all logic versioned and testable, and gives users a human-readable `tolvi recall` command as a bonus.

## How

- `tolvi recall` is a pure file-read path — no Anthropic API call, no embedding, designed to stay under 500 ms
- Two output formats: `--format human` (mirrors the /recall skill's `RECALL SUMMARY` block) and `--format hook-json` (emits `{"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "..."}}` as expected by Claude Code's SessionStart hook contract)
- `RecallConfig` struct added to the config layer: `session_count` (default 3), `decision_count` (default 10), `max_bytes` (default 8000), `include_patterns` (default false); two-tier override: CLI flag > `~/.config/tolvi/config.yaml` > compiled-in defaults
- `recallWriteHookJSON` assembles: a header line, the RECALL SUMMARY block, then full content of the most recent N session files, hard-capped at `max_bytes` with a truncation notice
- Status filter applied to decisions: skip `superseded`, `deprecated`, `draft` — same filter the /recall skill applies
- `session-recall.sh` hook is a two-liner: `command -v tolvi >/dev/null 2>&1 || exit 0` then `tolvi recall --format hook-json 2>/dev/null || exit 0`; silent fail means a missing binary or broken vault never blocks a Claude Code session
- Rejected: having the hook script read vault files directly — would be untestable, would duplicate recall logic outside the CLI version boundary

## Outcome

`tolvi recall` is shipped in the main CLI binary; the Claude Code SessionStart hook is a stable one-liner that calls it, and all recall logic lives in versioned, tested Go code.

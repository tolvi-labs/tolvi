---
tags: [decision, tolvi]
date: 2026-06-03
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: CLI / Claude Code integration
---

# Patterns excluded from default recall

**Date:** 2026-06-03
**Repo:** tolvi

## Why

The /recall skill and `tolvi recall` were enumerating and reading all pattern files on every session start. Patterns are timeless, rarely-changing reference — they don't tell you where a session left off. Loading ~100 files (~424 KB) on every session start added latency and consumed context window without delivering recall value.

## How

- `include_patterns` defaults to `false` in both `RecallConfig` (compiled-in default) and the /recall skill's summary block
- When false, neither `tolvi recall` nor the /recall skill enumerate or read `vault/patterns/`
- The recall summary output includes a single line: `Patterns: (not loaded at recall — query on demand with /ask or /semantic-recall)` so users know patterns exist and where to find them
- Users who want patterns at session start can set `include_patterns: true` in `~/.config/tolvi/config.yaml` or pass `--include-patterns` at the command line
- The same change was applied to the equivalent /recall Claude Code skill in the internal tooling repo

## Outcome

Patterns are excluded from the default recall path; the context-window cost of session startup dropped by ~424 KB while all pattern content remains available on demand.

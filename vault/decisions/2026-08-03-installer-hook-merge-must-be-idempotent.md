---
tags: [decision, tolvi, install, claude-code]
date: 2026-08-03
repo: tolvi
status: active
ticket: none
user_impact: high
product_area: Claude Code integration / install.sh
---

# Claude Code installer hook merge must be idempotent

**Date:** 2026-08-03
**Repo:** tolvi

## TL;DR

`integrations/claude-code/install.sh` merges hooks into `settings.json` with an unconditional `extend()`, so every re-run duplicates every hook. Re-running the installer — the normal path after an upgrade — makes `tolvi-recall` fire twice on session start and the `tolvi-sync` PreToolUse gate fire twice per `git commit`. **Fixed 2026-09-13.** `tolvi-solo` was the reference implementation.

## Why

The installer is expected to be re-runnable. Users re-run it after upgrading, after switching `--hooks-scope`, or just to repair a config. Today a second run silently corrupts the user's `settings.json` by duplicating hook entries rather than converging on the intended state, and nothing in the output tells them it happened. Duplicated hooks are user-visible as doubled recall output on every session start, and as the commit gate running twice on every commit — friction that reads as "Tolvi is buggy" rather than "the installer was run twice."

## How

The defect is in the Python heredoc inside `integrations/claude-code/install.sh`, in the merge loop that currently reads:

```python
existing = settings.setdefault("hooks", {})
for event, entries in fragment["hooks"].items():
    if event in existing:
        existing[event].extend(entries)   # <-- no membership guard
    else:
        existing[event] = entries
```

`tolvi-solo/install.sh` already solves this and should be copied rather than reinvented — it guards each append by checking whether an entry with the same hook `command` is already present, along the lines of `if not any(h.get("command") == recall_hook for h in ss)`. The tolvi fragment nests differently (entries carry a `matcher` plus a `hooks` list), so the guard needs to compare the inner `hooks[].command` values, the same shape `tolvi-solo` uses for its `PreToolUse` entry.

Reproduction, confirmed 2026-08-03 by extracting the real heredoc and running it twice against a temp settings file: after two runs the file holds `SessionStart: 2` and `PreToolUse: 2` where it should hold one of each. Running `tolvi-solo`'s block twice correctly leaves `1` and `1`.

Two adjacent notes for whoever picks this up. First, the permission-allowlist merge added to both installers on the same date is already idempotent and needs no change — it filters against the existing `allow` array before extending, so use it as the in-repo pattern if you would rather not cross-reference `tolvi-solo`. Second, a full fix should converge rather than merely skip: if a user switches `--hooks-scope`, stale entries pointing at the old hooks directory will linger, so consider matching on the hook basename and replacing the path instead of leaving both.

## Outcome

Fixed on 2026-09-13, as step 0 of [[2026-09-12-vault-topology-plan-of-record]], which could not proceed without it. The merge keys on each entry's hook script basename: any prior install of that script is dropped and the fresh entry re-added, so a re-run converges instead of accumulating, and an entry left pointing at a stale hooks directory is replaced rather than left alongside. Hooks belonging to other tools share no basename and survive untouched, which matters on this machine, where `SessionStart` already carries two entries that are not ours.

The fix went further than the skip guard this doc originally asked for: the second note below wanted convergence rather than mere deduplication, and that is what landed. `.github/scripts/test-install-hook-idempotency.sh` pins all three behaviors and is wired into `validate.yml`. It was written first and watched fail, reproducing `SessionStart: 2` and `PreToolUse: 2` on a second `--force` run.

See also: [[2026-06-03-tolvi-recall-as-hook-substrate]], [[2026-06-05-pretooluse-enforcement-for-vault-sync]]

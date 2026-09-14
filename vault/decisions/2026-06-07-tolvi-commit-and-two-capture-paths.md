---
tags: [decision, tolvi]
date: 2026-06-07
repo: tolvi
status: active
ticket: none
user_impact: low
product_area: CLI / Agent integration
---

# `tolvi commit` (CLI) and `/tolvi-commit` (skill): two capture paths by design

**Date:** 2026-06-07
**Repo:** tolvi

## Why

Capturing a working session should be effortless, but a CLI cannot synthesize one — it has no conversation and no LLM — while an agent skill can. Rather than collapse the two into one "smart" command, Tolvi ships both at opposite ends of a control/comprehensiveness tradeoff, so a user (or a CI pipeline) picks the right tool instead of getting surprised by the wrong behavior.

## How

- **`tolvi commit` (Go/cobra, the controlled path):** deterministic, no LLM. Gates on a session note existing for today (`vault/sessions/<date>.md` containing a `##` block — the same gate as the `tolvi-sync` pre-commit hook), auto-stages `vault/` only (never `git add -A`, so the user keeps control of which code lands), then runs `git commit`. No session note → `ErrNoSessionNote` → exit code 3 with guidance pointing at `tolvi sync session` or the skill. Flags: `-m/--message`, `--vault`. Implemented as `RunCommit` + `hasSessionNote` in `cli/internal/cli/commit.go`; unit + integration tests added; `go build`/`vet`/full suite green.
- **`/tolvi-commit` (Claude Code/Cursor skill, the comprehensive path):** runs the `/tolvi-sync` synthesis (reconstruct the whole session → decisions/patterns/session-log per the format spec, authority-gated), then `git add -A` + commit. Comprehensive and near-zero-effort, but non-deterministic and needs an agent in the loop.
- The generic synthesis behavior was also added to `SKILL.md`, and three slash commands (`tolvi-recall`, `tolvi-sync`, `tolvi-commit`) now ship in both `tolvi` and `tolvi-solo`, installed by their respective `install.sh`. Opening synthesis is consistent with the open-core boundary — see [[2026-06-07-tolvi-open-core-boundary-and-authority-gated-capture]].
- The OSS `/tolvi-commit` skill **omits AI/assistant attribution by default** (no `Co-Authored-By`, `Generated with Claude Code`, 🤖, etc.), with a post-commit verify-and-strip step. This is an intentional product stance from the Tolvi creator, documented in the skill itself: AI-assisted work is often only partly authored by the AI (sometimes not at all), so a blanket credit misrepresents authorship — users who disagree delete that section in their own copy.
- Rejected: a single command that auto-detects whether to synthesize. Non-determinism in a CLI is a footgun for CI and scripts; keeping the mechanical path predictable is the whole point of having two.

## Outcome

`tolvi commit` ships as the deterministic capture path and `/tolvi-commit` as the synthesized one, with docs across the tolvi README, the Claude Code integration README, and the marketing `cli` + `claude-code` pages explaining when to reach for each.

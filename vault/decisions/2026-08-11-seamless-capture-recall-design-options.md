---
tags: [decision, tolvi, future-project]
date: 2026-08-11
repo: tolvi
status: draft
ticket: none
user_impact: none
product_area: Claude Code integration / SDLC UX
---

# Seamless capture/recall: design options for cutting the remaining "training" burden

**Date:** 2026-08-11
**Repo:** tolvi

## Why

Brainstormed from a broader question (asked in a marketing-site session): how to make using Tolvi feel seamless rather than something a user has to be trained on. That question decomposes into three separable pieces, only the first of which was explored here:

1. **SDLC auto-integration for engineers** (this doc) — reduce reliance on the user remembering to invoke capture/recall manually.
2. **A standalone, non-git capture experience for non-technical users** — a different product surface entirely (no repo, no git as the anchor); explicitly deferred, not started.
3. **Marketing-site positioning of a "zero-training" Tolvi** — depends on (1) and (2) existing first; not started.

No decision was made. This records the option space for (1) so a future session can pick it up without re-deriving it.

## Current state (verified against `integrations/claude-code/` in this repo, 2026-08-11)

- `SessionStart` hook already auto-injects `tolvi recall` context on every session start, and runs the full `/tolvi-recall` on `/clear` specifically. Recall is **already ambient** — not part of the remaining gap.
- `PreToolUse(git commit)` hook (see [[2026-06-05-pretooluse-enforcement-for-vault-sync]]) blocks the commit if no session note exists for today, instructing Claude to write one first (i.e. run `/tolvi-sync`) — a block-then-retry round trip, visible to the user, on every commit that hasn't already been preceded by a sync.
- The `/tolvi` skill (write-capability / natural-language capture) is **not** auto-loaded — it must be typed once per session.
- This automation exists for the Claude Code integration only; Cursor/Aider/OpenHands/Continue don't have equivalent hooks.

## Option space explored

**Approach A — Proactive skill behavior + soft safety net (leaning recommendation, not decided).**
Teach `SKILL.md` to proactively run `/tolvi-sync` *before* attempting `git commit` whenever no note exists for today, so the block rarely fires in practice. Change the `PreToolUse` hook from a hard block to a warn-and-allow (commit proceeds either way; a missed vault entry is recoverable, a blocked commit is not acceptable friction). Net effect: normal case is invisible, worst case is a warned-but-successful commit.

**Approach B — Mechanical fallback sync.**
Keep the hook as the sole enforcement point; instead of blocking, have it call a new deterministic `tolvi sync session --auto` (commit message + changed files, no LLM) when no note exists, so something always lands. Simpler, always succeeds, but much lower signal than a real synthesis.

**Approach C — Decouple sync from commit entirely (Stop-hook).**
Sync on Claude Code session end (`Stop` event) instead of pre-commit. Matches the "session proposes, merge confirms" framing from [[2026-06-07-tolvi-open-core-boundary-and-authority-gated-capture]] more purely, but `Stop` fires on every turn, not once per session — needs debouncing that wasn't designed here.

**Leaning:** A, with B folded in as the actual safety-net mechanism (belt-and-suspenders: proactive synthesis in the common case, thin deterministic note if that's skipped, never a hard block).

**Tension to resolve before deciding:** [[2026-06-05-pretooluse-enforcement-for-vault-sync]] deliberately moved *away* from a soft nudge to a hard block, on the stated rationale that "nudges are still ignorable." Approach A's warn-and-allow is a return to a softer enforcement model and needs to either refute that rationale (e.g. because the proactive skill behavior makes the block redundant rather than the nudge being the only safeguard) or be reconciled with it explicitly, not silently reversed.

## Other answers given during the brainstorm (not yet built on)

- Scope: Claude Code only for this pass; no other agent integrations.
- Skill auto-load: leaning toward extending the `SessionStart` hook to also inject write-capability, so `/tolvi` never needs to be typed manually — not designed in detail.
- Failure mode: if silent sync can't complete, let the commit proceed regardless (log/warn only) rather than falling back to a block.

## Outcome

No implementation. This is a parking note for a future `writing-plans`-level design session on (1), plus separate future brainstorms for (2) and (3).

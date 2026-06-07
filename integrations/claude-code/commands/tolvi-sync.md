---
description: Synthesize the working session into vault docs (decisions, patterns, session log).
---

You are running /tolvi-sync. Turn this working session into durable, schema-conformant vault docs. This is the comprehensive capture path — it reconstructs the whole session, unlike `tolvi sync <type> <title>`, which captures a single note you already have in mind.

## What to capture — the authority gate

Capture what was tried *or* considered inside **this working session**, including reasoned rejections ("we tried X, Y broke, so we shipped Z") — a road not taken for a stated reason is some of the most valuable content. Do not import unqualified chatter from outside the session; a passing idea that was never weighed is noise.

## Steps

1. **Reconstruct** from the conversation: files changed, tickets, decisions made, patterns observed, and what is left open.
2. **Discover the vault** — walk up to the first `vault/.vault-meta.json`. Prefer the `tolvi` CLI, which discovers automatically; `--vault <path>` overrides.
3. **Session log** → `vault/sessions/<date>.md` (one file per day; append a block if it exists). Frontmatter: `tags: [session]`, `date`, `status: active`. Block shape: `## [HH:MM] Session — <summary>`, then `### What happened`, `### Files touched`, `### Left open`.
4. **Decisions** (if any) → `vault/decisions/<date>-<slug>.md`. Frontmatter: `tags: [decision]`, `date`, `repo`, `status` (optional: `ticket`, `user_impact`, `product_area`). Body: `# Title`, `**Date:**`, `**Repo:**`, then layered `## Why` (1–2 business-readable sentences), `## How` (depth scales with technical weight — include rejected alternatives and *why* each was rejected), `## Outcome` (1 sentence). Keep Why/Outcome short; depth lives only in How.
5. **Patterns** (if any) → `vault/patterns/<slug>.md` (no date prefix). Frontmatter: `tags: [pattern]`, `status: active`. Append a new example if the file already exists.
6. **Cross-link** related docs with `[[slug]]` within the same vault; add a `See also: [[...]]` line to the session block.

## Writing mechanics

Prefer `tolvi sync <type> <title> --body "..."` per doc — it does atomic write, frontmatter validation, slug derivation, and same-day session append. If the CLI is unavailable, use the Write tool and validate the frontmatter against the rules above before writing.

Confirm what was written.

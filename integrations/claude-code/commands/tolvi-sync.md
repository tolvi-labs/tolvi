---
description: Synthesize the working session into vault docs (decisions, patterns, session log).
---

You are running /tolvi-sync. Turn this working session into durable, schema-conformant vault docs. This is the comprehensive capture path — it reconstructs the whole session, unlike `tolvi sync <type> <title>`, which captures a single note you already have in mind.

<!-- PREFLIGHT:BEGIN -->
## Preflight — say it once if the CLI is missing

Before the steps below, check whether the `tolvi` CLI is available:

    command -v tolvi

**If it is present**, use it. It is one invocation, it discovers the vault itself, and it does not raise a permission prompt per file.

**If it is absent**, fall back to reading the vault directly (the steps below work either way), and tell the user exactly once per conversation:

> `!` tolvi CLI not on PATH. Reading the vault directly, which works but skips
> semantic retrieval and costs one shell call per step. To fix:
> `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest`, then
> `export PATH="$PATH:$(go env GOPATH)/bin"`. Verify with `tolvi doctor`.

Do not tell the user to run `tolvi doctor` as the fix here. This branch only runs when the binary is unreachable, so `tolvi doctor` is unreachable too; it is the verification step after the install, not the remedy. Point at `tolvi doctor` only when the binary exists and something else is wrong.

**Once per conversation means once.** If you have already reported this in the current conversation, do not repeat it — later commands in the same session stay quiet. You know what you have already said; no marker file is needed. A user who has chosen not to install the CLI should not be told four times in one session, because a warning repeated that often stops being read.

Never silently degrade. The fallback path is legitimate and produces real answers, but the user has to learn once that they are on it, or a broken install looks identical to a working one.
<!-- PREFLIGHT:END -->

## What to capture — the authority gate

Capture what was tried *or* considered inside **this working session**, including reasoned rejections ("we tried X, Y broke, so we shipped Z") — a road not taken for a stated reason is some of the most valuable content. Do not import unqualified chatter from outside the session; a passing idea that was never weighed is noise.

## Steps

1. **Reconstruct** from the conversation: files changed, tickets, decisions made, patterns observed, and what is left open.
2. **Discover the vault** — walk up to the first `vault/.vault-meta.json`. Prefer the `tolvi` CLI, which discovers automatically; `--vault <path>` overrides.

   **Vault routing** — ask the CLI where each doc belongs rather than reading any config yourself:

       tolvi roots --session-note

   Roots are declared once in a machine-local `~/.config/tolvi/roots.json`; a repo commits only its identity (`workspace`, `repo`, optional `product`) in `vault/.vault-meta.json`.
   - Session notes go where that command says. Where the workspace declares an org root that is `<org-root>/sessions/YYYY-MM-DD-<repo>.md`, never the repo's own `vault/`. With no `roots.json` the vault is in single-root mode and the note belongs in `vault/sessions/YYYY-MM-DD.md`, which is what an external contributor wants: their note ships with their PR.
   - Decisions default to the repo's own `vault/decisions/`, which is public and contributor-facing. Write one to the org root with `visibility: private` whenever it is internal or strategic — business, cross-repo coordination, unreleased products, or roadmap. When in doubt, write it private.
   - A `roots.json` that exists but lacks the root a doc needs refuses the write rather than falling back to the public vault. Absent config is a legitimate contributor state; a gap in present config is a misconfiguration, and falling back is how an internal note gets published.

3. **Session log** → `vault/sessions/<date>.md` (one file per day; append a block if it exists). Frontmatter: `tags: [session]`, `date`, `status: active`. Block shape: `## [HH:MM] Session — <summary>`, then `### What happened`, `### Files touched`, `### Left open`.
4. **Decisions** (if any) → `vault/decisions/<date>-<slug>.md`. Frontmatter: `tags: [decision]`, `date`, `repo`, `status` (optional: `ticket`, `user_impact`, `product_area`). Body: `# Title`, `**Date:**`, `**Repo:**`, then layered `## Why` (1–2 business-readable sentences), `## How` (depth scales with technical weight — include rejected alternatives and *why* each was rejected), `## Outcome` (1 sentence). Keep Why/Outcome short; depth lives only in How.
5. **Patterns** (if any) → `vault/patterns/<slug>.md` (no date prefix). Frontmatter: `tags: [pattern]`, `status: active`. Append a new example if the file already exists.
6. **Cross-link** related docs with `[[slug]]` within the same vault; add a `See also: [[...]]` line to the session block.

## Writing mechanics

Prefer `tolvi sync <type> <title> --body "..."` per doc — it does atomic write, frontmatter validation, slug derivation, and same-day session append. If the CLI is unavailable, use the Write tool and validate the frontmatter against the rules above before writing.

Confirm what was written.

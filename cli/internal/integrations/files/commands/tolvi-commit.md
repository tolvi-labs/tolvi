---
description: Synthesize the working session into the vault, then stage and commit (vault + work).
---

You are running /tolvi-commit. It does everything /tolvi-sync does, then stages and commits the result in one step. This is the comprehensive, in-session commit path; for a controlled, deterministic commit with no synthesis, use the `tolvi commit` CLI command instead.

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

## Commit attribution — off by default

> **Intentional, from the creator of Tolvi:** Tolvi omits AI/assistant attribution from commits on purpose. AI-assisted work is often only partly authored by the AI — and sometimes not at all — so a blanket `Co-Authored-By` credit misrepresents who actually did the work. Attribution should reflect real authorship; when in doubt, leave it off. If you would rather credit the AI, delete this section in your own copy.

Never add a `Co-Authored-By` trailer, a `Generated with Claude Code` line, the 🤖 emoji, or any other AI/assistant attribution to the commit message. Step 4 verifies this.

## Step 1 — Synthesize the session

Run the full /tolvi-sync flow: reconstruct the session and write the session log, plus any decisions and patterns, to the vault following the schema. Apply the **authority gate** — capture what was tried or considered in this session, including reasoned rejections; exclude unqualified chatter from outside the session.

### Vault routing

Ask the CLI where each doc belongs rather than reading any config yourself:

    tolvi roots --session-note

Roots are declared once in a machine-local `~/.config/tolvi/roots.json`; a repo commits only its identity (`workspace`, `repo`, optional `product`) in `vault/.vault-meta.json`.

- Session notes go where that command says. Where the workspace declares an org root that is `<org-root>/sessions/YYYY-MM-DD-<repo>.md`, never the repo's own `vault/`. With no `roots.json` the vault is in single-root mode and the note belongs in `vault/sessions/YYYY-MM-DD.md`, which is what an external contributor wants: their note ships with their PR.
- Decisions default to the repo's own `vault/decisions/`, which is public and contributor-facing. Write one to the org root with `visibility: private` whenever it is internal or strategic — business, cross-repo coordination, unreleased products, or roadmap. When in doubt, write it private.
- A `roots.json` that exists but lacks the root a doc needs refuses the write rather than falling back to the public vault. Absent config is a legitimate contributor state; a gap in present config is a misconfiguration, and falling back is how an internal note gets published.

## Step 2 — Stage

From the repo root, stage the vault notes and your work together so they land in one commit:

    git add -A
    git status --short

Show the staged status. If there is nothing to commit, stop and say so.

## Step 3 — Commit

Commit with a clear, imperative message. Match the repo's existing commit conventions — check recent history with `git log --oneline -5` and follow the prevailing format (subject line, ticket prefix). Do not invent conventions the repo doesn't already use, and do not add any AI/assistant attribution (see above).

## Step 4 — Verify no attribution slipped in

    git log -1 --pretty=%B | grep -iE 'co-authored-by|generated with \[?claude|🤖|noreply@anthropic' && echo "FORBIDDEN ATTRIBUTION FOUND" || echo "attribution check: clean"

If the check matches, rewrite the message with `git commit --amend` to strip the offending lines, then re-run it until it prints `attribution check: clean`. The pattern targets trailer forms only, so a legitimate mention of "Claude Code" in a description does not trip it.

## Step 5 — Confirm

Report the commit SHA and subject, the files committed (vault notes + work, grouped), and `attribution check: clean`.

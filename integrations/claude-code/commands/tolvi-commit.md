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
> semantic retrieval and costs one shell call per step. Fix: `tolvi doctor`,
> or install with `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest`
> and add `$(go env GOPATH)/bin` to your PATH.

**Once per conversation means once.** If you have already reported this in the current conversation, do not repeat it — later commands in the same session stay quiet. You know what you have already said; no marker file is needed. A user who has chosen not to install the CLI should not be told four times in one session, because a warning repeated that often stops being read.

Never silently degrade. The fallback path is legitimate and produces real answers, but the user has to learn once that they are on it, or a broken install looks identical to a working one.
<!-- PREFLIGHT:END -->

## Commit attribution — off by default

> **Intentional, from the creator of Tolvi:** Tolvi omits AI/assistant attribution from commits on purpose. AI-assisted work is often only partly authored by the AI — and sometimes not at all — so a blanket `Co-Authored-By` credit misrepresents who actually did the work. Attribution should reflect real authorship; when in doubt, leave it off. If you would rather credit the AI, delete this section in your own copy.

Never add a `Co-Authored-By` trailer, a `Generated with Claude Code` line, the 🤖 emoji, or any other AI/assistant attribution to the commit message. Step 4 verifies this.

## Step 1 — Synthesize the session

Run the full /tolvi-sync flow: reconstruct the session and write the session log, plus any decisions and patterns, to the vault following the schema. Apply the **authority gate** — capture what was tried or considered in this session, including reasoned rejections; exclude unqualified chatter from outside the session.

### Vault routing (public repos with a private vault)

While running that sync flow, before writing any session note or decision, check for a local routing config file `vault/.vault-routing.local.json` in the repo being committed (it is git-ignored and present only on internal-dev checkouts of a public repo). If it exists, read its `private_vault` path and route:

- Session notes ALWAYS go to `<private_vault>/sessions/YYYY-MM-DD-<workspace>.md` (workspace-suffixed to avoid cross-repo collisions) — NEVER into the public repo's own `vault/`.
- Decisions default to the repo's own `vault/decisions/` (public, contributor-facing). Write a decision to `<private_vault>/decisions/` with `visibility: private` in its frontmatter instead whenever it is internal or strategic — business, cross-repo coordination, unreleased products, or roadmap. When in doubt, write it private.
- Never write a session note or a `visibility: private` decision into the public repo's `vault/`.

If the config file is absent (an external contributor, or a private repo), behave exactly as before: everything goes to the local `vault/`, so contributor PRs keep feeding the public vault normally. The `tolvi` CLI reads the same `vault/.vault-routing.local.json` itself, so both capture paths route identically with no extra flags; `--open-source` and `--private-vault <path>` still override it, and `tolvi sync --private` marks a decision private.

## Step 2 — Stage

From the repo root, stage the vault notes and your work together so they land in one commit:

```bash
git add -A
git status --short
```

Show the staged status. If there is nothing to commit, stop and say so.

## Step 3 — Commit

Commit with a clear, imperative message. Match the repo's existing commit conventions — check recent history with `git log --oneline -5` and follow the prevailing format (subject line, ticket prefix). Do not invent conventions the repo doesn't already use, and do not add any AI/assistant attribution (see above).

## Step 4 — Verify no attribution slipped in

```bash
git log -1 --pretty=%B | grep -iE 'co-authored-by|generated with \[?claude|🤖|noreply@anthropic' && echo "FORBIDDEN ATTRIBUTION FOUND" || echo "attribution check: clean"
```

If the check matches, rewrite the message with `git commit --amend` to strip the offending lines, then re-run it until it prints `attribution check: clean`. The pattern targets trailer forms only, so a legitimate mention of "Claude Code" in a description does not trip it.

## Step 5 — Confirm

Report the commit SHA and subject, the files committed (vault notes + work, grouped), and `attribution check: clean`.

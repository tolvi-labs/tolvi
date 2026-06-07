---
description: Synthesize the working session into the vault, then stage and commit (vault + work).
---

You are running /tolvi-commit. It does everything /tolvi-sync does, then stages and commits the result in one step. This is the comprehensive, in-session commit path; for a controlled, deterministic commit with no synthesis, use the `tolvi commit` CLI command instead.

## Commit attribution — off by default

> **Intentional, from the creator of Tolvi:** Tolvi omits AI/assistant attribution from commits on purpose. AI-assisted work is often only partly authored by the AI — and sometimes not at all — so a blanket `Co-Authored-By` credit misrepresents who actually did the work. Attribution should reflect real authorship; when in doubt, leave it off. If you would rather credit the AI, delete this section in your own copy.

Never add a `Co-Authored-By` trailer, a `Generated with Claude Code` line, the 🤖 emoji, or any other AI/assistant attribution to the commit message. Step 4 verifies this.

## Step 1 — Synthesize the session

Run the full /tolvi-sync flow: reconstruct the session and write the session log, plus any decisions and patterns, to the vault following the schema. Apply the **authority gate** — capture what was tried or considered in this session, including reasoned rejections; exclude unqualified chatter from outside the session.

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

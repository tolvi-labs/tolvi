---
description: Surface recent sessions and active decisions from the vault before starting work.
---

You are running /tolvi-recall. Surface vault context before doing any work.

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

If the `tolvi` CLI is available, run `tolvi recall` and present its output. That is one invocation and the preferred path — prefer it whenever the binary exists.

Otherwise read the vault directly, using the commands below **verbatim**. They are loop-free on purpose: a `for f in *.md; do ... done` loop is not prefix-matchable by Claude Code's permission system, so every loop raises a permission prompt on every recall. Each command here is a single invocation covered by an ordinary `Bash(rg:*)` / `Bash(ls:*)` allow rule. Do not "simplify" them back into a loop.

1. **Discover the vault** — walk up to the first `vault/.vault-meta.json`:

       ls vault/.vault-meta.json ../vault/.vault-meta.json ../../vault/.vault-meta.json 2>/dev/null | head -1

   If nothing prints, say so and stop.

2. **Check for a private vault** — an internal-dev checkout routes its session notes elsewhere:

       cat vault/.vault-routing.local.json 2>/dev/null

   If this prints a `private_vault` path, this is a routed vault: session notes live there, not in the local `vault/sessions/`, and decisions are split across both roots. Read the `workspace` value out of `vault/.vault-meta.json` — it is the suffix routed notes are named with — and use the routed commands in steps 3 and 4. If nothing prints, use the plain ones.

3. **Sessions** — date-named files, so the newest sorts last:

       ls -1 vault/sessions/ | tail -3

   Routed vaults instead read the private root, filtered to this workspace's own notes, since every repo in the org shares it:

       ls -1 <private_vault>/sessions/ | rg -- '-<workspace>\.md$' | tail -3

   Either way, read the newest file with the Read tool (no shell needed) and surface its latest `## [HH:MM] Session — ...` heading plus any `### Left open` items.

4. **Decisions** — one call covers status and title across every decision:

       rg -N --no-heading -H --sort path '^(status:|# )' vault/decisions/ | tail -40

   Routed vaults run that against `<private_vault>/decisions/` as well, adding `repo:` to the pattern so you can tell which repo each one belongs to:

       rg -N --no-heading -H --sort path '^(status:|repo:|# )' <private_vault>/decisions/ | tail -60

   Output interleaves `path:status:` with `path:# Title` per file. Skip any whose `status` is `superseded`, `deprecated`, or `draft` (missing = `active`), and from the private root keep only those whose `repo` matches `<workspace>` — the rest belong to sibling repos. Surface up to ~10 active across both roots, newest first, as `slug — title`.

5. **Patterns** — not loaded at recall; query on demand with `tolvi ask`.

Output:

```
RECALL SUMMARY
──────────────────────────────────────────
Last session:  [date — heading]
Left open:
  [bullet per item, or none]
Decisions:     [N relevant | none]
  [slug — title  (status: X if not active)]
──────────────────────────────────────────
```

Then ask what to focus on.

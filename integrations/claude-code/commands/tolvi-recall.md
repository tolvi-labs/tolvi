---
description: Surface recent sessions and active decisions from the vault before starting work.
---

You are running /tolvi-recall. Surface vault context before doing any work.

If the `tolvi` CLI is available, run `tolvi recall` and present its output. That is one invocation and the preferred path — prefer it whenever the binary exists.

Otherwise read the vault directly, using the commands below **verbatim**. They are loop-free on purpose: a `for f in *.md; do ... done` loop is not prefix-matchable by Claude Code's permission system, so every loop raises a permission prompt on every recall. Each command here is a single invocation covered by an ordinary `Bash(rg:*)` / `Bash(ls:*)` allow rule. Do not "simplify" them back into a loop.

1. **Discover the vault** — walk up to the first `vault/.vault-meta.json`:

       ls vault/.vault-meta.json ../vault/.vault-meta.json ../../vault/.vault-meta.json 2>/dev/null | head -1

   If nothing prints, say so and stop.

2. **Sessions** — date-named files, so the newest sorts last:

       ls -1 vault/sessions/ | tail -3

   Then read the newest file with the Read tool (no shell needed) and surface its latest `## [HH:MM] Session — ...` heading plus any `### Left open` items.

3. **Decisions** — one call covers status and title across every decision:

       rg -N --no-heading -H --sort path '^(status:|# )' vault/decisions/ | tail -40

   Output interleaves `path:status:` with `path:# Title` per file. Skip any whose `status` is `superseded`, `deprecated`, or `draft` (missing = `active`). Surface up to ~10 active, newest first, as `slug — title`.

4. **Patterns** — not loaded at recall; query on demand with `tolvi ask`.

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

---
description: Surface recent sessions and active decisions from the vault before starting work.
---

You are running /tolvi-recall. Surface vault context before doing any work.

If the `tolvi` CLI is available, run `tolvi recall` and present its output. Otherwise read the vault directly:

1. **Discover the vault** — walk up from the current directory to the first `vault/.vault-meta.json`. If none is found, say so and stop.
2. **Sessions** — list `vault/sessions/*.md`, newest first. Surface the last one or two by date plus the text of the latest `## [HH:MM] Session — ...` heading, and any `### Left open` items from the most recent file.
3. **Decisions** — list `vault/decisions/*.md`. Skip any whose frontmatter `status` is `superseded`, `deprecated`, or `draft` (missing = `active`). Surface up to ~10 active, newest first, as `slug — title`.
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

# Vault conventions

This document describes how vault content is organized on disk, what frontmatter each doc type carries, and the small set of conventions that consumers (CLI, server, agents) rely on. The normative version of these rules lives in [`spec/tolvi-format-v1.md`](../spec/tolvi-format-v1.md); this document is the contributor-facing companion that explains the *why*.

## 1. Vault directory layout

Each repo that uses Tolvi has a `vault/` directory at the repo root:

```
<repo>/vault/
  .vault-meta.json
  decisions/                YYYY-MM-DD-slug.md
  sessions/                 YYYY-MM-DD.md (one file per day, multiple H2 session blocks)
  patterns/                 slug.md  (no date prefix — patterns are timeless)
```

`.vault-meta.json` declares the workspace identity and the embedding model used locally:

```json
{
  "workspace": "<workspace-slug>",
  "embedding_model": "nomic-embed-text",
  "schema_version": 1
}
```

`schema_version` is the vault-format version, currently `1`. Consumers refuse to operate on vaults whose `schema_version` they do not understand.

## 2. Status enum

Every doc carries a `status:` field in frontmatter. The enum has six values and is frozen for `tolvi-format-v1`:

| Status | Meaning | Surfaced by default? |
|---|---|---|
| `active` | Current. Default for new docs. | Yes |
| `in-progress` | Decision made, implementation still landing. | Yes |
| `superseded` | Replaced by a newer decision; `superseded_by:` links forward. | No |
| `deprecated` | No longer applicable; left for history. | No |
| `draft` | Work-in-progress, not yet authoritative. | No |
| `historical` | Preserved for context but not actionable. | Yes |

Adding a value to this enum is a `tolvi-format-v2` change, not a v1 patch.

**Surfacing note (informative).** The "surfaced by default" column describes default *retrieval* behavior — what `/v1/search` and `tolvi recall` return without explicit status filters. UI presentation of returned results (for example, rendering a status badge for `historical` docs) is a consumer concern, not part of the format spec.

## 3. Frontmatter schemas

Frontmatter is YAML between `---` fences at the top of the file. Schemas are broken out by doc type below; the machine-readable JSON Schemas live in [`spec/schemas/`](../spec/schemas/).

### Universally required (all doc types)

```yaml
---
tags: [<doc-type>, ...]
status: active
---
```

### Sessions

In addition to the universal fields:

```yaml
date: YYYY-MM-DD          # required; matches the filename date
```

### Decisions

In addition to the universal fields:

```yaml
date: YYYY-MM-DD          # required; matches the filename date
repo: <repo-slug>         # required; the repo this decision binds to
ticket: <free-form>       # optional; issue tracker ref (e.g. "PROJ-123", a URL, or "none")
supersedes: <slug>        # optional; backward link to the doc this replaces
superseded_by: <slug>     # optional; forward link (set when status: superseded)
```

`ticket:` is intentionally free-form so that any issue tracker (or none) can be referenced. A consumer that wants to link out can detect URL-shaped values and treat them as such.

### Patterns

Patterns add only optional fields:

```yaml
languages: [...]          # optional
frameworks: [...]         # optional
```

Patterns are intentionally timeless — there is no `date` or `repo` field, because patterns describe approaches that outlive any single decision or repo.

### Genericization notes

Some prior reference implementations included `product_area` and `user_impact` frontmatter fields. These were intentionally not adopted: they were domain-specific and did not generalize well to a public spec. Use `tags:` for categorization instead. Implementations that genuinely need additional fields may add them under an `x-*` namespace (for example, `x-team: payments`), which the spec ignores by design.

## 4. Wiki-link syntax

Vault docs cross-reference each other using wiki-link syntax:

- `[[slug]]` — link to a doc in the current repo's vault, where `slug` is the filename without extension.
- `[[repo:slug]]` — cross-repo link, resolved by the aggregator (locally) or by the server (over HTTP).

Citations returned by `/v1/ask` MUST use this syntax. This keeps responses interoperable with Obsidian and other wiki-link-aware tools, and avoids inventing a parallel link format for chat responses.

## 5. Decision template

Decision docs follow a three-section body, in order:

```markdown
## Why
Context and forces. What's the problem? What constraints apply?

## How
The decision itself. What did we choose, and what trade-off does that buy?

## Outcome
Observable result. Updated when the decision lands or is superseded.
```

The template is short on purpose. A decision that needs more than three sections is usually two decisions.

## 6. Aggregator pattern (recommended, not automated in v1)

Engineers who work across multiple repos often want a single Obsidian vault that views all of them at once. v1 does not automate this — the recipe is a few shell commands:

```bash
mkdir -p ~/tolvi-vault
cd ~/tolvi-vault
ln -s ~/path/to/repo-a/vault repo-a
ln -s ~/path/to/repo-b/vault repo-b
```

Then open `~/tolvi-vault/` in Obsidian (or any markdown editor that follows symlinks). Cross-repo `[[repo:slug]]` links resolve correctly because each linked subdirectory is named after its repo.

A `tolvi unify` command to automate this — including pruning stale links and detecting moved repos — is on the v1.x roadmap. It was deferred from v1 because the manual recipe works and exercising it informs the eventual command's design.

## 7. Optimizing vaults for agents (recommended)

Vault content exists so coding agents can answer decision-rationale questions cheaply and accurately. Two conventions, both additive to the format and both optional, make a measurable difference on agent token cost and answer quality. See [`adr/0003-vault-index-and-tldr-system.md`](./adr/0003-vault-index-and-tldr-system.md) for the architectural decision and the benchmark behind these recommendations.

### TL;DR block on long decision docs

Decision docs over roughly 5 KB SHOULD carry a `## TL;DR` block placed immediately after the frontmatter and before the body. Below 5 KB the block is optional — small docs are already cheap to read in full.

```markdown
---
tags: [decision, ...]
status: active
date: 2026-05-22
repo: <repo-slug>
---

## TL;DR

<Decision in one sentence>. <Why, ≤120 chars>. Rejected: <alternative(s), ≤120 chars>.

## Why
...
```

Budget the entire block, including the `## TL;DR` heading, to **≤500 bytes** (roughly 125 tokens). The discipline is the point: a TL;DR that won't fit in 500 bytes usually signals a decision whose rationale needs sharpening.

When the decision considered no alternatives — or considered them implicitly without writing them down — end the block with the literal marker:

```
Rejected: (none documented).
```

The marker makes the absence visible to both readers and tooling. Silent omission is non-conformant with the convention because a missing `Rejected:` line is indistinguishable from a forgotten one.

**Example (≤500 bytes):**

```markdown
## TL;DR

Adopt PASETO v4 (public, Ed25519) for service-to-service tokens. Removes the JWT `alg:` confusion class of bugs and verifies ~60% faster than RS256 in load tests. Rejected: stay on JWT with allow-listed algorithms (brittle); custom token format (NIH, security review tax).
```

### Vault index in agent conventions files

Every agent integration loads a per-tool conventions file at session start — `CLAUDE.md` for Claude Code, `.cursorrules` for Cursor, `CONVENTIONS.md` for Aider, `.openhands_instructions` for OpenHands, `.continuerules` for Continue. Injecting a generated index of vault docs into that file lets the agent resolve "where is the decision about X" to an exact filename without filesystem search.

The index is a delimited block within the conventions file:

```markdown
<!-- VAULT-INDEX:START repo=<name> generated=YYYY-MM-DD -->
## Vault Index

*Auto-generated. Re-run the index generator to refresh.*

### Decisions

- `2026-05-09-vault-format-v1-contract.md` — Vault format v1 contract
- `2026-05-22-vault-index-and-tldr-system.md` — Vault index and TL;DR system

### Patterns

- `pattern-slug.md` — Title extracted from the doc

<!-- VAULT-INDEX:END -->
```

Recommended rules for generators:

- Extract each line's title from the doc's H1 if present; otherwise the first H2; otherwise fall back to the filename slug with hyphens replaced by spaces.
- Sort decisions by filename descending (which, given the `YYYY-MM-DD-slug.md` naming rule in Section 1, sorts newest-first).
- Sort patterns by filename ascending (patterns are timeless; alphabetical is the natural order).
- Apply the default status filter from `spec/tolvi-format-v1.md` Section 9: exclude `superseded`, `deprecated`, and `draft` from the index.
- Be idempotent — replace the content between the `VAULT-INDEX:START` and `VAULT-INDEX:END` markers in place; append the block at the end of the file if no markers exist.

When to skip the index:

- Vaults under roughly 10 decision docs — the ~1–2 KB the index loads into every agent session is not worth the lookup speedup at that scale.
- Conventions files that are already at risk of exceeding the agent's context budget for system content — measure first.

### Regeneration

The index rots the moment a new doc lands. Wire regeneration into one of: a `pre-commit` hook, a session-end skill, or a CI job that opens a PR when drift is detected. A future `tolvi vault-index` CLI subcommand (see ROADMAP) is intended to be a one-line hook that adopters can wire into any of these triggers without depending on a specific agent platform.

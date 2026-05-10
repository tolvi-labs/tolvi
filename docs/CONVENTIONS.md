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

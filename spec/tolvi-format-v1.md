# Tolvi format v1

**Version:** 1.0.0 — 2026-05-10

## 1. Status

Stable. Breaking changes require `tolvi-format-v2` and accompanying migration tooling. The schema files in `./schemas/` are the machine-readable conformance test for this spec; where this prose disagrees with those schemas, the schemas win.

> **RFC 2119 keywords.** The words MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY in this document carry their RFC 2119 meanings. MUST is a hard requirement; SHOULD is a strong recommendation that may be deviated from with documented reason; MAY denotes a permitted but optional behavior. Sections labeled "(informative)" describe defaults and rationale and are not subject to conformance tests.

## 2. Vault directory layout (normative)

A Tolvi vault is a directory containing a `.vault-meta.json` file at its root and three subdirectories holding the v1 doc types:

```
<repo>/vault/
  .vault-meta.json
  decisions/
    YYYY-MM-DD-slug.md
  sessions/
    YYYY-MM-DD.md
  patterns/
    slug.md
```

Any directory containing a `.vault-meta.json` file at its root MUST be treated as a Tolvi vault. The three subdirectories `decisions/`, `sessions/`, and `patterns/` MUST be present for v1 conformance, even when empty. Implementations MAY ignore additional sibling subdirectories so that future doc types can be added without breaking v1 readers; implementations MUST NOT reject a vault solely because unknown subdirectories are present.

A vault's location on disk is not normative. Vaults typically live at `<repo>/vault/`, but tooling MUST locate vaults by the presence of `.vault-meta.json`, not by path.

## 3. `.vault-meta.json` (normative)

The vault metadata file declares the workspace identity and local indexing configuration.

| Field | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | yes | Workspace this vault belongs to. Used by servers for multi-tenant isolation. |
| `embedding_model` | string | yes | Identifier of the embedding model used to index this vault locally. The default value SHOULD be `nomic-embed-text` (Ollama). |
| `schema_version` | integer | yes | Vault format version. For `tolvi-format-v1`, this MUST be `1`. |

The authoritative form is `./schemas/vault-meta.json`. Canonical example:

```json
{
  "workspace": "<workspace-slug>",
  "embedding_model": "nomic-embed-text",
  "schema_version": 1
}
```

Implementations MUST refuse to operate on a vault whose `schema_version` they do not understand, rather than guessing.

## 4. File naming rules (normative)

- **Decisions** — `decisions/YYYY-MM-DD-slug.md`. The `YYYY-MM-DD` prefix is the date the decision was first written, not the date it was last modified.
- **Sessions** — `sessions/YYYY-MM-DD.md`. There MUST be exactly one session file per day per vault. Multiple session blocks within a day are recorded as `## [HH:MM] Session — <summary>` H2 headings inside that single file.
- **Patterns** — `patterns/slug.md`. Patterns carry no date prefix; they are intentionally timeless.

Filename slugs MUST match the regular expression `[a-z0-9]([a-z0-9-]*[a-z0-9])?` — that is, lowercase ASCII alphanumerics with single hyphens between them, never as the first or last character, and never doubled (`foo--bar` is non-conformant; `foo-` is non-conformant). The slug component (excluding any `YYYY-MM-DD-` prefix and the `.md` extension) MUST NOT exceed 80 characters. Date prefixes MUST be ISO-8601 (`YYYY-MM-DD`). Filenames MUST end in `.md`.

## 5. Frontmatter (normative)

Every vault doc begins with YAML frontmatter delimited by `---` lines. Field-level rules are defined by `./schemas/decision.json`, `./schemas/session.json`, and `./schemas/pattern.json`; this section summarizes them in prose.

**All doc types** require:

- `tags` — array of strings, MUST contain at least one element. The first tag SHOULD identify the doc type (`decision`, `session`, or `pattern`) but additional tags are free-form. Tag string comparison is case-sensitive; implementations MUST NOT normalize tag case during ingest. Duplicate tags within a single doc are non-conformant.
- `status` — enum value, see Section 6.

**Sessions** additionally require:

- `date` — ISO-8601 (`YYYY-MM-DD`). MUST match the filename date.

**Decisions** additionally require:

- `date` — ISO-8601, MUST match the filename date.
- `repo` — string, the repo slug this decision binds to.

Decisions optionally accept:

- `ticket` — free-form string. Any issue tracker reference, URL, or the literal `none` is valid; consumers MAY detect URL-shaped values and render them as links.
- `supersedes` — slug of the doc this one replaces.
- `superseded_by` — slug of the doc that replaces this one. Required when `status` is `superseded` (see Section 8).

**Patterns** optionally accept:

- `languages` — array of strings.
- `frameworks` — array of strings.

Patterns intentionally have no `date` or `repo` field, because patterns describe approaches that outlive any single decision.

**Custom fields.** Implementations MAY add custom fields under an `x-*` namespace (for example, `x-internal-priority: high`). Implementations MUST ignore unknown `x-*` fields rather than rejecting the document. This is the forward-compatibility escape hatch: domain-specific extensions live under `x-*`, the core spec stays small.

Future format versions (`tolvi-format-v2` and later) MUST NOT claim names beginning with `x-`. The `x-*` namespace is reserved permanently for implementation extensions.

**Unknown fields.** Implementations MUST reject documents that contain unknown top-level frontmatter fields not beginning with `x-`. This includes typos (`statuss:`), removed v0 experimental fields, and v2 fields appearing in v1 readers. Strict rejection prevents silent semantic drift; tolerant readers would let typos accumulate undetected.

Implementations MUST reject documents whose frontmatter fails its type schema with a clear error pointing at the offending field.

## 6. Status enum (normative)

The status enum is **frozen at six values for v1**. Adding a value requires `tolvi-format-v2`.

| Status | Meaning | Default surfacing |
|---|---|---|
| `active` | Current. Default for new docs. | Surfaced |
| `in-progress` | Decision made, implementation still landing. | Surfaced |
| `superseded` | Replaced by a newer decision; `superseded_by` MUST link forward. | Hidden |
| `deprecated` | No longer applicable; left for history. | Hidden |
| `draft` | Work-in-progress, not yet authoritative. | Hidden |
| `historical` | Preserved for context but not actionable. | Surfaced |

Status values are case-sensitive (`active` is conformant; `Active` is not). Implementations MUST reject documents whose `status` field contains a value not in the six-element enum, with a clear error pointing at the offending value.

Implementations MUST default `/v1/search`, `/v1/ask`, and any recall-style query endpoints to exclude documents whose status is `superseded`, `deprecated`, or `draft`. Callers MAY override this default per-query (for example, an `include_status=any` parameter, or an explicit list of statuses to include).

The "default surfacing" column describes default *retrieval* behavior. UI presentation of returned results — for example, rendering a status badge on `historical` docs — is a consumer concern, not part of this spec.

## 7. Wiki-link syntax (normative)

Vault docs cross-reference each other with wiki-link syntax:

- `[[slug]]` — links to a doc in the same vault, by filename slug (without the date prefix or `.md` extension). For example, `[[adopt-postgres]]` resolves to `decisions/2026-01-12-adopt-postgres.md` if such a file exists.
- `[[repo:slug]]` — cross-vault link. The `repo:` prefix is a workspace-scoped repo identifier that matches the `workspace` field of another vault's `.vault-meta.json`, or — when an aggregator merges multiple vaults — the aggregator's repo namespace.

Citations returned by synthesis-style endpoints (for example, `POST /v1/ask`) MUST use this syntax for cited references. This keeps responses interoperable with Obsidian and other wiki-link-aware tools, and avoids inventing a parallel link format for chat responses.

Implementations MUST resolve broken wiki-links non-fatally: log a warning, render the link as plain text, and continue. Implementations MUST NOT abort ingest because of a broken wiki-link. Implementations MAY emit a list of broken links via a diagnostic command (for example, `tolvi doctor`).

**Slug collisions across doc types.** When a `[[slug]]` reference is ambiguous because the same slug exists in more than one of `decisions/`, `sessions/`, `patterns/` (after stripping any date prefix), implementations MUST treat the link as broken and emit a diagnostic naming both candidate paths. Implementations MUST NOT silently pick one — different implementations would pick differently, and vault content would render inconsistently.

**Unknown `repo:` prefix.** When a `[[repo:slug]]` reference targets a `repo:` value that does not match any vault known to the resolver, implementations MUST treat the link as broken (warning, not error) using the same handling as a broken intra-vault link. This keeps cross-repo authoring symmetric with single-repo authoring: a missing target is always a warning, never an ingest failure.

## 8. Cross-reference rules (normative)

Supersession is bidirectional. When doc A is superseded by doc B, all three of the following MUST hold:

- A's frontmatter sets `status: superseded`.
- A's frontmatter sets `superseded_by: <B-slug>`.
- B's frontmatter sets `supersedes: <A-slug>`.

These three updates MUST be applied atomically. When the vault is stored in git, "atomically" means a single commit; storage backends without commit semantics MUST provide an equivalent all-or-nothing guarantee. Tooling SHOULD enforce atomicity — for example, a `tolvi status --supersede <new-slug>` command that updates both files in one operation.

Implementations MUST validate supersession bidirectionality during ingest. A missing back-reference MUST result in either a hard rejection or a loud warning surfaced to the operator; silent acceptance is non-conformant.

## 9. RAG defaults (informative)

The values in this section are *defaults* an implementation SHOULD apply when no override is configured. They are not a conformance requirement; an implementation MAY tune them based on its workload.

- **Recency multiplier.** Apply `(0.8 + 0.2 × exp(-age_days / 180))` to similarity scores at retrieval time. Newer docs are favored, but older docs are not entirely buried — the floor of `0.8` ensures that durable older content remains reachable. `age_days` is the integer number of whole days between the current UTC date and the doc's `date:` frontmatter field. For patterns (which carry no `date:` field), `age_days` is computed from a storage-tracked creation timestamp when available — for git-backed vaults, the date of the commit that introduced the file — falling back to the filesystem mtime when no storage-tracked timestamp is available.
- **Session document down-weight.** Multiply session-doc similarity scores by `0.7` so they do not crowd out durable decision and pattern content.
- **Default status filter.** Exclude `superseded`, `deprecated`, and `draft`. Section 6 mandates this as the default behavior; this section confirms it as part of the recommended defaults context.

These are starting points, not final answers; tuning candidates that emerge from real workloads will inform `tolvi-format-v2`.

## 10. Embedding model defaults (informative)

- **Local CLI.** `nomic-embed-text` (Ollama, 768-dimensional vectors). Configured via the `embedding_model` field in `.vault-meta.json`.
- **Server.** The embedding model is deployment configuration, not part of this spec. Self-hosters choose based on available infrastructure (Ollama, hosted API providers, self-hosted GPU inference).

Implementations MUST handle the case where `embedding_model` differs across vaults that are aggregated together by selecting exactly one of the following two behaviors. No third option is conformant.

1. **Re-embed.** The aggregator re-embeds all content to a single canonical model and uses the resulting vectors for retrieval.
2. **Refuse.** The aggregator refuses to merge vaults with mismatched `embedding_model` values and surfaces a clear error to the operator.

Mixing vectors from different embedding models — silently or with a warning — is non-conformant. The resulting similarity scores are not comparable, and ranking outputs would be implementation-defined in a way that makes vault content non-portable.

## 11. Versioning rules (normative)

`tolvi-format-v1` is stable. Breaking changes require a `tolvi-format-v2` revision, which will live at `/spec/tolvi-format-v2.md` alongside this document.

Implementations declare the format versions they support; the wire-level handshake details are deferred and tracked in `/docs/OPEN_QUESTIONS.md`. For v1, all conformant implementations MUST support `tolvi-format-v1`.

When v2 ships, migration tooling will be published in the same release. Implementations SHOULD support both v1 and v2 for at least one format-version cycle after v2 ships, to allow vault content to migrate without forcing a synchronized ecosystem upgrade.

A change is **breaking** (and therefore requires v2) if it could cause a v1-conformant vault to become non-conformant under the new rules, or could cause a v1-conformant implementation to misbehave on the new content. Adding new optional fields, new informative defaults, or new diagnostics is non-breaking and lands in v1.x.

## 12. Conformance (normative)

Conformance has two parts: vault conformance (a property of content on disk) and implementation conformance (a property of a tool that reads or writes vaults).

### Vault conformance

A vault is conformant with `tolvi-format-v1` if and only if all of the following hold:

1. Every file in `decisions/`, `sessions/`, and `patterns/` validates against its respective JSON Schema (`./schemas/decision.json`, `./schemas/session.json`, `./schemas/pattern.json`).
2. The `.vault-meta.json` file validates against `./schemas/vault-meta.json`.
3. The directory layout matches Section 2.
4. All supersession references are bidirectional per Section 8.

Wiki-link resolution is intentionally not a vault-conformance criterion: a vault may legitimately contain links that target docs not yet written. Wiki-link handling is an implementation-conformance concern (see below).

### Implementation conformance

An implementation is conformant with `tolvi-format-v1` if it satisfies all of the following:

1. It produces only vault-conformant output.
2. It accepts every vault-conformant input without error.
3. It handles broken wiki-links per Section 7 (warnings, not errors).
4. It applies the default status filter per Section 6 (excluding `superseded`, `deprecated`, `draft` from default queries).
5. It validates supersession bidirectionality per Section 8 (rejection or loud warning; not silent acceptance).
6. It rejects unknown non-`x-*` frontmatter fields per Section 5.

The schemas in `./schemas/` are authoritative. If this prose disagrees with the schemas, the schemas win.

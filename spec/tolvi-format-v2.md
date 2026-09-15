# Tolvi format v2

**Version:** 2.0.0 — 2026-09-14

## 1. Status

Current. Supersedes [`tolvi-format-v1`](./tolvi-format-v1.md), which is frozen. The schema files in `./schemas/` are the machine-readable conformance test for this spec; where this prose disagrees with those schemas, the schemas win.

**What changed from v1, and why.** v1 modeled a vault as belonging to a workspace and nothing more, and it carried no way for a vault to say which repo within that workspace it was, or which product family that repo shared a brain with. Tooling filled the gap out-of-band, and every consumer that did so derived its own answer; they disagreed silently, which is the failure this revision exists to end. v2 adds that identity to the vault and moves the question of *where* a doc is stored out of the format entirely: a vault declares who it is, and the machine reading it decides where its content lives.

Only one change is breaking: `schema_version` MUST now be `2`, which by the rule in Section 11 makes every v1 vault non-conformant under v2. Every other change in this revision is the addition of optional fields and is non-breaking in isolation.

> **RFC 2119 keywords.** The words MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY in this document carry their RFC 2119 meanings. MUST is a hard requirement; SHOULD is a strong recommendation that may be deviated from with documented reason; MAY denotes a permitted but optional behavior. Sections labeled "(informative)" describe defaults and rationale and are not subject to conformance tests.

## 2. Vault directory layout (normative)

A Tolvi vault is a directory containing a `.vault-meta.json` file at its root and three subdirectories holding the three doc types:

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

Any directory containing a `.vault-meta.json` file at its root MUST be treated as a Tolvi vault. The three subdirectories `decisions/`, `sessions/`, and `patterns/` MUST be present for conformance, even when empty. Implementations MAY ignore additional sibling subdirectories so that future doc types can be added without breaking existing readers; implementations MUST NOT reject a vault solely because unknown subdirectories are present.

A vault's location on disk is not normative. Vaults typically live at `<repo>/vault/`, but tooling MUST locate vaults by the presence of `.vault-meta.json`, not by path.

## 3. `.vault-meta.json` (normative)

The vault metadata file declares the workspace identity and local indexing configuration.

| Field | Type | Required | Description |
|---|---|---|---|
| `workspace` | string | yes | Workspace this vault belongs to: the container, and the unit servers key multi-tenant isolation on. |
| `repo` | string | no | The repo this vault belongs to: the member within that workspace. A container vault serves a whole workspace and names no repo. |
| `product` | string | no | A product whose repos share a vault root. Declared only where sibling repos genuinely share a brain. |
| `embedding_model` | string | yes | Identifier of the embedding model used to index this vault locally. The default value SHOULD be `nomic-embed-text` (Ollama). |
| `schema_version` | integer | yes | Vault format version. For `tolvi-format-v2`, this MUST be `2`. |

`workspace` and `repo` are two dimensions, not synonyms: uniqueness is keyed on the pair, so a `repo` value is meaningful only within its `workspace`. Implementations MUST NOT collapse them into a single identifier.

**Storage location is not part of this format.** A vault says who it is; it MUST NOT declare where any other vault lives. Implementations that route documents between vaults MUST take that mapping from configuration outside the vault, so that a vault which is published cannot leak the location of one that is not.

The authoritative form is `./schemas/vault-meta.json`. Canonical example:

```json
{
  "workspace": "<workspace-slug>",
  "repo": "<repo-slug>",
  "embedding_model": "nomic-embed-text",
  "schema_version": 2
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
- `user_impact` — free-form string describing who feels this decision and how much. Deliberately not an enum: the useful answer is usually a phrase, and a fixed scale invites picking a label over saying what changes.
- `product_area` — free-form string naming the part of the product this decision governs, as the authoring team names it. Free-form because the areas differ per project and a shared enum would be wrong for everyone.
- `visibility` — `public` or `private`. See below.

**Patterns** optionally accept:

- `languages` — array of strings.
- `frameworks` — array of strings.
- `visibility` — `public` or `private`. See below.

Patterns intentionally have no `date` or `repo` field, because patterns describe approaches that outlive any single decision.

**Visibility.** `visibility` marks a doc as internal. It is advisory to the format and meaningful only to implementations that store documents across more than one vault: such an implementation SHOULD keep a `private` doc out of any vault it treats as publishable. Absent or `public` carries no constraint. Sessions deliberately do not accept this field — an implementation that separates internal from published content already treats session logs as internal, so a per-session flag would be a second answer to a question already settled.

**Custom fields.** Implementations MAY add custom fields under an `x-*` namespace (for example, `x-internal-priority: high`). Implementations MUST ignore unknown `x-*` fields rather than rejecting the document. This is the forward-compatibility escape hatch: domain-specific extensions live under `x-*`, the core spec stays small.

Future format versions (`tolvi-format-v2` and later) MUST NOT claim names beginning with `x-`. The `x-*` namespace is reserved permanently for implementation extensions.

**Unknown fields.** Implementations MUST reject documents that contain unknown top-level frontmatter fields not beginning with `x-`. This includes typos (`statuss:`), removed v0 experimental fields, and fields from a newer format version appearing in an older reader. Strict rejection prevents silent semantic drift; tolerant readers would let typos accumulate undetected.

Implementations MUST reject documents whose frontmatter fails its type schema with a clear error pointing at the offending field.

## 6. Status enum (normative)

The status enum is **frozen at six values**, unchanged from v1. Adding a value requires `tolvi-format-v3`.

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

`tolvi-format-v2` is current. Breaking changes require a `tolvi-format-v3` revision, which will live at `/spec/tolvi-format-v3.md` alongside this document.

Implementations declare the format versions they support; the wire-level handshake details are deferred and tracked in `/docs/OPEN_QUESTIONS.md`. For v2, all conformant implementations MUST support `tolvi-format-v2`.

Implementations SHOULD support both v1 and v2 for at least one format-version cycle, to allow vault content to migrate without forcing a synchronized ecosystem upgrade. An implementation that reads only v2 MUST reject a v1 vault with an error naming the version it found and the migration required, never by guessing.

A change is **breaking** (and therefore requires a new major format version) if it could cause a conformant vault to become non-conformant under the new rules, or could cause a conformant implementation to misbehave on the new content. Adding new optional fields, new informative defaults, or new diagnostics is non-breaking and lands in a point revision. Note that requiring a new `schema_version` value is itself breaking by this rule, which is why this revision is v2 rather than v1.1.

### Migrating a v1 vault to v2 (normative)

The migration is mechanical and touches only `.vault-meta.json`; no document body or frontmatter changes. For each vault:

1. Set `schema_version` to `2`.
2. Add `repo`, naming the repo this vault belongs to within its `workspace`. Omit it only for a container vault that serves a whole workspace.
3. Where `workspace` holds a repo name rather than the container it belongs to, correct it. v1 left this ambiguous and implementations derived it inconsistently; v2 does not.
4. Optionally add `product` where sibling repos share a vault root.

```bash
jq '{workspace: "<org>", repo: "<repo>", embedding_model, schema_version: 2}' \
  vault/.vault-meta.json > tmp && mv tmp vault/.vault-meta.json
```

No automated migrator ships with this revision, because the change is a four-field rewrite of one small JSON file per vault and an automated tool would be harder to audit than the command above. Implementations MAY provide one. What implementations MUST NOT do is migrate a vault silently as a side effect of reading it: a vault's format version is the user's to change.

## 12. Conformance (normative)

Conformance has two parts: vault conformance (a property of content on disk) and implementation conformance (a property of a tool that reads or writes vaults).

### Vault conformance

A vault is conformant with `tolvi-format-v2` if and only if all of the following hold:

1. Every file in `decisions/`, `sessions/`, and `patterns/` validates against its respective JSON Schema (`./schemas/decision.json`, `./schemas/session.json`, `./schemas/pattern.json`).
2. The `.vault-meta.json` file validates against `./schemas/vault-meta.json`.
3. The directory layout matches Section 2.
4. All supersession references are bidirectional per Section 8.

Wiki-link resolution is intentionally not a vault-conformance criterion: a vault may legitimately contain links that target docs not yet written. Wiki-link handling is an implementation-conformance concern (see below).

### Implementation conformance

An implementation is conformant with `tolvi-format-v2` if it satisfies all of the following:

1. It produces only vault-conformant output.
2. It accepts every vault-conformant input without error.
3. It handles broken wiki-links per Section 7 (warnings, not errors).
4. It applies the default status filter per Section 6 (excluding `superseded`, `deprecated`, `draft` from default queries).
5. It validates supersession bidirectionality per Section 8 (rejection or loud warning; not silent acceptance).
6. It rejects unknown non-`x-*` frontmatter fields per Section 5.

The schemas in `./schemas/` are authoritative. If this prose disagrees with the schemas, the schemas win.

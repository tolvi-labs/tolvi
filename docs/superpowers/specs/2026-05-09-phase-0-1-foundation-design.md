# Tolvi Phase 0+1 Foundation — Design Spec

**Status:** Draft (pending approval)
**Date:** 2026-05-09
**Owner:** Alan
**Sub-project of:** Tolvi v1 product roadmap, Phases 0 and 1 (private)
**Reference:** Internal reference materials reviewed prior to drafting (private)

---

## 1. Goal

Bootstrap the `tolvi-labs/tolvi` monorepo to the point where Phase 2 (server) and Phase 3 (CLI) can begin in parallel against a locked, public format contract.

Concretely, after this work merges:

- The repo is structured, licensed, and OSS-ready
- The `tolvi-format-v1` contract is published and machine-validated
- A 10–15 doc synthetic sample vault exists for demos, tests, and CI
- Architecture, conventions, and known open questions are documented as ADRs
- CI validates the contract on every PR
- No language-specific code (Go or TS) exists yet — those phases are deliberately deferred

## 2. Non-goals

Explicitly out of scope for this sub-project:

- Any Go code (Phase 3)
- Any TypeScript server code, Postgres schema, Drizzle migrations (Phase 2)
- Docker compose for self-host (Phase 2)
- Agent skill files (Phase 4)
- TypeScript/Python SDKs (Phase 5)
- Documentation site / Mintlify (Phase 5)
- Any commercial / hosted features (Phase 9)
- An automated aggregator equivalent to `~/Developer-Vault/` for end users — Tolvi v1 ships the per-repo + server model; the local aggregator pattern is documented in CONVENTIONS.md as a recommended-but-manual setup, with automation deferred to v1.x

## 3. Delivery approach

**Single foundational PR** on a feature branch (`feat/foundation`), opened against `main`. Per the project's PR-size guidance (documented in `CONTRIBUTING.md`), PRs should normally stay under 500 lines, but Phase 0+1 is bootstrapping — there is nothing to incrementally add to. The exception is documented in `CONTRIBUTING.md`:

> **PR size:** Keep PRs reviewable in under 30 minutes — typically <500 lines of diff. Bootstrapping or generated artifacts (sample vaults, JSON schemas, generated specs) are exempt and should be split by *type of change* rather than by line count.

The PR is structured as a sequence of atomic commits the reviewer reads in order:

| Commit | Scope | Approx. diff |
|---|---|---|
| 1 | Apache 2.0 LICENSE, NOTICE (attribution), .gitignore, .editorconfig | ~30 lines |
| 2 | Placeholder README, CONTRIBUTING (incl. brand-isolation contributor rule + PR-size guidance), CODE_OF_CONDUCT, SECURITY, NAMESPACE.md, ROADMAP.md | ~280 lines |
| 3 | `.github/` issue + PR templates, `CODEOWNERS` | ~80 lines |
| 4 | Monorepo dirs (`/cli`, `/server`, `/spec`, `/skills`, `/docs`, `/examples`) with stub READMEs explaining what each will hold | ~60 lines |
| 5 | `/docs/ARCHITECTURE.md`, `/docs/CONVENTIONS.md`, `/docs/OPEN_QUESTIONS.md` | ~500 lines |
| 6 | `/docs/adr/0001-architecture-overview.md`, `/docs/adr/0002-vault-format-v1-contract.md`, `/docs/adr/README.md` | ~250 lines |
| 7 | `/spec/tolvi-format-v1.md` | ~400 lines |
| 8 | `/spec/schemas/{decision,session,pattern,vault-meta}.json` | ~200 lines |
| 9 | `/examples/sample-vault/` — `.vault-meta.json` + 10–15 synthetic decision/session/pattern files | ~400 lines |
| 10 | `.github/workflows/validate.yml` + `.github/scripts/extract-frontmatter.js` (Node helper for JSON-schema validation) + `.github/scripts/brand-isolation-check.sh` (CI guard against parent-org references outside NOTICE) | ~140 lines |
| 11 | Pre-commit hooks (`.pre-commit-config.yaml`): trailing whitespace, EOF newline, JSON syntax, frontmatter required-fields check | ~40 lines |

Total estimated diff: ~2,400 lines across 11 commits. Reviewer reads commit-by-commit; each commit is independently meaningful.

Note: this design spec itself (`docs/superpowers/specs/2026-05-09-phase-0-1-foundation-design.md`) is also part of the foundational PR — included as a transparency artifact so external readers can see how Phase 0+1 was scoped and decided. It belongs in commit 5 alongside `ARCHITECTURE.md` and friends.

## 4. Repo skeleton

```
tolvi/
├── LICENSE                          Apache 2.0
├── NOTICE                           Attribution (verbatim text fixed by Tolvi Labs)
├── README.md                        Placeholder — what + 5-min quickstart link.
│                                    Will be rewritten at end of project.
├── CONTRIBUTING.md                  Dev setup, PR conventions (incl. <500-line rule + bootstrapping exception)
├── CODE_OF_CONDUCT.md               Contributor Covenant 2.1
├── SECURITY.md                      Vulnerability reporting; placeholder email until security@tolvi.dev exists
├── NAMESPACE.md                     Pre-flight tracker — table of 9 items, status + timestamp
├── ROADMAP.md                       Public phase list; milestones-only (no internal dates)
├── CODEOWNERS                       @alator-ta on everything for now
├── .gitignore                       Go, Node, OS junk, IDE, .DS_Store, .tolvi/cache, /examples/sample-vault/.tolvi-index/, tolvi-build-plan.md (private working doc)
├── .editorconfig                    Standard
├── .pre-commit-config.yaml          Trailing whitespace, EOF newline, JSON syntax, frontmatter check
│
├── .github/
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   ├── feature_request.md
│   │   └── config.yml
│   ├── PULL_REQUEST_TEMPLATE.md
│   └── workflows/
│       └── validate.yml             Markdown lint + JSON schema validation + link check
│
├── cli/                             Go CLI (Phase 3 ships code)
│   └── README.md                    "What this dir will contain. See /docs/ARCHITECTURE.md."
│
├── server/                          TypeScript Fastify server (Phase 2 ships code)
│   └── README.md                    Same shape.
│
├── spec/                            Public format contract
│   ├── tolvi-format-v1.md
│   ├── schemas/
│   │   ├── decision.json
│   │   ├── session.json
│   │   ├── pattern.json
│   │   └── vault-meta.json
│   └── README.md                    "Spec is normative; schemas are machine-readable form."
│
├── skills/                          Agent integrations (Phase 4 ships content)
│   └── README.md
│
├── docs/                            Contributor docs (NOT the public docs site — that's Phase 5)
│   ├── ARCHITECTURE.md
│   ├── CONVENTIONS.md
│   ├── OPEN_QUESTIONS.md
│   ├── adr/
│   │   ├── README.md                ADR template + index
│   │   ├── 0001-architecture-overview.md
│   │   └── 0002-vault-format-v1-contract.md
│   └── superpowers/specs/           Design specs from process-driven planning
│                                    Committed publicly as transparency artifacts.
│
└── examples/
    └── sample-vault/
        ├── .vault-meta.json
        ├── decisions/               6 synthetic decision docs
        ├── sessions/                3 synthetic session-day files
        └── patterns/                4 synthetic pattern docs
```

## 5. Architecture spec (`/docs/ARCHITECTURE.md`)

Three layers — source, local arm, server arm — generalized for self-host:

```
┌─────────────────────────────────────────────────────────────┐
│  Source: per-repo vaults (in git)                           │
│    <repo>/vault/{decisions,sessions,patterns}/              │
│    <repo>/vault/.vault-meta.json                            │
└──────────────────┬──────────────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
        ▼                     ▼
┌─────────────────┐   ┌─────────────────────────────────────┐
│ Local arm:      │   │ Server arm (self-hostable):         │
│  tolvi CLI      │   │  Fastify + Postgres + pgvector      │
│  + sqlite-vec   │   │  Docker compose                     │
│                 │   │                                     │
│  Single-binary  │   │  POST /v1/sync (CLI publish target) │
│  Go, Ollama     │   │  POST /v1/search, /v1/ask           │
│  embedding      │   │  API key auth per workspace         │
│  default        │   │                                     │
└─────────────────┘   └─────────────────────────────────────┘
```

Sections:

1. **System surfaces and ownership** — CLI owns the local index and capture. Server owns the multi-tenant index and HTTP API. Format spec is the only contract crossing the boundary.
2. **Trust + auth model** — API key per workspace, hashed at rest (`pgcrypto` or `argon2`), scoped to ingest+search. No user accounts in v1. CLI sends `Authorization: Bearer <key>` to the server. Local-only operation requires no key.
3. **Data flow** — local capture (`tolvi sync`) writes the doc to `<repo>/vault/`; CLI computes `content_hash`; on `tolvi publish` (or scheduled CI sync) the doc POSTs to `/v1/sync`; server is idempotent on `(repo_id, path, content_hash)`; chunking + embedding happen server-side; index updates are read-after-write consistent within a single API call.
4. **Component boundaries** — explicit non-shared library: Go CLI implements `tolvi-format-v1` parsing/validation in Go; TS server does the same in TS. The format spec is the shared contract. Trade-off documented: code duplication is intentional (cross-language reach > shared lib convenience).
5. **What's deferred** — multi-arm shared library, web dashboard, OIDC/SSO, billing, OpenAPI-generated SDKs (all Phase 9). Aggregator automation (a unified per-engineer symlink fan-out across multiple repo vaults) is documented but not automated in v1; manual setup instructions live in `/docs/CONVENTIONS.md`.
6. **Self-host story** — `docker compose up` brings up Postgres (+pgvector) + server. CLI points at `http://localhost:3000` by default. Documented end-to-end in Phase 2's deliverable.

## 6. Conventions spec (`/docs/CONVENTIONS.md`)

Generalized from prior reference patterns. Removes deployment- and domain-specific fields; preserves the validated structure.

### 6.1 Vault directory layout

```
<repo>/vault/
  .vault-meta.json
  decisions/                YYYY-MM-DD-slug.md
  sessions/                 YYYY-MM-DD.md (one file per day, multiple H2 session blocks)
  patterns/                 slug.md  (no date prefix — patterns are timeless)
```

`.vault-meta.json` schema:

```json
{
  "workspace": "<workspace-slug>",
  "embedding_model": "nomic-embed-text",
  "schema_version": 1
}
```

### 6.2 Status enum (six values)

| Status | Meaning | Surfaced by default? |
|---|---|---|
| `active` | Current. Default for new docs. | Yes |
| `in-progress` | Decision made, implementation still landing. | Yes |
| `superseded` | Replaced by a newer decision; `superseded_by:` links forward. | No |
| `deprecated` | No longer applicable; left for history. | No |
| `draft` | Work-in-progress, not yet authoritative. | No |
| `historical` | Preserved for context but not actionable. | Yes |

This enum is **frozen for `tolvi-format-v1`**. Adding a value would be a `tolvi-format-v2` change.

**Surfacing note (informative):** The "surfaced by default" column describes default *retrieval* behavior — what `/v1/search` and `tolvi recall` return without explicit status filters. UI presentation (e.g. rendering a status badge for `historical` results) is a consumer concern, not part of the format spec.

### 6.3 Frontmatter schemas

**Universally required** (all doc types):

```yaml
---
tags: [<doc-type>, ...]
status: active
---
```

**Sessions** (additionally required):

```yaml
date: YYYY-MM-DD          # required; matches the filename date
```

**Decisions** (additionally required + optional):

```yaml
date: YYYY-MM-DD          # required; matches the filename date
repo: <repo-slug>         # required; the repo this decision binds to
ticket: <free-form>       # optional; issue tracker ref (any tracker; e.g. "PROJ-123", a URL, or "none")
supersedes: <slug>        # optional; backward link
superseded_by: <slug>     # optional; forward link (set when status: superseded)
```

Genericization notes:
- `product_area: <free-text>` → DROPPED. Use `tags:` for categorization. Domain-specific fields don't belong in a general-purpose spec.
- `user_impact: <enum>` → DROPPED. Same reason. Implementations may add custom fields under an `x-*` namespace if needed.
- `ticket:` is kept but free-form (any tracker), not tied to a specific issue-tracker.

**Patterns** (additionally optional):

```yaml
languages: [...]          # optional
frameworks: [...]         # optional
```

Patterns are intentionally timeless — no `date` or `repo` field required. They live in one canonical home per workspace (the convention for Tolvi's canonical home for cross-repo patterns is deferred to a later phase).

### 6.4 Wiki-link syntax

- `[[slug]]` — link to a doc in the current repo's vault
- `[[repo:slug]]` — cross-repo link (resolved by aggregator or server)
- Citations in `/v1/ask` responses MUST use this syntax for consistency with Obsidian and other vault consumers

### 6.5 Decision template

Three sections, in order:

```markdown
## Why
Context and forces. What's the problem? What constraints apply?

## How
The decision itself. What did we choose, and what trade-off does that buy?

## Outcome
Observable result. Updated when the decision lands or is superseded.
```

This template encodes a structure validated by prior production use.

### 6.6 Aggregator pattern (recommended, not automated in v1)

For engineers who want a unified cross-repo Obsidian view:

```bash
mkdir -p ~/tolvi-vault
cd ~/tolvi-vault
ln -s ~/path/to/repo-a/vault repo-a
ln -s ~/path/to/repo-b/vault repo-b
```

Then open `~/tolvi-vault/` as an Obsidian vault. v1.x will add `tolvi unify` to automate this; v1 documents the manual recipe.

## 7. Format spec (`/spec/tolvi-format-v1.md`)

Public contract. ~400 lines. Sections:

1. **Status** — "Stable as of v1.0. Breaking changes require `tolvi-format-v2` and migration tooling."
2. **Vault directory layout** — normative
3. **`.vault-meta.json`** — normative; JSON Schema reference
4. **File naming rules** — normative
5. **Frontmatter** — normative; per-type JSON Schema references
6. **Status enum** — normative; the six values, default-surface behavior
7. **Wiki-link syntax** — normative
8. **Cross-reference rules** — normative; bidirectional supersession requirement
9. **RAG defaults (informative)** — recency `(0.8 + 0.2 × exp(-age_days/180))`, session ×0.7, default status filter excludes `superseded|deprecated|draft`. **Implementations MAY tune; the contract is the *defaults*, not a requirement to use them.**
10. **Embedding model defaults (informative)** — `nomic-embed-text` (768 dims) for the local CLI, configurable via `.vault-meta.json`. Server-side embedding model is deployment configuration, not part of the spec.
11. **Versioning rules** — `tolvi-format-v2` will require migration tooling; spec docs will live at `/spec/tolvi-format-v2.md`; implementations declare supported versions in their handshake.
12. **Conformance** — "A vault is conformant with `tolvi-format-v1` if every file passes the JSON Schema for its type and the directory layout matches Section 2."

## 8. JSON Schemas (`/spec/schemas/`)

Four files, JSON Schema Draft 2020-12:

- `vault-meta.json` — workspace, embedding_model, schema_version
- `decision.json` — frontmatter schema for decision docs
- `session.json` — frontmatter schema for session-day files
- `pattern.json` — frontmatter schema for pattern docs

Schemas are referenced from `/spec/tolvi-format-v1.md` and validated against the sample vault in CI.

## 9. Sample vault (`/examples/sample-vault/`)

Synthetic, no real data from any TA project. Realistic shape:

| Type | Count | Examples |
|---|---|---|
| decisions | 6 | "Choose Postgres over MySQL", "Adopt feature flags via OpenFeature", "Migration runbook for the 2026-04 schema change", "Switch from REST to gRPC for service-to-service", "Pin Node version with Volta", "Standardize commit message format" |
| sessions | 3 | Three session-day files showing single-block + multi-block patterns; cross-link to decisions |
| patterns | 4 | "Idempotent migrations", "Tracing context propagation", "Feature-flag rollout", "Postgres advisory locks" |

Two of the 6 decisions are deliberately marked `status: superseded` (with a forward link) and `status: historical` so the sample exercises status filtering.

## 10. ADRs (`/docs/adr/`)

ADR convention: numbered sequentially, each is a short markdown doc with `Status / Context / Decision / Consequences` sections. Tolvi dogfoods this convention from PR #1 — every architectural decision in this spec gets an ADR entry.

`0001-architecture-overview.md` — captures the three-layer architecture choice and the Go-CLI-vs-TS-server boundary.

`0002-vault-format-v1-contract.md` — captures the format-spec-first decision and the freeze of the status enum + scoring defaults.

`README.md` — explains what an ADR is, links to upstream ADR resources, provides a template.

## 11. Open questions (`/docs/OPEN_QUESTIONS.md`)

Carries forward unresolved items identified during reference review, plus Tolvi-specific ones:

1. **Loud-fail on Ollama down** — `tolvi sync` should silently succeed even if indexing fails (writes-first principle); should `tolvi doctor` warn loudly when Ollama is down so users know the index is drifting?
2. **Cross-repo type sharing** — when vault frontmatter schemas drift across consumer repos, how is that resolved? (The format spec freeze helps; `tolvi lint --cross-repo` tooling deferred.)
3. **Vault content audit cadence** — no formal review cycle today for marking decisions superseded. Should `tolvi doctor` flag decisions older than N months with no activity?
4. **Index size growth** — production observations suggest ~6.5 MB / ~280 docs is comfortable for in-memory loading; what's the threshold where chunking + storage strategy needs to change?
5. **Secrets pre-commit lint** — vault content can leak credentials if engineers paste from terminals. Should Tolvi ship a `pre-commit` hook that scans for API keys, tokens, and high-entropy strings?
6. **Multi-tenant isolation in Phase 2** — what's the boundary between workspaces in the same Postgres? Per-row `workspace_id` filtering vs row-level security vs per-workspace schema?
7. **Aggregator automation** — `tolvi unify` is on the roadmap but deferred. When does it ship?
8. **CLI ↔ server format-version handshake** — the spec mentions implementations declare supported versions. What does the handshake actually look like?

## 12. CI (`.github/workflows/validate.yml`)

Single workflow, runs on every PR + push to main:

```yaml
name: validate
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Markdown lint
        uses: DavidAnson/markdownlint-cli2-action@v16
        with:
          globs: '**/*.md'
      - name: JSON Schema validation (sample vault frontmatter against schemas)
        run: |
          npx --yes ajv-cli@5 validate \
            -s spec/schemas/decision.json \
            -d 'examples/sample-vault/decisions/*.md' \
            --extract-frontmatter
          # ...repeat for session, pattern, vault-meta
      - name: Link checker
        uses: lycheeverse/lychee-action@v2
        with:
          args: --no-progress --exclude-mail '**/*.md'
      - name: Brand-isolation guard
        run: |
          # Fail if any tracked file outside NOTICE references the parent org name.
          # Uses a tightly-scoped allowlist (NOTICE only) and a block-list of
          # parent-org and deployment-specific terms.
          ./.github/scripts/brand-isolation-check.sh
```

(The JSON-schema-against-frontmatter step needs a small extractor — likely a 20-line Node script committed to `.github/scripts/extract-frontmatter.js`. The brand-isolation guard is a small bash script at `.github/scripts/brand-isolation-check.sh` — exits non-zero if forbidden terms appear outside NOTICE.)

Pre-commit hooks (`.pre-commit-config.yaml`) cover the local fast-feedback loop: trailing whitespace, EOF newline, JSON syntax, basic frontmatter required-fields check, and a local copy of the brand-isolation guard so authors catch leaks before CI does.

No Go or TS workflows yet — added in Phase 2/3 when those dirs ship code.

## 13. NOTICE file (attribution) and brand-isolation rule

The repo ships a single `NOTICE` file at the root containing exactly this text:

```
Tolvi
Copyright 2026 Torres Atlantic, LLC

This product is developed by Tolvi Labs, the open-source arm of
Torres Atlantic, LLC. Licensed under the Apache License, Version 2.0.
```

This is the **only** location in the repo where the parent organization name appears. All other content — code, docs, specs, examples, sample vault, READMEs, code comments, commit messages, branch names, PR descriptions — is brand-neutral and reads as a self-contained product.

**Three layers of enforcement:**

1. **Documentation** — `CONTRIBUTING.md` includes a "Brand isolation" section explaining the rule to external contributors: "All content in this repo (except the NOTICE file) is brand-neutral. Do not reference the parent organization, sibling products, or specific deployments in code, docs, examples, sample content, comments, commit messages, branch names, or PR descriptions. The brand-isolation guard in CI will reject PRs that violate this."

2. **Pre-commit** — `.pre-commit-config.yaml` runs `.github/scripts/brand-isolation-check.sh` locally so authors catch leaks before pushing.

3. **CI** — `.github/workflows/validate.yml` runs the same script as a hard gate. Block-list of forbidden terms is maintained in the script. Allowlist (where forbidden terms are tolerated) is short and explicit: `NOTICE` (the attribution file), the script itself (it contains the patterns it searches for), and the `docs/superpowers/` directory prefix (engineering-process artifacts that may legitimately discuss the rule, including showing the pattern list).

## 14. Pre-flight tracker (`NAMESPACE.md`)

| # | Item | Status | Timestamp | Notes |
|---|---|---|---|---|
| 1 | GitHub org `tolvi-labs` | ✅ Claimed | 2026-05-09 (verified by user this session) | — |
| 2 | `tolvi.dev` domain | ☐ TODO | — | Register at registrar of choice |
| 3 | `tolvi.com` domain | ☐ TODO | — | Check availability; acquire if reasonable |
| 4 | `tolvilabs.com` / `tolvilabs.dev` | ☐ TODO | — | Secondary, for the Labs brand |
| 5 | PyPI placeholder `tolvi==0.0.0` | ☐ TODO | — | Publish before public launch |
| 6 | npm placeholder `tolvi@0.0.0` | ☐ TODO | — | Publish before public launch |
| 7 | USPTO TESS trademark search | ☐ TODO | — | Software/SaaS classes; "Tolvi" + "Tolvi Labs" |
| 8 | Twitter/X handle `@tolvi` or `@tolvilabs` | ☐ TODO | — | Try both |
| 9 | LinkedIn company page Tolvi Labs | ☐ TODO | — | Reserve, don't activate |

Tracker is checked in so progress is visible in git history. Public — no sensitive data.

## 15. What success looks like (definition of done for this sub-project)

- [ ] PR `feat/foundation` opened against `main` with all 11 commits
- [ ] CI (`validate.yml`) passes green on the PR
- [ ] `npx ajv-cli` validates every sample vault doc's frontmatter against its schema
- [ ] `lychee` finds zero broken links in the docs
- [ ] All 11 commits are independently meaningful (reviewer can `git log --oneline` and understand the structure)
- [ ] User reviews the PR and approves
- [ ] User merges (squash or merge-commit, user's choice)
- [ ] Phase 2 (server) and Phase 3 (CLI) brainstorming sessions can begin against the locked spec

## 16. Sequencing into the rest of the roadmap

After this PR merges:

```
Phase 0+1 ─┬─► Phase 2 (server)    ─┐
           │                        ├─► Phase 4 (agent integrations)
           └─► Phase 3 (CLI)        ─┘
                                    └─► Phase 5 (SDK + docs site)
                                    └─► Phase 6 (soft launch)
                                    └─► Phase 7 (iterate)
                                    └─► Phase 8 (public launch)
                                    └─► Phase 9 (commercial — deferred)
```

Phase 2 and Phase 3 can run in parallel against the locked `tolvi-format-v1` contract. Each gets its own brainstorm → design → plan → implementation cycle.

---

## Appendix A — What this spec deliberately doesn't decide

These are scoped to later phases and left undecided here on purpose:

- Server-side embedding model and provider (Phase 2 decision)
- Postgres schema details (Phase 2)
- Go CLI command set surface — `tolvi init/sync/recall/ask/doctor/unify/publish/status` listed on the roadmap but final command shape is Phase 3's
- sqlite-vec wire format and on-disk layout (Phase 3)
- Claude Code skill file format (Phase 4)
- TypeScript SDK surface (Phase 5)
- Docs site framework — Mintlify vs Docusaurus (Phase 5)
- Telemetry opt-in / metrics shape (Phase 6)

## Appendix B — Reference materials reviewed

Internal reference materials reviewed prior to drafting this spec are tracked privately and not enumerated here. Per project policy, no source code or content from those references is included in the Tolvi repo; only convention shape, design patterns, and lessons-learned were absorbed and re-expressed in Tolvi's own voice.

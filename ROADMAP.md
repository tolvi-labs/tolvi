# Roadmap

The public roadmap for Tolvi v1. Internal sequencing details are tracked privately.

## Status legend

- ✅ Shipped
- 🚧 In progress
- ⏭️ Next up
- 📅 Planned
- 💤 Deferred

## Phases

### Phase 0+1 — Foundation ✅

- Repo skeleton (LICENSE, NOTICE, contributor docs)
- Architecture spec, conventions spec, ADR setup
- `tolvi-format-v1` public format contract + JSON schemas
- Synthetic sample vault for demos and CI validation
- CI guards: markdown lint, schema validation, brand isolation, link checker

### Phase 2 — Server core ✅

- TypeScript + Fastify on Node 20
- Postgres + pgvector for the index
- Multi-tenant API key auth at the workspace level
- Reversible Drizzle migrations
- Docker compose for self-host
- OpenAPI spec generated from route schemas (committed to `spec/openapi.json`)

### Phase 3 — Tolvi CLI ✅

- Go, single static binary, cross-platform (darwin / linux / windows × amd64 / arm64)
- `tolvi init`, `sync`, `ask`, `version`
- CAG architecture for local use (whole vault → Anthropic context via prompt caching)
- GoReleaser-driven distribution
- `recall`, `doctor`, `unify`, `publish`, `status` were deferred from the original plan; the v1 CLI scope is locked to `init`, `sync`, `ask`, and `version`

### Phase 3.x — CLI follow-ups

- ✅ `tolvi precommit install` — non-blocking git pre-commit nudge that flags commits touching dependency manifests, infra config, tooling config, or large diffs
- 📅 `tolvi vault-index` — regenerate the per-vault index block in an agent conventions file (`CLAUDE.md`, `.cursorrules`, etc.) from doc frontmatter and titles; idempotent; intended for use as a pre-commit hook or session-end step. Convention is documented in [`docs/CONVENTIONS.md`](./docs/CONVENTIONS.md) Section 7 and [`docs/adr/0003-vault-index-and-tldr-system.md`](./docs/adr/0003-vault-index-and-tldr-system.md)
- ✅ OpenAPI response schemas on all v1 data-plane routes (`spec/openapi.json` now describes both requests and responses end-to-end; unblocks Phase 5 SDK)

### Phase 4 — Agent integrations ✅

- Claude Code skill (Tier 1 — deep): `/tolvi` slash command with format-spec awareness, CLI orchestration, behavioral rules
- Cursor `.cursorrules` template (Tier 2 — light)
- Aider, OpenHands, Continue skeleton conventions (Tier 3)

### Phase 5 — TypeScript SDK and docs ✅

- ✅ Phase 5.A: `@tolvi-labs/sdk` — hand-written `Tolvi` client over `openapi-typescript`-generated types covering all v1 data-plane endpoints, typed error hierarchy, vitest + contract coverage tests, CI + release workflows (npm publish with provenance on `sdk-v*` tags).
- ✅ Phase 5.B: Documentation site live at `tolvilabs.com/docs` — Tier 1 narrative pages (quickstart, install, CLI, SDK, vault format, agent integrations).
- ✅ Phase 5.C: folded into the docs site — the SDK page documents the typed client and `spec/openapi.json` is the API contract (no separate generated reference renderer).

### Phase 6 — Soft launch 🚧

- ✅ `v0.1.0` tagged release on GitHub — cross-platform CLI binaries (macOS / Linux / Windows) + checksums
- ✅ Public repository (Apache 2.0)
- ✅ Hosted docs site live at `tolvilabs.com/docs`
- 📅 Homebrew tap, npm/PyPI publishes, Docker Hub image
- 📅 5–15 friendly users from the maintainer's network
- 📅 Daily metric tracking

### Phase 7 — Iterate 📅

- Whatever the first wave of users hits

### Phase 8 — Public launch 📅

- Show HN with a strong title biased toward concrete pain
- Posts in relevant developer communities

### Phase 9 — Commercial layer 💤

Triggered by adoption signal — see the README for details on how the OSS and any future hosted offering coexist.

## Versioning

The vault format is independently versioned (`tolvi-format-v1`, `tolvi-format-v2`, …). Format-version compatibility is documented in [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md).

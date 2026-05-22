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
- `recall`, `doctor`, `unify`, `publish`, `status` deferred from the original plan; see [`docs/superpowers/specs/2026-05-14-phase-3-cli-design.md`](./docs/superpowers/specs/2026-05-14-phase-3-cli-design.md) for the locked v1 scope

### Phase 3.x — CLI follow-ups

- ✅ `tolvi precommit install` — non-blocking git pre-commit nudge that flags commits touching dependency manifests, infra config, tooling config, or large diffs
- 📅 OpenAPI response schemas on data-plane routes (deferred from Phase 5 design; tracked separately)

### Phase 4 — Agent integrations ✅

- Claude Code skill (Tier 1 — deep): `/tolvi` slash command with format-spec awareness, CLI orchestration, behavioral rules
- Cursor `.cursorrules` template (Tier 2 — light)
- Aider, OpenHands, Continue skeleton conventions (Tier 3)

### Phase 5 — TypeScript SDK and docs ⏭️

- `@tolvi-labs/sdk` on npm
- Documentation site at `tolvilabs.com/tolvi`
- API reference auto-generated from OpenAPI
- Migration guides

### Phase 6 — Soft launch 📅

- Apache 2.0 release on GitHub
- Homebrew tap, npm/PyPI publishes, Docker Hub image
- 5–15 friendly users from the maintainer's network
- Daily metric tracking

### Phase 7 — Iterate 📅

- Whatever the first wave of users hits

### Phase 8 — Public launch 📅

- Show HN with a strong title biased toward concrete pain
- Posts in relevant developer communities

### Phase 9 — Commercial layer 💤

Triggered by adoption signal — see the README for details on how the OSS and any future hosted offering coexist.

## Versioning

The vault format is independently versioned (`tolvi-format-v1`, `tolvi-format-v2`, …). Format-version compatibility is documented in [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md).

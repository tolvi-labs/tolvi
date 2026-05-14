# Roadmap

The public roadmap for Tolvi v1. Internal sequencing details are tracked privately.

## Status legend

- ✅ Shipped
- 🚧 In progress
- ⏭️ Next up
- 📅 Planned
- 💤 Deferred

## Phases

### Phase 0+1 — Foundation 🚧

- Repo skeleton (LICENSE, NOTICE, contributor docs)
- Architecture spec, conventions spec, ADR setup
- `tolvi-format-v1` public format contract + JSON schemas
- Synthetic sample vault for demos and CI validation
- CI guards: markdown lint, schema validation, brand isolation, link checker

### Phase 2 — Server core ⏭️

- TypeScript + Fastify on Node 20
- Postgres + pgvector for the index
- Multi-tenant API key auth at the workspace level
- Reversible Drizzle migrations
- Docker compose for self-host

### Phase 3 — Tolvi CLI ⏭️

- Go, single static binary, cross-platform
- `tolvi init`, `sync`, `recall`, `ask`, `doctor`, `unify`, `publish`, `status`
- Embedded sqlite-vec local index
- Ollama embedding model by default

### Phase 4 — Agent integrations 📅

- Claude Code skill files
- Cursor `.cursorrules` template
- Aider, OpenHands, Continue (skeleton integrations)

### Phase 5 — TypeScript SDK and docs 📅

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

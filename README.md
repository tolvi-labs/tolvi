# Tolvi

Capture engineering decisions where they happen. Query them later in plain English.

> **Status:** Pre-release. The repo skeleton, format spec, and CI guards are in place; CLI and server land in subsequent phases. See [`ROADMAP.md`](./ROADMAP.md) for the public phase plan.

## What's here today

- [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md) — the public vault format contract
- [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) — system architecture
- [`docs/CONVENTIONS.md`](./docs/CONVENTIONS.md) — vault content conventions
- [`examples/sample-vault/`](./examples/sample-vault/) — synthetic vault that validates against the spec

## What's coming

- `tolvi` CLI (Go, single binary) — `init`, `sync`, `recall`, `ask`, `doctor`, `publish`
- `tolvi` server (TypeScript, Fastify, Postgres + pgvector) — multi-tenant, self-hostable
- Agent integrations (Claude Code, Cursor, others)
- TypeScript and Python SDKs

## License

[Apache 2.0](./LICENSE). See [`NOTICE`](./NOTICE) for attribution.

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). External contributors should read the **Brand isolation** section before opening a PR.

# Tolvi

Capture engineering decisions where they happen. Query them later in plain English.

> **Status:** Pre-1.0 — code shipped through Phase 4, no official release tagged yet. CLI + server + agent integrations are functional on `main`; TypeScript SDK and docs site (Phase 5) and launch ops (Phase 6) remain. See [`ROADMAP.md`](./ROADMAP.md).

## What Tolvi is

A per-repo engineering knowledge vault — decisions, sessions, and patterns stored as Markdown with YAML frontmatter under `<repo>/vault/`. A CLI lets you write and read it from the terminal; a server lets a team share an indexed view; agent integrations (Claude Code, Cursor, Aider, OpenHands, Continue) teach AI tools to use the vault first-class.

Two architectures, one format:

- **Local arm (CLI)** uses CAG — whole vault → Anthropic context via prompt caching. Zero infrastructure beyond an API key.
- **Server arm** uses RAG — pgvector + Ollama embeddings, multi-tenant, self-hostable via Docker Compose. For teams who outgrow the local context window or want a shared index.

The vault format (`tolvi-format-v1`) is the contract between the two arms and the only thing agents need to learn.

## Quickstart

The "drop into a repo" wedge:

```bash
# 1. Install the CLI (Go required; or grab a release binary)
go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest

# 2. Set your Anthropic API key (or write ~/.config/tolvi/config.yaml)
export ANTHROPIC_API_KEY=sk-ant-...

# 3. In any repo:
tolvi init
tolvi sync decision "Why we chose Postgres" --body "pgvector + JSON support tipped it"
tolvi ask "what did we decide about Postgres"
```

For Claude Code users, the skill at [`integrations/claude-code/`](./integrations/claude-code/) lets you do the same thing in natural language inside a Claude Code session (`/tolvi` slash command).

For optional pre-commit nudges that flag commits touching decision-likely files (deps, infra, tooling, large diffs):

```bash
tolvi precommit install
```

## What's shipped

| Surface | Where | Status |
|---|---|---|
| **Format spec** `tolvi-format-v1` | [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md), [`spec/schemas/`](./spec/schemas/) | ✅ |
| **CLI** (`init`, `sync`, `ask`, `precommit`, `version`) | [`cli/`](./cli/) | ✅ Phase 3 + 3.x |
| **Server** (Fastify + Postgres + pgvector, multi-tenant, OpenAPI) | [`server/`](./server/), [`spec/openapi.json`](./spec/openapi.json) | ✅ Phase 2 |
| **Claude Code skill** (Tier 1 — `/tolvi` slash command) | [`integrations/claude-code/`](./integrations/claude-code/) | ✅ Phase 4 |
| **Cursor `.cursorrules`** (Tier 2) | [`integrations/cursor/`](./integrations/cursor/) | ✅ Phase 4 |
| **Aider / OpenHands / Continue** skeletons (Tier 3) | [`integrations/aider/`](./integrations/aider/), [`integrations/openhands/`](./integrations/openhands/), [`integrations/continue/`](./integrations/continue/) | ✅ Phase 4 |
| **Sample vault** (synthetic, validates against the format spec) | [`examples/sample-vault/`](./examples/sample-vault/) | ✅ |

## What's coming

- **TypeScript SDK** for the server's HTTP API (auto-generated from `spec/openapi.json`)
- **Docs site** at [`tolvilabs.com/tolvi`](https://tolvilabs.com/tolvi) (not live yet)
- **First tagged release** — `cli-v0.1.0`, `server-v0.1.0` — once Phase 5 lands
- **Distribution channels** — Homebrew tap, npm publish, Docker Hub image (Phase 6)

## Repository layout

```
tolvi-labs/tolvi/
├── spec/                # Format spec + JSON Schemas + generated OpenAPI
├── cli/                 # Go CLI (single static binary)
├── server/              # TypeScript Fastify server (Docker Compose self-host)
├── integrations/        # Per-agent integration files
├── examples/sample-vault/  # Synthetic vault for demos + CI validation
├── docs/                # Architecture, conventions, ADRs, design specs
└── .github/             # CI workflows + scripts
```

Each subdirectory has its own README with detailed install and usage notes.

## Architecture

See [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) (note: some sections describe the original v1 plan; the CAG-vs-RAG split and Phase 5+ details aren't there yet — slated for the docs site work).

## License

[Apache 2.0](./LICENSE). See [`NOTICE`](./NOTICE) for attribution.

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). External contributors should read the **Brand isolation** section before opening a PR.

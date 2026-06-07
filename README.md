# Tolvi

Capture engineering decisions where they happen. Query them later in plain English.

> **Status:** Pre-1.0 - `v0.1.1` released. The CLI, server, TypeScript SDK, agent integrations, and the [docs site](https://tolvilabs.com/docs) are shipped; Homebrew tap and npm SDK publish are in progress. See [`ROADMAP.md`](./ROADMAP.md).

## What Tolvi is

A per-repo engineering knowledge vault - decisions, sessions, and patterns stored as Markdown with YAML frontmatter under `<repo>/vault/`. A CLI lets you write and read it from the terminal; a server lets a team share an indexed view; agent integrations (Claude Code, Cursor, Aider, OpenHands, Continue) teach AI tools to use the vault first-class.

Two architectures, one format:

- **Local arm (CLI)** uses CAG - whole vault → Anthropic context via prompt caching. Zero infrastructure beyond an API key.
- **Server arm** uses RAG - pgvector + Ollama embeddings, multi-tenant, self-hostable via Docker Compose. For teams who outgrow the local context window or want a shared index.

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

## Two ways to capture

Tolvi captures at two altitudes, on purpose — they are complementary, not redundant:

- **Mechanical (CLI)** — `tolvi sync` writes a single note you already have in mind; `tolvi commit` stages `vault/` and runs `git commit`, gated on a session note existing for today. Deterministic, no LLM, scriptable — what you commit is exactly what is there. Use it in CI, in hooks, or when you want control and no surprises.
- **Synthesized (skill)** — inside a Claude Code or Cursor session, `/tolvi-sync` reconstructs the *whole* working session into decisions, patterns, and a session log, and `/tolvi-commit` does that and then commits. Comprehensive and near-zero effort, but it needs an agent in the loop and is non-deterministic.

Rule of thumb: **mechanical for known, controlled capture; the skill for synthesizing the messy reality of a working session.** The skill captures what a qualified actor tried or considered (including reasoned rejections) — the high-signal record a Slack thread or ticket can't give you.

## What's shipped

| Surface | Where | Status |
|---|---|---|
| **Format spec** `tolvi-format-v1` | [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md), [`spec/schemas/`](./spec/schemas/) | ✅ |
| **CLI** (`init`, `sync`, `ask`, `recall`, `commit`, `precommit`, `version`) | [`cli/`](./cli/) | ✅ Phase 3 + 3.x |
| **Server** (Fastify + Postgres + pgvector, multi-tenant, OpenAPI) | [`server/`](./server/), [`spec/openapi.json`](./spec/openapi.json) | ✅ Phase 2 |
| **TypeScript SDK** `@tolvi-labs/sdk` (typed client over the server's HTTP API) | [`sdk/`](./sdk/) | ✅ Phase 5.A |
| **Claude Code skill** (Tier 1 - `/tolvi` slash command) | [`integrations/claude-code/`](./integrations/claude-code/) | ✅ Phase 4 |
| **Cursor `.cursorrules`** (Tier 2) | [`integrations/cursor/`](./integrations/cursor/) | ✅ Phase 4 |
| **Aider / OpenHands / Continue** skeletons (Tier 3) | [`integrations/aider/`](./integrations/aider/), [`integrations/openhands/`](./integrations/openhands/), [`integrations/continue/`](./integrations/continue/) | ✅ Phase 4 |
| **Sample vault** (synthetic, validates against the format spec) | [`examples/sample-vault/`](./examples/sample-vault/) | ✅ |
| **Docs site** (narrative guides + reference) | [`tolvilabs.com/docs`](https://tolvilabs.com/docs) | ✅ Phase 5.B |

## What's coming

- **Distribution channels** - Homebrew tap (token pending), npm SDK publish, Docker Hub image (Phase 6)

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

See [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) (note: some sections describe the original v1 plan; the fuller CAG-vs-RAG treatment lives in the docs site).

## License

[Apache 2.0](./LICENSE). See [`NOTICE`](./NOTICE) for attribution.

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). External contributors should read the **Brand isolation** section before opening a PR.

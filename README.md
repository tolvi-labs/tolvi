# Tolvi

Capture engineering decisions where they happen. Query them later in plain English.

> **Status:** Pre-1.0 - `v0.1.1` released. The CLI, server, TypeScript SDK, agent integrations, and the [docs site](https://tolvilabs.com/docs) are shipped; Homebrew tap and npm SDK publish are in progress. See [`ROADMAP.md`](./ROADMAP.md).

## What Tolvi is

A per-repo engineering knowledge vault - decisions, sessions, and patterns stored as Markdown with YAML frontmatter under `<repo>/vault/`. A CLI lets you write and read it from the terminal; a server lets a team share an indexed view; agent integrations (Claude Code, Codex, Cursor, Gemini CLI, OpenHands, Aider) teach AI tools to use the vault first-class.

Two architectures, one format:

- **Local arm (CLI)** uses CAG - whole vault → Anthropic context via prompt caching. Zero infrastructure beyond an API key.
- **Server arm** uses RAG - pgvector + Ollama embeddings, multi-tenant, self-hostable via Docker Compose. For teams who outgrow the local context window or want a shared index.

The vault format (`tolvi-format-v2`) is the contract between the two arms and the only thing agents need to learn.

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
tolvi recall                    # what was I doing? recent sessions + active decisions, no API call

# 4. Check the setup, and see where docs are routed:
tolvi doctor
tolvi doctor --json             # same checks, machine-readable, for scripts and tools
tolvi roots

# 5. See every repo on this machine that has a vault:
tolvi repos list
tolvi repos scan ~/src          # register ones that were never registered
```

Role packs give a new vault templates tuned to a kind of work. `tolvi packs list` shows the six the binary carries, and `tolvi init --pack engineer` provisions with one. The packs are owned by [tolvi-solo](https://github.com/tolvi-labs/tolvi-solo) and vendored here, pinned byte-identical by a parity check, so a `brew install` can provision a pack with no sibling checkout and solo keeps ownership.

For Claude Code users, the skill at [`skills/tolvi/`](./skills/tolvi/) lets you do the same thing in natural language inside a Claude Code session (`/tolvi` slash command). From a checkout, `bash skills/tolvi/install.sh --with-hooks` installs it. With no checkout, which is what `brew install` gives you, the binary carries the same files:

```bash
tolvi integrations install --with-hooks
```

It writes the skill and the three slash commands, wires the session hooks, and allowlists reads only: `tolvi sync` and `tolvi commit` stay a conscious per-call approval. Nothing is overwritten without `--force`, and running it twice changes nothing.

`tolvi repos` keeps a machine-local index at `~/.config/tolvi/repos.json`, beside `roots.json` and never committed. `tolvi init` registers a repo as it provisions it, and `tolvi repos scan <dir>` picks up ones that predate the index. The index is a cache over each repo's own `.vault-meta.json`, so entries are verified when they are read: a repo you moved is listed as stale rather than trusted, and `tolvi repos forget <path>` drops it when you say so. Losing the file costs nothing, because a scan rebuilds it.

`tolvi roots` is worth knowing early if you use more than one repo. Roots are declared once per machine in `~/.config/tolvi/roots.json` and never committed, so a repo commits only who it is and the machine decides where its docs live. With no `roots.json` everything stays in the repo's own `vault/`, which is the default and what a contributor wants.

For optional pre-commit nudges that flag commits touching decision-likely files (deps, infra, tooling, large diffs):

```bash
tolvi precommit install
```

## The stack

Every Tolvi tool installs from one marketplace:

```bash
/plugin marketplace add tolvi-labs/tolvi
/plugin install tolvi-guild        # plan against your codebase
/plugin install tolvi-bastion      # harden the plan before code
/plugin install tolvi-magellan     # compile it into an executable task DAG
```

`tolvi` itself, and `tolvi-solo` for the server-free single-builder setup, install the same way. Each plugin also ships manifests for Codex, Cursor, Gemini, Devin and Hermes, so the skills work outside Claude Code too.

## Two ways to capture

Tolvi captures at two altitudes, on purpose. They are complementary, not redundant:

- **Mechanical (CLI)**: `tolvi sync` writes a single note you already have in mind; `tolvi commit` stages `vault/` and runs `git commit`, gated on a session note existing for today. Deterministic, no LLM, scriptable: what you commit is exactly what is there. Use it in CI, in hooks, or when you want control and no surprises.
- **Synthesized (skill)**: inside a Claude Code or Cursor session, `/tolvi-sync` reconstructs the *whole* working session into decisions, patterns, and a session log, and `/tolvi-commit` does that and then commits. Comprehensive and near-zero effort, but it needs an agent in the loop and is non-deterministic.

Rule of thumb: **mechanical for known, controlled capture; the skill for synthesizing the messy reality of a working session.** The skill captures what a qualified actor tried or considered (including reasoned rejections), the high-signal record a Slack thread or ticket can't give you.

## What's shipped

| Surface | Where | Status |
|---|---|---|
| **Format spec** `tolvi-format-v2` | [`spec/tolvi-format-v2.md`](./spec/tolvi-format-v2.md), [`spec/schemas/`](./spec/schemas/) | ✅ |
| **CLI** (`init`, `sync`, `ask`, `recall`, `commit`, `precommit`, `version`) | [`cli/`](./cli/) | ✅ Phase 3 + 3.x |
| **Server** (Fastify + Postgres + pgvector, multi-tenant, OpenAPI) | [`server/`](./server/), [`spec/openapi.json`](./spec/openapi.json) | ✅ Phase 2 |
| **TypeScript SDK** `@tolvi-labs/sdk` (typed client over the server's HTTP API) | [`sdk/`](./sdk/) | ✅ Phase 5.A |
| **Claude Code skill** (Tier 1 - `/tolvi` slash command) | [`skills/tolvi/`](./skills/tolvi/) | ✅ Phase 4 |
| **Codex / Cursor / Gemini CLI / OpenHands** on the shared skill (Tier 2) | [`integrations/codex/`](./integrations/codex/), [`integrations/cursor/`](./integrations/cursor/), [`integrations/gemini/`](./integrations/gemini/), [`integrations/openhands/`](./integrations/openhands/) | ✅ Phase 4, reworked 2026-09-16 |
| **Aider** conventions file (Tier 3) | [`integrations/aider/`](./integrations/aider/) | ✅ Phase 4 |
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

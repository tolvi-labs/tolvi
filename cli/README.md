# Tolvi CLI

> **Status:** Phase 3 (not yet shipped). Track progress in [`ROADMAP.md`](../ROADMAP.md).

This directory will hold the `tolvi` command-line interface — a single static Go binary, cross-platform (darwin / linux / windows × amd64 / arm64).

## Planned commands

- `tolvi init` — scaffold `.tolvi/` in the current repo
- `tolvi sync` — capture session context as a vault doc
- `tolvi status <doc> <new_status>` — update lifecycle status with supersession tracking
- `tolvi recall <query>` — local lexical search over the current repo's vault
- `tolvi ask <query>` — semantic search + LLM-synthesized answer
- `tolvi doctor` — diagnostic: config, vault structure, embedding model, server connectivity
- `tolvi unify` — generate the cross-repo unified Obsidian view
- `tolvi publish` — push vault content to a configured Tolvi server (optional)

See [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) for the local-arm component in context.

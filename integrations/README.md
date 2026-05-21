# Tolvi integrations

Per-agent configuration files for using a Tolvi vault from your AI coding tool of choice.

## Status

| Integration | Tier | Status | Directory |
|---|---|---|---|
| Claude Code | 1 — deep | ✅ shipped | [`claude-code/`](./claude-code/) |
| Cursor | 2 — light | ✅ shipped | [`cursor/`](./cursor/) |
| Aider | 3 — skeleton | ✅ shipped | [`aider/`](./aider/) |
| OpenHands | 3 — skeleton | ✅ shipped | [`openhands/`](./openhands/) |
| Continue | 3 — skeleton | ✅ shipped | [`continue/`](./continue/) |

**Tier 1 — deep:** Custom skill files with slash commands, format-spec awareness, and CLI orchestration. The agent can read, write, and ask questions of the vault as a first-class workflow.

**Tier 2 — light:** Static configuration (e.g., `.cursorrules`) that teaches the agent about the vault format. No tool wiring; the agent uses its own primitives.

**Tier 3 — skeleton:** Per-tool README snippets or convention files showing the agent the vault layout. Symbolic — proves Tolvi is agent-agnostic without per-tool investment.

## Common conventions

All integrations assume:

- The `tolvi` CLI is installed and in `$PATH` (or the agent degrades gracefully).
- A vault exists at `<repo>/vault/` with a valid `.vault-meta.json` (created by `tolvi init`).
- The Tolvi format spec is at [`spec/tolvi-format-v1.md`](../spec/tolvi-format-v1.md).

## Adding a new integration

1. Create `integrations/<agent-name>/`.
2. Add the agent's primary config artifact (skill, rules, conventions doc).
3. Add a `README.md` documenting install + uninstall.
4. Update this top-level table.
5. If the integration ships installable scripts, add a CI smoke test under `.github/scripts/`.

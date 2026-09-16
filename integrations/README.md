# Tolvi integrations

Per-agent setup for using a Tolvi vault from your AI coding tool of choice.

## Status

| Integration | Tier | Status | Directory |
|---|---|---|---|
| Claude Code | 1 (deep) | ✅ shipped | [`skills/tolvi/`](../skills/tolvi/) |
| Codex | 2 (shared skill) | ✅ shipped | [`codex/`](./codex/) |
| Cursor | 2 (shared skill) | ✅ shipped | [`cursor/`](./cursor/) |
| OpenHands | 2 (shared skill) | ✅ shipped | [`openhands/`](./openhands/) |
| Aider | 3 (conventions file) | ✅ shipped | [`aider/`](./aider/) |

Continue was removed on 2026-09-16 because Continue was discontinued upstream.

**Tier 1 (deep):** the Tolvi skill, plus the Claude Code slash commands `/tolvi-recall`, `/tolvi-sync` and `/tolvi-commit`, plus session hooks that recall the vault when a session starts and stage vault changes before each commit.

**Tier 2 (shared skill):** the same `SKILL.md` as Tier 1, installed into a project's `.agents/skills/tolvi/` with `install.sh --agents`, for agents that support the Agent Skills standard. The agent reads, writes and asks questions of the vault through the skill. There are no slash commands or hooks.

**Tier 3 (conventions file):** a compact recap of the vault format that the agent loads as a conventions file. There is no skill; the agent uses its own tools.

## Common conventions

All integrations assume:

- The `tolvi` CLI is installed and in `$PATH` (or the agent degrades gracefully).
- A vault exists at `<repo>/vault/` with a valid `.vault-meta.json` (created by `tolvi init`).
- The Tolvi format spec is at [`spec/tolvi-format-v2.md`](../spec/tolvi-format-v2.md).

## Adding a new integration

1. If the agent supports Agent Skills, add no new config artifact: create `integrations/<agent-name>/README.md` documenting `install.sh --agents` for that agent, where it looks for skills, update, uninstall, and how the install was verified.
2. Otherwise, create `integrations/<agent-name>/` with the agent's primary config artifact (rules or conventions file) and a `README.md` documenting install and uninstall.
3. Update the table above.
4. If the integration ships installable scripts, add a CI smoke test under `.github/scripts/`.

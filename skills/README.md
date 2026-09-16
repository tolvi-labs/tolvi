# Skills

Claude Code skills published by this repo, one directory per skill, which is where Claude Code expects to find them.

- [`tolvi/`](./tolvi/): the Tolvi skill. read, write, and ask questions of a vault in natural language, plus the `/tolvi-recall`, `/tolvi-sync` and `/tolvi-commit` slash commands and the session hooks that make capture ambient. Install with [`tolvi/install.sh`](./tolvi/install.sh).

The plugin declares itself in [`../.claude-plugin/plugin.json`](../.claude-plugin/plugin.json).

## Other agents

Integrations for agents that are not Claude Code live in [`../integrations/`](../integrations/). Codex, Cursor, Gemini CLI and OpenHands use this same skill, through a plugin, an extension, or `tolvi/install.sh --agents`; each integration README gives the paths for its agent. Aider takes a conventions file.

## Sibling skills

The rest of the stack ships from its own repos, each with its own manifest and `skills/` directory: [bastion](https://github.com/tolvi-labs/bastion), [guild](https://github.com/tolvi-labs/guild), and [magellan](https://github.com/tolvi-labs/magellan).

# Agent integrations

> **Status:** Phase 4 (not yet shipped). Track progress in [`ROADMAP.md`](../ROADMAP.md).

This directory will hold drop-in integration files for AI coding agents:

- `claude-code/` — skill files for Claude Code (capture-on-finish, recall-before-decide, ask-in-context)
- `cursor/` — `.cursorrules` template + workflow examples
- `aider/`, `openhands/`, `continue/` — skeleton integrations

Once these ship, adding three lines to a `CLAUDE.md` (or equivalent agent-config file) gives the project a queryable, version-controlled history of every architectural decision the agents make.

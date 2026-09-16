---
tags: [decision, tolvi, integrations, agent-skills, installer]
date: 2026-09-16
repo: tolvi
status: active
ticket: none
user_impact: high
product_area: Agent integrations
---

# Codex, Cursor and OpenHands share the Tolvi Agent Skill, installed by install.sh --agents

**Date:** 2026-09-16
**Repo:** tolvi

## Why

The Cursor, OpenHands and Continue integrations shipped rules files that their tools had moved away from, so following the install steps could leave an agent that never saw the Tolvi conventions. Codex, Cursor and OpenHands now all load skills in the Agent Skills format, which the Claude Code skill already uses, so one skill can give those agents the full Tolvi guidance instead of a thin rules file.

## How

- `skills/tolvi/SKILL.md` is the single skill for every skill-capable agent. Its wording no longer addresses Claude by name or by Claude Code tool names; the slash command, hook-json recall output and pre-commit hook passages are labelled Claude Code only rather than removed. Frontmatter is unchanged, and its description already routes on `vault/.vault-meta.json`.
- `skills/tolvi/install.sh --agents` copies `SKILL.md` to `<repo root>/.agents/skills/tolvi/SKILL.md`, the project location Codex, Cursor and OpenHands read. The repo root is the nearest ancestor containing `.git` (a file or a directory, so worktrees work); with no repository it installs under the current directory and warns on stderr.
- Agent installs always copy. A project install is meant to be committed, and a committed symlink into one person's tolvi checkout is broken for everyone else. The older Cursor, OpenHands and Aider READMEs offered that symlink.
- `--agents` installs no slash commands, stack skills or hooks, rejects `--with-hooks` and `--hooks-scope` with a "Claude Code only" message, refuses to overwrite without `--force`, and `--uninstall --agents` removes only the agent install. `--path` sets another base for a personal install, such as `~/.agents/skills`, which Codex, Cursor and OpenHands all read. Behavior without `--agents` is unchanged.
- `.github/scripts/test-install-claude-code-skill.sh` covers the repo-root walk, the no-git fallback and its warning, `--path`, the overwrite refusal, `--force` restoring an edited copy, the hook-flag rejection, plain `--uninstall` leaving an agent install alone, and `--uninstall --agents`. Each check was watched failing against a mutated installer.
- Tiers now describe what the agent gets: Tier 1 is Claude Code (skill, slash commands, session hooks); Tier 2 is Codex, Cursor and OpenHands (the shared skill); Tier 3 is Aider (`CONVENTIONS.md`, loaded with `read: CONVENTIONS.md`).
- Removed: `integrations/cursor/.cursorrules`, `integrations/openhands/.openhands_instructions`, and `integrations/continue/`, because Continue was discontinued upstream in 2026 and its docs had already moved to a rules directory. The Cursor and OpenHands READMEs tell existing users to delete the old files.
- Each integration README ends with a Verification section stating what was actually run. Codex CLI 0.154.0 was run end to end: with the skill installed it listed `tolvi` among its skills, it did not list it in an identical repo without the install, and it answered a vault question by reading the recorded decision. `tolvi ask` was not exercised in that run because no Anthropic API key was set; Codex read the vault files instead of surfacing the missing-key error the skill asks for. Cursor, OpenHands and Aider installs follow each tool's documentation and have not been run end to end.
- Rejected: one native rules file per agent (`.cursor/rules/tolvi.mdc`, an `AGENTS.md` snippet, an OpenHands microagent), because three thin files drift from the Tier 1 skill and from each other. Rejected: an `AGENTS.md` snippet alongside the skill, because it has to be merged into a user's own `AGENTS.md` without overwriting it. Rejected: a `tolvi integrate` CLI subcommand for now, because it moves a docs-and-installer change into CLI release work.
- Known limits, deferred: the project-scope hooks install still searches for `.git` as a directory, so it picks the wrong root inside a worktree, and plain `--uninstall` exits before removing slash commands when `SKILL.md` is already gone. Both predate this change and would alter behavior without `--agents`.
- The vault-index convention in `docs/CONVENTIONS.md` now names `AGENTS.md` as the always-on file for Codex, Cursor and OpenHands.

## Outcome

Codex, Cursor and OpenHands install the same Tolvi skill Claude Code uses with one installer flag, Codex is verified end to end, and no shipped integration points at a retired rules file.

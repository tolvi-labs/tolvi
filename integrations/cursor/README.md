# Tolvi: Cursor integration

The shared Tolvi skill for [Cursor](https://cursor.com). Cursor loads Agent Skills from `.agents/skills/` or `.cursor/skills/` in your project, so installing the skill there teaches Cursor's agent how to read, write and ask questions of a Tolvi vault.

## Tier

Cursor is a **Tier 2 (shared skill)** integration in the [Tolvi integrations tier list](../README.md). It gets the same `SKILL.md` as [Claude Code](../../skills/tolvi/), without the Claude Code slash commands or session hooks.

## Prerequisites

- [Cursor](https://cursor.com) installed.
- The `tolvi` CLI on your `PATH`:

  ```bash
  go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
  ```

  Or download a release binary from <https://github.com/tolvi-labs/tolvi/releases>.

- A Tolvi vault in your repo (`vault/.vault-meta.json`). Run `tolvi init` if you do not have one yet.

## Install

### Into a project

Clone the tolvi repo once, then run the installer from anywhere inside your project:

```bash
git clone https://github.com/tolvi-labs/tolvi /path/to/tolvi
cd /path/to/your-project
bash /path/to/tolvi/skills/tolvi/install.sh --agents
```

The installer finds your repository root and copies the skill to `.agents/skills/tolvi/SKILL.md`. Commit that directory so everyone on the team gets the skill.

To install it for yourself only, in every project:

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --agents --path ~/.cursor/skills
```

Cursor also reads `~/.agents/skills`, so `--path ~/.agents/skills` gives Cursor, Codex and OpenHands one shared personal install.

If you used the earlier Tolvi `.cursorrules` file, delete it from your repo root. It is no longer maintained.

### As a plugin

This repository is also a Cursor plugin. To install it locally, put a copy in Cursor's local plugins folder, then reload the window:

```bash
git clone https://github.com/tolvi-labs/tolvi ~/.cursor/plugins/local/tolvi
```

Run **Developer: Reload Window**, then check that Tolvi appears under **Customize**. To update, run `git pull` in `~/.cursor/plugins/local/tolvi` and reload; to uninstall, delete that folder and reload. The plugin install has not been verified yet; see Verification below. If you also installed the Claude Code session hooks and Cursor is set to include third-party configs, use one or the other in Cursor, not both, or recall context arrives twice.

## Use

Ask Cursor's agent about the vault in plain language. It uses the skill when a request matches its description, such as a question about decisions, sessions or patterns in a repo that has `vault/.vault-meta.json`. You can also type `/tolvi` in the agent chat to load it.

- "What did we decide about Postgres?"
- "Write down that we chose PASETO over JWT."
- "Show me the most recent session log."

## Update

The install is a copy, so after pulling the tolvi repo, re-run it with `--force`:

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --agents --force
```

## Uninstall

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --uninstall --agents
```

For a personal install, add the same `--path` you installed with.

## Troubleshooting

### Cursor does not use the skill

- Is the file at `.agents/skills/tolvi/SKILL.md` under your repository root? Check with `ls .agents/skills/tolvi/`.
- Cursor discovers skills when it starts, so reload the window after installing.
- Type `/tolvi` in the agent chat to load the skill directly.

### `tolvi: command not found`

The CLI is not on your `PATH`. Run `export PATH="$PATH:$(go env GOPATH)/bin"` and add it to your shell profile. Without the CLI the skill falls back to reading `vault/` directly, and you lose `tolvi ask`.

### `tolvi ask` errors about `ANTHROPIC_API_KEY`

Set the `ANTHROPIC_API_KEY` environment variable, or write `~/.config/tolvi/config.yaml`:

```yaml
anthropic_api_key: sk-ant-...
model: claude-sonnet-4-7
```

See the [CLI README](../../cli/) for the full config reference.

## Verification

This install follows the Cursor Agent Skills documentation at <https://cursor.com/docs/skills>, checked on 2026-09-16. It has not been run end to end with Cursor.

The plugin install follows the Cursor plugin documentation at <https://cursor.com/docs/plugins>, checked on 2026-09-16. It has not been run end to end with Cursor.

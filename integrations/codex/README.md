# Tolvi: Codex integration

The shared Tolvi skill for [Codex](https://developers.openai.com/codex). Codex loads Agent Skills from `.agents/skills/` in your project, so installing the skill there teaches Codex how to read, write and ask questions of a Tolvi vault.

## Tier

Codex is a **Tier 2 (shared skill)** integration in the [Tolvi integrations tier list](../README.md). It gets the same `SKILL.md` as [Claude Code](../../skills/tolvi/), without the Claude Code slash commands or session hooks.

## Prerequisites

- Codex installed.
- The `tolvi` CLI on your `PATH`:

  ```bash
  go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
  ```

  Or download a release binary from <https://github.com/tolvi-labs/tolvi/releases>.

- A Tolvi vault in your repo (`vault/.vault-meta.json`). Run `tolvi init` if you do not have one yet.

## Install

### As a plugin

```bash
codex plugin marketplace add tolvi-labs/tolvi
codex plugin add tolvi@tolvi
```

This installs the Tolvi skill for you in every project. Start a new Codex session to pick it up.

### Into a project

Clone the tolvi repo once, then run the installer from anywhere inside your project:

```bash
git clone https://github.com/tolvi-labs/tolvi /path/to/tolvi
cd /path/to/your-project
bash /path/to/tolvi/skills/tolvi/install.sh --agents
```

The installer finds your repository root and copies the skill to `.agents/skills/tolvi/SKILL.md`. Commit that directory so everyone on the team gets the skill.

To install it for yourself only, in every project, without the plugin:

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --agents --path ~/.agents/skills
```

## Use

Ask Codex about the vault in plain language. Codex uses the skill when a request matches its description, such as a question about decisions, sessions or patterns in a repo that has `vault/.vault-meta.json`. You can also name the Tolvi skill in your prompt.

- "What did we decide about Postgres?"
- "Write down that we chose PASETO over JWT."
- "Show me the most recent session log."

## Update

For the plugin, refresh the marketplace and reinstall:

```bash
codex plugin marketplace upgrade tolvi
codex plugin add tolvi@tolvi
```

A project install is a copy, so after pulling the tolvi repo, re-run it with `--force`:

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --agents --force
```

## Uninstall

For the plugin:

```bash
codex plugin remove tolvi@tolvi
codex plugin marketplace remove tolvi
```

For a project install:

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --uninstall --agents
```

For a personal install, add the same `--path` you installed with.

## Troubleshooting

### Codex does not use the skill

- Is the file at `.agents/skills/tolvi/SKILL.md` under your repository root? Check with `ls .agents/skills/tolvi/`.
- If Codex was already running when you installed, start a new session.
- Name the Tolvi skill in your prompt.

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

Verified end to end on 2026-09-16 with Codex CLI 0.154.0 (`npx @openai/codex`): installed with `install.sh --agents` in a repo with a vault, the agent listed the tolvi skill (and did not in the same repo without it), and answered a vault question by reading the recorded decision. `tolvi ask` was not exercised, because no Anthropic API key was set.

The plugin install was verified end to end on 2026-09-16 with Codex CLI 0.154.0: before the install the agent did not list the tolvi skill, and after `codex plugin add tolvi@tolvi` it did. The plugin was installed from a local checkout of this repository. Updating the plugin with `codex plugin marketplace upgrade tolvi` was not exercised.

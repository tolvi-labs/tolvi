# Tolvi — Cursor integration

A `.cursorrules` template that teaches [Cursor](https://cursor.com) about Tolvi vault conventions. When you ask Cursor questions about decisions, sessions, or patterns in a repo that has Tolvi installed, Cursor will use the `tolvi` CLI under the hood and cite vault content with `[[slug]]` references.

## Prerequisites

- [Cursor](https://cursor.com) installed.
- The `tolvi` CLI in your `$PATH`. Install with:

  ```bash
  go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
  ```

  Or download a release binary from <https://github.com/tolvi-labs/tolvi/releases>.

- A Tolvi vault in your repo (`vault/.vault-meta.json`). Run `tolvi init` if you don't have one yet.

## Install

Copy `.cursorrules` from this directory to the **root of your project** (the repo where you want Cursor to know about Tolvi):

```bash
# From the root of your project:
cp /path/to/tolvi-labs/tolvi/integrations/cursor/.cursorrules .cursorrules
```

Or symlink, if you want updates from `git pull` to flow through:

```bash
# From the root of your project:
ln -s /path/to/tolvi-labs/tolvi/integrations/cursor/.cursorrules .cursorrules
```

Commit the `.cursorrules` file alongside the rest of your repo so the rules are shared with everyone on the team.

## Use

There is no slash command to invoke — Cursor reads `.cursorrules` automatically as part of every interaction in this project. Once the file is in your repo, just ask Cursor about the vault in natural language:

- "What did we decide about Postgres?"
- "Write down that we chose PASETO over JWT — body: JWT's lack of true revocation made it unusable for our session model."
- "Show me the most recent session log."
- "This repo doesn't have a vault yet — set one up."

Cursor will invoke `tolvi ask`, `tolvi sync`, or `tolvi init` as appropriate.

## Update

- **Symlink install** — `git pull` on the `tolvi-labs/tolvi` checkout updates `.cursorrules` automatically.
- **Copy install** — re-run the `cp` command to refresh.

## Uninstall

Delete the `.cursorrules` file from your repo root.

```bash
rm .cursorrules
```

## Troubleshooting

### Cursor doesn't seem to know about Tolvi

- Is `.cursorrules` actually at the repo root? `ls -la .cursorrules`
- Did you reload the Cursor window after adding the file? Cursor reads `.cursorrules` at session start.
- Does the file contain the latest content? Compare against `integrations/cursor/.cursorrules` in this repo.

### `tolvi: command not found`

The CLI isn't in `$PATH`. Install via `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest` or grab a release binary. The rules teach Cursor to fall back to direct file operations when the CLI is missing, but the CLI gives you streaming output, citation verification, and atomic writes.

### `tolvi ask` errors about `ANTHROPIC_API_KEY`

The CLI needs your Anthropic API key. Either set the `ANTHROPIC_API_KEY` environment variable or write `~/.config/tolvi/config.yaml`:

```yaml
anthropic_api_key: sk-ant-...
model: claude-sonnet-4-7
```

See the [CLI README](../../cli/) for the full config reference.

## What's in this file

`.cursorrules` is a single plain-text file Cursor reads as context for every interaction in a project. The Tolvi version contains:

- The format spec (frontmatter, slug rules, status enum, wiki-link syntax)
- The CLI command reference (`tolvi ask`, `sync`, `init` with their flags)
- Behavioral rules — when to prefer the CLI versus direct file ops, when to cite
- Escape hatches — what to do when the CLI is missing, the vault doesn't exist, the API key isn't set, the vault is too large

The file is **read-only context**. It doesn't auto-run anything and doesn't proactively interrupt your conversations. Proactive nudges live in the separate `tolvi precommit` git-hook subcommand (planned).

## Tier

Cursor is a **Tier 2 — light** integration in the [Tolvi integrations tier list](../README.md). The agent uses its own primitives (Cursor's built-in read/edit/run tools) to invoke the CLI; there's no custom Tolvi-specific tool wiring. For a deeper integration with custom slash commands, see [Claude Code](../claude-code/) (Tier 1).

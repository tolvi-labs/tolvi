# Tolvi — Claude Code integration

A Claude Code skill that lets you read, write, and ask questions of a Tolvi vault from any Claude Code session.

Type `/tolvi` once at the start of a session, then ask the vault questions or capture decisions in natural language. Claude shells out to the `tolvi` CLI under the hood and verifies citations.

## Prerequisites

- [Claude Code](https://claude.ai/code) installed
- The `tolvi` CLI in your `$PATH`. Install with:

  ```bash
  go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
  ```

  Or download a release binary from <https://github.com/tolvi-labs/tolvi/releases>.

- A Tolvi vault in your repo (`vault/.vault-meta.json`). Run `tolvi init` if you don't have one yet.

## Install

From the root of your `tolvi-labs/tolvi` checkout:

```bash
cd integrations/claude-code
bash install.sh
```

Default mode is **symlink**, so `git pull` on this repo updates the skill automatically.

### Install options

- `--copy` — Deep-copy `SKILL.md` instead of symlinking. Use when you want isolation from `git pull` updates.
- `--path <dir>` — Override the install destination (default: `$HOME/.claude/skills`). The `tolvi/` subdirectory is created under this path.
- `--force` — Overwrite an existing install. Refuses by default to avoid clobbering customizations.

### Manual install (no script)

```bash
# From the root of your tolvi-labs/tolvi checkout:
mkdir -p ~/.claude/skills/tolvi
ln -s "$(pwd)/integrations/claude-code/SKILL.md" ~/.claude/skills/tolvi/SKILL.md
```

## Use

Open a Claude Code session in any repo with a Tolvi vault and type:

```
/tolvi
```

Claude loads the skill content and acknowledges briefly. Then ask in natural language:

- "What did we decide about Postgres?"
- "Write down that we chose PASETO over JWT — body: JWT's lack of true revocation made it unusable for our session model."
- "Show me the most recent session log."
- "This repo doesn't have a vault yet — set one up."

Claude shells out to `tolvi ask`, `tolvi sync`, or `tolvi init` as appropriate.

## Update

- **Symlink install** — run `git pull` in your `tolvi-labs/tolvi` checkout. The skill updates automatically; restart your Claude Code session or re-invoke `/tolvi` to pick up changes.
- **Copy install** — re-run `bash install.sh --copy --force` to refresh.

## Uninstall

```bash
cd integrations/claude-code
bash install.sh --uninstall
```

Removes `~/.claude/skills/tolvi/SKILL.md` and the `tolvi/` directory (if empty). If the directory contains user-added files, it's left in place and the script reports it on stderr.

## Troubleshooting

### Claude doesn't seem to know about Tolvi

- Did you type `/tolvi` at the start of the session? The skill is not auto-loaded.
- Is the symlink intact? Check: `ls -la ~/.claude/skills/tolvi/SKILL.md`
- Did the repo path move? If you cloned `tolvi-labs/tolvi` to a new location, re-run `bash install.sh --force`.

### `tolvi: command not found`

The CLI isn't in `$PATH`. Install via `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest` or grab a release binary. The skill teaches Claude to fall back to direct file operations when the CLI is missing, but the CLI gives you streaming output, citation verification, and atomic writes.

### `tolvi ask` errors about `ANTHROPIC_API_KEY`

The CLI needs your Anthropic API key. Either set the `ANTHROPIC_API_KEY` environment variable or write `~/.config/tolvi/config.yaml`:

```yaml
anthropic_api_key: sk-ant-...
model: claude-sonnet-4-7
```

See the [CLI README](../../cli/) for the full config reference.

## What's in this skill

The skill is a single file: `SKILL.md`. It contains:

- The Tolvi format spec (frontmatter, slug rules, status enum, wiki-link syntax)
- The CLI command reference (`tolvi ask`, `sync`, `init` with their flags)
- Behavioral rules — when to prefer the CLI versus direct file ops, when to cite, when to refuse
- Escape hatches — what to do when the CLI is missing, the vault doesn't exist, the API key isn't set, the vault is too large

The skill is **read-only context**. It doesn't auto-run anything, doesn't store state, and doesn't proactively interrupt your conversations. Proactive nudges live in the separate `tolvi precommit` git-hook subcommand (planned).

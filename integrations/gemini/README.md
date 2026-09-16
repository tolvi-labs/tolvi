# Tolvi: Gemini CLI integration

The Tolvi extension for [Gemini CLI](https://geminicli.com). The extension bundles the shared Tolvi skill and a short context file, which teach Gemini how to read, write and ask questions of a Tolvi vault.

## Tier

Gemini CLI is a **Tier 2 (shared skill)** integration in the [Tolvi integrations tier list](../README.md). It gets the same `SKILL.md` as [Claude Code](../../skills/tolvi/), without the Claude Code slash commands or session hooks.

## Prerequisites

- [Gemini CLI](https://geminicli.com) installed, with git available.
- The `tolvi` CLI on your `PATH` (recommended; without it the skill reads `vault/` directly):

  ```bash
  go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
  ```

  Or download a release binary from <https://github.com/tolvi-labs/tolvi/releases>.

- A Tolvi vault in your repo (`vault/.vault-meta.json`). Run `tolvi init` if you do not have one yet.

## Install

```bash
gemini extensions install https://github.com/tolvi-labs/tolvi --ref main
```

Gemini CLI then offers to install with git clone because this repository's releases do not carry the extension; answer yes. Without `--ref main`, Gemini CLI installs from the latest GitHub release, which does not contain the extension. The extension provides the Tolvi skill and loads `GEMINI-EXTENSION.md` as context in every session where it is active.

## Use

Ask Gemini about the vault in plain language. It uses the skill when a request matches its description, such as a question about decisions, sessions or patterns in a repo that has `vault/.vault-meta.json`.

- "What did we decide about Postgres?"
- "Write down that we chose PASETO over JWT."
- "Show me the most recent session log."

## Update

```bash
gemini extensions update tolvi
```

A git clone install updates from `main`.

## Uninstall

```bash
gemini extensions uninstall tolvi
```

## Troubleshooting

### Gemini does not use the skill

- Is the extension installed and enabled? Check with `gemini extensions list`.
- Start a new Gemini session after installing.
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

This install follows the Gemini CLI extension documentation at <https://geminicli.com/docs/extensions/reference/>, checked on 2026-09-16. It has not been run end to end with Gemini CLI. The install command with `--ref main` has not been run.

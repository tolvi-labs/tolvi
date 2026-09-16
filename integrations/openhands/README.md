# Tolvi: OpenHands integration

The shared Tolvi skill for [OpenHands](https://www.all-hands.dev). OpenHands supports the Agent Skills standard and loads repository skills from `.agents/skills/`, so installing the skill there teaches OpenHands how to read, write and ask questions of a Tolvi vault.

## Tier

OpenHands is a **Tier 2 (shared skill)** integration in the [Tolvi integrations tier list](../README.md). It gets the same `SKILL.md` as [Claude Code](../../skills/tolvi/), without the Claude Code slash commands or session hooks.

## Prerequisites

- [OpenHands](https://www.all-hands.dev) installed and configured.
- The `tolvi` CLI on your `PATH` in the environment where OpenHands runs:

  ```bash
  go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
  ```

  Or download a release binary from <https://github.com/tolvi-labs/tolvi/releases>.

- A Tolvi vault in your repo (`vault/.vault-meta.json`). Run `tolvi init` if you do not have one yet.

## Install

Clone the tolvi repo once, then run the installer from anywhere inside your project:

```bash
git clone https://github.com/tolvi-labs/tolvi /path/to/tolvi
cd /path/to/your-project
bash /path/to/tolvi/skills/tolvi/install.sh --agents
```

The installer finds your repository root and copies the skill to `.agents/skills/tolvi/SKILL.md`. Commit that directory so OpenHands finds it in every workspace built from the repo.

To install it for yourself only, in every project:

```bash
bash /path/to/tolvi/skills/tolvi/install.sh --agents --path ~/.agents/skills
```

If you used the earlier Tolvi `.openhands_instructions` file, delete it from your repo root. It is no longer maintained.

## Use

Ask OpenHands about the vault in plain language. It uses the skill when a request matches its description, such as a question about decisions, sessions or patterns in a repo that has `vault/.vault-meta.json`.

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

### OpenHands does not use the skill

- Is the file at `.agents/skills/tolvi/SKILL.md` under your repository root, and is it committed? Check with `git ls-files .agents/skills/tolvi/`.
- Start a new conversation, since OpenHands loads repository skills when a conversation starts.

### `tolvi: command not found`

The CLI is not on the `PATH` of the environment OpenHands runs in. Without it the skill falls back to reading `vault/` directly, and you lose `tolvi ask`.

### `tolvi ask` errors about `ANTHROPIC_API_KEY`

Set the `ANTHROPIC_API_KEY` environment variable, or write `~/.config/tolvi/config.yaml`:

```yaml
anthropic_api_key: sk-ant-...
model: claude-sonnet-4-7
```

See the [CLI README](../../cli/) for the full config reference.

## Verification

This install follows the OpenHands Skills documentation at <https://docs.openhands.dev/overview/skills>, checked on 2026-09-16. It has not been run end to end with OpenHands.

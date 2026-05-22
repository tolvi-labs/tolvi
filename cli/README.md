# Tolvi CLI

> **Status:** Phase 3 shipped. Single static Go binary, cross-platform (darwin / linux / windows × amd64 / arm64). Track upcoming work in [`ROADMAP.md`](../ROADMAP.md).

## Install

```bash
go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
```

Or download a pre-built binary from <https://github.com/tolvi-labs/tolvi/releases>.

Set your Anthropic API key via the `ANTHROPIC_API_KEY` env var or write `~/.config/tolvi/config.yaml`:

```yaml
anthropic_api_key: sk-ant-...
model: claude-sonnet-4-7
```

## Commands

### `tolvi init`

Provision `vault/{decisions,sessions,patterns}/` and `vault/.vault-meta.json` in the current repo. Workspace name derived from `git remote get-url origin` or cwd basename; override with `--workspace <name>`.

### `tolvi sync <type> <title>`

Create a new vault doc (`type` = `decision | session | pattern`). Opens `$EDITOR` for the body by default; `--body "..."` skips the editor. Sessions on a day that already has a file get a new block appended.

Key flags: `--slug`, `--status`, `--body`, `--no-edit`, `--print-path`, `--vault`.

### `tolvi ask <query>`

Stream a cited answer to the query. CAG-based: whole vault → Anthropic context (with prompt caching). Prints a `Sources` footer with verified `[[slug]]` references and file paths after the stream completes.

Key flags: `--no-stream`, `--json`, `--model`, `--include-status`, `--exclude-type`, `--vault`.

### Pre-commit hook

Tolvi includes an optional git pre-commit hook that prints a non-blocking nudge when staged changes match decision-likely patterns (dependency manifests, infra config, tooling config, or large diffs).

Install in any Tolvi-enabled repo:

```bash
tolvi precommit install
```

The hook is a 4-line shell shim written to `.git/hooks/pre-commit`. It:

- Always exits 0 — never blocks a commit
- Silently degrades to a no-op if the `tolvi` binary is removed from `$PATH`
- Honors `TOLVI_PRECOMMIT_QUIET=1` to silence the nudge per-shell

Flags for `tolvi precommit install`:

- `--force` — overwrite an existing non-tolvi hook
- `--append` — chain after an existing hook (preserves the previous content)
- `--repo <path>` — install into a specific repo's `.git/hooks/`

Remove with `tolvi precommit uninstall`. Refuses to remove a non-tolvi hook unless `--force`.

### `tolvi version`

Prints the binary version (baked at release time via `-ldflags`).

## See also

- [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) — the local-arm component in context
- [`../integrations/`](../integrations/) — per-agent integration files (Claude Code skill, Cursor `.cursorrules`, Aider/OpenHands/Continue conventions)

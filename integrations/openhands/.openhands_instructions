# Tolvi conventions

This project may use Tolvi: a per-repo engineering knowledge vault of decisions, sessions, and patterns stored as Markdown with YAML frontmatter under `<repo>/vault/`. When the user asks about decisions, sessions, patterns, or references `[[some-slug]]`, treat this as a Tolvi vault query and apply the conventions below.

If `<repo>/vault/.vault-meta.json` does not exist in this project, Tolvi is not active here — ignore these conventions and use normal tools.

## Vault structure

`<repo>/vault/{decisions,sessions,patterns}/*.md` plus `<repo>/vault/.vault-meta.json`. One workspace per repo.

## Format spec

### Frontmatter per doc type

- **decision** — `tags: [decision]`, `date: YYYY-MM-DD`, `repo: <name>`, `status: <enum>`. Optional: `ticket`, `user_impact`, `product_area`.
- **session** — `tags: [session]`, `date: YYYY-MM-DD`, `status: <enum>`.
- **pattern** — `tags: [pattern]`, `status: <enum>`. No date, no repo (patterns are timeless).

### Slug rules

Regex: `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`. Lowercase letters and digits only; hyphens in the middle only; no leading/trailing hyphen; no underscores; no uppercase; no spaces. Max 80 chars. Auto-derive from title via NFKD ASCII-fold + lowercase + non-alphanumeric runs to `-`.

### File paths

- decision: `vault/decisions/<date>-<slug>.md`
- session: `vault/sessions/<date>.md` (one file per day; new blocks append)
- pattern: `vault/patterns/<slug>.md` (no date prefix)

### Status enum

`active | in-progress | superseded | deprecated | draft | historical`. Default search surfaces `active | in-progress | historical`; the other three are filtered out unless requested.

### Wiki-link syntax

- `[[slug]]` — references a doc by slug in the same vault. Must resolve to a real file.
- `[[repo:slug]]` — cross-vault reference; format spec supports it but the CLI v1 does not surface it. Do not generate `[[repo:slug]]` citations.

## CLI

The `tolvi` binary is the substrate. Prefer the CLI over direct file operations when available.

- `tolvi ask <query>` — stream a cited answer; prints a `Sources` footer with verified `[[slug]]` references.
- `tolvi sync <type> <title>` — write a new vault doc with atomic write + frontmatter validation + slug auto-gen. Sessions on a day that already has a file get a new block appended.
- `tolvi init` — provision `vault/` with the three subdirs and `.vault-meta.json`.
- `tolvi precommit install` — adds a non-blocking git pre-commit nudge that flags commits touching dependency manifests, infra config, or large diffs. Suggests `tolvi sync decision "..."` when fired.

Cite vault content with `[[slug]]`. Use exact slugs that exist in the vault — verify before citing. Don't invent slugs.

## Escape hatches

- **CLI not in `$PATH`**: suggest `go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest` or a release binary from <https://github.com/tolvi-labs/tolvi/releases>. Fall back to direct file operations.
- **No vault in repo**: offer `tolvi init` with confirmation before running.
- **`tolvi ask` reports vault too large**: suggest `--exclude-type session` or migrate to the server arm.
- **No `ANTHROPIC_API_KEY`**: point the user at `~/.config/tolvi/config.yaml` or the env var.

For the full skill content (with worked examples and richer behavioral rules), see the canonical Claude Code skill at <https://github.com/tolvi-labs/tolvi/blob/main/integrations/claude-code/SKILL.md>.

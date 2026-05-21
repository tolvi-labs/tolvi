---
name: tolvi
description: Read, write, and ask questions of a Tolvi engineering vault. Use when working in a repo with vault/.vault-meta.json, or when the user mentions decisions, sessions, or patterns.
---

# Tolvi

Tolvi is a per-repo engineering knowledge vault — decisions, sessions, and patterns stored as Markdown with YAML frontmatter under `<repo>/vault/`. This skill teaches Claude Code how to read, write, and query a Tolvi vault, and when to use the `tolvi` CLI versus direct file operations.

When you (Claude) finish loading this skill, briefly acknowledge it (one sentence) and wait for the user's actual request. Do not auto-scan the vault or auto-invoke any CLI command.

## When to use this skill

**Triggers:**

- Explicit: user types `/tolvi`
- Implicit: user asks "what did we decide about X", "what session notes do we have on Y", "write a decision about Z", or references `[[some-slug]]` in a request
- Repo state: cwd or an ancestor contains `vault/.vault-meta.json`

**Anti-triggers:**

- The repo has no `vault/.vault-meta.json` — offer `tolvi init` instead, with confirmation before running
- The user is asking about *code*, not *project knowledge* — use Read/Grep tools normally

## Vault structure

```
<repo>/
└── vault/
    ├── .vault-meta.json      # workspace + embedding model + schema version
    ├── decisions/            # YYYY-MM-DD-<slug>.md
    ├── sessions/             # YYYY-MM-DD.md (one per day; multiple blocks append)
    └── patterns/             # <slug>.md (no date prefix; timeless)
```

One workspace per repo. `.vault-meta.json` is the marker file the CLI uses to discover the vault (walk-up from `$PWD` to filesystem root or `$HOME`).

## Format spec — compact reference

### Frontmatter shape per doc type

**Decision** — required fields:

```yaml
---
tags: [decision]
date: 2026-04-12
repo: my-project
status: active
---
```

Optional decision fields: `ticket`, `user_impact`, `product_area`.

**Session** — required fields:

```yaml
---
tags: [session]
date: 2026-04-12
status: active
---
```

**Pattern** — required fields (no date, no repo; patterns are timeless):

```yaml
---
tags: [pattern]
status: active
---
```

### Slug rules

Slug regex: `^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`

- Lowercase letters and digits only; hyphens allowed in the middle
- No leading hyphen, no trailing hyphen
- No underscores, no uppercase, no spaces
- Max 80 chars
- Single-character slugs (`a`, `1`) are allowed

Auto-derivation from a title: NFKD ASCII-fold → lowercase → non-alphanumeric runs replaced with `-` → trim leading/trailing `-` → cap at 80 → re-trim. The CLI does this automatically in `tolvi sync`.

### File path layout

- **Decision**: `vault/decisions/<date>-<slug>.md` (date is YYYY-MM-DD)
- **Session**: `vault/sessions/<date>.md` (one file per day; multiple session blocks append to the same file)
- **Pattern**: `vault/patterns/<slug>.md` (no date prefix)

### Status enum (six values)

| Status | When to use |
|---|---|
| `active` | Current, in effect |
| `in-progress` | Decided but implementation in flight |
| `superseded` | Replaced by a newer decision (point to the replacement in the body) |
| `deprecated` | No longer applies; kept for history |
| `draft` | Not finalized; ideas in progress |
| `historical` | Older context; valid but archival |

Search defaults to `active | in-progress | historical`. The other three are filtered out unless the caller passes `--include-status` explicitly.

### Wiki-link syntax

- `[[slug]]` — references a doc by slug in the same vault. Always resolvable to a file when the slug exists.
- `[[repo:slug]]` — cross-vault reference (the format spec supports it; the CLI v1 does not surface it in citations). Do not generate `[[repo:slug]]` citations in v1.

## CLI commands

The `tolvi` binary is the substrate. Shell out via the Bash tool. If `tolvi` is not in `$PATH`, suggest installation (see Escape hatches below).

### `tolvi ask <query>`

Stream a cited answer to the query. Tokens print to stdout as they arrive; a `Sources` footer prints after the stream completes with verified `[[slug]]` references, file paths, and statuses.

Key flags:

- `--no-stream` — buffer output instead of streaming (use in CI logs or environments that mangle ANSI)
- `--json` — emit structured JSON (`{answer, citations, model, tokens, ...}`) instead of human-readable text; buffered, no streaming
- `--model <name>` — override the configured Anthropic model
- `--include-status <csv>` — override the default status filter (e.g., `--include-status active,superseded`)
- `--exclude-type <csv>` — omit a doc type (e.g., `--exclude-type session` when the vault is too large)
- `--vault <path>` — override walk-up discovery

### `tolvi sync <type> <title>`

`type` is one of `decision | session | pattern`. Writes a new vault doc with auto-derived slug and valid frontmatter. Opens `$EDITOR` for the body by default.

Key flags:

- `--body "..."` — pass the body inline; skips the `$EDITOR` flow
- `--no-edit` — write a skeleton-only file (no body capture)
- `--slug <name>` — override the auto-derived slug
- `--status <value>` — frontmatter status (default: `active`)
- `--print-path` — output only the resulting path to stdout (useful for piping)
- `--vault <path>` — override walk-up discovery

**Session same-day behavior:** if `vault/sessions/<date>.md` already exists, `tolvi sync session` appends a new session block to the existing file rather than refusing or overwriting. The new block uses an HH:MM-prefixed `## [HH:MM] Session — ...` heading.

### `tolvi init`

One-time provision: creates `vault/decisions/`, `vault/sessions/`, `vault/patterns/`, and `vault/.vault-meta.json`. Refuses if the vault already exists. Workspace name derived from `git remote get-url origin` parse, falling back to `basename $PWD`. Override with `--workspace <name>`.

## Behavioral rules

These are *preferences*, not hard gates. Use judgment.

### Reading a specific vault doc → use the Read tool

For "show me the postgres decision" or any exact-doc request, read the file directly. Direct read is faster than shelling out for one file.

### Searching the vault → prefer `tolvi ask`

For semantic queries ("what did we decide about X", "any patterns for Y"), shell out to `tolvi ask`. The CLI handles CAG (whole vault → Anthropic context) plus citation verification plus streaming. Don't reimplement with `grep` unless the CLI is unavailable.

### Writing a new doc → prefer `tolvi sync`

The CLI handles atomic write, frontmatter validation against the embedded JSON Schema, slug auto-generation, and session same-day append. Composing markdown + Write tool directly is acceptable when:

- The CLI is missing from `$PATH`
- The user explicitly asks for a direct Write
- You're iterating on a draft the user has not yet committed to capturing

When using direct Write, validate the frontmatter mentally against the rules above before writing. Frontmatter-validation failures cause silent vault corruption that's annoying to debug.

### Citing vault content

When summarizing or quoting vault content in a response, cite with `[[slug]]`. Use exact slugs that exist in the vault — verify by reading the matching file before citing. Don't invent slugs.

### Pre-commit nudges live elsewhere

A separate `tolvi precommit` git-hook subcommand handles proactive "you may have made a decision worth capturing" prompts at commit time. It is not part of this skill. If the user asks "should I capture this as a decision?" mid-session, you may suggest `tolvi sync` directly; don't try to invoke the precommit hook from within Claude Code.

## Escape hatches

### `tolvi` binary not in `$PATH`

Suggest installation:

```bash
go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
```

Or download from <https://github.com/tolvi-labs/tolvi/releases>. Until the user installs it, fall back to direct file operations with manual frontmatter validation.

### No vault in repo

Offer to run `tolvi init`. If the user proposes a workspace name, pass it via `--workspace <name>`. Confirm before running — don't auto-init silently.

### `tolvi ask` reports vault too large

The CLI errors at approximately 180,000 estimated tokens. Suggest one of:

- `--exclude-type session` — drops noisy session logs from the prompt
- `--include-status active` — already the default; mention only if the user widened it
- Migrate to the server arm — a separate `tolvi-server` deployment outside the scope of this skill

### No `ANTHROPIC_API_KEY`

`tolvi ask` errors with a clear pointer to `~/.config/tolvi/config.yaml` and the `ANTHROPIC_API_KEY` environment variable. Surface that message verbatim to the user rather than paraphrasing.

### Vault has invalid frontmatter in a file

`tolvi ask` prints a warning on stderr and skips the invalid file. Suggest the user inspect the file via Read; offer to help fix the frontmatter against the rules above. The vault remains queryable for the other files.

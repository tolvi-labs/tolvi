# AGENTS.md

Guidance for coding agents working in this repo. Humans: see [`CONTRIBUTING.md`](./CONTRIBUTING.md).

## What this is

Tolvi is a git-native engineering vault: decisions, sessions and patterns as plain Markdown that agents read as context. Two arms share one contract. The Go CLI in `cli/` is the local arm and holds the whole vault in context; the TypeScript server in `server/` is the hosted arm and uses retrieval. Neither shares a parsing library with the other, on purpose: the spec is the only thing crossing the boundary.

## Build and test

```bash
npm ci                      # once, for the validation suite
npm run validate            # schemas, parity, brand, installers, preflight, prose, markdown
cd cli && go build ./... && go test ./...
cd server && npm ci && npm test
```

Run `npm run validate` before proposing any change. It is cheap and it catches the things review does not.

## The rule that matters most here

**A skipped check is not a passing check.** `validate:brand` needs a `BRAND_BLOCKLIST` value it does not ship, so locally it prints a loud banner and exits 0 having examined nothing. A green `npm run validate` is therefore not evidence this repo is clean; CI is. Never report a check as passing when it did not run. The same goes for any command whose failure is silent: confirm what it did, do not infer from its quiet.

## Conventions

- **The format spec is normative.** `spec/tolvi-format-v2.md` and the schemas in `spec/schemas/` win over any prose, including this file. v1 is frozen beside it.
- **The schemas exist three times**: `spec/schemas/` is the source, `cli/internal/format/schemas/` is embedded because Go cannot consume npm, and `@tolvi-labs/spec` vendors them for everyone else. `npm run validate:schema-parity` pins the first two. Edit `spec/schemas/` and copy.
- **One resolver owns routing.** `cli/internal/vault` decides where a doc lives. Write, read, the hooks and the skills all call it. If you find yourself deriving a vault path anywhere else, that is the bug.
- **Tests first.** Every guard in `.github/scripts/` was written by breaking it first and watching it fail. Do the same for new ones.
- **Prose:** no em dashes in README prose (`validate:prose` enforces it), and never a line break mid-sentence.

## What not to do

- Do not commit anything under `docs/superpowers/`. Plans and specs are pre-code artifacts; they are gitignored and must stay that way.
- Do not add a dependency field to `.claude-plugin/plugin.json` for the `tolvi` CLI. No such field exists, and an invented one reads as enforcement while doing nothing. The preflight block in `skills/tolvi/commands/_preflight.md` is the real mechanism.
- Do not name the parent company anywhere but `NOTICE`.

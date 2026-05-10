# Phase 0+1 Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bootstrap the `tolvi-labs/tolvi` monorepo to a state where Phase 2 (server) and Phase 3 (CLI) can begin in parallel against a locked, public format contract — with brand-isolation enforcement baked in from PR #1.

**Architecture:** Single foundational PR on a feature branch (`feat/foundation`) opened against `main`. Eleven atomic commits, each independently meaningful. No Go or TS code in this phase — only repo skeleton, contributor docs, format spec (`tolvi-format-v1`), JSON schemas, synthetic sample vault, and CI guards.

**Tech Stack:** Markdown + YAML + JSON Schema (Draft 2020-12). CI uses GitHub Actions, `markdownlint-cli2`, `ajv-cli`, `lychee`, plus a small Node helper and bash script. Pre-commit framework for local fast-feedback.

**Companion spec:** `docs/superpowers/specs/2026-05-09-phase-0-1-foundation-design.md` — read this before starting. The plan references spec sections by number for content; the spec is the canonical source of truth for what each artifact contains.

**Brand-isolation rule (critical):** Outside the single `NOTICE` file, no artifact may reference parent organizations, sibling products, or specific deployments. The CI guard added in Task 10 enforces this. See spec section 13 for the rule and `~/.claude/projects/-Users-alantorres-tolvi-labs-tolvi/memory/feedback_tolvi_brand_isolation.md` for the full rationale.

**Pre-execution setup:** Before starting Task 1, the implementing agent should:
1. Verify cwd is `/Users/alantorres/tolvi-labs/tolvi`
2. Verify on `main` with a clean working tree (apart from untracked `tolvi-build-plan.md` and the docs/superpowers/ tree which has been committed alongside this plan to main)
3. Create the feature branch: `git checkout -b feat/foundation`
4. Confirm Node 20+ and `npx` work (used for ajv-cli validation)

---

## File Structure (full inventory of artifacts created across all 11 commits)

```
tolvi/
├── LICENSE                                            (commit 1)
├── NOTICE                                             (commit 1)
├── .gitignore                                         (commit 1)
├── .editorconfig                                      (commit 1)
├── README.md                                          (commit 2 — overwrites existing single-line file)
├── CONTRIBUTING.md                                    (commit 2)
├── CODE_OF_CONDUCT.md                                 (commit 2)
├── SECURITY.md                                        (commit 2)
├── NAMESPACE.md                                       (commit 2)
├── ROADMAP.md                                         (commit 2)
├── CODEOWNERS                                         (commit 3)
├── .pre-commit-config.yaml                            (commit 11)
│
├── .github/
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md                              (commit 3)
│   │   ├── feature_request.md                         (commit 3)
│   │   └── config.yml                                 (commit 3)
│   ├── PULL_REQUEST_TEMPLATE.md                       (commit 3)
│   ├── workflows/
│   │   └── validate.yml                               (commit 10)
│   └── scripts/
│       ├── extract-frontmatter.js                     (commit 10)
│       └── brand-isolation-check.sh                   (commit 10)
│
├── cli/README.md                                      (commit 4 — stub for Phase 3)
├── server/README.md                                   (commit 4 — stub for Phase 2)
├── spec/
│   ├── README.md                                      (commit 4)
│   ├── tolvi-format-v1.md                             (commit 7)
│   └── schemas/
│       ├── vault-meta.json                            (commit 8)
│       ├── decision.json                              (commit 8)
│       ├── session.json                               (commit 8)
│       └── pattern.json                               (commit 8)
├── skills/README.md                                   (commit 4 — stub for Phase 4)
│
├── docs/
│   ├── ARCHITECTURE.md                                (commit 5)
│   ├── CONVENTIONS.md                                 (commit 5)
│   ├── OPEN_QUESTIONS.md                              (commit 5)
│   └── adr/
│       ├── README.md                                  (commit 6)
│       ├── 0001-architecture-overview.md              (commit 6)
│       └── 0002-vault-format-v1-contract.md           (commit 6)
│
└── examples/
    └── sample-vault/
        ├── .vault-meta.json                           (commit 9)
        ├── decisions/                                 (commit 9 — 6 files)
        ├── sessions/                                  (commit 9 — 3 files)
        └── patterns/                                  (commit 9 — 4 files)
```

`docs/superpowers/specs/2026-05-09-phase-0-1-foundation-design.md` and `docs/superpowers/plans/2026-05-09-phase-0-1-foundation.md` are committed to `main` alongside this plan, before the foundational PR work begins.

---

## Task 1 — LICENSE, NOTICE, .gitignore, .editorconfig

**Files:**
- Create: `LICENSE`
- Create: `NOTICE`
- Create: `.gitignore`
- Create: `.editorconfig`

- [ ] **Step 1: Fetch the canonical Apache 2.0 license text**

```bash
curl -fsSL https://www.apache.org/licenses/LICENSE-2.0.txt -o LICENSE
```

Expected: file written, exits 0. Verify size:

```bash
wc -c LICENSE
```

Expected: ~11,357 bytes.

- [ ] **Step 2: Verify LICENSE content is the Apache 2.0 text**

```bash
grep -q "Apache License" LICENSE && grep -q "Version 2.0" LICENSE && echo OK
```

Expected: prints `OK`.

- [ ] **Step 3: Create NOTICE with the verbatim attribution text**

Write to `NOTICE` (exact bytes — no trailing whitespace, single trailing newline):

```
Tolvi
Copyright 2026 Torres Atlantic, LLC

This product is developed by Tolvi Labs, the open-source arm of
Torres Atlantic, LLC. Licensed under the Apache License, Version 2.0.
```

- [ ] **Step 4: Verify NOTICE content matches exactly**

```bash
diff <(cat NOTICE) <(cat <<'EOF'
Tolvi
Copyright 2026 Torres Atlantic, LLC

This product is developed by Tolvi Labs, the open-source arm of
Torres Atlantic, LLC. Licensed under the Apache License, Version 2.0.
EOF
)
```

Expected: no output (files match), exit 0.

- [ ] **Step 5: Create .gitignore**

Write to `.gitignore`:

```gitignore
# OS junk
.DS_Store
Thumbs.db

# Editors / IDEs
.vscode/
.idea/
*.swp
*.swo

# Node / npm
node_modules/
npm-debug.log*

# Go
*.test
*.out
/cli/dist/
/cli/bin/

# Tolvi local caches
.tolvi/cache/
.tolvi-index/
/examples/sample-vault/.tolvi-index/

# Private working docs (not for the public repo)
tolvi-build-plan.md
```

- [ ] **Step 6: Create .editorconfig**

Write to `.editorconfig`:

```editorconfig
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 2

[*.go]
indent_style = tab
indent_size = 4

[Makefile]
indent_style = tab

[*.md]
trim_trailing_whitespace = false
```

- [ ] **Step 7: Verify build plan is gitignored**

```bash
git check-ignore tolvi-build-plan.md
```

Expected: prints `tolvi-build-plan.md` (file is correctly ignored), exit 0.

- [ ] **Step 8: Commit**

```bash
git add LICENSE NOTICE .gitignore .editorconfig
git status
```

Expected: only those four files staged. Confirm `tolvi-build-plan.md` is NOT in the staged list.

```bash
git commit -m "chore: add LICENSE (Apache 2.0), NOTICE, .gitignore, .editorconfig"
```

---

## Task 2 — Top-level project docs (README, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, NAMESPACE, ROADMAP)

**Files:**
- Modify: `README.md` (currently `# tolvi` placeholder; overwrite)
- Create: `CONTRIBUTING.md`
- Create: `CODE_OF_CONDUCT.md`
- Create: `SECURITY.md`
- Create: `NAMESPACE.md`
- Create: `ROADMAP.md`

- [ ] **Step 1: Overwrite README.md with placeholder content**

```markdown
# Tolvi

Capture engineering decisions where they happen. Query them later in plain English.

> **Status:** Pre-release. The repo skeleton, format spec, and CI guards are in place; CLI and server land in subsequent phases. See [`ROADMAP.md`](./ROADMAP.md) for the public phase plan.

## What's here today

- [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md) — the public vault format contract
- [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md) — system architecture
- [`docs/CONVENTIONS.md`](./docs/CONVENTIONS.md) — vault content conventions
- [`examples/sample-vault/`](./examples/sample-vault/) — synthetic vault that validates against the spec

## What's coming

- `tolvi` CLI (Go, single binary) — `init`, `sync`, `recall`, `ask`, `doctor`, `publish`
- `tolvi` server (TypeScript, Fastify, Postgres + pgvector) — multi-tenant, self-hostable
- Agent integrations (Claude Code, Cursor, others)
- TypeScript and Python SDKs

## License

[Apache 2.0](./LICENSE). See [`NOTICE`](./NOTICE) for attribution.

## Contributing

See [`CONTRIBUTING.md`](./CONTRIBUTING.md). External contributors should read the **Brand isolation** section before opening a PR.
```

- [ ] **Step 2: Create CONTRIBUTING.md**

Write to `CONTRIBUTING.md`:

````markdown
# Contributing to Tolvi

Thanks for your interest. Tolvi is an open-source project under [Apache 2.0](./LICENSE).

## Getting started

1. Fork the repo
2. Create a feature branch (`feat/<short-slug>`, `fix/<short-slug>`, or `docs/<short-slug>`)
3. Make your changes
4. Run pre-commit checks locally (see below)
5. Open a PR against `main`

## Brand isolation

All content in this repo is brand-neutral. The single exception is the [`NOTICE`](./NOTICE) file, which contains the project's attribution.

**Do not reference** parent organizations, sibling products, or specific deployments anywhere else — including code, docs, examples, sample content, code comments, commit messages, branch names, or PR descriptions.

The CI guard at `.github/scripts/brand-isolation-check.sh` enforces this. PRs that introduce forbidden references will fail CI and be rejected. The same script runs as a pre-commit hook so you can catch leaks locally.

If you are unsure whether content is brand-neutral, ask in your PR description and a maintainer will help.

## PR size

PRs should normally stay under **500 lines of diff** and be reviewable in **under 30 minutes**. If your work is bigger, split it into a stack of smaller PRs that can be reviewed and merged independently.

**Bootstrapping exception:** PRs that establish initial repo scaffolding, generated artifacts (sample vaults, JSON schemas, generated specs), or other irreducibly-large foundational content are exempt and should be split by *type of change* (atomic commits within one PR) rather than by line count.

## Local development

### Prerequisites

- Node 20+ (for CI scripts and validation)
- [`pre-commit`](https://pre-commit.com/) (`brew install pre-commit` or `pipx install pre-commit`)

### Set up pre-commit hooks

```bash
pre-commit install
```

This installs the hooks from `.pre-commit-config.yaml`. They run on every `git commit` and catch:
- Trailing whitespace, missing EOF newlines
- Invalid JSON
- Required-frontmatter-fields violations in vault content
- Brand-isolation violations

### Run all checks manually

```bash
pre-commit run --all-files
```

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <subject>

[optional body]
```

Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `ci`, `build`.

## Code of conduct

By participating, you agree to abide by the [Code of Conduct](./CODE_OF_CONDUCT.md).

## Reporting security issues

See [`SECURITY.md`](./SECURITY.md). Do not file public issues for security vulnerabilities.
````

- [ ] **Step 3: Fetch the Contributor Covenant 2.1 text**

```bash
curl -fsSL https://www.contributor-covenant.org/version/2/1/code_of_conduct/code_of_conduct.md -o CODE_OF_CONDUCT.md
```

Expected: file written, exits 0. Verify:

```bash
grep -q "Contributor Covenant" CODE_OF_CONDUCT.md && grep -q "version 2.1" CODE_OF_CONDUCT.md && echo OK
```

Expected: prints `OK`.

- [ ] **Step 4: Edit CODE_OF_CONDUCT.md to set the contact line**

The official template has `[INSERT CONTACT METHOD]` as a placeholder. Replace with a placeholder that is brand-neutral (final email lands once `tolvi.dev` is registered):

```bash
sed -i.bak 's|\[INSERT CONTACT METHOD\]|opening a private security advisory in this repository (until `conduct@tolvi.dev` is live)|g' CODE_OF_CONDUCT.md && rm CODE_OF_CONDUCT.md.bak
```

Verify:

```bash
grep -q "INSERT CONTACT METHOD" CODE_OF_CONDUCT.md && echo "FAIL: placeholder still present" || echo "OK: placeholder replaced"
```

Expected: prints `OK: placeholder replaced`.

- [ ] **Step 5: Create SECURITY.md**

Write to `SECURITY.md`:

```markdown
# Security policy

## Reporting a vulnerability

If you discover a security vulnerability in Tolvi, please **do not** open a public GitHub issue.

Instead, report it privately by [opening a security advisory](https://github.com/tolvi-labs/tolvi/security/advisories/new) in this repository. We will acknowledge within 48 hours.

Once `tolvi.dev` is registered, security reports may also be sent to `security@tolvi.dev`.

## Scope

The following are in scope:

- Vulnerabilities in code published in this repository
- Vulnerabilities in the published `tolvi` CLI binary, the `tolvi-server` Docker image, or the `@tolvi-labs/sdk` npm package
- Vulnerabilities that allow unauthorized read or write access to vault content via the `tolvi` server API

The following are out of scope (please file as regular issues):

- Theoretical vulnerabilities without proof-of-concept
- Issues in third-party dependencies (please report upstream)
- Issues in the user's own deployment configuration

## Supported versions

| Version | Supported |
|---|---|
| pre-1.0 | Best-effort; security fixes will land on `main` |

A formal supported-versions policy will be published when 1.0 ships.
```

- [ ] **Step 6: Create NAMESPACE.md**

Write to `NAMESPACE.md`:

```markdown
# Namespace pre-flight tracker

This file tracks the Day 0 namespace claims required before public launch. Status is updated as items are claimed.

| # | Item | Status | Timestamp | Notes |
|---|---|---|---|---|
| 1 | GitHub org `tolvi-labs` | ✅ Claimed | 2026-05-09 | — |
| 2 | `tolvi.dev` domain | ☐ TODO | — | Register at registrar of choice |
| 3 | `tolvi.com` domain | ☐ TODO | — | Check availability; acquire if reasonable |
| 4 | `tolvilabs.com` / `tolvilabs.dev` | ☐ TODO | — | Secondary, for the Labs brand |
| 5 | PyPI placeholder `tolvi==0.0.0` | ☐ TODO | — | Publish before public launch |
| 6 | npm placeholder `tolvi@0.0.0` | ☐ TODO | — | Publish before public launch |
| 7 | USPTO TESS trademark search | ☐ TODO | — | Software/SaaS classes; "Tolvi" + "Tolvi Labs" |
| 8 | Twitter/X handle `@tolvi` or `@tolvilabs` | ☐ TODO | — | Try both |
| 9 | LinkedIn company page Tolvi Labs | ☐ TODO | — | Reserve, don't activate |

To update: edit the row in place, change ☐ to ✅, fill in the timestamp, commit.
```

- [ ] **Step 7: Create ROADMAP.md**

Write to `ROADMAP.md`:

```markdown
# Roadmap

The public roadmap for Tolvi v1. Internal sequencing details are tracked privately.

## Status legend

- ✅ Shipped
- 🚧 In progress
- ⏭️ Next up
- 📅 Planned
- 💤 Deferred

## Phases

### Phase 0+1 — Foundation 🚧

- Repo skeleton (LICENSE, NOTICE, contributor docs)
- Architecture spec, conventions spec, ADR setup
- `tolvi-format-v1` public format contract + JSON schemas
- Synthetic sample vault for demos and CI validation
- CI guards: markdown lint, schema validation, brand isolation, link checker

### Phase 2 — Server core ⏭️

- TypeScript + Fastify on Node 20
- Postgres + pgvector for the index
- Multi-tenant API key auth at the workspace level
- Reversible Drizzle migrations
- Docker compose for self-host

### Phase 3 — Tolvi CLI ⏭️

- Go, single static binary, cross-platform
- `tolvi init`, `sync`, `recall`, `ask`, `doctor`, `unify`, `publish`, `status`
- Embedded sqlite-vec local index
- Ollama embedding model by default

### Phase 4 — Agent integrations 📅

- Claude Code skill files
- Cursor `.cursorrules` template
- Aider, OpenHands, Continue (skeleton integrations)

### Phase 5 — TypeScript SDK and docs 📅

- `@tolvi-labs/sdk` on npm
- Documentation site at `tolvi.dev`
- API reference auto-generated from OpenAPI
- Migration guides

### Phase 6 — Soft launch 📅

- Apache 2.0 release on GitHub
- Homebrew tap, npm/PyPI publishes, Docker Hub image
- 5–15 friendly users from the maintainer's network
- Daily metric tracking

### Phase 7 — Iterate 📅

- Whatever the first wave of users hits

### Phase 8 — Public launch 📅

- Show HN with a strong title biased toward concrete pain
- Posts in relevant developer communities

### Phase 9 — Commercial layer 💤

Triggered by adoption signal — see the README for details on how the OSS and any future hosted offering coexist.

## Versioning

The vault format is independently versioned (`tolvi-format-v1`, `tolvi-format-v2`, …). Format-version compatibility is documented in [`spec/tolvi-format-v1.md`](./spec/tolvi-format-v1.md).
```

- [ ] **Step 8: Verify all six files exist and contain no forbidden references**

```bash
ls -1 README.md CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md NAMESPACE.md ROADMAP.md
```

Expected: all six files listed, no errors.

```bash
grep -iE 'corvin|firebase|firestore|cloud run|gs://|healthcare|hipaa' README.md CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md NAMESPACE.md ROADMAP.md && echo "FAIL: forbidden term found" || echo "OK: clean"
```

Expected: prints `OK: clean`.

- [ ] **Step 9: Commit**

```bash
git add README.md CONTRIBUTING.md CODE_OF_CONDUCT.md SECURITY.md NAMESPACE.md ROADMAP.md
git commit -m "docs: add top-level README, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, NAMESPACE, ROADMAP"
```

---

## Task 3 — `.github/` issue + PR templates and CODEOWNERS

**Files:**
- Create: `.github/ISSUE_TEMPLATE/bug_report.md`
- Create: `.github/ISSUE_TEMPLATE/feature_request.md`
- Create: `.github/ISSUE_TEMPLATE/config.yml`
- Create: `.github/PULL_REQUEST_TEMPLATE.md`
- Create: `CODEOWNERS`

- [ ] **Step 1: Create the directory**

```bash
mkdir -p .github/ISSUE_TEMPLATE
```

- [ ] **Step 2: Create bug_report.md**

Write to `.github/ISSUE_TEMPLATE/bug_report.md`:

```markdown
---
name: Bug report
about: Report a defect in Tolvi
title: ''
labels: bug
assignees: ''
---

### What happened

A clear, minimal description of the bug.

### Expected behavior

What you expected instead.

### Reproduction

Steps to reproduce. A minimal example is most useful.

```bash
# the command(s) you ran
```

### Environment

- Tolvi version: (run `tolvi --version`)
- OS:
- Embedding backend (Ollama / OpenAI / other):

### Logs

```
Paste relevant logs here.
```
```

- [ ] **Step 3: Create feature_request.md**

Write to `.github/ISSUE_TEMPLATE/feature_request.md`:

```markdown
---
name: Feature request
about: Suggest a feature or improvement
title: ''
labels: enhancement
assignees: ''
---

### What problem does this solve

The user pain or workflow gap. Concrete is better than abstract.

### Proposed solution

What you'd like Tolvi to do. Mention any existing flags / commands you're modeling on.

### Alternatives considered

Other approaches you thought about and why this one is preferred.

### Additional context

Links to related discussions, prior art, or related ROADMAP items.
```

- [ ] **Step 4: Create config.yml**

Write to `.github/ISSUE_TEMPLATE/config.yml`:

```yaml
blank_issues_enabled: false
contact_links:
  - name: Security vulnerability
    url: https://github.com/tolvi-labs/tolvi/security/advisories/new
    about: Report security issues privately, not via public issues. See SECURITY.md.
```

- [ ] **Step 5: Create PULL_REQUEST_TEMPLATE.md**

Write to `.github/PULL_REQUEST_TEMPLATE.md`:

```markdown
## What this PR does

A short summary. One sentence is best.

## Why

The user-facing or maintainer-facing motivation. Link to the issue or roadmap item.

## How

Implementation highlights. Mention any non-obvious decisions.

## Testing

How was this verified? What new tests were added?

## Checklist

- [ ] PR is under 500 lines of diff (or is exempt per CONTRIBUTING.md "PR size" rules)
- [ ] All new code has tests
- [ ] `pre-commit run --all-files` passes locally
- [ ] No forbidden brand references introduced (CI will verify)
- [ ] Updated relevant docs (README, ROADMAP, ADRs) if applicable
```

- [ ] **Step 6: Create CODEOWNERS**

Write to `CODEOWNERS`:

```
# Default owner for everything in the repo
*       @alator-ta

# Spec changes need extra scrutiny — same owner today, separate line for future delegation
/spec/  @alator-ta
```

(Replace `@alator-ta` with the correct GitHub handle if different — verify with `git config user.name`.)

- [ ] **Step 7: Verify all five files exist**

```bash
ls -1 .github/ISSUE_TEMPLATE/bug_report.md .github/ISSUE_TEMPLATE/feature_request.md .github/ISSUE_TEMPLATE/config.yml .github/PULL_REQUEST_TEMPLATE.md CODEOWNERS
```

Expected: all five files listed.

- [ ] **Step 8: Commit**

```bash
git add .github/ CODEOWNERS
git commit -m "ci: add issue templates, PR template, CODEOWNERS"
```

---

## Task 4 — Monorepo directory stubs

**Files:**
- Create: `cli/README.md`
- Create: `server/README.md`
- Create: `spec/README.md`
- Create: `skills/README.md`
- Create: `examples/README.md`

These stubs reserve the dirs and tell readers what each will hold. No code yet.

- [ ] **Step 1: Create the directories**

```bash
mkdir -p cli server spec skills examples
```

- [ ] **Step 2: Write cli/README.md**

```markdown
# Tolvi CLI

> **Status:** Phase 3 (not yet shipped). Track progress in [`ROADMAP.md`](../ROADMAP.md).

This directory will hold the `tolvi` command-line interface — a single static Go binary, cross-platform (darwin / linux / windows × amd64 / arm64).

## Planned commands

- `tolvi init` — scaffold `.tolvi/` in the current repo
- `tolvi sync` — capture session context as a vault doc
- `tolvi status <doc> <new_status>` — update lifecycle status with supersession tracking
- `tolvi recall <query>` — local lexical search over the current repo's vault
- `tolvi ask <query>` — semantic search + LLM-synthesized answer
- `tolvi doctor` — diagnostic: config, vault structure, embedding model, server connectivity
- `tolvi unify` — generate the cross-repo unified Obsidian view
- `tolvi publish` — push vault content to a configured Tolvi server (optional)

See [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) for the local-arm component in context.
```

- [ ] **Step 3: Write server/README.md**

```markdown
# Tolvi Server

> **Status:** Phase 2 (not yet shipped). Track progress in [`ROADMAP.md`](../ROADMAP.md).

This directory will hold the `tolvi` HTTP server — TypeScript on Node 20, Fastify, Postgres + pgvector, multi-tenant via API keys at the workspace level.

## Planned API surface

- `POST /v1/documents` — ingest a vault doc (idempotent on `content_hash`)
- `GET /v1/documents` — list with filters
- `GET /v1/documents/:id` — fetch one
- `DELETE /v1/documents/:id` — soft delete
- `POST /v1/search` — semantic + filter search; returns ranked chunks with citations
- `POST /v1/ask` — search with LLM synthesis
- `POST /v1/sync` — batch ingest from CLI publish
- `GET /v1/repos` — list a workspace's repos

OpenAPI spec will be generated from Fastify route definitions and published to [`../spec/`](../spec/).

## Self-hosting

A `docker-compose.yml` will bring up Postgres (with pgvector) and the server in one command. See [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) for the server-arm component in context.
```

- [ ] **Step 4: Write spec/README.md**

```markdown
# Spec

The public format contract for Tolvi vaults.

## Files

- [`tolvi-format-v1.md`](./tolvi-format-v1.md) — the normative spec. **Read this first.**
- [`schemas/`](./schemas/) — machine-readable JSON Schemas (Draft 2020-12) for validating vault content.

## Stability

`tolvi-format-v1` is the stable v1 contract. Breaking changes require a `tolvi-format-v2` revision, which will live in this directory alongside v1, with documented migration tooling.

Every consumer (CLI, server, SDKs, third-party tools) implements parsing and validation against the spec. The spec is the only artifact crossing language boundaries — there is no shared parsing library.
```

- [ ] **Step 5: Write skills/README.md**

```markdown
# Agent integrations

> **Status:** Phase 4 (not yet shipped). Track progress in [`ROADMAP.md`](../ROADMAP.md).

This directory will hold drop-in integration files for AI coding agents:

- `claude-code/` — skill files for Claude Code (capture-on-finish, recall-before-decide, ask-in-context)
- `cursor/` — `.cursorrules` template + workflow examples
- `aider/`, `openhands/`, `continue/` — skeleton integrations

The marketing pitch this enables: "Add three lines to your CLAUDE.md, get a queryable, version-controlled history of every architectural decision your AI agents make."
```

- [ ] **Step 6: Write examples/README.md**

```markdown
# Examples

Synthetic content used for documentation, demos, and CI validation.

- [`sample-vault/`](./sample-vault/) — a complete `tolvi-format-v1` vault with 6 decisions, 3 session-day files, and 4 patterns. All content is fully synthetic; no real-world content is used. The CI workflow validates every file in this vault against the JSON Schemas in [`../spec/schemas/`](../spec/schemas/).

To validate locally:

```bash
npx ajv-cli@5 validate -s ../spec/schemas/decision.json -d 'sample-vault/decisions/*.md' --extract-frontmatter
```
```

- [ ] **Step 7: Verify all five stub READMEs exist and reference the right phases**

```bash
ls -1 cli/README.md server/README.md spec/README.md skills/README.md examples/README.md
```

Expected: all five files listed.

```bash
grep -lE 'corvin|firebase|firestore' cli/README.md server/README.md spec/README.md skills/README.md examples/README.md && echo "FAIL: forbidden term" || echo "OK: clean"
```

Expected: prints `OK: clean`.

- [ ] **Step 8: Commit**

```bash
git add cli/README.md server/README.md spec/README.md skills/README.md examples/README.md
git commit -m "docs: add stub READMEs for cli/, server/, spec/, skills/, examples/"
```

---

## Task 5 — Architecture, Conventions, Open Questions docs

**Files:**
- Create: `docs/ARCHITECTURE.md`
- Create: `docs/CONVENTIONS.md`
- Create: `docs/OPEN_QUESTIONS.md`

Content for these documents is the prose form of spec sections 5, 6, and 11. The plan reproduces the structure here; the implementing agent should write the prose using the spec as the source of truth.

- [ ] **Step 1: Write docs/ARCHITECTURE.md**

Use spec section 5 as the source of truth. The file MUST contain:

1. **Header:** `# Tolvi architecture` + 1-paragraph intro
2. **The 3-layer ASCII diagram** (verbatim from spec section 5)
3. **Section 1 — System surfaces and ownership** (CLI owns local index + capture; Server owns multi-tenant index + HTTP API; format spec is the only contract crossing the boundary)
4. **Section 2 — Trust + auth model** (API key per workspace, hashed at rest using argon2 or pgcrypto, scoped to ingest+search; no user accounts in v1; CLI sends `Authorization: Bearer <key>`; local-only operation requires no key)
5. **Section 3 — Data flow** (capture → CLI computes content_hash → POST /v1/sync → server idempotent on (repo_id, path, content_hash) → server-side chunking + embedding → pgvector → search/ask reads are read-after-write consistent within a single API call)
6. **Section 4 — Component boundaries** (no shared library; Go CLI parses tolvi-format-v1 in Go; TS server parses it in TS; trade-off: cross-language reach > shared lib convenience; format spec is the contract)
7. **Section 5 — What's deferred to Phase 9** (multi-arm shared library, web dashboard, OIDC/SSO, billing, OpenAPI-generated SDKs; aggregator automation deferred to v1.x)
8. **Section 6 — Self-host story** (`docker compose up` brings up Postgres+pgvector + server; CLI defaults to `http://localhost:3000`; documented end-to-end in Phase 2)

Word count target: 800–1,200 words.

Verification:

```bash
test -f docs/ARCHITECTURE.md && wc -w docs/ARCHITECTURE.md
```

Expected: file exists; word count in 800–1,200 range.

```bash
grep -iE 'corvin|firebase|firestore|gs://|healthcare|hipaa' docs/ARCHITECTURE.md && echo "FAIL: forbidden term" || echo "OK: clean"
```

Expected: prints `OK: clean`.

- [ ] **Step 2: Write docs/CONVENTIONS.md**

Use spec section 6 as the source of truth. The file MUST contain:

1. **Header:** `# Vault conventions`
2. **Section 1 — Vault directory layout** (the `vault/{decisions,sessions,patterns}/` tree + the `.vault-meta.json` example with workspace, embedding_model, schema_version=1)
3. **Section 2 — Status enum (six values)** (the table from spec section 6.2 verbatim, plus the surfacing note about UI presentation being a consumer concern)
4. **Section 3 — Frontmatter schemas** (universally required: tags, status; sessions add date; decisions add date, repo, ticket (optional, free-form e.g. "PROJ-123"), supersedes (optional), superseded_by (optional); patterns are timeless — no date or repo required, optional languages and frameworks; genericization notes for fields that were dropped)
5. **Section 4 — Wiki-link syntax** (`[[slug]]` for same-repo, `[[repo:slug]]` for cross-repo; citations in `/v1/ask` MUST use this syntax)
6. **Section 5 — Decision template** (the `## Why / ## How / ## Outcome` three-section template)
7. **Section 6 — Aggregator pattern (recommended, not automated in v1)** (the `mkdir + ln -s` recipe for engineers who want a unified Obsidian view; v1.x adds `tolvi unify` to automate)

Word count target: 600–1,000 words.

Verification:

```bash
test -f docs/CONVENTIONS.md && wc -w docs/CONVENTIONS.md
grep -iE 'corvin|firebase|firestore|gs://|healthcare|hipaa' docs/CONVENTIONS.md && echo "FAIL: forbidden term" || echo "OK: clean"
```

Expected: file exists; word count in range; prints `OK: clean`.

- [ ] **Step 3: Write docs/OPEN_QUESTIONS.md**

Use spec section 11 as the source of truth. Reproduce the 8 open questions verbatim:

1. Loud-fail on Ollama down
2. Cross-repo type sharing
3. Vault content audit cadence
4. Index size growth
5. Secrets pre-commit lint
6. Multi-tenant isolation in Phase 2
7. Aggregator automation
8. CLI ↔ server format-version handshake

Add a footer note: "This file is meant to grow. When you find a non-obvious open question while working on Tolvi, add it here. When a question gets resolved, move it to an ADR."

Verification:

```bash
test -f docs/OPEN_QUESTIONS.md
grep -c '^[0-9]\.' docs/OPEN_QUESTIONS.md
```

Expected: file exists; count is 8 (one numbered question per line opening with `1.` through `8.`).

- [ ] **Step 4: Commit**

```bash
git add docs/ARCHITECTURE.md docs/CONVENTIONS.md docs/OPEN_QUESTIONS.md
git commit -m "docs: add ARCHITECTURE, CONVENTIONS, OPEN_QUESTIONS"
```

---

## Task 6 — ADR setup (`docs/adr/`)

**Files:**
- Create: `docs/adr/README.md`
- Create: `docs/adr/0001-architecture-overview.md`
- Create: `docs/adr/0002-vault-format-v1-contract.md`

Tolvi dogfoods its own ADR convention from PR #1. ADRs are short, numbered, and use the `Status / Context / Decision / Consequences` structure.

- [ ] **Step 1: Create the directory and write the ADR README + template**

```bash
mkdir -p docs/adr
```

Write to `docs/adr/README.md`:

```markdown
# Architecture Decision Records

Tolvi tracks every meaningful architectural choice as an Architecture Decision Record (ADR). ADRs are short, numbered files that capture the *context* and *consequences* of a decision so future maintainers can understand why something is the way it is.

## Format

Each ADR is one markdown file named `NNNN-short-slug.md`, where `NNNN` is a zero-padded sequential number.

```markdown
# NNNN — Title

**Status:** [proposed | accepted | superseded by NNNN | deprecated]
**Date:** YYYY-MM-DD

## Context

What is the situation? What forces are at play?

## Decision

What did we choose to do, and what trade-off does that buy us?

## Consequences

What follows from this decision — both positive and negative? What becomes easier? What becomes harder?
```

## Index

| # | Title | Status |
|---|---|---|
| 0001 | [Architecture overview](./0001-architecture-overview.md) | accepted |
| 0002 | [Vault format v1 contract](./0002-vault-format-v1-contract.md) | accepted |

## Adding a new ADR

1. Pick the next sequential number
2. Copy the template above into `NNNN-short-slug.md`
3. Fill in all four sections — leave nothing as "TBD"
4. Add an entry to the index above
5. Commit in the same PR as the work the decision enables

## Superseding

When a decision is replaced, set the old ADR's status to `superseded by NNNN`, link to the replacement, and create the new ADR. The old file stays — it's part of the historical record.
```

- [ ] **Step 2: Write 0001-architecture-overview.md**

Write to `docs/adr/0001-architecture-overview.md`:

```markdown
# 0001 — Architecture overview

**Status:** accepted
**Date:** 2026-05-09

## Context

Tolvi is a developer knowledge tool that captures engineering decisions where they happen and surfaces them via natural-language search from CLI or web. The product needs to support both:

- A local-first workflow (engineer at their machine, capturing and querying a single repo's vault), and
- A team workflow (multiple engineers across multiple repos, querying a shared index hosted on infrastructure they control)

Several architectures could satisfy this:

1. **Single-arm cloud-only:** all queries go through a hosted server. Simple, but breaks the local-first promise and creates a hard dependency on the network.
2. **Single-arm local-only:** every engineer indexes locally; no server. Works for individuals but not teams.
3. **Two-arm with a shared library:** local CLI and server share an embedding/chunking library. Cleanest in theory but locks both arms to the same language.
4. **Two-arm with a shared spec, no shared library:** local CLI and server are independent implementations of the same vault format spec. Different languages allowed.

## Decision

Adopt option **4: two-arm with a shared format spec**.

- The CLI is written in Go and ships as a single static binary
- The server is written in TypeScript on Fastify, with Postgres + pgvector for the index
- Both arms parse and validate `tolvi-format-v1` independently
- The spec at `/spec/tolvi-format-v1.md` is the only artifact crossing the language boundary

## Consequences

**Positive:**

- The CLI is genuinely standalone — no Node.js runtime needed for local users
- The server can be optimized for its workload (Postgres-native indexing, pgvector for search) without compromising the CLI binary's size or startup time
- Adding a Python implementation, a Rust implementation, or a third-party reimplementation costs nothing in either arm — they all read the same spec
- The format spec gets the level of rigor it deserves — it's a public contract, not an implementation detail

**Negative:**

- Code duplication: chunking, embedding, frontmatter parsing all exist twice
- Risk of drift: a bug fix in the Go parser must be ported to the TypeScript parser
- Mitigation: the JSON Schemas in `/spec/schemas/` are the conformance tests both arms validate against; CI runs them on every change
```

- [ ] **Step 3: Write 0002-vault-format-v1-contract.md**

Write to `docs/adr/0002-vault-format-v1-contract.md`:

```markdown
# 0002 — Vault format v1 contract

**Status:** accepted
**Date:** 2026-05-09

## Context

Tolvi's vault format is a public contract that downstream tools — CLIs, servers, SDKs, third-party integrations — all rely on. Once users have vaults populated with `tolvi-format-v1` content, breaking changes are expensive: they require migration tooling and force every tool in the ecosystem to upgrade.

Several approaches to this contract were considered:

1. **No spec — let implementations diverge.** Bad: ecosystem fragmentation, no portability of vault content between tools.
2. **Loose spec — describe the shape but leave details unspecified.** Bad: implementations make different choices about edge cases (default status, missing fields, ordering rules) and content stops working when moved between tools.
3. **Strict spec with frozen defaults.** Best: implementations have a clear conformance target; the cost is committing early to numbers (recency curve, status filter defaults) before deployment data has accumulated.

## Decision

Adopt **option 3: strict spec with frozen defaults** for `tolvi-format-v1`. Specifically:

- The status enum is **frozen at six values**: `active`, `in-progress`, `superseded`, `deprecated`, `draft`, `historical`. Adding a value requires `tolvi-format-v2`.
- The frontmatter schema for each doc type (decision, session, pattern) is locked. Required vs optional fields are normative.
- The `.vault-meta.json` schema is locked with `schema_version: 1`.
- The recency multiplier `(0.8 + 0.2 × exp(-age_days/180))` and session-down-weight `× 0.7` are documented as **informative** defaults — implementations MAY tune. The defaults SHOULD apply when no tuning is configured.
- Wiki-link syntax (`[[slug]]`, `[[repo:slug]]`) is normative.

These numbers are adopted from prior reference deployments where they were validated in production.

## Consequences

**Positive:**

- Vault content moves cleanly between tools — a vault written by one CLI works in any conformant CLI
- The format spec gets test coverage automatically: every conformant implementation runs its parser against `/examples/sample-vault/` in CI
- Migration costs are predictable: format changes go through `tolvi-format-v2` with documented migration tooling

**Negative:**

- Some defaults will probably want tuning (e.g. the 180-day half-life on the recency multiplier). The tuning escape hatch lives in implementations, not in the spec.
- The status enum cannot grow without a v2 bump. If a real-world need emerges for a seventh status, it triggers a format revision.
- Mitigation: open questions tracking in [`../OPEN_QUESTIONS.md`](../OPEN_QUESTIONS.md) surfaces tuning candidates so v2 starts informed.
```

- [ ] **Step 4: Verify ADR files**

```bash
ls -1 docs/adr/README.md docs/adr/0001-architecture-overview.md docs/adr/0002-vault-format-v1-contract.md
grep -iE 'corvin|firebase|firestore|gs://|healthcare|hipaa' docs/adr/*.md && echo "FAIL" || echo "OK: clean"
```

Expected: all three files listed; prints `OK: clean`.

- [ ] **Step 5: Commit**

```bash
git add docs/adr/
git commit -m "docs: bootstrap ADR convention with 0001 (architecture) and 0002 (format contract)"
```

---

## Task 7 — `spec/tolvi-format-v1.md` (the public format contract)

**Files:**
- Create: `spec/tolvi-format-v1.md`

This is the most important deliverable of the foundational PR. The content follows spec section 7 (12 numbered sections). Word count target: 1,500–2,500 words.

- [ ] **Step 1: Write the spec file**

Create `spec/tolvi-format-v1.md` with the following 12 sections in order. Each section heading is `## N. <title>`. Use the spec doc (section 7) as the source of structure; the body prose is written fresh:

1. **Status** — "Stable as of v1.0. Breaking changes require `tolvi-format-v2` and migration tooling. The schema files in `./schemas/` are the machine-readable conformance test for this spec."

2. **Vault directory layout (normative)** — describe the `vault/` tree, `.vault-meta.json` marker, and the three subdirectories (`decisions/`, `sessions/`, `patterns/`).

3. **`.vault-meta.json` (normative)** — describe the schema fields (`workspace`, `embedding_model`, `schema_version`), with a JSON example and a reference to `./schemas/vault-meta.json`.

4. **File naming rules (normative)** — `decisions/YYYY-MM-DD-slug.md`, `sessions/YYYY-MM-DD.md` (one per day), `patterns/slug.md` (no date prefix).

5. **Frontmatter (normative)** — break out by doc type. Reference `./schemas/{decision,session,pattern}.json` for the exact field-level rules.

6. **Status enum (normative)** — the six-value table with surfaced-by-default behavior. Explicitly state: "implementations MUST default `/v1/search` and recall-style queries to exclude `superseded`, `deprecated`, and `draft`."

7. **Wiki-link syntax (normative)** — `[[slug]]` for same-repo references, `[[repo:slug]]` for cross-repo. State that citations in synthesis-style endpoints (e.g. `POST /v1/ask`) MUST use this syntax.

8. **Cross-reference rules (normative)** — bidirectional supersession requirement: when doc A is superseded by doc B, A's frontmatter MUST set `status: superseded` and `superseded_by: <B-slug>`, and B's frontmatter MUST set `supersedes: <A-slug>`. Atomic updates required.

9. **RAG defaults (informative)** — recency multiplier `(0.8 + 0.2 × exp(-age_days/180))`, session-doc score multiplier `× 0.7`, default status filter excluding `superseded|deprecated|draft`. State explicitly that implementations MAY tune; the contract is the *defaults*, not the requirement to use them. Include a one-line rationale: "These numbers are adopted from prior reference deployments where they were validated in production."

10. **Embedding model defaults (informative)** — `nomic-embed-text` (Ollama, 768 dims) for the local CLI; configurable via `.vault-meta.json`. Server-side embedding model is deployment configuration, not part of the spec.

11. **Versioning rules (normative)** — `tolvi-format-v2` will require migration tooling. Spec docs at `/spec/tolvi-format-v2.md`. Implementations declare supported versions in the handshake (handshake details deferred to a Phase 2 ADR).

12. **Conformance (normative)** — "A vault is conformant with `tolvi-format-v1` if (a) every file matches the JSON Schema for its type, and (b) the directory layout matches Section 2. The schemas in `./schemas/` are authoritative."

- [ ] **Step 2: Verify the spec exists, has the expected structure, and is brand-clean**

```bash
test -f spec/tolvi-format-v1.md
grep -c '^## [0-9]\+\.' spec/tolvi-format-v1.md
```

Expected: file exists; `grep -c` returns 12 (twelve numbered sections).

```bash
wc -w spec/tolvi-format-v1.md
```

Expected: word count in 1,500–2,500 range.

```bash
grep -iE 'corvin|firebase|firestore|gs://|healthcare|hipaa' spec/tolvi-format-v1.md && echo "FAIL" || echo "OK: clean"
```

Expected: prints `OK: clean`.

- [ ] **Step 3: Commit**

```bash
git add spec/tolvi-format-v1.md
git commit -m "spec: add tolvi-format-v1 public format contract"
```

---

## Task 8 — JSON Schemas (`spec/schemas/`)

**Files:**
- Create: `spec/schemas/vault-meta.json`
- Create: `spec/schemas/decision.json`
- Create: `spec/schemas/session.json`
- Create: `spec/schemas/pattern.json`

JSON Schema Draft 2020-12. The schemas are the machine-readable conformance test for `tolvi-format-v1`. CI in Task 10 validates the sample vault in Task 9 against these schemas.

- [ ] **Step 1: Create the directory**

```bash
mkdir -p spec/schemas
```

- [ ] **Step 2: Write spec/schemas/vault-meta.json**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://tolvi.dev/spec/schemas/vault-meta.json",
  "title": "Tolvi vault metadata",
  "description": "The .vault-meta.json marker file at the root of every Tolvi vault directory.",
  "type": "object",
  "required": ["workspace", "embedding_model", "schema_version"],
  "additionalProperties": false,
  "properties": {
    "workspace": {
      "type": "string",
      "minLength": 1,
      "description": "The workspace this vault belongs to. Used by the server for multi-tenant isolation."
    },
    "embedding_model": {
      "type": "string",
      "minLength": 1,
      "description": "The embedding model used to index this vault locally. Default: nomic-embed-text."
    },
    "schema_version": {
      "type": "integer",
      "const": 1,
      "description": "Tolvi vault format version. Currently always 1."
    }
  }
}
```

- [ ] **Step 3: Write spec/schemas/decision.json**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://tolvi.dev/spec/schemas/decision.json",
  "title": "Tolvi decision frontmatter",
  "description": "Frontmatter schema for files in vault/decisions/.",
  "type": "object",
  "required": ["tags", "date", "status", "repo"],
  "additionalProperties": true,
  "properties": {
    "tags": {
      "type": "array",
      "items": { "type": "string", "minLength": 1 },
      "minItems": 1,
      "description": "Tags MUST include 'decision'. Other tags are free-form."
    },
    "date": {
      "type": "string",
      "pattern": "^\\d{4}-\\d{2}-\\d{2}$",
      "description": "ISO-8601 date matching the filename date prefix."
    },
    "status": {
      "type": "string",
      "enum": ["active", "in-progress", "superseded", "deprecated", "draft", "historical"]
    },
    "repo": {
      "type": "string",
      "minLength": 1,
      "description": "The repo this decision binds to (slug, not URL)."
    },
    "ticket": {
      "type": "string",
      "description": "Free-form issue tracker reference. May be a tracker ID like 'PROJ-123', a URL, or 'none'."
    },
    "supersedes": {
      "type": "string",
      "description": "Slug of an older decision this one replaces."
    },
    "superseded_by": {
      "type": "string",
      "description": "Slug of a newer decision that replaces this one. Set when status is 'superseded'."
    }
  }
}
```

- [ ] **Step 4: Write spec/schemas/session.json**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://tolvi.dev/spec/schemas/session.json",
  "title": "Tolvi session frontmatter",
  "description": "Frontmatter schema for files in vault/sessions/. One file per day; multiple session blocks per file.",
  "type": "object",
  "required": ["tags", "date", "status"],
  "additionalProperties": true,
  "properties": {
    "tags": {
      "type": "array",
      "items": { "type": "string", "minLength": 1 },
      "minItems": 1,
      "description": "Tags MUST include 'session'. Other tags are free-form."
    },
    "date": {
      "type": "string",
      "pattern": "^\\d{4}-\\d{2}-\\d{2}$",
      "description": "ISO-8601 date matching the filename."
    },
    "status": {
      "type": "string",
      "enum": ["active", "in-progress", "superseded", "deprecated", "draft", "historical"]
    }
  }
}
```

- [ ] **Step 5: Write spec/schemas/pattern.json**

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://tolvi.dev/spec/schemas/pattern.json",
  "title": "Tolvi pattern frontmatter",
  "description": "Frontmatter schema for files in vault/patterns/. Patterns are intentionally timeless — no date or repo required.",
  "type": "object",
  "required": ["tags", "status"],
  "additionalProperties": true,
  "properties": {
    "tags": {
      "type": "array",
      "items": { "type": "string", "minLength": 1 },
      "minItems": 1,
      "description": "Tags MUST include 'pattern'. Other tags are free-form."
    },
    "status": {
      "type": "string",
      "enum": ["active", "in-progress", "superseded", "deprecated", "draft", "historical"]
    },
    "languages": {
      "type": "array",
      "items": { "type": "string", "minLength": 1 },
      "description": "Optional list of programming languages this pattern applies to."
    },
    "frameworks": {
      "type": "array",
      "items": { "type": "string", "minLength": 1 },
      "description": "Optional list of frameworks/libraries this pattern applies to."
    }
  }
}
```

- [ ] **Step 6: Verify all four schemas are valid JSON**

```bash
for f in spec/schemas/*.json; do
  node -e "JSON.parse(require('fs').readFileSync('$f','utf8'))" && echo "OK: $f"
done
```

Expected: prints `OK: spec/schemas/decision.json`, `OK: spec/schemas/pattern.json`, `OK: spec/schemas/session.json`, `OK: spec/schemas/vault-meta.json`.

- [ ] **Step 7: Verify schemas compile under ajv**

```bash
npx --yes ajv-cli@5 compile -s spec/schemas/vault-meta.json
npx --yes ajv-cli@5 compile -s spec/schemas/decision.json
npx --yes ajv-cli@5 compile -s spec/schemas/session.json
npx --yes ajv-cli@5 compile -s spec/schemas/pattern.json
```

Expected: each prints `schema spec/schemas/<name>.json is valid`.

- [ ] **Step 8: Commit**

```bash
git add spec/schemas/
git commit -m "spec: add JSON Schemas for vault-meta, decision, session, pattern"
```

---

## Task 9 — Synthetic sample vault (`examples/sample-vault/`)

**Files:**
- Create: `examples/sample-vault/.vault-meta.json`
- Create: 6 files under `examples/sample-vault/decisions/`
- Create: 3 files under `examples/sample-vault/sessions/`
- Create: 4 files under `examples/sample-vault/patterns/`

All content is fully synthetic. Generic engineering scenarios only — no domain-specific examples.

- [ ] **Step 1: Create directories**

```bash
mkdir -p examples/sample-vault/decisions examples/sample-vault/sessions examples/sample-vault/patterns
```

- [ ] **Step 2: Write .vault-meta.json**

`examples/sample-vault/.vault-meta.json`:

```json
{
  "workspace": "tolvi-examples",
  "embedding_model": "nomic-embed-text",
  "schema_version": 1
}
```

- [ ] **Step 3: Write the 6 decision files**

Use the following list. Each file has frontmatter + a `## Why / ## How / ## Outcome` body. The bodies should be 100–200 words of plausible, generic engineering prose. Two of the six are non-`active` to exercise status filtering.

| # | Filename | Status | Topic |
|---|---|---|---|
| 1 | `2026-04-12-choose-postgres-over-mysql.md` | active | Choosing Postgres for the primary store |
| 2 | `2026-04-15-feature-flags-via-openfeature.md` | active | Adopting OpenFeature for flag delivery |
| 3 | `2026-04-22-grpc-for-service-to-service.md` | active | Switching service-to-service from REST to gRPC |
| 4 | `2026-04-25-pin-node-with-volta.md` | active | Pinning Node version with Volta |
| 5 | `2026-04-28-jwt-vs-paseto-tokens.md` | superseded | (Earlier choice of JWT, superseded by PASETO) |
| 6 | `2026-05-02-paseto-tokens.md` | active | (Replacement: PASETO with link back to #5 via supersedes/superseded_by) |

Frontmatter template for each (replace bracketed values):

```yaml
---
tags: [decision, sample]
date: <YYYY-MM-DD>
status: <status>
repo: tolvi-examples
ticket: none
---
```

For decisions #5 and #6, add the supersession links:

```yaml
# In #5 (superseded):
status: superseded
superseded_by: 2026-05-02-paseto-tokens

# In #6 (active replacement):
supersedes: 2026-04-28-jwt-vs-paseto-tokens
```

- [ ] **Step 4: Write the 3 session files**

| Filename | Note |
|---|---|
| `2026-04-12.md` | Single session block — landed Postgres choice |
| `2026-04-22.md` | Two session blocks (morning + afternoon) — gRPC investigation + Volta pin spike |
| `2026-05-02.md` | Single session block — token rotation work that produced the supersession |

Each file uses this frontmatter:

```yaml
---
tags: [session, sample]
date: <YYYY-MM-DD>
status: active
---
```

Body uses the convention `## [HH:MM] Session — <one-line summary>` per block, followed by a 50–100-word narrative referencing the relevant decision via `[[slug]]`.

- [ ] **Step 5: Write the 4 pattern files**

| Filename | Topic | Languages / Frameworks |
|---|---|---|
| `idempotent-migrations.md` | Pattern for safely-rerunnable database migrations | none specified |
| `tracing-context-propagation.md` | OpenTelemetry context propagation across service hops | `["typescript", "go"]` |
| `feature-flag-rollout.md` | Progressive feature flag rollout (1% → 10% → 50% → 100%) | none specified |
| `postgres-advisory-locks.md` | Coordinating cross-process work via Postgres advisory locks | `["sql"]` |

Frontmatter template:

```yaml
---
tags: [pattern, sample]
status: active
---
```

(For files with `languages` or `frameworks`, add those fields to the frontmatter.)

Body: 150–300 words describing the pattern, when to use it, and a small code example.

- [ ] **Step 6: Verify the sample vault structure**

```bash
ls examples/sample-vault/decisions/ | wc -l
ls examples/sample-vault/sessions/ | wc -l
ls examples/sample-vault/patterns/ | wc -l
test -f examples/sample-vault/.vault-meta.json
```

Expected: 6, 3, 4, and the .vault-meta.json file exists.

- [ ] **Step 7: Validate the vault-meta against its schema**

```bash
npx --yes ajv-cli@5 validate -s spec/schemas/vault-meta.json -d examples/sample-vault/.vault-meta.json
```

Expected: prints `examples/sample-vault/.vault-meta.json valid`.

- [ ] **Step 8: Validate every decision/session/pattern file's frontmatter**

The `extract-frontmatter.js` helper from Task 10 doesn't exist yet, so for this task verify manually with a one-off Node command:

```bash
for f in examples/sample-vault/decisions/*.md; do
  node -e "
    const fs = require('fs');
    const txt = fs.readFileSync('$f', 'utf8');
    const m = txt.match(/^---\n([\s\S]*?)\n---/);
    if (!m) { console.error('FAIL: no frontmatter in $f'); process.exit(1); }
    console.log('OK: $f has frontmatter (' + m[1].split('\n').length + ' lines)');
  "
done
```

Expected: 6 `OK:` lines for decisions. Repeat the loop for `sessions/` (expect 3) and `patterns/` (expect 4).

- [ ] **Step 9: Verify no forbidden brand references**

```bash
grep -riE 'corvin|firebase|firestore|gs://|healthcare|hipaa|MRN|BAA|eligibility' examples/sample-vault/ && echo "FAIL" || echo "OK: clean"
```

Expected: prints `OK: clean`.

- [ ] **Step 10: Commit**

```bash
git add examples/sample-vault/
git commit -m "examples: add synthetic sample vault (6 decisions, 3 sessions, 4 patterns)"
```

---

## Task 10 — CI workflow + helper scripts

**Files:**
- Create: `.github/scripts/extract-frontmatter.js`
- Create: `.github/scripts/brand-isolation-check.sh`
- Create: `.github/workflows/validate.yml`

- [ ] **Step 1: Create the scripts directory**

```bash
mkdir -p .github/scripts
```

- [ ] **Step 2: Write the frontmatter extractor (Node, no deps)**

`.github/scripts/extract-frontmatter.js`:

```javascript
#!/usr/bin/env node
// Usage: extract-frontmatter.js <markdown-file>
// Reads a markdown file, extracts the YAML frontmatter between leading --- markers,
// converts it to JSON, and writes the JSON to stdout. Exits non-zero if no frontmatter found.

const fs = require('fs');
const path = require('path');

const file = process.argv[2];
if (!file) {
  console.error('Usage: extract-frontmatter.js <file.md>');
  process.exit(2);
}

const text = fs.readFileSync(file, 'utf8');
const match = text.match(/^---\n([\s\S]*?)\n---/);
if (!match) {
  console.error(`No frontmatter found in ${file}`);
  process.exit(1);
}

// Parse the YAML frontmatter ourselves — no external deps.
// Only handles: scalar strings, arrays of strings, integers, the canonical YAML our schemas expect.
const yaml = match[1];
const obj = {};
const lines = yaml.split('\n');

let i = 0;
while (i < lines.length) {
  const line = lines[i];
  if (!line.trim() || line.trim().startsWith('#')) { i++; continue; }
  const kvMatch = line.match(/^([a-zA-Z_][a-zA-Z0-9_]*):\s*(.*)$/);
  if (!kvMatch) { i++; continue; }
  const key = kvMatch[1];
  const rawValue = kvMatch[2].trim();

  if (rawValue === '') {
    // Multi-line array (block style)
    const arr = [];
    i++;
    while (i < lines.length && lines[i].match(/^\s*-\s+/)) {
      arr.push(lines[i].replace(/^\s*-\s+/, '').trim().replace(/^["']|["']$/g, ''));
      i++;
    }
    obj[key] = arr;
  } else if (rawValue.startsWith('[') && rawValue.endsWith(']')) {
    // Inline array
    const inner = rawValue.slice(1, -1).trim();
    obj[key] = inner === '' ? [] : inner.split(',').map(s => s.trim().replace(/^["']|["']$/g, ''));
    i++;
  } else if (/^-?\d+$/.test(rawValue)) {
    obj[key] = parseInt(rawValue, 10);
    i++;
  } else {
    obj[key] = rawValue.replace(/^["']|["']$/g, '');
    i++;
  }
}

process.stdout.write(JSON.stringify(obj, null, 2) + '\n');
```

Make it executable:

```bash
chmod +x .github/scripts/extract-frontmatter.js
```

- [ ] **Step 3: Verify the extractor on a known sample**

```bash
node .github/scripts/extract-frontmatter.js examples/sample-vault/decisions/2026-04-12-choose-postgres-over-mysql.md
```

Expected: prints a JSON object containing at minimum `tags`, `date`, `status`, `repo`. Exit 0.

- [ ] **Step 4: Write the brand-isolation guard script**

`.github/scripts/brand-isolation-check.sh`:

```bash
#!/usr/bin/env bash
# Brand-isolation check. Fails (exit 1) if any tracked file outside the allowlist
# references parent organizations, sibling products, or specific deployments.
#
# Allowlist entries are matched as path prefixes:
#   - "NOTICE" matches the file ./NOTICE exactly
#   - "docs/superpowers/" matches any file under that directory
#
# See CONTRIBUTING.md "Brand isolation" for the rule and rationale.
set -euo pipefail

# Block-list of forbidden patterns (case-insensitive). Add carefully — false positives
# annoy contributors; missed terms break the brand-isolation promise.
FORBIDDEN_PATTERNS=(
  'corvin'
  'corvin health'
  'torres atlantic'
  'ta-ai-tooling'
  'ta-wide'
  '\bfirebase\b'
  '\bfirestore\b'
  '\bcloud run\b'
  'gs://'
  'healthcare'
  'hipaa'
  '\bbaa\b'
  '\bphi\b'
  '\bmrn\b'
  'eligibility'
)

# Allowlist (file path or directory prefix). Files matching any entry are skipped.
#  - NOTICE: the official attribution file (only place the parent-org name belongs)
#  - This script: contains the patterns it searches for
#  - docs/superpowers/: engineering-process artifacts (specs, plans) that may
#    legitimately discuss the rule, including showing the pattern list. These
#    ship publicly as transparency artifacts.
ALLOWLIST_PATHS=(
  'NOTICE'
  '.github/scripts/brand-isolation-check.sh'
  'docs/superpowers/'
)

# Compile the forbidden-pattern regex once
PATTERN=$(IFS='|'; echo "${FORBIDDEN_PATTERNS[*]}")

is_allowlisted() {
  local file="$1"
  for allowed in "${ALLOWLIST_PATHS[@]}"; do
    # Directory prefix match (allowlist entry ends with /)
    if [[ "$allowed" == */ ]] && [[ "$file" == "$allowed"* ]]; then
      return 0
    fi
    # Exact file match
    if [ "$file" = "$allowed" ]; then
      return 0
    fi
  done
  return 1
}

FOUND=0
while IFS= read -r file; do
  if is_allowlisted "$file"; then continue; fi
  if grep -iIqE "$PATTERN" "$file" 2>/dev/null; then
    echo "FAIL: forbidden term in tracked file: $file"
    grep -iInE "$PATTERN" "$file" | head -5
    FOUND=1
  fi
done < <(git ls-files)

if [ "$FOUND" -ne 0 ]; then
  echo ""
  echo "Brand-isolation check failed. The forbidden terms above must be removed,"
  echo "or the file must be added to the allowlist in .github/scripts/brand-isolation-check.sh."
  echo "See CONTRIBUTING.md 'Brand isolation' section for details."
  exit 1
fi

echo "Brand-isolation check passed."
```

Make it executable:

```bash
chmod +x .github/scripts/brand-isolation-check.sh
```

- [ ] **Step 5: Run the brand-isolation guard against the current repo state**

```bash
.github/scripts/brand-isolation-check.sh
```

Expected: prints `Brand-isolation check passed.` Exit 0.

If FAIL: investigate and fix the leak before continuing. Do NOT add the file to the allowlist as a workaround unless the file genuinely cannot be brand-neutral (currently only NOTICE qualifies).

- [ ] **Step 6: Write the CI workflow**

`.github/workflows/validate.yml`:

```yaml
name: validate

on:
  push:
    branches: [main]
  pull_request:

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Markdown lint
        uses: DavidAnson/markdownlint-cli2-action@v16
        with:
          globs: '**/*.md'
          # Allow a config file later; for now, defaults are fine.

      - name: Compile JSON Schemas
        run: |
          for f in spec/schemas/*.json; do
            npx --yes ajv-cli@5 compile -s "$f"
          done

      - name: Validate sample vault — vault-meta
        run: |
          npx --yes ajv-cli@5 validate \
            -s spec/schemas/vault-meta.json \
            -d examples/sample-vault/.vault-meta.json

      - name: Validate sample vault — decisions
        run: |
          for md in examples/sample-vault/decisions/*.md; do
            node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
            npx --yes ajv-cli@5 validate -s spec/schemas/decision.json -d /tmp/fm.json
          done

      - name: Validate sample vault — sessions
        run: |
          for md in examples/sample-vault/sessions/*.md; do
            node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
            npx --yes ajv-cli@5 validate -s spec/schemas/session.json -d /tmp/fm.json
          done

      - name: Validate sample vault — patterns
        run: |
          for md in examples/sample-vault/patterns/*.md; do
            node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
            npx --yes ajv-cli@5 validate -s spec/schemas/pattern.json -d /tmp/fm.json
          done

      - name: Brand-isolation guard
        run: .github/scripts/brand-isolation-check.sh

      - name: Link checker
        uses: lycheeverse/lychee-action@v2
        with:
          args: --no-progress --exclude-mail '**/*.md'
```

- [ ] **Step 7: Verify the workflow YAML is syntactically valid**

```bash
node -e "
  const fs = require('fs');
  const yaml = fs.readFileSync('.github/workflows/validate.yml', 'utf8');
  // Cheap sanity check — actual schema validation happens on push.
  if (!yaml.includes('jobs:') || !yaml.includes('steps:')) {
    console.error('FAIL: workflow missing required keys');
    process.exit(1);
  }
  console.log('OK: validate.yml looks structurally sound');
"
```

Expected: prints `OK: validate.yml looks structurally sound`.

- [ ] **Step 8: Run all sample-vault validations locally to make sure they pass before CI does**

```bash
# vault-meta
npx --yes ajv-cli@5 validate -s spec/schemas/vault-meta.json -d examples/sample-vault/.vault-meta.json

# decisions
for md in examples/sample-vault/decisions/*.md; do
  node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
  npx --yes ajv-cli@5 validate -s spec/schemas/decision.json -d /tmp/fm.json
done

# sessions
for md in examples/sample-vault/sessions/*.md; do
  node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
  npx --yes ajv-cli@5 validate -s spec/schemas/session.json -d /tmp/fm.json
done

# patterns
for md in examples/sample-vault/patterns/*.md; do
  node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
  npx --yes ajv-cli@5 validate -s spec/schemas/pattern.json -d /tmp/fm.json
done
```

Expected: every line prints `<file> valid`. Any FAIL means the sample vault content needs to be fixed before CI will pass.

- [ ] **Step 9: Commit**

```bash
git add .github/scripts/extract-frontmatter.js .github/scripts/brand-isolation-check.sh .github/workflows/validate.yml
git commit -m "ci: add validate workflow with markdown lint, schema validation, brand guard, link checker"
```

---

## Task 11 — Pre-commit configuration

**Files:**
- Create: `.pre-commit-config.yaml`

- [ ] **Step 1: Write the pre-commit config**

`.pre-commit-config.yaml`:

```yaml
# Tolvi pre-commit hooks. Install with `pre-commit install` after `brew install pre-commit`.
# See CONTRIBUTING.md for details.
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
        exclude: '^LICENSE$|\.md$'
      - id: end-of-file-fixer
        exclude: '^LICENSE$'
      - id: check-json
      - id: check-yaml
      - id: check-merge-conflict

  - repo: local
    hooks:
      - id: brand-isolation
        name: brand-isolation guard
        entry: .github/scripts/brand-isolation-check.sh
        language: script
        pass_filenames: false
        always_run: true

      - id: validate-vault-meta
        name: validate sample-vault meta against schema
        entry: bash -c 'npx --yes ajv-cli@5 validate -s spec/schemas/vault-meta.json -d examples/sample-vault/.vault-meta.json'
        language: system
        pass_filenames: false
        files: '^(spec/schemas/vault-meta\.json|examples/sample-vault/\.vault-meta\.json)$'
```

- [ ] **Step 2: Install pre-commit hooks (if pre-commit is available)**

```bash
which pre-commit && pre-commit install || echo "pre-commit not installed; CONTRIBUTING.md tells contributors how"
```

Expected: either `pre-commit installed at .git/hooks/pre-commit` or the `pre-commit not installed` message. Both are acceptable.

- [ ] **Step 3: Run all hooks against all files (if pre-commit is installed)**

```bash
which pre-commit && pre-commit run --all-files || echo "skipped: pre-commit not installed"
```

Expected: every hook prints `Passed`. If any prints `Failed`, fix the underlying issue before committing.

- [ ] **Step 4: Commit**

```bash
git add .pre-commit-config.yaml
git commit -m "ci: add pre-commit config with brand-isolation guard and schema validation"
```

---

## Task 12 — Final pre-PR verification and PR creation

This task verifies the whole branch is in a shippable state and opens the PR.

- [ ] **Step 1: Confirm branch state**

```bash
git log --oneline main..HEAD
```

Expected: 11 commits listed (one per task above), in order.

- [ ] **Step 2: Run the full local validation suite end-to-end**

```bash
.github/scripts/brand-isolation-check.sh

for f in spec/schemas/*.json; do npx --yes ajv-cli@5 compile -s "$f"; done

npx --yes ajv-cli@5 validate -s spec/schemas/vault-meta.json -d examples/sample-vault/.vault-meta.json

for md in examples/sample-vault/decisions/*.md; do
  node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
  npx --yes ajv-cli@5 validate -s spec/schemas/decision.json -d /tmp/fm.json
done

for md in examples/sample-vault/sessions/*.md; do
  node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
  npx --yes ajv-cli@5 validate -s spec/schemas/session.json -d /tmp/fm.json
done

for md in examples/sample-vault/patterns/*.md; do
  node .github/scripts/extract-frontmatter.js "$md" > /tmp/fm.json
  npx --yes ajv-cli@5 validate -s spec/schemas/pattern.json -d /tmp/fm.json
done
```

Expected: every command exits 0; brand-isolation prints "passed"; every validation prints "valid".

- [ ] **Step 3: Confirm the build plan is NOT in any commit**

```bash
git log --all --oneline -- tolvi-build-plan.md
```

Expected: empty output. The file is gitignored and was never tracked.

```bash
git status --ignored | grep tolvi-build-plan.md
```

Expected: `tolvi-build-plan.md` shown under "Ignored files".

- [ ] **Step 4: Push the branch (only after user confirms ready)**

> **Pause here.** Do not push without explicit user approval. The user has not yet authorized any push to `origin`. Surface the branch state and wait for go-ahead.

When approved:

```bash
git push -u origin feat/foundation
```

- [ ] **Step 5: Open the PR (only after user confirms)**

```bash
gh pr create --title "feat: foundation — repo skeleton, format spec, sample vault, CI guards" --body "$(cat <<'EOF'
## What this PR does

Bootstraps the repo to a state where Phase 2 (server) and Phase 3 (CLI) can begin in parallel against a locked, public format contract.

## Why

Establishing the foundation in a single, reviewable PR lets the rest of the project move with confidence: the format spec is locked, the sample vault validates against the schemas in CI, and the brand-isolation guard catches contributor leaks at PR time rather than after merge.

## How

Eleven atomic commits, each self-contained:

1. LICENSE (Apache 2.0), NOTICE, .gitignore, .editorconfig
2. README, CONTRIBUTING (with brand-isolation rule), CODE_OF_CONDUCT, SECURITY, NAMESPACE, ROADMAP
3. .github issue + PR templates, CODEOWNERS
4. Stub READMEs for cli/, server/, spec/, skills/, examples/
5. docs/ARCHITECTURE.md, CONVENTIONS.md, OPEN_QUESTIONS.md
6. docs/adr/ — ADR template + 0001 (architecture), 0002 (format contract)
7. spec/tolvi-format-v1.md — the public format contract
8. spec/schemas/ — JSON Schemas for vault-meta, decision, session, pattern
9. examples/sample-vault/ — 6 decisions, 3 sessions, 4 patterns (synthetic)
10. .github/workflows/validate.yml + helper scripts (frontmatter extractor, brand-isolation guard)
11. .pre-commit-config.yaml

This PR is exempt from the under-500-line guideline as bootstrapping content (see CONTRIBUTING.md, "PR size").

## Testing

- `pre-commit run --all-files` passes locally
- All four JSON Schemas compile under `ajv-cli`
- Every file in `examples/sample-vault/` validates against its schema
- The brand-isolation guard passes (NOTICE is the only file referencing the parent organization)
- The link checker finds no broken intra-repo links

## Checklist

- [x] PR is exempt from <500-line rule (bootstrapping content per CONTRIBUTING.md)
- [x] All new code has tests / verification steps
- [x] `pre-commit run --all-files` passes locally
- [x] No forbidden brand references introduced (CI verifies)
- [x] Updated relevant docs

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 6: Confirm CI is green**

After PR creation, monitor:

```bash
gh pr checks --watch
```

Expected: all checks pass green. If any fail, fix the underlying issue, push, and re-watch.

---

## Definition of done

- [ ] All 11 commits exist on `feat/foundation` in order
- [ ] `pre-commit run --all-files` passes locally
- [ ] Every sample vault file validates against its JSON Schema locally
- [ ] Brand-isolation guard passes locally
- [ ] PR is opened against `main` with the description above
- [ ] All CI checks (`validate.yml`) are green on the PR
- [ ] User reviews the PR and approves
- [ ] User merges (squash or merge-commit, user's choice — not auto-merged)

After merge:

- [ ] `feat/foundation` branch is deleted
- [ ] Phase 2 (server) and Phase 3 (CLI) brainstorm sessions can begin against the locked spec

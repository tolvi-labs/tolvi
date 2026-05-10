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

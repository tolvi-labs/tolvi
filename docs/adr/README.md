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
| 0003 | [Vault index and TL;DR system](./0003-vault-index-and-tldr-system.md) | accepted |

## Adding a new ADR

1. Pick the next sequential number
2. Copy the template above into `NNNN-short-slug.md`
3. Fill in all four sections — leave nothing as "TBD"
4. Add an entry to the index above
5. Commit in the same PR as the work the decision enables

## Superseding

When a decision is replaced, set the old ADR's status to `superseded by NNNN`, link to the replacement, and create the new ADR. The old file stays — it's part of the historical record.

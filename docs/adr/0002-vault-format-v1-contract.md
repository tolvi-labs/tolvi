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

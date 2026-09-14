---
tags: [decision, tolvi, vault, agents]
date: 2026-05-22
repo: tolvi
status: active
ticket: none
---

# Tolvi adopts the vault index + TL;DR convention for its own vault

## TL;DR

Adopt the vault index + TL;DR convention from public ADR 0003 for `tolvi-labs/tolvi`'s own internal vault. Pre-commit-style index regeneration recommended once the `tolvi vault-index` CLI subcommand ships; until then, regeneration is manual. Rejected: wait until OSS adopters validate the convention (we authored it; using it is the closure).

## Why

ADR 0003 ([`docs/adr/0003-vault-index-and-tldr-system.md`](../../docs/adr/0003-vault-index-and-tldr-system.md)) recommends the TL;DR-block-on-long-decisions and vault-index-in-agent-conventions conventions for any Tolvi vault. Tolvi's own internal vault under `tolvi-labs/tolvi/vault/` is itself a Tolvi vault — making the convention authoritative for outsiders while not using it ourselves would be inconsistent.

The benchmark numbers behind ADR 0003 (27× wall-clock speedup, 30–50% token reduction on decision-rationale queries) were measured on a reference implementation against a vault of similar size and shape. The same conventions should pay off here once the vault grows beyond a handful of decisions.

## How

**TL;DR blocks.** Backfill `## TL;DR` blocks on any decision in `vault/decisions/` larger than ~5 KB. As of this decision, no decision in this vault exceeds that threshold — the convention applies forward-looking. Authors writing new long decisions follow the format documented in [`docs/CONVENTIONS.md`](../../docs/CONVENTIONS.md) Section 7.

**Index regeneration.** The Phase 3.x roadmap item `tolvi vault-index` will ship the CLI tool that regenerates the index block in this repo's `CLAUDE.md`. Until that lands, regeneration is manual — done via the global agent-side skill that already exists in the maintainer's tooling. The agent-side skill is not part of this OSS repo and not a dependency for OSS adopters.

**Trigger.** Once `tolvi vault-index` ships, wire it into the existing `tolvi precommit install` flow (or as its own pre-commit hook entry) so the index never drifts past one commit.

## Outcome

Internal adoption documented. The vault index has not been generated against this repo's `CLAUDE.md` yet because (a) this vault has very few decisions today, and (b) the CLI subcommand that would automate it has not yet shipped. Both gates open naturally — adoption becomes mechanical when the CLI lands and the vault crosses ~10 decisions.

See also: [[adr-0003-vault-index-and-tldr-system]] (the public architectural record this decision implements for this project).

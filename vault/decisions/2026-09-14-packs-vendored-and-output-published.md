---
tags: [decision, tolvi, packs, integrations, json-contract]
date: 2026-09-14
repo: tolvi
status: active
ticket: none
user_impact: high
product_area: CLI / packs / Claude Code integration
---

# Packs are vendored rather than owned, and the CLI's output becomes a published contract

**Date:** 2026-09-14
**Repo:** tolvi

## TL;DR

Four things make this CLI hard to consume programmatically: nothing enumerates the repos that have vaults, every command prints for humans only, pack selection lives in a different product, and a cask-installed binary carries no Claude Code integration at all. Closing them by pulling packs into this repo would have cost `tolvi-solo` its core deliverable, which an active decision forbids. Packs stay owned by solo and are vendored here instead, pinned byte-identical by a parity check. That is the same mechanism the integration files need, because `go:embed` cannot traverse upward out of the `cli/` module, so one convention this repo already uses seven times covers both.

## Why

Everything this CLI emits is written for a person reading a terminal. `doctor` prints check lines, `vault-health` prints findings, and neither has a machine-readable form, so a CI job, an editor extension, or any other tool wanting to act on the result has to parse formatted text that changes whenever the wording improves. There is also no way to ask which repos on a machine have vaults at all: `roots.json` declares org and product roots by design and explicitly refuses to name repo roots, since those resolve from the working directory.

Packs are the sharper problem. They are the obvious thing for `tolvi init` to offer, and the obvious way to offer them is to move them here. `2026-06-05-tolvi-solo-as-separate-product` is active and rejects exactly that, naming both "embedding solo configs inside the tolvi monorepo" and "a trimmed fork of the tolvi Go CLI", on the grounds that the schema packs add value "without requiring users to install a compiled binary". Making packs a Go CLI feature undoes that rationale directly. The contradiction surfaced from the vault before any code was written, which is the point of reading it first.

## How

- **Packs are vendored, not owned.** `tolvi-solo` keeps ownership and gains `packs/<name>/pack.json` as the machine-readable source of truth, with its README table checked against it. This repo vendors `pack.json` plus `templates/` into the `cli/` module tree, pinned byte-identical by a new check modelled on `.github/scripts/schema-copy-parity-check.sh`: bidirectional, failing both when a source has no vendored copy and when a vendored copy has no source, with remediation naming solo as the origin. `tolvi init --pack` and `tolvi packs list --json` read the vendored copies.
- **Pack metadata had no machine-readable form at all.** A pack is exactly `README.md` plus `templates/*.md`; names, statuses and descriptions live only in prose tables, and solo's installer discovers packs by a bare `ls`. `pack.json` is new work in solo rather than a port, and it gives solo's own installer a registry to read instead of a directory listing.
- **Integration files are embedded, because the cask ships exactly one artifact.** `Casks/tolvi.rb` declares a single `binary "tolvi"` and nothing else: no directory artifact, no caveats, no postflight. Support files cannot ride alongside without changing the cask's shape, so `tolvi integrations install` embeds them. They are copied into the `cli/` module first, because `cli/internal/format/schemas.go` already documents the constraint that forces it: "Go's //go:embed cannot traverse upward from the source file's directory."
- **The command installs only what this repo owns.** Its own skill, hooks, and three slash commands. Stack skills stay out, per `2026-09-12-stack-skills-live-in-their-product-repos`, because solo installs them as symlinks into sibling checkouts and a binary cannot recreate that honestly. It detects sibling repos and reports what it found without wiring anything, keeping the discovery value without assuming a directory layout it cannot guarantee.
- **JSON output becomes a published contract.** Doctor, vault-health, repos-list and packs-list shapes get schemas in `spec/schemas/`, published through `@tolvi-labs/spec`, embedded and pinned by the existing schema-copy parity check. Consumers pin to something documented rather than to Go struct tags, and contract tests get an authoritative thing to assert against. Subject to the `$id` rule in `2026-05-14-jsonschema-ids-locked-tolvilabs`.
- **The registry is a cache, never a second routing rule.** `tolvi init` appends to a machine-local `~/.config/tolvi/repos.json`, mirroring how `roots.json` is declared and never committed. `tolvi repos list --json` returns per-repo summary computed by the resolver, so a consumer reads rather than re-derives. The cost of the alternative is on record: a fifth consumer of the routing rule re-derived it and had never indexed a routed session note.
- **Hook merging stays idempotent**, per `2026-08-03-installer-hook-merge-must-be-idempotent`, and the permission allowlist adds reads only. Solo's installer deliberately withholds `tolvi sync` and `tolvi commit` from the allowlist so writes stay a conscious per-call approval, and that choice carries over.
- **Prefer a behavioral check over asserting on source text.** Solo's strongest check provisions a throwaway git repo, runs the installer for real, and validates the vault it produced. Its comment records why: asserting on the heredoc's text would pass while the vault it makes stays unreadable. The pack and integration checks follow that shape.

## Resolved gaps

- *An active decision rejects folding packs into this repo. How is the contradiction resolved?* Honor the decision. Packs stay owned by solo and are vendored here, pinned by a parity check.
- *`packs list --json` has no data source. Where does pack metadata come from?* A new `pack.json` per pack in solo, with the README table checked against it.
- *An embedded binary cannot symlink stack skills. What does `integrations install` cover?* This repo's own integration only, plus detect-and-report for sibling checkouts.
- *Should the JSON shapes be published or internal?* Published and versioned in `spec/`, so the contract outlives any one consumer.

## Outcome

The CLI becomes consumable by tools rather than only by people, and no sibling product loses its identity to get there. What remains is additive: a registry file, four schemas, a `pack.json` per pack, two vendoring checks, and one new command.

## Amendment, 2026-09-23

A crucible on this work pressed two of the choices above. Both stand, with their reasoning recorded here so the same objections are not re-argued from scratch.

**Publishing four schemas, challenged against `2026-09-13-meta-v2-bumps-schema-version-not-id`.** That decision refused to buy stability for consumers that do not exist, and it verified the claim rather than assuming it. The objection was that publishing schemas for output nothing consumes yet repeats the mistake it warned about. It does not: that decision tested for a precondition, and the precondition is now met. A consumer is committed, with an approved design, and it reads all four shapes. The principle is intact and the situation is different, which is the distinction worth keeping. If the consumer were dropped, the schemas would become exactly the speculative stability that decision refuses.

**`tolvi init` writing machine-global state.** `init` is a per-repo command, and appending to `~/.config/tolvi/repos.json` makes it touch machine-wide state, which is a new failure surface for a command people already rely on. The contract: **the registry write is best-effort.** A failure warns, names `tolvi repos scan` as the repair, and never fails `init`. The vault is the deliverable; the registry is a cache over vaults that a full scan can rebuild from nothing. An `init` that succeeded at provisioning and then reported failure because an index could not be updated would be lying about what happened.

**The integration-tree duplication in the fourth bullet above was challenged and upheld, for a reason that changed mid-flight.** The crucible argued the copy into `cli/` was optional because relocation was available. It was, until `c7d3cad` shipped `.claude-plugin/plugin.json` and moved the tree to `skills/tolvi/`, which the plugin loader requires at the repo root. The duplication is now forced rather than chosen. See `2026-09-23-duplicate-and-pin-is-a-mitigation-not-a-default` for when that mitigation applies, and for the `go:embed` `all:` trap that the underscore-prefixed `_preflight.md` sets.

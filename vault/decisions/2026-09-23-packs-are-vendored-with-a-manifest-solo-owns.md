---
tags: [decision, tolvi, packs, tolvi-solo, vendoring]
date: 2026-09-23
repo: tolvi
status: active
ticket: none
user_impact: high
product_area: CLI / packs
---

# Packs get a manifest in tolvi-solo and are vendored here, so neither repo loses what it owns

**Date:** 2026-09-23
**Repo:** tolvi (with the manifest half in tolvi-solo)

## TL;DR

`tolvi init --pack <name>` and `tolvi packs list` now work from a binary with no checkout beside it. Packs stay owned by tolvi-solo, which gained `packs/<name>/pack.json` as their machine-readable source of truth; this repo vendors `pack.json` plus `templates/` and pins the copies with a parity check. This is the plan `2026-09-14-packs-vendored-and-output-published` set, carried out.

## Why

Pack selection is the obvious thing for `tolvi init` to offer and the obvious way to offer it was to move packs here, which an active decision forbids: `2026-06-05-tolvi-solo-as-separate-product` names "embedding solo configs inside the tolvi monorepo" as rejected, on the grounds that solo's packs add value without requiring a compiled binary. Vendoring gets the binary what it needs while leaving that intact.

A second problem sat underneath it. A pack was a directory of Markdown and nothing else: names, statuses and descriptions lived only in prose tables, and solo's installer discovered packs with a bare `ls`. Nothing could read a pack without parsing Markdown, so there was no honest thing to vendor.

## How

- **`packs/<name>/pack.json` in tolvi-solo is the source of truth**, declaring status, summary, verticals, and every template with what to use it for. It was generated from the README tables that already held that information, so the manifests started as a record of what was true rather than as new claims.
- **Solo's own check keeps the three in step**: a manifest, its templates on disk, and the README table, failing in both directions. A template added on disk, a template listed but gone, a status contradicting the table, and a table row for a pack that has no directory are all caught. Solo's installer now lists packs from the manifests instead of a directory listing.
- **This repo vendors `pack.json` and `templates/` only.** A pack's README is prose for people reading the solo repo, and the CLI has no use for it.
- **The copies are pinned bidirectionally** by `.github/scripts/pack-copy-parity-check.sh`: a pack shipped by solo and never vendored fails, as does a vendored pack solo no longer ships. It skips cleanly with no sibling checkout, so a contributor with only this repo is not blocked, and CI checks out both. This is duplicate-and-pin as `2026-09-23-duplicate-and-pin-is-a-mitigation-not-a-default` allows it: the duplication is forced by the binary needing files that another repo owns, and the check lives beside the copies.
- **An existing template is never replaced.** `Install` skips a file already in `vault/templates/` and does not report it as written, because people edit these and a re-run that replaced an edited template would lose work.
- **An unknown pack is reported, not rolled back.** The vault is provisioned before templates are written, so the error names the packs that exist and leaves the vault in place rather than deleting something a person may already be writing into.
- **Vendored trees are excluded from markdownlint.** A pinned copy cannot be linted here: any fix this repo's rules demanded would break the parity check against a repo with its own rules.
- **`packs list --json` is published** as `spec/schemas/packs-list.json` under the `outputs` key, validated from this repo's own tests.

## Outcome

A brew-installed `tolvi` can list six role packs and provision a vault with any of them, tolvi-solo still owns them, and drift between the two repos fails a check rather than reaching a user.

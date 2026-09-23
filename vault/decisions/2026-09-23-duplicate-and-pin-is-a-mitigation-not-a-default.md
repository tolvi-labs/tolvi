---
tags: [decision, tolvi, repo-structure, go-embed, conventions]
date: 2026-09-23
repo: tolvi
status: active
ticket: none
user_impact: none
product_area: Repo structure / build conventions
---

# Duplicate-and-pin is a mitigation for forced duplication, never a default

**Date:** 2026-09-23
**Repo:** tolvi

## TL;DR

This repo keeps two byte-identical copies of the same files in more than one place, each pinned by a check. That is a mitigation for duplication something else forces, not a pattern to reach for. It applies only when a published artifact or a toolchain constraint requires a file in a second location and no relocation removes the requirement. Where relocation is available, relocate. A pinned copy still rots when the check is the only thing holding it, so the check has to live with the copies.

## Why

A crucible on the `tolvi repos` and `integrations install` work challenged copying the Claude Code integration tree into `cli/` so `go:embed` can reach it, arguing the duplication was optional because nothing outside this repo needed the files where they sat. Relocation was the better answer and was accepted. Before it could be written, `c7d3cad` shipped `.claude-plugin/plugin.json` and moved the tree to `skills/tolvi/`, which the plugin loader requires at the repo root. Relocation became impossible and the duplication became forced, so the answer flipped back. The principle did not change; only which branch of it applied. Recording it stops the next reviewer from re-running that argument from scratch, and stops the flip being read as "duplication is fine here".

## How

- **The test is whether relocation removes the requirement.** Two locations are forced when an external consumer reads a published path, or when a toolchain cannot reach across the layout, and moving the file breaks the other consumer. If one location can serve both, there is no decision to make: move it.
- **The constraint that forces it here is the plugin manifest.** Claude Code's plugin loader requires `skills/` at the repo root, which is where `c7d3cad` put the integration tree. `cli/internal/format/schemas.go:20` records the other half: "Go's //go:embed cannot traverse upward from the source file's directory." A file cannot be at the repo root for the loader and inside `cli/` for the compiler at the same time, so the embedded copy is a copy.
- **A forced copy is pinned by a check, and the check ships with it.** `spec/schemas/` and `cli/internal/format/schemas/` are the worked example: `schema-copy-parity-check.sh` fails when either side drifts, and it runs in `npm run validate`. A new copy that cannot name the check that pins it is not ready to land.
- **The check has to be in the repo that has the copies.** `tolvi-solo/commands/_preflight.md:4` tells a reader to "re-run `.github/scripts/preflight-sync-check.sh`", and that script does not exist in tolvi-solo; it lives here. The convention rotted at the seam between two repos, which is precisely where a pinned copy is least likely to be checked.
- **Underscore-prefixed files need `all:`.** `go:embed` skips `_` and `.` prefixed names unless the pattern says `all:`, and `skills/tolvi/commands/_preflight.md` is deliberately underscore-prefixed. A bare `//go:embed` over that tree would omit it silently. It likely needs no embedding at all, since the preflight block is physically duplicated into each command file and `_preflight.md` is only the drift check's source, but the omission should be a decision rather than an accident.

## Outcome

Forced duplication is allowed, named, and pinned by a check that lives beside the copies; optional duplication is still refused, and relocation remains the first answer whenever it is available.

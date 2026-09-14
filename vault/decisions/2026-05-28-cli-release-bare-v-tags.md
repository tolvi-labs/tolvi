---
tags: [decision, tolvi]
date: 2026-05-28
repo: tolvi
status: active
ticket: none
user_impact: low
product_area: CLI release / distribution
---

# CLI releases use bare `v*` tags, not `cli-v*`

**Date:** 2026-05-28
**Repo:** tolvi

## Why

The first CLI release was about to be cut against a `cli-v0.1.0` tag, but that tag scheme silently breaks the release tooling. Picking the tag format wrong on a first release is costly to unwind, so the scheme was settled deliberately before tagging. This is a purely technical/release-engineering decision with only a minor user-facing edge (how a pinned `go install` resolves).

## How

- **The blocker:** OSS GoReleaser (v2.16.0, the `distribution: goreleaser` action) cannot parse a `cli-` prefix as a semantic version — it fails at "parsing tag" with `invalid semantic version`. Proven with a local dry-run against a throwaway `cli-v0.1.0` tag.
- **Rejected — `monorepo.tag_prefix`:** GoReleaser's prefix-strip feature is **Pro-only**; the OSS binary errors with `field monorepo not found in type config.Project`. Not viable for a free OSS release pipeline.
- **Rejected — `GORELEASER_CURRENT_TAG=${GITHUB_REF_NAME#cli-}` + `--skip=validate`:** this makes the tag parse, but GoReleaser then builds the GitHub release against the *stripped* name `v0.1.0` (creating a phantom `v0.1.0` tag alongside the pushed `cli-v0.1.0`), and `--skip=validate` drops the clean-tree/tag-on-commit guard. It also leaves the git tag (`cli-v0.1.0`) diverging from the binary's reported version (`v0.1.0`). Discarded as fighting the tooling to preserve a prefix GoReleaser discards anyway.
- **Chosen — bare `v*`:** changed the `cli-release.yml` trigger `cli-v*` → `v*` and tagged `v0.1.0`. GoReleaser parses it with zero hacks, the GitHub release attaches to the real pushed tag, and `-ldflags "-X main.version={{.Tag}}"` bakes `v0.1.0` into the binary. The SDK keeps its `sdk-v*` trigger; `v*` and `sdk-v*` globs do not collide (a `sdk-v*` tag starts with `s`).
- **Supplementary `cli/v0.1.0` tag:** the CLI is a Go submodule (`github.com/tolvi-labs/tolvi/cli`, no root `go.mod`), so Go's submodule versioning needs `cli/vX.Y.Z`-form tags for `go install .../cli/cmd/tolvi@v0.1.0` to resolve a pinned version. A bare `v0.1.0` maps (in Go's model) to a non-existent root module. So a `cli/v0.1.0` tag was pushed at the same commit purely for `go install` resolution; it starts with `c` and does not re-trigger the `v*` release workflow.
- **Known limitation:** `go install` builds carry no GoReleaser ldflags, so `tolvi version` prints `dev` for them regardless of tag; only the released archive binaries report `v0.1.0`.

## Outcome

`tolvi CLI v0.1.0` is published as a real GitHub release (5 platform archives + checksums) cut from a bare `v0.1.0` tag, with a companion `cli/v0.1.0` Go-module tag in place for pinned `go install` once the repo is public.

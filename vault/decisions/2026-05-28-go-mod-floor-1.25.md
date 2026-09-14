---
tags: [decision, tolvi]
date: 2026-05-28
repo: tolvi
status: active
ticket: none
user_impact: low
product_area: CLI build / distribution
---

# CLI go.mod floor set to Go 1.25.0 (down from 1.26.2)

**Date:** 2026-05-28
**Repo:** tolvi

## Why

The CLI declared a minimum Go version of 1.26.2 — the newest possible patch at the time — which would block anyone not on that exact-latest Go from building or `go install`ing the tool, with no actual reason. Lowering the bar widens who can adopt the CLI without changing any behavior. Purely a build/compatibility cleanup; no runtime surface.

## How

- A CI break (the `cli` workflow failing with `go.mod requires go >= 1.26.2 (running go 1.22; GOTOOLCHAIN=local)`) prompted auditing what actually requires 1.26.2.
- Dependency-floor analysis with `go list -m -f '{{.GoVersion}} {{.Path}}' all`: the highest *dependency* requirements are `golang.org/x/text` v0.37.0 and `golang.org/x/sync` v0.20.0 at **1.25.0**, then `github.com/anthropics/anthropic-sdk-go` v1.43.0 at **1.23.0**; everything else is lower. The `1.26.2` on the main module was arbitrary toolchain-set noise, not dependency-driven.
- **Chosen — `go 1.25.0`:** the real floor with zero dependency changes. It also matches Go's official two-release support window (1.25 + 1.26), so it only excludes already-EOL Go. Verified `go build`/`go vet`/`go test ./...` all green on a freshly downloaded real `go1.25.0` toolchain (not just the local 1.26.3).
- **Rejected — `1.23.0`:** reachable only by downgrading `x/text` + `x/sync`, for the marginal benefit of supporting EOL Go 1.23/1.24. Not worth the dependency churn.
- **Rejected — below 1.23:** would require downgrading the Anthropic SDK itself (the core dependency). Off the table.
- Paired with a CI change: all three `setup-go` steps now use `go-version-file: cli/go.mod` instead of a hardcoded `go-version`, so the workflows track this floor automatically and can't drift again. See [[ci-setup-go-version-from-go-mod]].

## Outcome

The CLI now builds and `go install`s on Go 1.25+ (down from 1.26.2-only), verified on a real `go1.25.0` toolchain, with CI reading the requirement straight from `go.mod`.

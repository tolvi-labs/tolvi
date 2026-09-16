---
tags: [decision, tolvi]
date: 2026-05-27
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: Release / Distribution
---

# Tolvi v0.1.0 launch scope — channels in, channels deferred

**Date:** 2026-05-27
**Repo:** tolvi

## Why

Phase 6 is a soft launch to a small invited cohort. For an audience that small and that close to the project, every additional distribution channel beyond the essentials adds setup work (new repos, new tokens, new external accounts) that delays the first ship without meaningfully changing whether those particular users can install. The decision picks the minimum channel set that satisfies the cohort's actual install habits and explicitly defers the rest to before the wider public launch (Phase 8 / Show HN), when impersonation defense and registry coverage start to matter materially.

## How

**In scope for v0.1.0:**

- **GitHub release binaries (CLI, via GoReleaser).** Baseline — already wired in `.github/workflows/cli-release.yml`. Non-negotiable.
- **`@tolvi-labs/sdk` on npm with provenance.** `.github/workflows/sdk-release.yml` is fully wired (tag `sdk-v*` → `npm publish --provenance --access public` + GitHub release). Marginal cost: creating the `tolvi-labs` npm org and setting the `NPM_TOKEN` repo secret. Local pre-flight confirmed `npm publish --dry-run` produces a valid 15.4 kB tarball with 47 files.
- **Homebrew tap.** `brew install tolvi-labs/tap/tolvi` is the install path Mac devs reach for first. GoReleaser generates the formula nearly for free given a tap repo and a PAT scoped to push to it. The friendly-user cohort skews Mac-dev, so the value-to-effort ratio is high.

**Deferred (not v0.1.0):**

- **Docker Hub server image.** Highest setup cost (Docker Hub org + access-token secret + a publish workflow that doesn't exist yet — `.github/workflows/server.yml` is build-only, no publish step) for the lowest payoff in this cohort. Self-hosters can build from `docker-compose.yml` and the existing Dockerfile in the meantime. Revisit before Phase 8 when self-host gets pitched publicly.
- **PyPI `tolvi==0.0.0` defensive squat** (`NAMESPACE.md` #5). There is no Python package in the product (CLI is Go, SDK is TypeScript). This is purely impersonation defense, and the risk window opens at Phase 8 visibility, not at a soft launch to known names.
- **npm-unscoped `tolvi` defensive squat** (`NAMESPACE.md` #6). The real package is the scoped `@tolvi-labs/sdk`. Cheap to claim later when going public; not on the v0.1.0 critical path.

**Rejected alternatives:**

- *"Full Phase 6 now" — all four channels at v0.1.0.* Rejected: the Docker Hub publish workflow doesn't exist, so this would block tagging on writing a workflow that benefits ~0 of the v0.1.0 cohort.
- *"Binaries only — defer npm SDK too."* Rejected: the SDK pipeline is already wired and locally green; deferring it would leave a working channel idle and ship a release with no programmatic-access story.
- *"Skip Homebrew, ship a `curl | sh` installer instead."* Rejected: `brew install` is what the cohort expects, GoReleaser's brews block is the documented path, and a bespoke installer is more code to maintain than a tap repo.

## Outcome

The launch runbook has a fixed channel list for v0.1.0 — GitHub binaries + npm SDK + Homebrew tap — and an explicit deferred list (Docker Hub, PyPI squat, npm-unscoped squat) scheduled for the pre-Phase-8 sweep. The deferred items stay tracked in `NAMESPACE.md` and the ROADMAP without entering the v0.1.0 critical path.

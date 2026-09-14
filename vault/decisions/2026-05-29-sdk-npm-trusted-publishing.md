---
tags: [decision, tolvi]
date: 2026-05-29
repo: tolvi
status: in-progress
ticket: none
user_impact: low
product_area: SDK / Release
---

# Publish the SDK to npm via Trusted Publishing (OIDC), bootstrapped by a manual first publish

**Date:** 2026-05-29
**Repo:** tolvi

## Why

The `@tolvi-labs/sdk` TypeScript client needs to ship to npm so people can install it. npm now warns against long-lived automation tokens in CI and steers toward Trusted Publishing, so we chose the path that leaves no publishing secret in the repo.

## How

- The release workflow `sdk-release.yml` was already wired: on a `sdk-v*` tag it verifies the tag matches `package.json` version, regenerates types and drift-checks `src/types.gen.ts`, typechecks/tests/builds, then `npm publish --provenance --access public`, then cuts a GitHub release. It already has `id-token: write`.
- Blockers found during recon: the `@tolvi-labs` npm scope does not exist yet (registry returns 404), no `NPM_TOKEN` secret is configured, and there are no `sdk-v*` tags.
- Decision: use Trusted Publishing (OIDC) rather than a stored `NPM_TOKEN`. With OIDC the workflow needs only `id-token: write` (already present), the `NPM_TOKEN`/`NODE_AUTH_TOKEN` is removed entirely, and provenance is generated automatically (the `--provenance` flag becomes unnecessary).
- The catch, unlike PyPI: npm cannot publish a package's *first* version over OIDC — trusted-publisher config lives on the package's settings page, which doesn't exist until the package has been published once. So `0.1.0` must be bootstrapped another way. We chose a one-time manual publish (`npm login` + `npm publish` from a machine) over a throwaway-token CI publish; the only cost is that `0.1.0` itself won't carry a provenance attestation. Rejected the temp-token route because it means storing the exact secret npm warns about, even briefly.
- Migration after bootstrap: configure the trusted publisher on npmjs.com (org `tolvi-labs`, repo `tolvi`, workflow `sdk-release.yml`, action `npm publish`); update `sdk-release.yml` to bump the runner to Node ≥ 22.14.0 and npm ≥ 11.5.1 (trusted-publishing minimums — Node 22 ships npm 10.x, so add `npm install -g npm@latest`), and drop the `NODE_AUTH_TOKEN` env. Every release after `0.1.0` then publishes over OIDC with automatic provenance and no secret.
- See [[2026-05-28-cli-release-bare-v-tags]] for the analogous CLI release tag scheme, and [[npm-trusted-publishing-oidc]] for the reusable pattern.

## Outcome

The release path is decided and the workflow is wired; the SDK is one manual bootstrap publish plus a trusted-publisher config away from OIDC-based releases with zero stored npm credentials.

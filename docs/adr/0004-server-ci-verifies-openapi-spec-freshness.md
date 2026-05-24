# 0004 — Server CI verifies `spec/openapi.json` freshness against route schemas

**Status:** accepted
**Date:** 2026-05-24

## Context

`spec/openapi.json` is the canonical wire contract for the Tolvi API — consumed by the TypeScript SDK's `openapi-typescript` type generator and any third-party client codegen. The file is generated from the live Fastify route schemas by `server/src/scripts/openapi-dump.ts` (exposed as `npm run openapi:dump`) and committed to the repo.

Until this decision, nothing in CI verified that the committed `spec/openapi.json` matched what would be regenerated from the current server route schemas. A contributor who modified a route's zod schema and forgot to re-run `openapi:dump` would silently desync the spec from reality. Downstream consumers — the SDK type generator most immediately — would then propagate the staleness as wrong types, broken contract tests, or runtime mismatches.

This gap was tracked as OPEN_QUESTIONS #11 from 2026-05-24 until this decision resolved it.

## Decision

Add a drift-check step to `.github/workflows/server.yml` that runs `npm run openapi:dump` and fails the workflow if `git diff --exit-code spec/openapi.json` reports a non-zero diff:

```yaml
- name: Regenerate spec/openapi.json from route schemas
  working-directory: server
  run: npm run openapi:dump
- name: Fail on spec drift
  run: |
    if ! git diff --exit-code spec/openapi.json; then
      echo "::error::spec/openapi.json is out of date with the live server route schemas."
      echo "Regenerate locally: (cd server && npm run openapi:dump) and commit the result."
      exit 1
    fi
```

The drift check runs immediately after `npm ci`, before typecheck and tests, so contributors get the fastest possible feedback when they've forgotten to regenerate.

The workflow's `paths` filter is extended to also trigger on `spec/openapi.json` changes, so that direct edits to the spec — which shouldn't happen, but might — also run the drift check.

## Consequences

**Positive:**

- A whole class of "spec is stale, types are wrong, things break later" failures is gone. Contributors get a loud CI failure with the exact command to fix.
- The SDK's own type-drift check (`.github/workflows/sdk.yml`) now has a sturdy upstream guarantee: the spec it consumes is fresh.
- The pattern is symmetric across server and SDK — both have a `regenerate-then-diff` step. The convention is portable to any future generator we add.

**Negative:**

- Adds ~10s to every server CI run (one `tsx` invocation + a `git diff`). Acceptable; the workflow is already minutes long.
- Couples the server workflow to the existence of `spec/openapi.json` at the repo root. The path is documented elsewhere as the canonical location; this just enforces it.
- Mitigation if `openapi:dump` becomes non-deterministic: investigate the generator. The fix should be in the generator, not in the CI check.

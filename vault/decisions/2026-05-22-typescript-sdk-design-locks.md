---
tags: [decision, tolvi]
date: 2026-05-22
repo: tolvi
status: in-progress
ticket: none
user_impact: medium
product_area: TypeScript SDK
---

# Phase 5.A TypeScript SDK design locks (mid-brainstorm)

**Date:** 2026-05-22
**Repo:** tolvi

## Why

The Phase 5.A TypeScript SDK brainstorm is mid-flight; three foundational design decisions are already locked. Capturing them here so the brainstorm can resume from a fresh context without re-litigating answered questions. Once the brainstorm completes, the full design lives in `docs/superpowers/specs/2026-05-22-typescript-sdk-design.md` (not yet written at sync time); this decision file is the bridge.

## How

**Status:** brainstorm in flight as of this sync. Three of ~6 expected clarifying questions answered; Section 1 of design presented; awaiting user approval to proceed with Section 2.

### Locked decisions (do NOT re-debate)

- **Code generation approach: hand-written client over auto-generated types.** Use `openapi-typescript` to generate type definitions only from `spec/openapi.json`. Hand-write a thin `Tolvi` class exposing resource-grouped methods (`client.documents.create`, `client.search.query`, etc.). Rejected alternatives: full auto-gen via `openapi-typescript-codegen` (ugly API surface like `postV1Documents`, bigger bundle); `openapi-fetch` typed-fetch wrapper (less SDK-shaped — users write URL paths as strings). The hand-written client per-route maintenance burden is acceptable at 8 current routes; adding a new method when the server grows is trivial.
- **Repo location: `sdk/` subdirectory in `tolvi-labs/tolvi` monorepo.** Alongside existing `cli/`, `server/`, `integrations/`. One repo, one CI pipeline (new `sdk` workflow alongside the existing `cli` / `server` / `validate`), one PR for SDK + server changes when they're coupled. Rejected: separate `tolvi-labs/sdk` or `tolvi-labs/typescript-sdk` repo — overhead not justified pre-1.0 / solo. Easy to extract later if a separate contributor surface is needed.
- **Runtime target: universal (Node 18+, Bun, Deno, modern browsers), ESM-only.** Uses only Web APIs (`fetch`, `AbortController`); no Node-specific deps. `engines.node >= 18` for built-in fetch. ESM-only is the modern default. Rejected: Node-only (misses browser/edge consumers); dual ESM+CJS (doubles build complexity for a CJS consumer base that doesn't exist yet); browser-only (irrelevant).

### Design decisions presented in Section 1 (awaiting approval, not yet locked but unlikely to change)

- **Two-layer architecture:** user code → hand-written `Tolvi` class with resource-grouped methods → auto-generated types from `spec/openapi.json` → `fetch()` → server.
- **Package name:** `@tolvi-labs/sdk` (per existing roadmap; no change).
- **Method surface:** resource-grouped: `client.documents.{create,list,get,delete}`, `client.sync.batch`, `client.repos.list`, `client.search.query`, `client.ask`. (The last one stays top-level since "ask" isn't really a REST resource.)
- **Auth:** bearer token in `Authorization` header, set once at construction: `new Tolvi({ apiKey, baseUrl })`. No default `baseUrl` — explicit since the server isn't hosted.
- **Errors:** typed error classes per documented status code (`TolviValidationError` for 400, `TolviNotFoundError` for 404, `TolviEmbeddingUnavailableError` for 503 from search/ask, etc.). Thrown on non-2xx.
- **AbortController support:** per-method `{ signal?: AbortSignal }` option for cancellation.
- **Types regeneration:** `npm run gen:types` script in `sdk/package.json` runs `openapi-typescript spec/openapi.json -o sdk/src/types.gen.ts`. CI runs the gen and asserts no drift (so SDK types never silently lag behind server changes).
- **Testing:** vitest + a mock `fetch` for unit tests of all client methods + a contract test that parses `spec/openapi.json` and asserts the SDK's `client.*` surface covers every documented operation.

### Out-of-v1 (explicit deferrals)

- **Streaming `ask` responses.** The CLI streams; the SDK buffers. Streaming SDK methods need a different shape (async iterator) and v1 keeps it simple. Add when users ask.
- **Auto-retry / backoff.** Lean client; users add `p-retry` or similar.
- **Pagination helpers.** `list` returns raw `{ documents, next_cursor: null }`; server-side pagination isn't even implemented yet.
- **CommonJS dual export.** ESM-only.
- **OpenAPI codegen for methods.** Types only — methods are hand-written.
- **Server-version capability negotiation.** README documents which server version each SDK release is tested against.
- **Built-in caching / memoization.** Users add their own layer.

### Resumption checklist

After clearing context, to resume:

1. Read `docs/superpowers/specs/2026-05-21-tolvi-precommit-hook-design.md` (most recent shipped spec — sets the brainstorm-skill cadence)
2. Read this decision file (the three locked answers)
3. Re-read the session log entry for 2026-05-22 (`vault/sessions/2026-05-22.md`)
4. Pick up brainstorming-skill flow at Section 2 (repo layout) → ... → Section N → write spec at `docs/superpowers/specs/2026-05-22-typescript-sdk-design.md` → invoke `superpowers:writing-plans`

## Outcome

Three SDK design decisions are locked and resumable from a fresh context. Brainstorm can pick up at design-presentation Section 2 without re-asking the codegen / repo-location / runtime-target questions. Full spec doc has not yet been written; this decision is its precursor.

## Update 2026-05-24 — brainstorm completed; spec + plan finalized; Task 1 shipped

The mid-brainstorm pause captured above has resolved. Brainstorm was resumed in a fresh context and walked Sections 2–8 with the user one section at a time. All design decisions are now finalized in the spec document below; this decision file is left in place for chronological traceability (the 2026-05-22 locks were the entry condition for the resumed brainstorm) but is no longer the active reference — read the spec for current state.

**Finalized artifacts:**

- **Design spec:** `docs/superpowers/specs/2026-05-22-typescript-sdk-design.md` (a local engineering artifact, not committed) — 687 lines, 11 sections (§0 summary through §10 cross-references). Covers repo layout, API surface, error hierarchy, type-gen + drift detection, testing strategy, CI + release flow, and docs scope. The three 2026-05-22 locks (hand-written client over generated types, `sdk/` subdirectory, universal ESM-only runtime) are recorded as §1 "Foundational decisions (locked)" and were not re-debated.

- **Implementation plan:** `docs/superpowers/plans/2026-05-22-typescript-sdk.md` (a local engineering artifact, not committed) — 2,726 lines, 18 TDD-shaped tasks + a final code review. Each task has exact file paths, real code in every step, verification commands with expected output, and ends with `git add` + `git status` (manual-commit convention; implementer subagents never run `git commit`).

- **Implementation status:** Task 1 (sdk/ scaffolding) shipped via subagent-driven development on 2026-05-24. Both stage reviews (spec compliance + code quality) approved. Tasks 2–18 + final review remain queued.

**Additions to Section 1 that were locked during the resumed brainstorm** (formally settled rather than "presented but unlikely to change"):

- Two-layer architecture (Tolvi class → generated types → fetch → server)
- Package name `@tolvi-labs/sdk`
- Eager resource construction (not lazy)
- No default request timeout (users opt in via AbortSignal.timeout)
- No auto-retry, no pagination helpers, no streaming (deferred per §3.5 / §7.6)
- Subclass-per-status error hierarchy: `TolviValidationError` (400), `TolviAuthError` (401), `TolviNotFoundError` (404), `TolviEmbeddingUnavailableError` (503), `TolviUnknownAPIError` (other) + `TolviConnectionError` + `TolviAbortError`. Wrap DOM `AbortError` as `TolviAbortError` for consistent `instanceof TolviError` catches. Defer `TolviRateLimitError` until the server actually rate-limits.
- `openapi-typescript ^7.x` with `--alphabetize`; `src/types.gen.ts` is **committed to git**, not gitignored. CI drift check runs `npm run gen:types && git diff --exit-code`.
- vitest + mock fetch for unit tests (one file per resource); contract test parses `spec/openapi.json` and asserts SDK covers every documented op via a maintained mapping table with an `intentionallyUnexposed` allowlist for `/healthz` + `/readyz`. Defer integration-against-real-server tests for v0.1.0.
- Two GitHub Actions workflows mirroring the existing `cli.yml` / `cli-release.yml` split: `sdk.yml` (push/PR CI with drift detection) and `sdk-release.yml` (`sdk-v*` tags → npm publish with `--provenance --access public` via OIDC + GitHub release).
- `tsc` only build, no bundler.
- JSDoc + `@example` block on every public method; defer `sdk/examples/` directory for v0.1.0.

**New constraints captured to OPEN_QUESTIONS during the brainstorm:**

- #10: `gen:types` reads `../spec/openapi.json` so the `sdk/` package is monorepo-coupled. Mitigation if extracted: download a release artifact or fetch `/openapi.json` from a running server. README MUST note the monorepo dependency.
- #11: server CI doesn't verify `spec/openapi.json` freshness against route schemas — separate ticket needed.

**Status stays `in-progress`** until v0.1.0 ships (Tasks 2–18 + final review + first `sdk-v0.1.0` tag). At that point flip to `active` or `historical` and unlink the "active reference" deprecation note above.

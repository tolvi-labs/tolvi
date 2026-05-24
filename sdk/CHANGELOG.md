# Changelog

All notable changes to `@tolvi-labs/sdk` are documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] — 2026-05-24

### Added

- Initial release. Hand-written `Tolvi` client over `openapi-typescript`-generated types.
- Resource-grouped methods covering all v1 data-plane endpoints: `client.documents.{create,list,get,delete}`, `client.sync.batch`, `client.repos.list`, `client.search.query`, and top-level `client.ask`.
- Typed error hierarchy: `TolviError` (abstract base), `TolviAPIError` with five status-coded subclasses (`TolviValidationError`/400, `TolviAuthError`/401, `TolviNotFoundError`/404, `TolviEmbeddingUnavailableError`/503, `TolviUnknownAPIError`/other), `TolviConnectionError` (pre-response failures), `TolviAbortError` (signal aborted).
- `ErrorEnvelope` forward-compat hatch: every `TolviAPIError` exposes the full server envelope via `.body`, so future server-added fields (e.g. `error.details`) are accessible without an SDK update.
- Universal runtime: ESM-only; Node 18+, Bun, Deno, modern browsers. Zero runtime dependencies — depends only on platform `fetch` + `AbortController`.
- `Tolvi` constructor accepts a `fetch` override for testability and a `userAgent` override (default `tolvi-sdk/${VERSION}`).
- `AbortSignal` forwarding on every method via `{ signal?: AbortSignal }` option. No default timeout — opt in via `AbortSignal.timeout(ms)`.
- `x-request-id` extraction from response headers, surfaced on `TolviAPIError.requestId`.
- 16 public type aliases re-exported from the package root (e.g. `CreateDocumentRequest`, `Document`, `FullDocument`, `SyncBatchResponse`, `SearchResult`, `AskResponse`, `Citation`).
- JSDoc + `@example` blocks on every public method.
- Vitest unit tests with mock `fetch` (44 tests) plus a contract test that parses `spec/openapi.json` and asserts the SDK covers every documented operation (3 tests).
- `sdk.yml` CI workflow with type-drift detection. `sdk-release.yml` publish workflow with npm provenance via OIDC, triggered on `sdk-v*` tags.

### Tested against

- `tolvi-server v0.1.0`

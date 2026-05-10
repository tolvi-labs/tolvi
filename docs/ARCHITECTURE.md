# Tolvi architecture

Tolvi is a vault-format, a CLI, and a server that share a single contract. Engineers write decisions, sessions, and patterns into a per-repo `vault/` directory. The CLI indexes that vault locally for `recall` and `ask`. The server provides the same retrieval over HTTP for multi-repo and team use. This document describes how those pieces fit together, where the trust boundaries are, and what is intentionally out of scope for v1.

## The three layers

```
┌─────────────────────────────────────────────────────────────┐
│  Source: per-repo vaults (in git)                           │
│    <repo>/vault/{decisions,sessions,patterns}/              │
│    <repo>/vault/.vault-meta.json                            │
└──────────────────┬──────────────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
        ▼                     ▼
┌─────────────────┐   ┌─────────────────────────────────────┐
│ Local arm:      │   │ Server arm (self-hostable):         │
│  tolvi CLI      │   │  Fastify + Postgres + pgvector      │
│  + sqlite-vec   │   │  Docker compose                     │
│                 │   │                                     │
│  Single-binary  │   │  POST /v1/sync (CLI publish target) │
│  Go, Ollama     │   │  POST /v1/search, /v1/ask           │
│  embedding      │   │  API key auth per workspace         │
│  default        │   │                                     │
└─────────────────┘   └─────────────────────────────────────┘
```

The source layer is plain markdown in git. Both arms read from it. Neither arm owns it.

## 1. System surfaces and ownership

The CLI owns local capture and the local index. `tolvi sync` writes a doc into the repo's `vault/` directory and updates the embedded sqlite-vec index. `tolvi recall` and `tolvi ask` read from that local index. The CLI never requires a server to be useful — local-only operation is a first-class mode.

The server owns the multi-tenant index and the HTTP API. It accepts documents over `POST /v1/sync`, chunks and embeds them, stores them in Postgres with pgvector, and serves `POST /v1/search` and `POST /v1/ask`. It does not write to anyone's `vault/` directory; the CLI (or CI, or any other client implementing the format) is the source of writes.

The format spec is the only contract crossing the boundary. Both arms implement parsing and validation against `spec/tolvi-format-v1.md`. There is no shared parsing library. (See section 4.)

## 2. Trust and auth model

Authentication in v1 is a per-workspace API key. The server stores keys hashed at rest using either `pgcrypto` or argon2. Keys are scoped to two operations: ingest (writes via `/v1/sync`) and search (reads via `/v1/search` and `/v1/ask`). There are no user accounts, no roles, and no per-doc permissions in v1.

The CLI sends the key as `Authorization: Bearer <key>` on every server call. Local-only operation requires no key — there is no anonymous bind to the local index, because the local index lives on the engineer's filesystem under their own user account. The trust boundary is the OS, not the application.

Workspace isolation in shared Postgres is an open question (see [`OPEN_QUESTIONS.md`](./OPEN_QUESTIONS.md)). The current default is per-row `workspace_id` filtering enforced in the query layer.

## 3. Data flow

A typical capture-and-query cycle:

1. An engineer (or an agent) calls `tolvi sync` with a doc body and frontmatter.
2. The CLI writes the markdown file to `<repo>/vault/<type>/<filename>.md` and computes a `content_hash` over the file's bytes.
3. The CLI updates the local sqlite-vec index. Embedding runs against Ollama by default.
4. On `tolvi publish` (or via a scheduled CI sync), the doc is `POST`ed to `/v1/sync` along with `(repo_id, path, content_hash)`.
5. The server treats `(repo_id, path, content_hash)` as the idempotency key — the same triple submitted twice is a no-op.
6. The server chunks the doc, embeds each chunk, and stores chunks + vectors in Postgres + pgvector.
7. Subsequent `/v1/search` and `/v1/ask` calls see the new doc immediately. Index updates are read-after-write consistent within a single API call.

If the local embed step fails (Ollama down), `tolvi sync` still writes the file — the writes-first principle. The local index is allowed to drift; `tolvi doctor` is the recovery mechanism.

## 4. Component boundaries

The CLI is Go, single static binary, cross-platform. The server is TypeScript on Fastify, Node 20+. They do not share a parsing library. Each implements `tolvi-format-v1` parsing and validation in its own language, against the published JSON Schemas under `spec/schemas/`.

This is a deliberate trade-off. A shared library written in either language would constrain the other to embed a runtime, and would push toward a polyglot toolchain in every consumer. Code duplication of a small, well-specified parser is the smaller cost. The spec — including the JSON Schemas — is the contract.

The same boundary applies to future SDKs and third-party clients. Anything that conforms to `tolvi-format-v1` is a valid producer. Anything that can read JSON Schema + markdown frontmatter is a valid consumer.

## 5. What's deferred to Phase 9

The following are explicitly not in v1:

- A multi-arm shared library or generated parser.
- A web dashboard for browsing or editing vaults.
- OIDC, SSO, or any user-account auth model.
- Billing, quotas, or commercial metering.
- OpenAPI-generated SDKs in additional languages.

The aggregator pattern (a unified per-engineer view across multiple repo vaults) is documented in [`CONVENTIONS.md`](./CONVENTIONS.md) but is a manual `mkdir + ln -s` recipe in v1. A `tolvi unify` command to automate it is on the v1.x roadmap.

## 6. Self-host story

The intended self-host path is `docker compose up`. The compose file brings up Postgres with the pgvector extension and the Tolvi server, and exposes the API on `http://localhost:3000`. The CLI defaults to that address, so a self-hosted setup requires no CLI configuration beyond an API key.

The end-to-end self-host walkthrough — from `git clone` through first `/v1/ask` response — lands as a deliverable in Phase 2 alongside the server itself. The server is built to run in any environment that can host a Postgres instance and a Node process; there is no dependency on a specific cloud provider or managed service.

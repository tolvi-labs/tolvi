# Tolvi Server

> **Status:** Phase 2 (not yet shipped). Track progress in [`ROADMAP.md`](../ROADMAP.md).

This directory will hold the `tolvi` HTTP server — TypeScript on Node 20, Fastify, Postgres + pgvector, multi-tenant via API keys at the workspace level.

## Planned API surface

- `POST /v1/documents` — ingest a vault doc (idempotent on `content_hash`)
- `GET /v1/documents` — list with filters
- `GET /v1/documents/:id` — fetch one
- `DELETE /v1/documents/:id` — soft delete
- `POST /v1/search` — semantic + filter search; returns ranked chunks with citations
- `POST /v1/ask` — search with LLM synthesis
- `POST /v1/sync` — batch ingest from CLI publish
- `GET /v1/repos` — list a workspace's repos

OpenAPI spec will be generated from Fastify route definitions and published to [`../spec/`](../spec/).

## Self-hosting

A `docker-compose.yml` will bring up Postgres (with pgvector) and the server in one command. See [`../docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) for the server-arm component in context.

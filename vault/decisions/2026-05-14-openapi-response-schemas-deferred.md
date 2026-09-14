---
tags: [decision, tolvi]
date: 2026-05-14
repo: tolvi
status: active
ticket: none
user_impact: low
product_area: API Contract
---

# OpenAPI response schemas deferred — v1 follow-up before SDK generation

**Date:** 2026-05-14
**Repo:** tolvi

## Why

PR B ships `spec/openapi.json`, generated from the live Fastify route schemas and committed to the repo. The intent is for the OpenAPI document to be the consumable contract for SDK generation (Phase 5 TypeScript SDK) and any third-party client. Today, the generated document describes every endpoint's **request shape** correctly but emits empty response schemas (`"responses": { "200": { "description": "Default Response" } }`) on every v1 data-plane route. This is a partial contract — useful for "what endpoints exist and what do they accept" but useless for "what do they return."

## How

- **Why it's empty:** `fastify-type-provider-zod`'s `jsonSchemaTransform` derives the OpenAPI schema from each route's zod schema declarations. Tasks 14 (documents), 22 (search), and 24 (ask) declared `schema: { body: SomeZodSchema }` but **not** `schema: { response: { 200: SomeZodSchema } }`. With no response zod schema declared, the transform has nothing to emit — so the response section gets a default placeholder.
- **Affected routes:** `/v1/documents` (POST, GET, GET /:id, DELETE), `/v1/sync` (POST), `/v1/repos` (GET), `/v1/search` (POST), `/v1/ask` (POST). Health routes (`/healthz`, `/readyz`) declare inline response shapes and document correctly.
- **Why deferred:** authoring response zod schemas for each route is significant scope — ~5 endpoints × non-trivial response shapes (e.g., `/v1/ask` returns answer + citations[] + search_results[] + tokens). Doing it as part of Task 25 would have ballooned that task and re-opened the route files committed in earlier tasks of PR B. Better to land PR B and add response schemas as a focused follow-up.
- **Tracked in code:** comment block at the top of `server/src/scripts/openapi-dump.ts` documents the limitation. Commit message for Task 25 also calls it out.
- **What gets the v1 SDK unblocked:** add `response: { 200: ResponseSchema }` to every route's `schema` declaration, regenerate `spec/openapi.json`, commit. Probably one focused PR. No runtime behavior change — only the generated document gets richer.
- **Status is `in-progress`** rather than `active` because the decision *not* to do the work is itself a placeholder — the work needs to happen before Phase 5.

## Outcome

`spec/openapi.json` is committed and describes the API surface from the request side; response shapes are knowingly absent and documented as such. SDK generation is blocked on this follow-up; the rest of the v1 contract is solid.

**Update 2026-05-22:** Follow-up shipped. `server/src/routes/_responses.ts` extracted shared response primitives (`ErrorEnvelope`, `IsoTimestamp`, `HeadingPath`); response zod schemas added to documents (4 routes), sync, repos, search, and ask. `spec/openapi.json` grew 482 → 1,324 lines; all 8 data-plane endpoints now have populated response content. SDK generation unblocked.

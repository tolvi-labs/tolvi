// Shared response shapes for the v1 HTTP API.
//
// fastify-type-provider-zod uses these for two things:
//   1. OpenAPI generation (spec/openapi.json) — populates the response
//      side of each path entry. Without these, generated specs only
//      describe requests, blocking SDK generation.
//   2. Response serialization — fields not in the schema are dropped
//      from the JSON output. The schemas below mirror what each route
//      actually returns; missing a field is a contract bug.

import { z } from 'zod';

// ErrorEnvelope — the canonical 4xx/5xx response shape across the API.
// Routes return `{ error: { code, message } }` on documented failures
// (404 not_found, 503 embedding_unavailable / llm_unavailable, etc.).
export const ErrorEnvelope = z.object({
  error: z.object({
    code: z.string(),
    message: z.string(),
  }),
});

// IsoTimestamp — Postgres TIMESTAMP columns surface as ISO 8601 strings
// at JSON-serialization time. Allow both string and Date so the schema
// works whether the route returns a Date object or a pre-serialized
// string (fastify converts Date → string during reply.send).
export const IsoTimestamp = z.union([z.string(), z.date()]);

// HeadingPath — chunk.heading_path is a Postgres text[] column that can
// be null when the chunk isn't inside a heading-bounded section.
export const HeadingPath = z.array(z.string()).nullable();

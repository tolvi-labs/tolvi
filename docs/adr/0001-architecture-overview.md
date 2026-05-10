# 0001 — Architecture overview

**Status:** accepted
**Date:** 2026-05-09

## Context

Tolvi is a developer knowledge tool that captures engineering decisions where they happen and surfaces them via natural-language search from CLI or web. The product needs to support both:

- A local-first workflow (engineer at their machine, capturing and querying a single repo's vault), and
- A team workflow (multiple engineers across multiple repos, querying a shared index hosted on infrastructure they control)

Several architectures could satisfy this:

1. **Single-arm cloud-only:** all queries go through a hosted server. Simple, but breaks the local-first promise and creates a hard dependency on the network.
2. **Single-arm local-only:** every engineer indexes locally; no server. Works for individuals but not teams.
3. **Two-arm with a shared library:** local CLI and server share an embedding/chunking library, both written in the same language. Cleanest in theory, but the language choice forces a real loss in one arm — picking Go gives up the TypeScript server's ecosystem fit (Fastify, Drizzle, pgvector clients, npm SDK story); picking TypeScript gives up the CLI's single-static-binary distribution and Node-runtime-free install.
4. **Two-arm with a shared spec, no shared library:** local CLI and server are independent implementations of the same vault format spec. Different languages allowed.

## Decision

Adopt option **4: two-arm with a shared format spec**.

- The CLI is written in Go and ships as a single static binary
- The server is written in TypeScript on Fastify, with Postgres + pgvector for the index
- Both arms parse and validate `tolvi-format-v1` independently
- The spec at `/spec/tolvi-format-v1.md` is the only artifact crossing the language boundary

## Consequences

**Positive:**

- The CLI is genuinely standalone — no Node.js runtime needed for local users
- The server can be optimized for its workload (Postgres-native indexing, pgvector for search) without compromising the CLI binary's size or startup time
- Adding a Python implementation, a Rust implementation, or a third-party reimplementation costs nothing in either arm — they all read the same spec
- The format spec gets the level of rigor it deserves — it's a public contract, not an implementation detail

**Negative:**

- Code duplication: chunking, embedding-input normalization, frontmatter parsing, and wiki-link resolution all exist twice
- Risk of drift: a bug fix or behavior change in one arm must be ported to the other
- Partial mitigation: the JSON Schemas in `/spec/schemas/` catch frontmatter-level drift — both arms parse against the same schemas in CI. Chunking and embedding-input parity is a residual risk in v1 with no automated guard; it's tracked as an open question for v1.x (likely solved with golden-output tests against `/examples/sample-vault/`).

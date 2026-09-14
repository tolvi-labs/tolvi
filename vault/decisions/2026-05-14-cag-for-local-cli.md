---
tags: [decision, tolvi]
date: 2026-05-14
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: CLI / Local Dev
---

# CAG (Context-Augmented Generation) for local CLI; RAG stays cloud-only

**Date:** 2026-05-14
**Repo:** tolvi

## Why

The original Tolvi plan called for a local engineer-facing CLI that ran the same retrieval-augmented setup as the cloud server: local sqlite-vec index, local Ollama for embeddings, local LLM optional. That's a lot of machinery to ask an engineer to set up before they can ask their own vault a question. For typical engineering vaults (hundreds of docs, ~tens of thousands of tokens), the entire vault fits comfortably inside a modern frontier-model context window. The simplicity win is enormous: zero local setup beyond an API key, and a working `tolvi ask <query>` on day one.

## How

- **Local CLI (Phase 3) uses CAG:** walks `<repo>/vault/` at query time, concatenates every `.md` file, sends the whole vault as a cached system-prompt block alongside the user query in a single Anthropic API call. Response comes back with citations the LLM already produced; CLI just renders them.
- **Cloud server (Phase 2) stays RAG:** unchanged from PR A/PR B. pgvector, Ollama embeddings, in-SQL ranking, citation verification at the route layer. This split is intentional — the cloud arm scales past the context window for teams with large vaults; the CLI is the wedge for individual engineers and small teams.
- **No format-spec change.** `tolvi-format-v1` is storage-shape (frontmatter, slug, status, body, wiki-links). It doesn't care whether the consumer indexes with embeddings or stuffs into context. Same vault works for both arms.
- **Three sub-decisions locked at the same time (so Phase 3 brainstorm doesn't re-litigate them):**
  1. CAG strategy: whole vault → context, every query. Not selective TOC-then-fetch, not hybrid. Simpler is better while it fits the window.
  2. Local LLM: Anthropic API via user-supplied `ANTHROPIC_API_KEY` (env or `~/.config/tolvi/config.yaml`). No Ollama fallback in v1.
  3. Cross-cutting impact: none. Server arm unchanged.
- **What this drops from the original Phase 3 plan:** sqlite-vec local index, local Ollama embedding dependency, local-index build/refresh tooling, the implied `tolvi index` command. The CLI dep surface shrinks to the Anthropic SDK + a file walker.
- **What this adds:** Anthropic prompt-caching wiring (cache the system-prompt vault block; per-query message is small + uncached, so repeat queries within the 5-minute cache window pay ~10% of full input cost), and a context-window guard that warns or errors if total vault tokens exceed a safe threshold (~150K to leave headroom for the answer).
- **What stays:** `tolvi sync` (write a session/decision/pattern with frontmatter), `tolvi recall` (local lexical search — substring/regex, no embeddings needed), `tolvi status` (lifecycle helpers), `tolvi publish` (POSTs to server `/v1/sync` if a team is using the cloud arm).
- **Rejected alternative — selective context filtering (LLM picks docs to load based on a TOC):** added latency and an extra LLM round-trip for marginal token savings at the scales we care about. If a vault outgrows 150K tokens we tell the engineer to use the cloud arm; we don't build a half-RAG locally.
- **Rejected alternative — local Ollama embedding + sqlite-vec:** the simplicity loss isn't worth it for v1. Every install is a 270MB-plus model pull, every machine needs an Ollama service running. If users want offline-and-private, they can self-host the server arm.
- **Known limitation:** no offline support in CLI v1. Goes online for every query. Acceptable for the wedge use case (engineers with internet); revisit if real demand surfaces.

## Outcome

The Phase 3 CLI design just got substantially simpler — one external dependency (Anthropic), zero local services, a single `tolvi ask` command that works the moment `ANTHROPIC_API_KEY` is set. The two-arm split (local CAG, cloud RAG) is now the canonical architecture; Phase 3 brainstorm starts from this premise without re-debate.

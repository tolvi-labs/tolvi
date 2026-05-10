# Open questions

Unresolved design questions. Some carry forward from the work that informed Tolvi's initial shape; others are Tolvi-specific and surfaced during foundation design. None block v1, but each will need an answer before the corresponding code lands.

1. **Loud-fail on Ollama down.** `tolvi sync` succeeds even if the local embedding step fails — the writes-first principle keeps capture robust. The open question is whether `tolvi doctor` should warn loudly when Ollama is unreachable so that users learn quickly that the local index is drifting, rather than discovering it the next time `tolvi recall` returns stale results.

2. **Cross-repo type sharing.** When vault frontmatter schemas drift across consumer repos (for example, one repo adds a custom `x-*` field that another doesn't recognize), how is that resolved? The format spec freeze helps, but tooling like `tolvi lint --cross-repo` to detect drift across an aggregator view is deferred.

3. **Vault content audit cadence.** There is no formal review cycle today for marking decisions superseded when reality has moved on. Should `tolvi doctor` flag decisions older than N months with no inbound or outbound link activity, as a prompt to triage them?

4. **Index size growth.** Production observations from prior reference implementations suggest ~6.5 MB and ~280 docs is comfortable for in-memory loading. What's the threshold where the chunking and storage strategy needs to change — and should the CLI warn before reaching it?

5. **Secrets pre-commit lint.** Vault content can leak credentials when engineers paste from terminals or logs. Should Tolvi ship a `pre-commit` hook that scans staged vault files for API keys, tokens, and high-entropy strings before they reach git?

6. **Multi-tenant isolation in Phase 2.** What's the boundary between workspaces sharing a single Postgres instance? Per-row `workspace_id` filtering in the query layer is the simplest answer; row-level security pushes the check to the database; per-workspace schema is the strongest isolation but the most operationally heavy. Each has trade-offs and the choice should be made before the server code lands.

7. **Aggregator automation.** `tolvi unify` is on the roadmap but deferred from v1. When does it ship — driven by user pain, by a specific milestone, or by the maintainer reaching the threshold of repeating the manual recipe?

8. **CLI ↔ server format-version handshake.** The format spec mentions that implementations declare supported versions. What does the handshake actually look like on the wire — a header, a field on every request, a one-time `/v1/capabilities` exchange? And what happens when the CLI supports a newer format version than the server?

---

This file is meant to grow. When you find a non-obvious open question while working on Tolvi, add it here. When a question gets resolved, capture the answer as an Architecture Decision Record under [`./adr/`](./adr/) — the ADR template and existing decisions live there — then remove the question from this file in the same PR.

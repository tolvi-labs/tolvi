# 0003 — Vault index and TL;DR system

**Status:** accepted
**Date:** 2026-05-22

## Context

A coding agent answering "why was X decided?" defaults to reading source code. For decision-rationale queries this fails three ways:

1. **Wasted discovery cost.** Each session burns thousands of tokens on filesystem-level search cycles (`grep`, `find`, file listings) before reaching the decision doc that actually has the answer.
2. **Wrong-but-plausible answers.** Source code carries the *what* but not the *why*. Without the decision doc in context, the agent credibly proposes patterns that the team has already rejected.
3. **Long reads when short would do.** A 30 KB decision doc costs roughly 7,500 tokens to read in full; a 500-byte summary costs roughly 125. Most "what was decided" questions terminate cleanly on the summary.

A Tolvi vault is the right corpus to fix all three — decision docs already exist and already capture the rationale. The gap is that nothing tells the agent the vault is there, and nothing surfaces a cheap-to-read summary so the agent can decide whether to open the full file.

A reference implementation tested two complementary additions: a per-vault index injected into the agent's conventions file, and a short structured summary at the top of long decision docs. A 3-arm benchmark on decision-rationale queries measured a 27× wall-clock speedup and a 30–50% reduction in tokens consumed per query, with no loss of answer quality. The mechanism is small, portable, and additive to `tolvi-format-v1` — no breaking changes to the format itself.

## Decision

Recommend two complementary conventions for Tolvi vaults. Neither is normative in `tolvi-format-v1`; both are documented as contributor-facing guidance in [`docs/CONVENTIONS.md`](../CONVENTIONS.md).

### 1. TL;DR block on long decision docs

A short summary block placed immediately after the YAML frontmatter and before the body. Required on docs over roughly 5 KB; optional on shorter docs.

```markdown
---
tags: [decision, ...]
status: active
...
---

## TL;DR

<Decision in one sentence>. <Why, ≤120 chars>. Rejected: <alternative(s), ≤120 chars>.
```

If no alternatives were considered, the block ends with `Rejected: (none documented).` — the literal marker stands in for the absent content so its absence is visible to readers and tools, rather than ambiguous.

Total budget: ≤500 bytes including the heading. The discipline is the point — when a TL;DR cannot be written in 500 bytes, the decision under-specified its rationale and is worth re-examining.

### 2. Auto-generated vault index injected into agent conventions

Each agent integration loads a per-tool conventions file at session start — `CLAUDE.md` for Claude Code, `.cursorrules` for Cursor, `CONVENTIONS.md` for Aider, and so on. A delimited block listing every decision and pattern in the vault is injected into that file:

```markdown
<!-- VAULT-INDEX:START repo=<name> generated=<ISO date> -->
## Vault Index

### Decisions

- `2026-05-09-vault-format-v1-contract.md` — Vault format v1 contract
- `2026-05-22-vault-index-and-tldr-system.md` — Vault index and TL;DR system
- ...

### Patterns

- `slug.md` — Pattern title from the doc
- ...
<!-- VAULT-INDEX:END -->
```

The block is generated from each doc's filename plus its top-of-body title (H1 if present; first H2 or filename slug otherwise). It is idempotent — re-running the generator replaces the contents between the markers, or appends the block if absent.

The agent reads the index out of the conventions file (which is already loaded into context every session, so the marginal cost is zero on cache hit) and resolves "where is the decision about X" to an exact filename without filesystem search. The index sits in prompt cache, so the per-session overhead amortizes to roughly 10% of nominal after the first turn.

### Regeneration

The generator is a tooling concern, not a format concern. Reference implementations exist as agent slash commands and shell scripts; a future `tolvi vault-index` CLI subcommand is on the roadmap so OSS adopters have an in-tree tool that does not depend on any specific agent platform. The regenerator MUST be run on session-end (or as part of pre-commit, depending on workflow) — without a refresh trigger, the index rots the moment a new doc lands.

### Alternatives considered

- **Auto-summarize at query time.** Have the agent open every decision doc and summarize on demand. Rejected: defeats the cost argument — the agent still pays to read each candidate doc. A pre-computed summary is the entire point.
- **Frontmatter `summary:` field.** Move the TL;DR into structured frontmatter. Rejected: long-form rationale doesn't sit cleanly in YAML, and the markdown body keeps the doc readable in any markdown editor without specialized tooling. The `## TL;DR` block is a *body* convention, not a frontmatter one.
- **Vector-only retrieval (no index).** Rely entirely on semantic search to surface the right doc. Rejected: works above roughly 1,000 chunks where naive keyword search starts to fail, but at typical vault sizes (≤300 docs) keyword/index lookup is competitive and far cheaper. Semantic search remains useful as a complement, not a replacement.
- **Per-doc title-page metadata file.** A sidecar `.toml` per decision listing title, summary, status. Rejected: doubles file count, splits the source of truth, and adds a step every author must do. The TL;DR block keeps everything in one file.

## Consequences

**Positive:**

- Decision-rationale queries terminate on summary content, not full-doc reads — measured 27× wall-clock speedup and 30–50% token reduction on a reference benchmark.
- The agent stops proposing already-rejected patterns, because the rejected alternatives appear in the TL;DR block of the relevant doc.
- The format spec is unchanged — `tolvi-format-v1` vaults remain conformant without TL;DR blocks or an index. Adoption is per-vault and per-agent.
- The convention is portable across agent platforms; it depends only on each agent's habit of loading a conventions file at session start.

**Negative:**

- Authors carry a small additional discipline (write the TL;DR; keep it under 500 bytes). The 5 KB threshold is a guideline, not enforced by any current tooling.
- The index must be regenerated or it rots. Workflows that don't wire regeneration into a trigger (pre-commit, session-end) will accumulate drift. Mitigation: the planned `tolvi vault-index` CLI subcommand is designed to be a one-line hook.
- The index loads a few KB into every agent session. Below roughly 10 decision docs the overhead is not worth it — small vaults should skip the index until they grow.
- The benchmark numbers are from a single reference implementation on a vault of a few hundred docs. Behavior at larger scales is not yet measured.

## Adoption

This decision is itself adopted for Tolvi's own internal vault, tracked in a maintainer-only decision under `vault/decisions/` (the `vault/` directory is gitignored — see [`docs/CONVENTIONS.md`](../CONVENTIONS.md) Section 1 for the layout that adopters should mirror). Public guidance for adopters lives in [`docs/CONVENTIONS.md`](../CONVENTIONS.md) Section 7.

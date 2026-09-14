---
tags: [decision, tolvi]
date: 2026-05-14
repo: tolvi
status: active
ticket: none
user_impact: low
product_area: Ask API
---

# /v1/ask citations[] uses flatMap to surface cross-repo slug ambiguity

**Date:** 2026-05-14
**Repo:** tolvi

## Why

Tolvi vaults can contain multiple repos in the same workspace (`workspaces:repos = 1:N`). It's legitimate — even expected — for two repos in the same workspace to have a decision file with the same slug (e.g., both `backend` and `mobile` have `decisions/postgres-vs-mongo.md` → both have slug `postgres-vs-mongo`). When the LLM cites `[[postgres-vs-mongo]]` in an answer, the API needs to be honest about which document is being referenced. The plan's original code silently picked the first match and emitted a single (possibly wrong) `document_id` to the consumer.

## How

- **Schema baseline:** `documents.slug_idx` is a plain (non-unique) index on `(workspace_id, slug)`. Uniqueness lives at `(workspace_id, repo_id, doc_type, slug)`. So same workspace can legitimately have multiple documents with the same slug across different repos.
- **The bug it would have shipped (caught in Task 24 review):** plan code was

  ```ts
  const citations = citedSlugs.map((slug) => {
    const r = results.find((x) => x.slug === slug)!;
    return { slug: r.slug, doc_type: r.docType, document_id: r.documentId };
  });
  ```

  `Array.prototype.find` returns the first match. If search returned two `postgres-vs-mongo` hits (one per repo), only one made it into `citations[]` — silently. The `document_id` returned was effectively random (first-match by ranking order).
- **Fix:** flatMap returning all matching results per cited slug.

  ```ts
  const citations = citedSlugs.flatMap((slug) =>
    results
      .filter((r) => r.slug === slug)
      .map((r) => ({ slug: r.slug, doc_type: r.docType, document_id: r.documentId })),
  );
  ```

  Non-breaking: the citation shape `{ slug, doc_type, document_id }` is unchanged. The array just gets longer when ambiguity exists. Consumers can dedup or render all sources; either is honest.
- **Why not change the citation shape?** Adding a `repo` field would break the existing v1 contract that's now in `spec/openapi.json`. flatMap is non-breaking and the loss-of-information case (same slug across repos) is the only one where the consumer needs the second entry to disambiguate.
- **Why not teach the LLM `[[repo:slug]]` syntax?** The format spec supports it; the system prompt does not. Doing it later — when there's user-facing demand — is cheap. Adding it now expands `SYSTEM_PROMPT` (which is the cached prompt block) and adds parsing in `extractCitations` for marginal v1 benefit.
- **Test impact:** the existing Task 24 integration test (`expect(body.citations).toHaveLength(1)`) still passes because the test workspace only ingests one document. The fix changes behavior only when ambiguity exists.

## Outcome

Cross-repo slug collisions in `/v1/ask` responses now surface as multiple citation entries instead of one-and-silently-wrong. The API contract gains a small honest "this could mean several docs" signal at zero shape cost.

---
tags: [decision, tolvi]
date: 2026-05-14
repo: tolvi
status: active
ticket: none
user_impact: none
product_area: Format Spec
---

# JSON Schema $id URLs locked to tolvilabs.com pre-1.0

**Date:** 2026-05-14
**Repo:** tolvi

## Why

JSON Schema `$id` URIs are stable identifiers. Once a third-party vault references one, changing it requires a versioned format migration — every consumer has to handle both URIs. Tolvi's four schemas (`vault-meta.json`, `decision.json`, `session.json`, `pattern.json`) shipped in PR #1 with placeholder `$id` URLs at `https://tolvi.dev/spec/schemas/<X>.json`. The actual registered domain is `tolvilabs.com`, not `tolvi.dev`. Doing the swap now — before anyone outside this repo references the schemas — is a free correction; doing it after first external use is a breaking spec change.

## How

- **New canonical form:** `https://tolvilabs.com/tolvi/spec/schemas/<X>.json` (path-style under the Labs umbrella, leaving room for sibling Labs projects to sit alongside at `tolvilabs.com/<project>/...`).
- **Files changed (4):** `spec/schemas/{vault-meta,decision,session,pattern}.json` — one line each, just the `$id` field.
- **Why path-style not subdomain?** One domain + one cert + one DNS record. Subdomains per project would require certificate-and-DNS work for every Labs project. Path-style is cheap to host (static site at `tolvilabs.com` with subdirectories) and the schemas are static JSON files anyway.
- **Same commit also updated:** `CODE_OF_CONDUCT.md` (conduct@), `SECURITY.md` (security@), `ROADMAP.md` (docs site URL), `NAMESPACE.md` (table restructured — `tolvilabs.com` promoted to row 2 ✅ Claimed 2026-05-11; `tolvi.dev` and `tolvi.com` marked ❌ Not pursued with rationale). Total: 8 files, +10/-10 lines.
- **Deliberately NOT updated:** `docs/superpowers/specs/2026-05-09-phase-0-1-foundation-design.md` and `docs/superpowers/plans/2026-05-09-phase-0-1-foundation.md`. These are historical artifacts — point-in-time records of what was planned. Rewriting them to retcon the domain would erase signal about what assumption was held when.
- **Verification ran:** `npm run validate` (Ajv against all 13 sample-vault files still green), brand-isolation guard green, markdown lint green.
- **Schemas-as-shipped never actually published.** PR #1 merged but no vault outside this repo has consumed those `$id` URLs yet. The "free correction" window is wide open.

## Outcome

Spec schemas now reference the actually-owned domain. The `$id` URIs are stable for v1.0 and beyond; any future change is now a deliberate versioned migration, not a "we should have used the right domain" cleanup.

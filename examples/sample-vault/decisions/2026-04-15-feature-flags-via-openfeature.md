---
tags: [decision, sample, feature-flags]
date: 2026-04-15
status: active
repo: tolvi-examples
ticket: none
---

## Why

Three services were each shipping their own ad-hoc feature flag implementation: one read flags from environment variables, one queried a config service on every request, and one used a vendor SDK directly. The duplication was costing us time on every cross-service rollout, and the vendor SDK had spread far enough that swapping it out would have been a multi-week effort. We wanted a single abstraction that all services could share without coupling us to any one vendor.

## How

We adopted OpenFeature as the in-process flag API. Each service depends only on the OpenFeature SDK; the backend is wired up at startup based on configuration. For local development the no-op backend returns defaults, which keeps tests deterministic. In staging and production we use a single hosted backend for now, but the abstraction means we can swap it later without touching call sites. Rollouts follow our [[feature-flag-rollout]] pattern.

## Outcome

The first new flag added under OpenFeature took fifteen minutes end-to-end, including the dashboard configuration. The vendor-specific SDK has been removed from two of the three services; the third is scheduled for the next release. No regressions have been reported.

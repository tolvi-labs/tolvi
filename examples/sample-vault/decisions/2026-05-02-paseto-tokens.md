---
tags: [decision, sample, auth, security]
date: 2026-05-02
status: active
repo: tolvi-examples
ticket: none
supersedes: jwt-vs-paseto-tokens
---

## Why

The earlier choice in [[jwt-vs-paseto-tokens]] had two persistent problems: the JWT specification's algorithm-agility surface (especially the `alg: none` and key-confusion families) kept showing up in our security checklists as a class of bug we had to manually exclude, and benchmarking showed RS256 verification dominating the hot path on the auth-checking middleware. We wanted a token format that removes the algorithm-confusion class of bugs by construction and that verifies in less CPU.

## How

We adopted PASETO v4 (public, Ed25519-signed). The library surface is intentionally small: there is no algorithm field in the header, so there is no `alg: none`. Tokens still carry an opaque user identifier, scopes, and standard timing claims. Migration runs side-by-side: the auth service issues both formats during a two-week cutover and resource servers accept either. After the cutover, JWT support is removed in a single coordinated release.

## Outcome

Verification CPU on the middleware dropped by roughly 60% in load tests. The cutover began on 2026-05-02 and is on track to complete inside the planned window. No security review item involving algorithm confusion has appeared since.

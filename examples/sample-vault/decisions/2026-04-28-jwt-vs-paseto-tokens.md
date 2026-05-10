---
tags: [decision, sample, auth, security]
date: 2026-04-28
status: superseded
repo: tolvi-examples
ticket: none
superseded_by: paseto-tokens
---

## Why

We needed a token format for short-lived authentication tokens issued by the auth service and verified by every other service. The team's prior experience was with JWT, and the ecosystem support is excellent. We picked JWT primarily because libraries existed in every language we use, and because JWT is what most developers reach for by default. We documented the choice with the understanding that we would revisit it once we had a real benchmark.

## How

We chose RS256-signed JWTs with a five-minute expiry, a JWKS endpoint for key rotation, and `kid` headers so that consumers could verify against the right key during a rotation window. Tokens carried only an opaque user identifier, scopes, and the standard `iss`, `aud`, `iat`, and `exp` claims. We deliberately avoided putting any user-identifying data into the body.

## Outcome

Superseded by [[paseto-tokens]] on 2026-05-02 after benchmarking and operational review. Two issues drove the change: the `alg: none` family of footguns kept showing up in security review checklists, and the verification cost on the hot path was higher than we wanted. The replacement decision documents the migration plan and the new cutover behavior in detail.

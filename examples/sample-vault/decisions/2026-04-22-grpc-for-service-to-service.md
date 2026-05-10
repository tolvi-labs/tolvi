---
tags: [decision, sample, rpc, networking]
date: 2026-04-22
status: active
repo: tolvi-examples
ticket: none
---

## Why

Our internal services were communicating over JSON-over-HTTP with hand-written client wrappers in each consumer. Two recent incidents traced back to silent drift between server response shapes and client expectations: a renamed field made it to production without breaking any test, because the clients tolerated unknown fields and the missing field defaulted to a falsy value. We wanted a contract format that fails loudly on mismatch and a code generation story that makes drift hard to introduce.

## How

We standardized on gRPC for service-to-service traffic. Protocol buffers live in a shared `proto/` directory; generated stubs are published per-language as part of CI. Public-facing edges remain HTTP/JSON via a thin gateway. We adopted server reflection in non-production environments to keep `grpcurl` ergonomic for debugging. Tracing uses our [[tracing-context-propagation]] pattern so that a request crossing the HTTP-to-gRPC boundary keeps its trace context intact.

## Outcome

The first migrated path saw a 40% reduction in payload size and end-to-end latency dropped by roughly 12ms at the median. More importantly, two field-rename PRs have already been caught at compile time in downstream consumers — exactly the failure mode we were trying to surface earlier.

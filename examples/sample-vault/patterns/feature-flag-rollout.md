---
tags: [pattern, sample, feature-flags, release]
status: active
---

# Progressive feature flag rollout

A new feature reaches production in stages, not all at once. The standard ladder is `1% → 10% → 50% → 100%`, with at least one full business day between each step and a documented rollback path at every rung. The point of the ladder is not the specific percentages — it is to give the team time to look at telemetry and customer reports at a sample size that is large enough to detect problems but small enough that a problem is recoverable.

The flag itself should be evaluated as close to the user-visible code path as possible. Deep flag checks scattered across many files become impossible to clean up; a single check at the entry point is cheap to remove later. Every flag should have an owner and a planned removal date in its description — otherwise flags accumulate forever.

A typical rollout cadence:

| Stage | Audience | Duration | Exit criteria |
|---|---|---|---|
| 1% | Random sample | 24h | Error rate within noise floor; no new alert |
| 10% | Random sample | 48h | Latency within budget; no support escalations |
| 50% | Random sample | 48h | Same telemetry as 10%, plus dashboard sanity check |
| 100% | All users | indefinite | Stable for 7 days, then schedule flag removal |

When to use this: any user-facing behavior change with non-trivial blast radius. When not to use this: trivial bug fixes, internal-only changes, and infra-side toggles where a binary cutover is safer (for example, switching backend storage destinations).

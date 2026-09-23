---
tags: [decision, vuln, <system-slug>]
date: YYYY-MM-DD
system: <system-slug>
status: active
asset: none
exposure: unknown
severity: none
reachable: unknown
disposition: none
tracker: none
---

# [Triage decision for <finding class>]

## Why
[What the finding is, in the general terms a reader needs to understand the decision, and why it needed a judgement call rather than a routine fix.]

## How
- **Finding class:** [the category — injection, authz bypass, deserialization, dependency CVE — not a reproduction]
- **Severity and score:** [rating and the framework used]
- **Reachability:** [whether the vulnerable path is actually reachable in this deployment, and what gates it]
- **Disposition:** [fix now / fix next cycle / accept / not applicable] — [the reason]
- **Compensating controls:** [what reduces the risk in the meantime]
- **Disclosure posture:** [internal only, coordinated disclosure, embargo date, or reporter expectations]
- **Tracker:** [the internal ticket holding the reproduction and the technical detail]

## Outcome
[The finding is [disposition] at [severity], reachable / not reachable, tracked as [id].]

> This note records the triage decision. The reproduction, payload, and unpatched specifics stay in the tracker — a vault that gets committed is the wrong place for them.

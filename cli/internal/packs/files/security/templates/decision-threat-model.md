---
tags: [decision, threat-model, <system-slug>]
date: YYYY-MM-DD
system: <system-slug>
status: active
asset: <what-you-are-protecting>
exposure: unknown
review: YYYY-MM-DD
---

# [Threat model for <system>]

## Why
[What is worth protecting here and why this system warrants an explicit model. A threat model that covers everything covers nothing, so name the asset rather than the technology.]

## How
- **Asset:** [what an attacker actually wants — the data, the money, the access, the uptime]
- **Trust boundaries:** [where untrusted input crosses into trusted code, and what sits on each side]
- **Adversary assumed:** [who you are defending against, and their capability and motivation]
- **In scope:** [the threats this model addresses]
- **Explicitly out of scope:** [threats you are deliberately not defending against, and why — a supply-chain compromise, a malicious insider, a nation-state]
- **Risks accepted:** [what you are knowingly living with, and who signed off]
- **Review date:** [when this gets revisited, since a threat model goes stale as the system grows]

## Outcome
[The model covers [threats] against [asset], explicitly excludes [threats], and is reviewed on [date].]

> Decisions and rationale only. Exploitable specifics belong in the tracker, never in the vault.

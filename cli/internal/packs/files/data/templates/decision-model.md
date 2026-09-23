---
tags: [decision, model, <project-slug>]
date: YYYY-MM-DD
project: <project-slug>
status: active
metric: none
dataset: none
baseline: none
serving: none
---

# [The approach chosen, stated as a commitment]

## Why
[What the model has to do, and the constraint that ruled out the obvious choice — latency budget, cost ceiling, explainability requirement, or training data you do not have.]

## How
- **Chosen:** [approach or architecture] — [one-line reason]
- **Beat:** [the baseline or incumbent, and by how much on which metric]
- **Rejected:** [alternative] — [specific reason: cost, latency, data hunger, opacity]
- **Trade-off accepted:** [what you gave up — interpretability, tail accuracy, inference cost, retraining burden]
- **Serving shape:** [batch or realtime, and the latency and cost envelope it must hold]
- **Failure modes:** [where it is known to be weak, and what you are not claiming it does]

## Outcome
[The project serves [approach] at [metric], accepting [trade-off], replacing [baseline].]

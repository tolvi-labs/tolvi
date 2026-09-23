---
tags: [decision, dataset, <project-slug>]
date: YYYY-MM-DD
project: <project-slug>
status: active
metric: none
dataset: <dataset-slug>
license: unknown
pii: unknown
---

# [Adopt / Rebuild / Retire] [dataset]

## Why
[What this dataset is for and why it is the right source. If retiring: what made it unfit — drift, licence, leakage, or a bias you found.]

## How
- **Source and provenance:** [where it came from, how it was collected, and who owns it]
- **Licence and permitted use:** [SPDX or terms, and whether commercial or model-training use is allowed]
- **PII posture:** [what personal data it contains, what was removed or hashed, and what that means for where it may be stored]
- **Excluded on purpose:** [rows, classes, sources, or time ranges you dropped, and why — this is the part that explains a model's blind spots later]
- **Splits:** [train/val/test strategy, and how leakage between them is prevented]
- **Known problems:** [imbalance, label noise, drift, coverage gaps]

## Outcome
[The project now trains on [dataset] under [licence], excluding [what], split by [strategy].]

> Record PII posture, never PII. No sample records, identifiers, or raw rows in the vault.

---
tags: [decision, tolvi, spec, json-contract, doctor]
date: 2026-09-23
repo: tolvi
status: active
ticket: none
user_impact: medium
product_area: CLI / published contracts
---

# Output schemas ship beside the format schemas but version with the CLI, not with the format

**Date:** 2026-09-23
**Repo:** tolvi

## TL;DR

`doctor` and `doctor vault-health` now have `--json`, and their shapes are published as JSON Schemas. Those schemas describe what a command prints, which changes when the CLI changes, while the package's major version already means the vault format version. Mixing them would make one of the two meanings a lie, so `spec/index.json` advertises them under a separate `outputs` key with its own compatibility promise: additive within a package major, and a breaking change ships as a new file rather than an edit.

## Why

Everything the CLI printed was written for a person reading a terminal, so any tool acting on a result had to parse formatted text that changes whenever the wording improves. `--json` closes that. Publishing the shape rather than leaving it implicit is what makes it a contract a consumer can pin to, per `2026-09-14-packs-vendored-and-output-published`.

Publishing them raised a versioning problem that decision did not settle. `spec/README.md` states that the package's major version *is* the format version: `@tolvi-labs/spec@^2` is a pin on tolvi-format-v2, and `spec-package-check.sh` fails the build when `package.json`, `index.json` and the schemas disagree. An output shape has nothing to do with the vault format, so it cannot use that major to signal a break. Listed under the same key, a consumer pinning the format would silently receive output schemas it never agreed to.

## How

- **A separate `outputs` key in `spec/index.json`.** `schemas` stays the vault format; `outputs` holds `doctor` and `vault-health`. `spec-package-check.sh` now verifies both keys resolve to files that exist, and its comment records why the two are kept apart.
- **The promise, written where a consumer will read it:** output schemas are extended additively within a package major, and a breaking change ships as a new file rather than as a silent edit to an existing one. A pin on the format never drags an incompatible output shape along with it.
- **Check ids are the stable key, names are display text.** `Check` gained an `ID` (`path`, `vault`, `api_key`, `claude_permissions`), assigned in `RunDoctor` beside the ordering rather than inside each constructor, so a check cannot ship without one and an id cannot drift from the slot it labels. Renaming a `name` is a wording change; renaming an `id` is a breaking change to the contract.
- **Every payload names the CLI that produced it.** `tolvi_version` is in both shapes, so a consumer needing a newer contract can say so instead of parsing output it cannot trust.
- **Findings are always an array.** A clean vault encodes `"findings": []`, never `null`, so a consumer iterates unconditionally. There is a test for exactly that, because Go's zero slice would have encoded as `null` and nobody would have noticed until a downstream crash.
- **Text and JSON never interleave.** Under `--json` the report writer is `io.Discard` and JSON owns stdout alone. Exit codes are unchanged, so a caller can read either the process code or the `ok` field.
- **The contract is enforced from this repo.** Five Go tests assert the wire shape and validate real command output against the embedded schemas, so a field rename that only updates a struct fails here rather than in a consumer. The embedded copies are pinned to `spec/schemas/` by the existing `schema-copy-parity-check.sh`, which needed no change.
- **Rejected: tagging the domain structs.** `Check` and `HealthReport` are internal types that change with the checks; giving them wire tags would make every internal rename a published break. The emitters build separate structs whose tags are the contract.

## Outcome

`tolvi doctor --json` and `tolvi doctor vault-health --json` emit schema-validated JSON, published in `spec/` under a key whose versioning rule matches what the schemas actually track.

# Spec

The public format contract for Tolvi vaults.

## Files

- [`tolvi-format-v1.md`](./tolvi-format-v1.md) — the normative spec. **Read this first.**
- [`schemas/`](./schemas/) — machine-readable JSON Schemas (Draft 2020-12) for validating vault content.

## Stability

`tolvi-format-v1` is the stable v1 contract. Breaking changes require a `tolvi-format-v2` revision, which will live in this directory alongside v1, with documented migration tooling.

Every consumer (CLI, server, SDKs, third-party tools) implements parsing and validation against the spec. The spec is the only artifact crossing language boundaries — there is no shared parsing library.

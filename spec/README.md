# Spec

The public format contract for Tolvi vaults.

## Files

- [`tolvi-format-v2.md`](./tolvi-format-v2.md): the normative spec. **Read this first.**
- [`tolvi-format-v1.md`](./tolvi-format-v1.md): the previous version, frozen and preserved as the contract v1 vaults were written against.
- [`schemas/`](./schemas/): machine-readable JSON Schemas (Draft 2020-12) for validating vault content.

## Stability

`tolvi-format-v2` is the current contract; `tolvi-format-v1` is frozen and sits alongside it. Breaking changes require a `tolvi-format-v3` revision in this directory, with documented migration tooling. The schemas in `./schemas/` track the current version.

Every consumer (CLI, server, SDKs, third-party tools) implements parsing and validation against the spec. The spec is the only artifact crossing language boundaries. There is no shared parsing library.

# Spec

The public format contract for Tolvi vaults.

## Files

- [`tolvi-format-v2.md`](./tolvi-format-v2.md): the normative spec. **Read this first.**
- [`tolvi-format-v1.md`](./tolvi-format-v1.md): the previous version, frozen and preserved as the contract v1 vaults were written against.
- [`schemas/`](./schemas/): machine-readable JSON Schemas (Draft 2020-12). `index.json` splits them by what they describe: `schemas` validates vault content, and `outputs` describes what the CLI prints with `--json`.

## Stability

`tolvi-format-v2` is the current contract; `tolvi-format-v1` is frozen and sits alongside it. Breaking changes require a `tolvi-format-v3` revision in this directory, with documented migration tooling. The schemas in `./schemas/` track the current version.

**Output schemas move with the CLI, not with the format.** The package's major version is the format version, so an output shape cannot use it to signal a break. Those schemas are therefore extended additively within a package major, and a breaking change ships as a new file rather than as an edit to an existing one. A consumer pinned to `@tolvi-labs/spec@^2` for the format is never handed an output shape that stopped being what it validated against.

Every consumer (CLI, server, SDKs, third-party tools) implements parsing and validation against the spec. The spec is the only artifact crossing language boundaries. There is no shared parsing library.

## Vendoring the schemas

The schemas are published as [`@tolvi-labs/spec`](https://www.npmjs.com/package/@tolvi-labs/spec) so other repos can depend on the contract instead of copying it:

```bash
npm i -D @tolvi-labs/spec
```

```js
import index from "@tolvi-labs/spec";              // { format, schemaVersion, schemas }
import meta from "@tolvi-labs/spec/schemas/vault-meta.json";
```

**The package's major version is the format version.** `@tolvi-labs/spec@^2` is a pin on `tolvi-format-v2`, and `.github/scripts/spec-package-check.sh` fails the build if `package.json`, `index.json`, and the `schema_version` the schemas pin ever disagree. A 3.x package shipping v2 schemas would make the version meaningless to every consumer, which is the failure that check exists to prevent.

Copying the schemas into another repo is the thing this replaces. A copy with no guard drifts: the set served at `tolvilabs.com/tolvi/spec/schemas/` was hand-copied once and spent months serving v1 after the repo had shipped v2, so the published contract disagreed with the implementation and nothing was positioned to notice.

### Releasing

Bump `spec/package.json`, then tag:

```bash
git tag spec-v2.0.1 && git push origin spec-v2.0.1
```

`spec-release.yml` verifies the tag matches the version, verifies the version matches the format, validates the schemas, and publishes with provenance.

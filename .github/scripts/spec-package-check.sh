#!/usr/bin/env bash
# Fails (exit 1) if spec/package.json drifts from the format it ships.
#
# @tolvi-labs/spec exists so other repos can vendor the schemas instead of
# reading a sibling checkout. That only works if the version means something:
# the package's major version IS the format version, so `@tolvi-labs/spec@^2`
# is a pin on tolvi-format-v2. A 3.x package carrying v2 schemas, or an
# index.json that names a format the schemas do not implement, breaks that
# contract quietly for every consumer.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

fail=0

pkg_major="$(node -p "require('./spec/package.json').version.split('.')[0]")"
index_version="$(node -p "require('./spec/index.json').schemaVersion")"
index_format="$(node -p "require('./spec/index.json').format")"
schema_const="$(node -p "require('./spec/schemas/vault-meta.json').properties.schema_version.const")"

if [ "$pkg_major" != "$schema_const" ]; then
  echo "✗ spec/package.json major is $pkg_major but vault-meta.json pins schema_version $schema_const"
  fail=1
fi
if [ "$index_version" != "$schema_const" ]; then
  echo "✗ spec/index.json says schemaVersion $index_version but vault-meta.json pins $schema_const"
  fail=1
fi
if [ "$index_format" != "tolvi-format-v$schema_const" ]; then
  echo "✗ spec/index.json says format $index_format, which does not match schema_version $schema_const"
  fail=1
fi
if [ ! -f "spec/tolvi-format-v$schema_const.md" ]; then
  echo "✗ no spec/tolvi-format-v$schema_const.md for the version the schemas pin"
  fail=1
fi

# Every schema the manifest advertises must actually ship, under either key.
# `schemas` is the vault format, pinned to the package major. `outputs` is the
# CLI's --json contract, which moves with the CLI instead: it is extended
# additively within a package major, and a breaking change ships as a new file
# rather than as a silent edit, so a pin on the format never drags an
# incompatible output shape along with it.
node -e '
const idx = require("./spec/index.json");
const fs = require("fs");
let bad = 0;
for (const key of ["schemas", "outputs"]) {
  for (const [name, rel] of Object.entries(idx[key] ?? {})) {
    if (!fs.existsSync("spec/" + rel.replace("./", ""))) {
      console.log(`✗ index.json advertises ${key}.${name} at ${rel}, which does not exist`);
      bad++;
    }
  }
}
process.exit(bad ? 1 : 0);
' || fail=1

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "See .github/scripts/spec-package-check.sh for why the versions must agree."
  exit 1
fi
echo "✓ @tolvi-labs/spec $pkg_major.x ships tolvi-format-v$schema_const, and index.json agrees"

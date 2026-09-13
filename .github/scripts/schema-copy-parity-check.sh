#!/usr/bin/env bash
# schema-copy-parity-check.sh — fails (exit 1) if the published schemas in
# spec/schemas/ and the copies the CLI embeds from cli/internal/format/schemas/
# have drifted apart.
#
# Two copies exist for good reasons: spec/ is the published artifact a third
# party reads, and the CLI embeds its own so the binary needs no network or
# working tree. Nothing kept them in step, though, and only the spec/ copy is
# validated by validate:schemas — so editing one and forgetting the other ships
# a CLI that disagrees with its own published spec, silently.
set -euo pipefail

fail=0

for published in spec/schemas/*.json; do
  name="$(basename "$published")"
  embedded="cli/internal/format/schemas/$name"
  if [ ! -f "$embedded" ]; then
    echo "✗ $name: published in spec/schemas/ but not embedded in cli/internal/format/schemas/"
    fail=1
    continue
  fi
  if ! diff -q "$published" "$embedded" >/dev/null; then
    echo "✗ $name: spec/schemas/ and cli/internal/format/schemas/ have drifted"
    diff "$published" "$embedded" || true
    fail=1
    continue
  fi
  echo "✓ $name: copies agree"
done

for embedded in cli/internal/format/schemas/*.json; do
  name="$(basename "$embedded")"
  if [ ! -f "spec/schemas/$name" ]; then
    echo "✗ $name: embedded in the CLI but not published in spec/schemas/"
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "Schema copies must be byte-identical. Edit spec/schemas/ and copy to"
  echo "cli/internal/format/schemas/, or see .github/scripts/schema-copy-parity-check.sh."
  exit 1
fi

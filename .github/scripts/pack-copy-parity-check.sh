#!/usr/bin/env bash
# pack-copy-parity-check.sh — fails (exit 1) when the packs vendored at
# cli/internal/packs/files/ have drifted from their source in tolvi-solo.
#
# tolvi-solo owns the packs. `2026-06-05-tolvi-solo-as-separate-product` is
# active and rejects moving them here, so this repo vendors them: `tolvi init
# --pack` and `tolvi packs list` read the vendored copies, and the binary needs
# no sibling checkout. A copy nothing pins is a copy that drifts, so this is
# the pin.
#
# Only pack.json and templates/ are vendored. A pack's README is prose for
# people reading the solo repo, and the CLI has no use for it.
#
# Skips cleanly when the tolvi-solo sibling checkout is absent, so it never
# blocks a contributor who has only this repo. CI checks out both.
#
# To sync after a change in solo:
#   rsync -a --delete --include='*/' --include='pack.json' --include='templates/***' \
#     --exclude='*' ../tolvi-solo/packs/ cli/internal/packs/files/
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

src="../tolvi-solo/packs"
dst="cli/internal/packs/files"

if [ ! -d "$src" ]; then
  echo "pack parity: no tolvi-solo sibling checkout; skipping."
  exit 0
fi

fail=0

# Vendored copy against source, both directions: a pack added in solo and never
# vendored is as wrong as a vendored pack solo no longer ships.
for pack_dir in "$src"/*/; do
  name="$(basename "$pack_dir")"
  [ -f "$pack_dir/pack.json" ] || continue   # a pack with no manifest is solo's check to make
  if [ ! -d "$dst/$name" ]; then
    echo "✗ $name: shipped by tolvi-solo but not vendored here"
    fail=1
    continue
  fi
  if ! diff -r "$pack_dir/templates" "$dst/$name/templates" >/dev/null 2>&1; then
    echo "✗ $name: templates have drifted from tolvi-solo"
    diff -r "$pack_dir/templates" "$dst/$name/templates" | head -10 || true
    fail=1
  fi
  if ! diff -q "$pack_dir/pack.json" "$dst/$name/pack.json" >/dev/null 2>&1; then
    echo "✗ $name: pack.json has drifted from tolvi-solo"
    fail=1
  fi
  [ "$fail" -eq 0 ] && echo "✓ $name: vendored copy agrees with tolvi-solo"
done

for vendored in "$dst"/*/; do
  name="$(basename "$vendored")"
  if [ ! -f "$src/$name/pack.json" ]; then
    echo "✗ $name: vendored here but tolvi-solo does not ship it"
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "tolvi-solo owns the packs. Re-sync with:"
  echo "  rsync -a --delete --include='*/' --include='pack.json' --include='templates/***' \\"
  echo "    --exclude='*' ../tolvi-solo/packs/ cli/internal/packs/files/"
  exit 1
fi

#!/usr/bin/env bash
# Preflight-sync check. Fails (exit 1) if any slash command's PREFLIGHT block
# has drifted from the canonical copy in commands/_preflight.md.
#
# Every command that prefers the `tolvi` CLI carries the same preflight
# instructions. Four hand-maintained copies is how the marketing site's stack
# sections drifted into four different answers, so these are pinned instead.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
src="skills/tolvi/commands/_preflight.md"
[ -f "$src" ] || { echo "✗ missing $src"; exit 1; }

extract() { sed -n '/<!-- PREFLIGHT:BEGIN -->/,/<!-- PREFLIGHT:END -->/p' "$1"; }

canonical="$(extract "$src")"
[ -n "$canonical" ] || { echo "✗ $src has no PREFLIGHT block"; exit 1; }

fail=0
for f in skills/tolvi/commands/*.md; do
  [ "$(basename "$f")" = "_preflight.md" ] && continue
  got="$(extract "$f")"
  if [ -z "$got" ]; then
    echo "⚠ $f: no PREFLIGHT block (skipping; add one if it uses the CLI)"
    continue
  fi
  if [ "$got" != "$canonical" ]; then
    echo "✗ $f: PREFLIGHT block has drifted from $src"
    diff <(echo "$canonical") <(echo "$got") | head -12 || true
    fail=1
  else
    echo "✓ $f: preflight in sync"
  fi
done

[ "$fail" -eq 0 ] || { echo ""; echo "Re-copy the block from $src."; exit 1; }

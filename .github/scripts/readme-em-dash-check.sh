#!/usr/bin/env bash
# README em-dash check. Fails (exit 1) on any em dash in README prose.
#
# Per the house rule, READMEs, site docs, and marketing pages take spaced
# hyphens and no em dashes, with no "unless it reads better" escape hatch.
# Two exemptions survive because they are not prose:
#
#   - inside a fenced code block: literal tool output, ASCII diagrams,
#     directory trees, and code comments document real formats.
#   - a lone em dash in a table cell: the "not applicable" glyph.
#
# Written as a check rather than left to review because the previous rule was
# a per-sentence judgement call, which is how 192 of them accumulated.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

fail=0
while IFS= read -r f; do
  hits="$(awk '
    /^[[:space:]]*```/ { infence = !infence; next }
    !infence && /—/ {
      if ($0 ~ /\|[[:space:]]*—[[:space:]]*\|/) next
      printf "  %s:%d  %s\n", FILENAME, FNR, substr($0, 1, 100)
    }
  ' "$f")"
  if [ -n "$hits" ]; then
    echo "✗ em dash in README prose:"
    printf '%s\n' "$hits"
    fail=1
  fi
done < <(git ls-files '*README.md')

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "Use a spaced hyphen, a colon, a comma, or a full stop."
  echo "Code fences and table N/A glyphs are exempt; see this script for why."
  exit 1
fi
echo "✓ READMEs: no em dashes in prose"

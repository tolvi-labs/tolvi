#!/usr/bin/env bash
# Brand-isolation check. Fails (exit 1) if any tracked file outside the allowlist
# matches the forbidden-term regex supplied at runtime via the BRAND_BLOCKLIST
# environment variable.
#
# The pattern is provided by the environment (a GitHub Actions repository variable
# in CI; see the "Brand-isolation guard" step in .github/workflows/validate.yml)
# rather than hardcoded here, so the public source never enumerates the protected
# terms. To run locally, export BRAND_BLOCKLIST first.
#
# Allowlist entries are matched as follows:
#   - exact path: matches only the named file (e.g. "NOTICE")
#   - path ending in /: matches any file under that directory
#
# See CONTRIBUTING.md "Brand isolation" for the rule and rationale.
set -euo pipefail

if [ -z "${BRAND_BLOCKLIST:-}" ]; then
  echo "BRAND_BLOCKLIST is not set; skipping brand-isolation check."
  echo "Set the BRAND_BLOCKLIST repo variable (a pipe-separated grep -E pattern) to enforce it."
  exit 0
fi
PATTERN="$BRAND_BLOCKLIST"

# Allowlist (file path or directory prefix). Files matching any entry are skipped.
#   - NOTICE: the official attribution file (the one place the parent-org name belongs)
#   - This script: provided so an incidental value never trips the check on itself
ALLOWLIST_PATHS=(
  'NOTICE'
  '.github/scripts/brand-isolation-check.sh'
)

is_allowlisted() {
  local file="$1"
  for allowed in "${ALLOWLIST_PATHS[@]}"; do
    if [[ "$allowed" == */ ]] && [[ "$file" == "$allowed"* ]]; then
      return 0
    fi
    if [ "$file" = "$allowed" ]; then
      return 0
    fi
  done
  return 1
}

FOUND=0
while IFS= read -r file; do
  if is_allowlisted "$file"; then continue; fi
  if grep -iIqE "$PATTERN" "$file" 2>/dev/null; then
    echo "FAIL: forbidden term in tracked file: $file"
    grep -iInE "$PATTERN" "$file" | head -5
    FOUND=1
  fi
done < <(git ls-files)

if [ "$FOUND" -ne 0 ]; then
  echo ""
  echo "Brand-isolation check failed. The forbidden terms above must be removed,"
  echo "or the file must be added to the allowlist in .github/scripts/brand-isolation-check.sh."
  echo "See CONTRIBUTING.md 'Brand isolation' section for details."
  exit 1
fi

echo "Brand-isolation check passed."

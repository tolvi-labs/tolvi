#!/usr/bin/env bash
# Brand-isolation check. Fails (exit 1) if any tracked file outside the allowlist
# references parent organizations, sibling products, or specific deployments.
#
# Allowlist entries are matched as follows:
#   - exact path: matches only the named file (e.g. "NOTICE")
#   - path ending in /: matches any file under that directory (e.g. "docs/superpowers/")
#
# See CONTRIBUTING.md "Brand isolation" for the rule and rationale.
set -euo pipefail

# Block-list of forbidden patterns (case-insensitive). Add carefully — false positives
# annoy contributors; missed terms break the brand-isolation promise.
FORBIDDEN_PATTERNS=(
  'corvin'
  'corvin health'
  'torres atlantic'
  'ta-ai-tooling'
  'ta-wide'
  '\bfirebase\b'
  '\bfirestore\b'
  '\bcloud run\b'
  'gs://'
  'healthcare'
  'hipaa'
  '\bbaa\b'
  '\bphi\b'
  '\bmrn\b'
  'eligibility'
)

# Allowlist (file path or directory prefix). Files matching any entry are skipped.
#   - NOTICE: the official attribution file (only place the parent-org name belongs)
#   - This script: contains the patterns it searches for
#   - docs/superpowers/: engineering-process artifacts (specs, plans) that may
#     legitimately discuss the rule, including showing the pattern list. These
#     ship publicly as transparency artifacts.
ALLOWLIST_PATHS=(
  'NOTICE'
  '.github/scripts/brand-isolation-check.sh'
  'docs/superpowers/'
)

# Compile the forbidden-pattern regex once
PATTERN=$(IFS='|'; echo "${FORBIDDEN_PATTERNS[*]}")

is_allowlisted() {
  local file="$1"
  for allowed in "${ALLOWLIST_PATHS[@]}"; do
    # Directory prefix match (allowlist entry ends with /)
    if [[ "$allowed" == */ ]] && [[ "$file" == "$allowed"* ]]; then
      return 0
    fi
    # Exact file match
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

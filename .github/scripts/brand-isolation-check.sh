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
  # Loud on purpose. This check is the only one in the suite that cannot run
  # without a secret, so a local `npm run validate` exits 0 having never
  # checked a thing. That is easy to read as "brand isolation passed" when it
  # means "brand isolation was not examined", and reading it the wrong way is
  # how forbidden terms reached a public repo in September 2026.
  echo ""
  echo "  ############################################################"
  echo "  #  BRAND-ISOLATION CHECK SKIPPED — NOT PASSED              #"
  echo "  #                                                          #"
  echo "  #  BRAND_BLOCKLIST is unset, so nothing was examined.      #"
  echo "  #  A clean run here is NOT evidence this repo is clean.    #"
  echo "  #  CI enforces it; a green local validate does not.        #"
  echo "  #                                                          #"
  echo "  #  To check locally, export BRAND_BLOCKLIST (the pattern   #"
  echo "  #  lives in the repo variable of the same name).           #"
  echo "  ############################################################"
  echo ""
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

# Commit messages are published too, and this check never looked at them. A
# message naming a protected term is as public as a file containing one, and
# harder to remove later: fixing it means rewriting history rather than editing
# a file. Two such messages sat in public history for months while CI stayed
# green, including, with some irony, the message of the commit that removed
# forbidden terms from the files.
while IFS= read -r line; do
  sha="${line%% *}"
  if git log -1 --pretty='%s%n%b' "$sha" | grep -iqE "$PATTERN" 2>/dev/null; then
    echo "FAIL: forbidden term in commit message: $sha"
    git log -1 --pretty='%s%n%b' "$sha" | grep -inE "$PATTERN" | head -3 | sed 's/^/    /'
    FOUND=1
  fi
done < <(git log --all --pretty='%H %s')

if [ "$FOUND" -ne 0 ]; then
  echo ""
  echo "Brand-isolation check failed. The forbidden terms above must be removed,"
  echo "or, for a commit message, stripped with a git filter-repo --replace-message pass."
  echo "or the file must be added to the allowlist in .github/scripts/brand-isolation-check.sh."
  echo "See CONTRIBUTING.md 'Brand isolation' section for details."
  exit 1
fi

echo "Brand-isolation check passed."

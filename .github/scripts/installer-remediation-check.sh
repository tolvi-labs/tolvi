#!/usr/bin/env bash
# Installer-remediation check. Fails (exit 1) if an installer tells the user to
# run `go install` without also telling them to put the resulting binary on PATH.
#
# `go install` writes to $(go env GOPATH)/bin, which is frequently not on PATH.
# An install that mentions only the install command therefore looks successful
# while leaving every skill and hook that shells out to `tolvi` on the silent
# fallback path. The PATH export is the half people miss, so it is the half
# this check pins.
#
# tolvi-solo carries a deliberate duplicate of this remediation, because it is
# a standalone clone and cannot depend on this repo. Each repo pins its own copy.
set -euo pipefail

fail=0

check_file() {
  local file="$1"
  [ -f "$file" ] || return 0
  grep -q 'go install github.com/tolvi-labs/tolvi' "$file" || return 0

  if ! grep -q 'go env GOPATH' "$file"; then
    echo "✗ $file: mentions 'go install' but never 'go env GOPATH'."
    echo "    A user who follows it ends up with an unreachable binary."
    fail=1
    return 0
  fi
  if ! grep -qE 'export PATH|PATH=' "$file"; then
    echo "✗ $file: mentions 'go install' but never a PATH export."
    fail=1
    return 0
  fi
  echo "✓ $file: go install remediation names the PATH export"
}

while IFS= read -r f; do
  check_file "$f"
done < <(git ls-files '*install.sh' 'integrations/**/*.sh')

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "Installer remediation must name both commands. See"
  echo ".github/scripts/installer-remediation-check.sh for why."
  exit 1
fi

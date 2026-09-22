#!/usr/bin/env bash
# Vault-health parity check. The checks exist twice on purpose:
#
#   - Go, in cli/internal/cli/vaulthealth.go, embedded in `tolvi doctor`.
#     Native so doctor has no Python dependency; shelling out would
#     give doctor its own silent-degradation path, which is the failure it
#     exists to detect.
#   - Python, in tolvi-solo's skills/vault-health/, for users who run the
#     vault without installing the CLI at all.
#
# Two implementations drift. This runs both over a fixture with a known defect
# in every dimension and fails if their findings or grades disagree.
#
# Skips cleanly when the tolvi-solo sibling checkout or python3 is absent, so it
# never blocks a contributor who only has this repo.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

PY="../tolvi-solo/skills/vault-health/scripts/vault_health.py"
if [ ! -f "$PY" ]; then
  echo "vault-health parity: no tolvi-solo sibling checkout; skipping."
  exit 0
fi
if ! command -v python3 &>/dev/null; then
  echo "vault-health parity: python3 not installed; skipping."
  exit 0
fi

fixture="$(mktemp -d)"
trap 'rm -rf "$fixture"' EXIT
v="$fixture/vault"
mkdir -p "$v/decisions" "$v/sessions" "$v/patterns"
echo '{"workspace":"t","embedding_model":"nomic-embed-text","schema_version":1}' > "$v/.vault-meta.json"

# One defect per dimension, plus a clean note.
printf -- '---\ndate: 2026-01-01\nstatus: active\n---\n\n# Untagged note\n'                              > "$v/decisions/2026-01-01-untagged.md"
printf -- '---\ntags: [d]\nstatus: wibble\n---\n\n# Bad status\n'                                        > "$v/decisions/2026-01-02-badstatus.md"
printf -- '---\ntags: [d]\ndate: YYYY-MM-DD\nrepo: <repo-slug>\n---\n\n# Unfilled template\n'            > "$v/decisions/2026-01-03-unfilled.md"
printf -- '---\ntags: [d]\nstatus: active\n---\n\n# Choose Postgres\n'                                   > "$v/decisions/2026-01-04-dupe-a.md"
printf -- '---\ntags: [d]\nstatus: active\n---\n\n# Choose Postgres\n'                                   > "$v/decisions/2026-01-05-dupe-b.md"
printf -- '---\ntags: [d]\nstatus: active\n---\n\n# Escapes\n\nIt says caf\\u00e9 here.\n'               > "$v/decisions/2026-01-06-escaped.md"
printf -- '---\ntags: [d]\nstatus: active\n---\n\n# A clean one\n\nNothing wrong here.\n'                > "$v/decisions/2026-01-07-clean.md"

# The Python script parses frontmatter with its own standard-library subset of
# YAML, so both implementations must also parse a clean note written in every
# frontmatter shape real vaults use.
printf -- '---\ntags:\n  [\n    d,\n    e(f),\n  ]\ndate: "2026-01-08"\nstatus: active # set by hand\nmetadata:\n  type: decision\n  related:\n  - a\n  - b\n---\n\n# Every shape\n' > "$v/decisions/2026-01-08-shapes.md"

# Compare the facts both implementations must agree on, not their layout:
# the overall grade, each per-dimension grade, and which files carry findings.
# Report formatting differs between them by design.
#
# Every extraction is `|| true`-guarded: under `set -o pipefail` a grep that
# matches nothing would otherwise abort the script with no explanation.
overall() { printf '%s' "$1" | grep -oE 'Overall grade: [A-F][+-]?' | head -1 || true; }
dims() {
  printf '%s' "$1" \
    | grep -oE '(tags|lifecycle|dedup|completeness|unicode) [A-F][+-]?' \
    | sort -u || true
}
files() {
  printf '%s' "$1" \
    | grep -oE 'decisions/[A-Za-z0-9._-]+\.md' \
    | sort -u || true
}

go_out="$(cd cli && go run ./cmd/tolvi doctor vault-health --vault "$v" 2>&1 || true)"
py_out="$(python3 "$PY" "$v" 2>&1 || true)"

fail=0
compare() {
  local label="$1" a="$2" b="$3"
  if [ "$a" != "$b" ]; then
    echo "✗ vault-health parity: $label differs."
    diff <(printf '%s\n' "$a") <(printf '%s\n' "$b") | head -20 || true
    fail=1
  else
    echo "✓ $label agrees"
  fi
}

compare "overall grade"     "$(overall "$py_out")" "$(overall "$go_out")"
compare "per-dimension grades" "$(dims "$py_out")" "$(dims "$go_out")"
compare "files with findings"  "$(files "$py_out")" "$(files "$go_out")"

if [ "$fail" -ne 0 ]; then
  echo ""
  echo "Both implementations must report the same findings and grades."
  echo "See cli/internal/cli/vaulthealth.go for the port and why it exists."
  exit 1
fi

echo ""
echo "vault-health parity: Go and Python agree."
overall "$go_out"

#!/usr/bin/env bash
# test-install-hook-idempotency.sh — CI guard for the Claude Code hook merge.
#
# The installer is expected to be re-runnable: users re-run it after an upgrade,
# after switching --hooks-scope, or just to repair a config. A merge that appends
# unconditionally silently corrupts settings.json on the second run, which the
# user sees as doubled recall output on every session start and the commit gate
# firing twice on every commit. That reads as "Tolvi is buggy" rather than "the
# installer was run twice", so the merge is pinned here rather than left to review.
#
# Exercises:
#   - a second --force run leaves exactly one SessionStart and one PreToolUse entry
#   - an entry pointing at a stale hooks directory converges onto the new path
#     instead of lingering alongside it
#   - hooks belonging to other tools are never touched
#
# Runs in a temp HOME so it doesn't touch the real environment.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
INSTALLER="$REPO_ROOT/integrations/claude-code/install.sh"

if [[ ! -f "$INSTALLER" ]]; then
  echo "FAIL: installer not found at $INSTALLER" >&2
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "FAIL: jq is required for this check" >&2
  exit 1
fi

TEST_HOME=$(mktemp -d)
trap 'rm -rf "$TEST_HOME"' EXIT

SETTINGS="$TEST_HOME/.claude/settings.json"
HOOKS_DIR="$TEST_HOME/.claude/hooks/tolvi"

run_install() {
  HOME="$TEST_HOME" bash "$INSTALLER" --with-hooks --hooks-scope user "$@" >/dev/null
}

assert_count() {
  local event="$1" expected="$2" actual
  actual="$(jq --arg e "$event" '.hooks[$e] | length' "$SETTINGS")"
  if [[ "$actual" != "$expected" ]]; then
    echo "FAIL: hooks.$event has $actual entries, expected $expected" >&2
    jq '.hooks' "$SETTINGS" >&2
    exit 1
  fi
}

echo "→ TEST_HOME=$TEST_HOME"

# --- Case 1: re-running the installer must converge, not accumulate ---
echo "→ Running install twice (second with --force, the documented upgrade path)"
run_install
run_install --force

assert_count SessionStart 1
assert_count PreToolUse 1
echo "✓ Second run left one SessionStart and one PreToolUse entry"

# --- Case 2: an entry pointing at a stale hooks directory must converge ---
echo "→ Seeding a stale-path entry and re-installing"
jq --arg old "/old/path/tolvi-recall" \
   '.hooks.SessionStart = [{matcher: "startup", hooks: [{type: "command", command: $old}]}]
    | .hooks.PreToolUse = []' \
   "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"

run_install --force

assert_count SessionStart 1
STALE="$(jq -r '[.hooks.SessionStart[].hooks[].command] | map(select(startswith("/old/path"))) | length' "$SETTINGS")"
if [[ "$STALE" != "0" ]]; then
  echo "FAIL: a stale hooks path survived the merge" >&2
  jq '.hooks.SessionStart' "$SETTINGS" >&2
  exit 1
fi
CURRENT="$(jq -r --arg want "$HOOKS_DIR/tolvi-recall" \
  '[.hooks.SessionStart[].hooks[].command] | map(select(. == $want)) | length' "$SETTINGS")"
if [[ "$CURRENT" != "1" ]]; then
  echo "FAIL: expected exactly one entry pointing at $HOOKS_DIR/tolvi-recall, got $CURRENT" >&2
  jq '.hooks.SessionStart' "$SETTINGS" >&2
  exit 1
fi
echo "✓ Stale entry replaced rather than duplicated"

# --- Case 3: another tool's hooks must survive untouched ---
echo "→ Seeding a third-party hook and re-installing"
jq '.hooks.SessionStart += [{matcher: "startup", hooks: [{type: "command", command: "/other/tool/hook"}]}]' \
   "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"

run_install --force

FOREIGN="$(jq -r '[.hooks.SessionStart[].hooks[].command] | map(select(. == "/other/tool/hook")) | length' "$SETTINGS")"
if [[ "$FOREIGN" != "1" ]]; then
  echo "FAIL: third-party hook count is $FOREIGN, expected 1 (must be neither dropped nor duplicated)" >&2
  jq '.hooks.SessionStart' "$SETTINGS" >&2
  exit 1
fi
assert_count SessionStart 2
echo "✓ Third-party hook preserved alongside a single tolvi entry"

echo "✓ All Claude Code hook idempotency checks passed."

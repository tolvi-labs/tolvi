#!/usr/bin/env bash
# test-install-claude-code-skill.sh — CI smoke test for the Claude Code
# skill installer.
#
# Exercises:
#   - default symlink install lands at $HOME/.claude/skills/tolvi/SKILL.md
#   - the resulting file is readable
#   - --uninstall removes the file and the empty directory
#   - --copy produces a regular file, not a symlink
#
# Runs in a temp HOME so it doesn't touch the real environment.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
INSTALLER="$REPO_ROOT/integrations/claude-code/install.sh"
SOURCE_SKILL="$REPO_ROOT/integrations/claude-code/SKILL.md"

if [[ ! -f "$INSTALLER" ]]; then
  echo "FAIL: installer not found at $INSTALLER" >&2
  exit 1
fi
if [[ ! -f "$SOURCE_SKILL" ]]; then
  echo "FAIL: SKILL.md not found at $SOURCE_SKILL" >&2
  exit 1
fi

TEST_HOME=$(mktemp -d)
trap 'rm -rf "$TEST_HOME"' EXIT

echo "→ TEST_HOME=$TEST_HOME"
echo "→ Running default (symlink) install"
HOME="$TEST_HOME" bash "$INSTALLER" >/dev/null

DEST_FILE="$TEST_HOME/.claude/skills/tolvi/SKILL.md"
if [[ ! -L "$DEST_FILE" ]]; then
  echo "FAIL: expected $DEST_FILE to be a symlink" >&2
  exit 1
fi
if [[ ! -r "$DEST_FILE" ]]; then
  echo "FAIL: $DEST_FILE is not readable" >&2
  exit 1
fi

# Verify the symlink resolves to the canonical SKILL.md.
RESOLVED="$(readlink "$DEST_FILE")"
if [[ "$RESOLVED" != "$SOURCE_SKILL" ]]; then
  echo "FAIL: symlink resolves to $RESOLVED, expected $SOURCE_SKILL" >&2
  exit 1
fi
echo "✓ Symlink install verified"

echo "→ Running --uninstall"
HOME="$TEST_HOME" bash "$INSTALLER" --uninstall >/dev/null

if [[ -e "$DEST_FILE" || -L "$DEST_FILE" ]]; then
  echo "FAIL: $DEST_FILE still exists after --uninstall" >&2
  exit 1
fi
if [[ -d "$TEST_HOME/.claude/skills/tolvi" ]]; then
  echo "FAIL: tolvi/ directory still exists after --uninstall" >&2
  exit 1
fi
echo "✓ Uninstall verified"

echo "→ Running --copy install"
HOME="$TEST_HOME" bash "$INSTALLER" --copy >/dev/null

if [[ -L "$DEST_FILE" ]]; then
  echo "FAIL: --copy should produce a regular file, got symlink at $DEST_FILE" >&2
  exit 1
fi
if [[ ! -f "$DEST_FILE" ]]; then
  echo "FAIL: --copy did not produce a file at $DEST_FILE" >&2
  exit 1
fi
echo "✓ Copy install verified"

# Re-uninstall to leave temp HOME clean.
HOME="$TEST_HOME" bash "$INSTALLER" --uninstall >/dev/null

echo "✓ All Claude Code skill installer smoke checks passed."

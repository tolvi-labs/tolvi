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
INSTALLER="$REPO_ROOT/skills/tolvi/install.sh"
SOURCE_SKILL="$REPO_ROOT/skills/tolvi/SKILL.md"

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

echo "✓ Claude Code skill installer smoke checks passed."

# --- --agents mode: the shared Agent Skill for Codex, Cursor and OpenHands ---
# Runs under its own HOME so a stray write into ~/.claude is detectable; the
# Claude Code checks above leave ~/.claude/commands behind in TEST_HOME.
AGENT_HOME=$(mktemp -d)
trap 'rm -rf "$TEST_HOME" "$AGENT_HOME"' EXIT

AGENT_REPO="$AGENT_HOME/project"
mkdir -p "$AGENT_REPO/src/deep"
git -C "$AGENT_REPO" init -q
AGENT_FILE="$AGENT_REPO/.agents/skills/tolvi/SKILL.md"

echo "→ Running --agents install from a subdirectory of a git repo"
( cd "$AGENT_REPO/src/deep" && HOME="$AGENT_HOME" bash "$INSTALLER" --agents >/dev/null )
if [[ -L "$AGENT_FILE" || ! -f "$AGENT_FILE" ]]; then
  echo "FAIL: expected a regular file (not a symlink) at $AGENT_FILE" >&2
  exit 1
fi
if ! cmp -s "$AGENT_FILE" "$SOURCE_SKILL"; then
  echo "FAIL: $AGENT_FILE differs from $SOURCE_SKILL" >&2
  exit 1
fi
if [[ -e "$AGENT_HOME/.claude" ]]; then
  echo "FAIL: --agents wrote Claude Code files under HOME ($AGENT_HOME/.claude exists)" >&2
  exit 1
fi
echo "✓ --agents install lands a copy at the repository root and nothing under HOME"

echo "→ Running --agents again without --force"
if STDERR=$( cd "$AGENT_REPO" && HOME="$AGENT_HOME" bash "$INSTALLER" --agents 2>&1 >/dev/null ); then
  echo "FAIL: a second --agents install without --force should exit non-zero" >&2
  exit 1
fi
if [[ "$STDERR" != *"already exists"* ]]; then
  echo "FAIL: expected an 'already exists' refusal, got: $STDERR" >&2
  exit 1
fi
if ! cmp -s "$AGENT_FILE" "$SOURCE_SKILL"; then
  echo "FAIL: the refused install changed $AGENT_FILE" >&2
  exit 1
fi
echo "✓ --agents refuses to overwrite without --force"

echo "→ Running --agents --with-hooks"
HOOKS_REPO="$AGENT_HOME/hooks-project"
mkdir -p "$HOOKS_REPO"
git -C "$HOOKS_REPO" init -q
if STDERR=$( cd "$HOOKS_REPO" && HOME="$AGENT_HOME" bash "$INSTALLER" --agents --with-hooks 2>&1 >/dev/null ); then
  echo "FAIL: --agents --with-hooks should exit non-zero" >&2
  exit 1
fi
if [[ "$STDERR" != *"Claude Code only"* ]]; then
  echo "FAIL: expected a 'Claude Code only' refusal, got: $STDERR" >&2
  exit 1
fi
if [[ -e "$HOOKS_REPO/.agents" || -e "$AGENT_HOME/.claude" ]]; then
  echo "FAIL: the refused --agents --with-hooks install wrote files" >&2
  exit 1
fi
echo "✓ --agents rejects --with-hooks"

echo "→ Running --agents outside any git repository"
NOGIT_DIR="$AGENT_HOME/no-git"
mkdir -p "$NOGIT_DIR"
( cd "$NOGIT_DIR" && HOME="$AGENT_HOME" bash "$INSTALLER" --agents >/dev/null )
if [[ ! -f "$NOGIT_DIR/.agents/skills/tolvi/SKILL.md" ]]; then
  echo "FAIL: without a git repo, --agents should install under the current directory" >&2
  exit 1
fi
echo "✓ --agents falls back to the current directory"

echo "→ Running --agents --path for a personal install"
PERSONAL="$AGENT_HOME/personal-skills"
HOME="$AGENT_HOME" bash "$INSTALLER" --agents --path "$PERSONAL" >/dev/null
if [[ -L "$PERSONAL/tolvi/SKILL.md" || ! -f "$PERSONAL/tolvi/SKILL.md" ]]; then
  echo "FAIL: expected a regular file at $PERSONAL/tolvi/SKILL.md" >&2
  exit 1
fi
echo "✓ --agents --path installs a copy at the given base"

echo "→ Running --uninstall --agents"
( cd "$AGENT_REPO/src/deep" && HOME="$AGENT_HOME" bash "$INSTALLER" --uninstall --agents >/dev/null )
if [[ -e "$AGENT_FILE" || -d "$AGENT_REPO/.agents/skills/tolvi" ]]; then
  echo "FAIL: --uninstall --agents left $AGENT_REPO/.agents/skills/tolvi behind" >&2
  exit 1
fi
if [[ -e "$AGENT_HOME/.claude" ]]; then
  echo "FAIL: --uninstall --agents touched HOME" >&2
  exit 1
fi
echo "✓ --uninstall --agents removes the project install"

echo "✓ All skill installer smoke checks passed."

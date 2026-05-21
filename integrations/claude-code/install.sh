#!/usr/bin/env bash
# install.sh — install the Tolvi Claude Code skill.
#
# Default: symlink integrations/claude-code/SKILL.md → ~/.claude/skills/tolvi/SKILL.md
# so `git pull` updates land automatically.
#
# Flags:
#   --copy            Deep-copy instead of symlinking (isolate from repo updates)
#   --uninstall       Remove the installed file + directory
#   --path <dir>      Override install destination (default: ~/.claude/skills)
#   --force           Overwrite an existing install (refuses by default)
#   -h, --help        Print usage

set -euo pipefail

# Resolve the directory this script lives in, so we can find SKILL.md
# regardless of where the user invoked the script from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_SKILL="$SCRIPT_DIR/SKILL.md"

DEFAULT_BASE="${HOME}/.claude/skills"
DEST_BASE="$DEFAULT_BASE"
MODE="symlink"
ACTION="install"
FORCE="false"

usage() {
  cat <<EOF
Usage: bash install.sh [--copy] [--uninstall] [--path <dir>] [--force]

Default: symlink integrations/claude-code/SKILL.md into
         \$HOME/.claude/skills/tolvi/SKILL.md so that 'git pull' on the
         tolvi-labs/tolvi repo updates the skill automatically.

Flags:
  --copy          Deep-copy SKILL.md instead of symlinking.
  --uninstall     Remove the installed file and the tolvi/ skill directory.
  --path <dir>    Install destination base (default: \$HOME/.claude/skills).
                  The tolvi/ subdirectory is created under this path.
  --force         Overwrite an existing install. Refuses by default to
                  avoid clobbering user customizations.
  -h, --help      Show this help.

EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --copy)       MODE="copy";        shift ;;
    --uninstall)  ACTION="uninstall"; shift ;;
    --path)       DEST_BASE="$2";     shift 2 ;;
    --force)      FORCE="true";       shift ;;
    -h|--help)    usage; exit 0 ;;
    *)            echo "install.sh: unknown flag: $1" >&2; usage; exit 1 ;;
  esac
done

DEST_DIR="$DEST_BASE/tolvi"
DEST_FILE="$DEST_DIR/SKILL.md"

if [[ "$ACTION" == "uninstall" ]]; then
  if [[ ! -e "$DEST_FILE" && ! -L "$DEST_FILE" ]]; then
    echo "install.sh: nothing to uninstall at $DEST_FILE"
    # Still try to clean up an empty tolvi/ dir if it exists.
    if [[ -d "$DEST_DIR" ]]; then
      if rmdir "$DEST_DIR" 2>/dev/null; then
        echo "✓ Removed empty $DEST_DIR"
      fi
    fi
    exit 0
  fi

  # Verify we're only removing what we installed.
  if [[ -L "$DEST_FILE" ]]; then
    LINK_TARGET="$(readlink "$DEST_FILE")"
    echo "✓ Removing $DEST_FILE (was symlink → $LINK_TARGET)"
    rm "$DEST_FILE"
  elif [[ -f "$DEST_FILE" ]]; then
    echo "✓ Removing $DEST_FILE (was copy)"
    rm "$DEST_FILE"
  else
    echo "install.sh: $DEST_FILE is neither a symlink nor a regular file; refusing to remove" >&2
    exit 1
  fi

  # Try to remove the tolvi/ dir if it's empty.
  if [[ -d "$DEST_DIR" ]]; then
    if rmdir "$DEST_DIR" 2>/dev/null; then
      echo "✓ Removed empty $DEST_DIR"
    else
      echo "install.sh: $DEST_DIR is not empty (contains user-added files); leaving in place" >&2
    fi
  fi
  exit 0
fi

# --- install path ---

if [[ ! -f "$SOURCE_SKILL" ]]; then
  echo "install.sh: cannot find source SKILL.md at $SOURCE_SKILL" >&2
  exit 1
fi

if [[ -e "$DEST_FILE" || -L "$DEST_FILE" ]]; then
  if [[ "$FORCE" != "true" ]]; then
    echo "install.sh: $DEST_FILE already exists. Re-run with --force to overwrite." >&2
    exit 1
  fi
  rm "$DEST_FILE"
fi

mkdir -p "$DEST_DIR"
echo "✓ Detected Claude Code skill directory: $DEST_BASE"

if [[ "$MODE" == "symlink" ]]; then
  ln -s "$SOURCE_SKILL" "$DEST_FILE"
  echo "✓ Symlinked $DEST_FILE → $SOURCE_SKILL"
else
  cp "$SOURCE_SKILL" "$DEST_FILE"
  echo "✓ Copied SKILL.md to $DEST_FILE"
fi

if [[ ! -r "$DEST_FILE" ]]; then
  echo "install.sh: $DEST_FILE is not readable after install" >&2
  exit 1
fi
echo "✓ Verifying: $DEST_FILE is readable ✓"

cat <<EOF

Next steps:
  - In any Claude Code session, type /tolvi to load the skill.
  - The CLI binary should be in \$PATH. Verify with: tolvi version
  - If you don't have it yet:
      go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest
EOF

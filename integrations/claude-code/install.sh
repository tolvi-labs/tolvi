#!/usr/bin/env bash
# install.sh — install the Tolvi Claude Code skill and optional session hooks.
#
# Default: symlink integrations/claude-code/SKILL.md → ~/.claude/skills/tolvi/SKILL.md
# so `git pull` updates land automatically.
#
# Flags:
#   --copy                  Deep-copy instead of symlinking (isolate from repo updates)
#   --uninstall             Remove the installed skill file + directory
#   --path <dir>            Override install destination (default: ~/.claude/skills)
#   --force                 Overwrite an existing install (refuses by default)
#   --with-hooks            Also install session-recall + commit-sync hooks
#   --hooks-scope <scope>   user (default) | project  — where hooks are wired:
#                             user:    ~/.claude/settings.json (all Tolvi repos)
#                             project: .claude/settings.json  (this repo only)
#   -h, --help              Print usage

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
WITH_HOOKS="false"
HOOKS_SCOPE=""   # "user" | "project" — empty means prompt

usage() {
  cat <<EOF
Usage: bash install.sh [--copy] [--uninstall] [--path <dir>] [--force]
                       [--with-hooks] [--hooks-scope user|project]

Default: symlink integrations/claude-code/SKILL.md into
         \$HOME/.claude/skills/tolvi/SKILL.md so that 'git pull' on the
         tolvi-labs/tolvi repo updates the skill automatically.

Flags:
  --copy                  Deep-copy SKILL.md instead of symlinking.
  --uninstall             Remove the installed skill file and the tolvi/ directory.
  --path <dir>            Install destination base (default: \$HOME/.claude/skills).
                          The tolvi/ subdirectory is created under this path.
  --force                 Overwrite an existing install. Refuses by default to
                          avoid clobbering user customizations.
  --with-hooks            Also install Claude Code session hooks:
                            • SessionStart  → runs 'tolvi recall' before each session
                            • PostToolUse   → nudges /sync-session after git commits
  --hooks-scope <scope>   Where to wire the hooks:
                            user    — \$HOME/.claude/settings.json (activates in all
                                      repos with a vault — recommended)
                            project — .claude/settings.json in the current directory
                                      (this repo only; committable)
                          Omit to be prompted interactively.
  -h, --help              Show this help.

EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --copy)         MODE="copy";        shift ;;
    --uninstall)    ACTION="uninstall"; shift ;;
    --path)         DEST_BASE="$2";     shift 2 ;;
    --force)        FORCE="true";       shift ;;
    --with-hooks)   WITH_HOOKS="true";  shift ;;
    --hooks-scope)  HOOKS_SCOPE="$2";   shift 2 ;;
    -h|--help)      usage; exit 0 ;;
    *)              echo "install.sh: unknown flag: $1" >&2; usage; exit 1 ;;
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

# --- hook install function ---
# Installs session-recall.sh + commit-sync-nudge.sh and wires them into
# the Claude Code settings.json at the chosen scope.
install_hooks() {
  local scope="$HOOKS_SCOPE"

  # Prompt if scope not specified and we are interactive.
  if [[ -z "$scope" ]]; then
    if [[ -t 0 ]]; then
      echo ""
      echo "Where should Claude Code hooks be wired?"
      echo "  [1] Your account (~/.claude/settings.json)"
      echo "      → Hooks activate in any repo that has a vault/  (recommended)"
      echo "  [2] This project only (.claude/settings.json)"
      echo "      → Explicit per-repo opt-in; file can be committed"
      echo ""
      read -r -p "Choice [1]: " scope_choice
      case "$scope_choice" in
        2) scope="project" ;;
        *) scope="user" ;;
      esac
    else
      # Non-interactive (piped / CI): default to user-level.
      echo "install.sh: non-interactive — defaulting hooks scope to 'user'"
      scope="user"
    fi
  fi

  # Validate scope value.
  case "$scope" in
    user|project) ;;
    *) echo "install.sh: --hooks-scope must be 'user' or 'project', got: $scope" >&2; exit 1 ;;
  esac

  # Determine destination directories.
  local hooks_dest settings_file
  if [[ "$scope" == "user" ]]; then
    hooks_dest="${HOME}/.claude/hooks/tolvi"
    settings_file="${HOME}/.claude/settings.json"
  else
    # Project-level: resolve the project root (where .git lives, or cwd).
    local project_root
    project_root="$(pwd)"
    # Walk up to find .git
    local d="$project_root"
    while [[ "$d" != "/" ]]; do
      if [[ -d "$d/.git" ]]; then
        project_root="$d"
        break
      fi
      d="$(dirname "$d")"
    done
    hooks_dest="${project_root}/.claude/hooks/tolvi"
    settings_file="${project_root}/.claude/settings.json"
  fi

  # Copy hook scripts to destination.
  mkdir -p "$hooks_dest"
  cp "$SCRIPT_DIR/hooks/session-recall.sh"    "$hooks_dest/session-recall.sh"
  cp "$SCRIPT_DIR/hooks/commit-sync-nudge.sh" "$hooks_dest/commit-sync-nudge.sh"
  chmod +x "$hooks_dest/session-recall.sh" "$hooks_dest/commit-sync-nudge.sh"
  echo "✓ Installed hook scripts → $hooks_dest/"

  # Merge hooks.json (with __HOOKS_DIR__ substituted) into settings.json.
  if ! command -v python3 >/dev/null 2>&1; then
    echo "install.sh: python3 not found — skipping settings.json merge." >&2
    echo "  Manually add the hooks from $SCRIPT_DIR/hooks.json to $settings_file" >&2
    echo "  replacing __HOOKS_DIR__ with $hooks_dest" >&2
    return 1
  fi

  python3 - "$settings_file" "$SCRIPT_DIR/hooks.json" "$hooks_dest" <<'PYEOF'
import json, os, sys

settings_path, hooks_tpl_path, hooks_dir = sys.argv[1], sys.argv[2], sys.argv[3]

# Load (or create) the settings file.
if os.path.exists(settings_path):
    with open(settings_path) as f:
        settings = json.load(f)
else:
    os.makedirs(os.path.dirname(settings_path) or ".", exist_ok=True)
    settings = {}

# Load and interpolate the hooks template.
with open(hooks_tpl_path) as f:
    fragment_text = f.read().replace("__HOOKS_DIR__", hooks_dir)
fragment = json.loads(fragment_text)

# Merge: extend existing event-type arrays rather than overwriting.
existing = settings.setdefault("hooks", {})
for event, entries in fragment["hooks"].items():
    if event in existing:
        existing[event].extend(entries)
    else:
        existing[event] = entries

with open(settings_path, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")

print(f"✓ Merged hooks into {settings_path}")
PYEOF

  echo ""
  echo "Hooks installed ($scope scope)."
  if [[ "$scope" == "user" ]]; then
    echo "  Recall fires on every session start in any repo with a vault/."
    echo "  Sync nudge fires after every git commit in any repo with a vault/."
  else
    echo "  Recall and sync-nudge fire in this project only."
    echo "  Commit $settings_file to share with teammates."
  fi
}

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

if [[ "$WITH_HOOKS" == "true" ]]; then
  install_hooks
fi

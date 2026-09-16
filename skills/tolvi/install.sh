#!/usr/bin/env bash
# install.sh — install the Tolvi Claude Code skill and optional session hooks.
#
# Default: symlink skills/tolvi/SKILL.md → ~/.claude/skills/tolvi/SKILL.md
# so `git pull` updates land automatically.
#
# Flags:
#   --copy                  Deep-copy instead of symlinking (isolate from repo updates)
#   --uninstall             Remove the installed skill file + directory
#   --path <dir>            Override install destination (default: ~/.claude/skills)
#   --force                 Overwrite an existing install (refuses by default)
#   --with-hooks            Also install tolvi-recall + tolvi-sync hooks
#   --hooks-scope <scope>   user (default) | project  — where hooks are wired:
#                             user:    ~/.claude/settings.json (all Tolvi repos)
#                             project: .claude/settings.json  (this repo only)
#   --agents                Install the shared Agent Skill for Codex, Cursor and
#                             OpenHands: copy SKILL.md into <repo root>/.agents/skills/tolvi/
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
AGENTS="false"
PATH_SET="false"
NO_GIT_ROOT="false"

usage() {
  cat <<EOF
Usage: bash install.sh [--copy] [--uninstall] [--path <dir>] [--force]
                       [--with-hooks] [--hooks-scope user|project] [--agents]

Default: symlink skills/tolvi/SKILL.md into
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
                            • SessionStart  → tolvi-recall fires before each session
                            • PreToolUse    → tolvi-sync warns when the vault is unsynced
  --hooks-scope <scope>   Where to wire the hooks:
                            user    — \$HOME/.claude/settings.json (activates in all
                                      repos with a vault — recommended)
                            project — .claude/settings.json in the current directory
                                      (this repo only; committable)
                          Omit to be prompted interactively.
  --agents                Install the shared Agent Skill for Codex, Cursor and
                          OpenHands instead of the Claude Code install. Copies
                          SKILL.md into <repo root>/.agents/skills/tolvi/, where
                          the repo root is the nearest ancestor containing .git
                          (or the current directory). Never symlinks, and installs
                          no slash commands, stack skills or hooks. --path sets a
                          different base, such as ~/.agents/skills. Combine with
                          --uninstall to remove it.
  -h, --help              Show this help.

EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --copy)         MODE="copy";        shift ;;
    --uninstall)    ACTION="uninstall"; shift ;;
    --path)         DEST_BASE="$2";     PATH_SET="true"; shift 2 ;;
    --force)        FORCE="true";       shift ;;
    --with-hooks)   WITH_HOOKS="true";  shift ;;
    --hooks-scope)  HOOKS_SCOPE="$2";   shift 2 ;;
    --agents)       AGENTS="true";      shift ;;
    -h|--help)      usage; exit 0 ;;
    *)              echo "install.sh: unknown flag: $1" >&2; usage; exit 1 ;;
  esac
done

# Nearest ancestor of the current directory that contains .git, or the current
# directory when there is none. Agents look for .agents/skills at the repo root.
find_project_root() {
  local d
  d="$(pwd)"
  while [[ "$d" != "/" ]]; do
    if [[ -e "$d/.git" ]]; then
      echo "$d"
      return 0
    fi
    d="$(dirname "$d")"
  done
  pwd
}

if [[ "$AGENTS" == "true" ]]; then
  if [[ "$WITH_HOOKS" == "true" || -n "$HOOKS_SCOPE" ]]; then
    echo "install.sh: --with-hooks and --hooks-scope are Claude Code only and cannot be combined with --agents" >&2
    exit 1
  fi
  # A project install is committed, and a committed symlink into one person's
  # tolvi checkout is broken for everyone else, so agent installs always copy.
  MODE="copy"
  if [[ "$PATH_SET" != "true" ]]; then
    root="$(find_project_root)"
    if [[ ! -e "$root/.git" ]]; then
      NO_GIT_ROOT="true"
      echo "install.sh: no git repository found above $(pwd); installing the skill under the current directory" >&2
    fi
    DEST_BASE="$root/.agents/skills"
  fi
fi

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

  # An agents install never placed slash commands or stack skills.
  if [[ "$AGENTS" == "true" ]]; then
    exit 0
  fi

  # Remove installed slash commands (only our own files).
  if [[ -d "$SCRIPT_DIR/commands" ]]; then
    for src in "$SCRIPT_DIR/commands"/*.md; do
      cmd_dest="${HOME}/.claude/commands/$(basename "$src")"
      if [[ -L "$cmd_dest" || -f "$cmd_dest" ]]; then
        rm "$cmd_dest" && echo "✓ Removed command: /$(basename "${cmd_dest%.md}")"
      fi
    done
  fi

  # Remove installed stack skills (only our own symlinks/copies).
  for name in tolvi-bastion tolvi-guild; do
    sk_dest="${HOME}/.claude/skills/$name"
    if [[ -L "$sk_dest" ]]; then
      rm "$sk_dest" && echo "✓ Removed stack skill: /$name (was symlink)"
    elif [[ -d "$sk_dest" ]]; then
      rm -rf "$sk_dest" && echo "✓ Removed stack skill: /$name (was copy)"
    fi
  done
  exit 0
fi

# --- command install function ---
# Symlinks (or copies, with --copy) the tolvi slash commands into
# ~/.claude/commands/ so /tolvi-recall, /tolvi-sync, /tolvi-commit are available.
install_commands() {
  local src_dir="$SCRIPT_DIR/commands"
  local dest_dir="${HOME}/.claude/commands"
  [[ -d "$src_dir" ]] || return 0
  mkdir -p "$dest_dir"
  local f name dest
  for f in "$src_dir"/*.md; do
    name="$(basename "$f")"
    # Underscore-prefixed files are shared includes (e.g. _preflight.md), not
    # slash commands; installing one would create a bogus /_preflight.
    [[ "$name" == _* ]] && continue
    dest="$dest_dir/$name"
    if [[ -e "$dest" || -L "$dest" ]]; then
      if [[ "$FORCE" != "true" ]]; then
        echo "  ⚠ $dest exists — skipping (re-run with --force to overwrite)"
        continue
      fi
      rm "$dest"
    fi
    if [[ "$MODE" == "symlink" ]]; then
      ln -s "$f" "$dest"
    else
      cp "$f" "$dest"
    fi
    echo "✓ Installed command: /${name%.md}"
  done
}

# --- stack skill install ---
# Symlinks (or copies, with --copy) sibling Tolvi stack skills — tolvi-bastion and
# tolvi-guild — from repos cloned alongside this one, into ~/.claude/skills/ so
# /tolvi-bastion and /tolvi-guild ship as part of the suite.
install_stack_skills() {
  local repo_root parent dest_base
  repo_root="$(cd "$SCRIPT_DIR/../.." && pwd)"   # skills/tolvi → repo root
  parent="$(dirname "$repo_root")"               # the tolvi-labs/ workspace
  dest_base="${HOME}/.claude/skills"
  mkdir -p "$dest_base"
  local name src dest
  for name in tolvi-bastion tolvi-guild; do
    case "$name" in
      tolvi-bastion) src="$parent/bastion/skills/tolvi-bastion" ;;
      tolvi-guild)   src="$parent/guild/skills/tolvi-guild" ;;
    esac
    dest="$dest_base/$name"
    if [[ ! -d "$src" ]]; then
      echo "  ⚠ $name: source not found at $src — clone tolvi-labs/${name#tolvi-} alongside tolvi (skipping)"
      continue
    fi
    if [[ -e "$dest" || -L "$dest" ]]; then
      if [[ "$FORCE" != "true" ]]; then
        echo "  ⚠ $dest exists — skipping (re-run with --force to overwrite)"
        continue
      fi
      rm -rf "$dest"
    fi
    if [[ "$MODE" == "symlink" ]]; then
      ln -s "$src" "$dest"
    else
      cp -R "$src" "$dest"
    fi
    echo "✓ Installed stack skill: /$name"
  done
}

# --- hook install function ---
# Installs tolvi-recall + tolvi-sync and wires them into
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
  cp "$SCRIPT_DIR/hooks/tolvi-recall" "$hooks_dest/tolvi-recall"
  cp "$SCRIPT_DIR/hooks/tolvi-sync"   "$hooks_dest/tolvi-sync"
  chmod +x "$hooks_dest/tolvi-recall" "$hooks_dest/tolvi-sync"
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

# Merge: converge on the intended state rather than appending blindly. The
# installer is re-run after every upgrade, so an unconditional extend() would
# duplicate each hook on the second run — doubling recall output on session
# start and firing the commit gate twice per commit.
def hook_scripts(entry):
    """Basenames of the hook scripts an entry invokes (e.g. {'tolvi-recall'})."""
    return {
        os.path.basename(h["command"])
        for h in entry.get("hooks", [])
        if h.get("command")
    }


existing = settings.setdefault("hooks", {})
for event, entries in fragment["hooks"].items():
    current = existing.setdefault(event, [])
    for entry in entries:
        scripts = hook_scripts(entry)
        # Drop any prior install of these same scripts, including one left
        # pointing at a stale hooks directory by an earlier --hooks-scope, then
        # re-add the fresh entry. Hooks belonging to other tools share no
        # basename with ours and are left untouched.
        current[:] = [e for e in current if not hook_scripts(e) & scripts]
        current.append(entry)

# Allowlist the read-only tolvi subcommands so /tolvi-recall and `tolvi ask`
# stop raising a permission prompt on every use. Writes are deliberately NOT
# allowlisted: `tolvi sync` and `tolvi commit` mutate the vault and the git
# history, so they should stay a conscious per-call approval.
allow = settings.setdefault("permissions", {}).setdefault("allow", [])
added = [r for r in ("Bash(tolvi recall:*)", "Bash(tolvi ask:*)") if r not in allow]
allow.extend(added)

with open(settings_path, "w") as f:
    json.dump(settings, f, indent=2)
    f.write("\n")

print(f"✓ Merged hooks into {settings_path}")
if added:
    print(f"✓ Allowlisted {', '.join(added)} (read-only; sync/commit still prompt)")
PYEOF

  echo ""
  echo "Hooks installed ($scope scope)."
  if [[ "$scope" == "user" ]]; then
    echo "  tolvi-recall fires on every session start in any repo with a vault/."
    echo "  tolvi-sync warns when today's session note is missing, and never blocks a commit."
  else
    echo "  tolvi-recall and tolvi-sync fire in this project only."
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
if [[ "$AGENTS" == "true" ]]; then
  echo "✓ Agent skill directory: $DEST_BASE"
else
  echo "✓ Detected Claude Code skill directory: $DEST_BASE"
fi

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

if [[ "$AGENTS" != "true" ]]; then
  install_commands
  install_stack_skills

  if [[ "$WITH_HOOKS" == "true" ]]; then
    install_hooks
  fi
fi

# Verify the binary is reachable BY NAME, not merely installed. `go install`
# succeeds into $(go env GOPATH)/bin, which is often not on PATH, so an install
# can look complete while every skill and hook that shells out to `tolvi`
# silently falls back to reading the vault directly. Naming only the install
# command is not a fix; the PATH export is the half people miss.
check_cli() {
  if command -v tolvi &>/dev/null; then
    echo ""
    echo "✓ tolvi CLI: $(command -v tolvi)"
    echo "  Run 'tolvi doctor' to check the rest of your setup."
    return 0
  fi

  cat <<EOF

! tolvi CLI not found on PATH.

  The skills still work without it: they fall back to reading vault/ directly.
  You lose 'tolvi ask' and the single-invocation recall path.

  To install it:
      go install github.com/tolvi-labs/tolvi/cli/cmd/tolvi@latest

  Then make it reachable, which 'go install' does not do for you:
      export PATH="\$PATH:\$(go env GOPATH)/bin"     # add to ~/.zshenv, not ~/.zshrc

  Verify with:  tolvi doctor
EOF
}

if [[ "$AGENTS" == "true" ]]; then
  echo ""
  echo "Next steps:"
  if [[ "$PATH_SET" != "true" && "$NO_GIT_ROOT" != "true" ]]; then
    echo "  - Commit ${DEST_DIR} so everyone on the team gets the skill."
  fi
  echo "  - Codex, Cursor and OpenHands load the skill when a request matches its description."
else
  cat <<EOF

Next steps:
  - In any Claude Code session, type /tolvi to load the skill.
EOF
fi

check_cli

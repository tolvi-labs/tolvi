#!/usr/bin/env bash
# test-recall-hook-formats.sh: CI guard for the SessionStart recall hook.
#
# The same hook script serves Claude Code and Cursor. Claude Code reads
# hookSpecificOutput.additionalContext; Cursor's sessionStart reads
# additional_context. A hook that answers Cursor in Claude Code's shape is
# silently ignored, and a change that alters Claude Code's output breaks every
# existing install, so both shapes are pinned here.
#
# The fixture runs with a minimal PATH and a temp HOME and GOPATH, so the hook
# takes its direct vault read path unless a test puts a stub tolvi on PATH.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
HOOK="$REPO_ROOT/skills/tolvi/hooks/tolvi-recall"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

FIXTURE="$WORK/repo"
mkdir -p "$FIXTURE/vault/sessions" "$FIXTURE/vault/decisions" "$WORK/home" "$WORK/elsewhere" "$WORK/novault"
printf '{"workspace":"fixture","repo":"fixture","embedding_model":"nomic-embed-text","schema_version":2}\n' > "$FIXTURE/vault/.vault-meta.json"
printf -- '---\ntags: [session]\ndate: 2026-09-01\nstatus: active\n---\n\n## [09:00] Session\n' > "$FIXTURE/vault/sessions/2026-09-01.md"
printf -- '---\ntags: [decision]\ndate: 2026-09-01\nrepo: fixture\nstatus: active\n---\n\n# Use Postgres for the primary datastore\n' > "$FIXTURE/vault/decisions/2026-09-01-use-postgres.md"

# Preflight: find_tolvi in the hook also checks /usr/local/bin and
# /opt/homebrew/bin directly, which env -i PATH="...:/usr/bin:/bin" below
# cannot hide. A real tolvi binary at either path would make the fallback
# checks in this test find it instead of exercising the direct vault read
# they pin, and fail in a way that looks like a hook bug rather than a
# machine-specific condition.
for sys_bin in /usr/local/bin/tolvi /opt/homebrew/bin/tolvi; do
  if [[ -x "$sys_bin" ]]; then
    echo "FAIL: $sys_bin exists, and the recall hook finds it even with a minimal PATH, so this test cannot exercise the direct vault read it pins. Run it in CI or on a machine without that binary." >&2
    exit 1
  fi
done

CLAUDE_STARTUP='{"hook_event_name":"SessionStart","source":"startup"}'
CLAUDE_CLEAR='{"hook_event_name":"SessionStart","source":"clear"}'
CURSOR_PAYLOAD="{\"hook_event_name\":\"sessionStart\",\"cursor_version\":\"2.6.0\",\"composer_mode\":\"agent\",\"workspace_roots\":[\"$FIXTURE\"]}"

EXPECTED_CLAUDE_STARTUP='{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"Recent sessions:\n- 2026-09-01\nActive decisions:\n- Use Postgres for the primary datastore\n\n"}}'
EXPECTED_CLAUDE_CLEAR=$'{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"The conversation context was just cleared (/clear). Before anything else in your next response, run the /tolvi-recall command in full and present its complete RECALL SUMMARY, not a lightweight summary. If the next user message is already a task, run recall first, then continue with that task."}}\n'

# run_hook <dir> <payload> [extra PATH prefix]
run_hook() {
  local dir="$1" payload="$2" prefix="${3:-}"
  ( cd "$dir" && printf '%s' "$payload" | env -i PATH="${prefix}/usr/bin:/bin" HOME="$WORK/home" GOPATH="$WORK/home/go" bash "$HOOK" )
}

fail() { echo "FAIL: $*" >&2; exit 1; }

echo "→ Claude Code startup payload"
OUT="$(run_hook "$FIXTURE" "$CLAUDE_STARTUP"; printf x)"; OUT="${OUT%x}"
[[ "$OUT" == "$EXPECTED_CLAUDE_STARTUP" ]] || fail "Claude Code startup output changed. Got: $OUT"
printf '%s' "$OUT" | python3 -c 'import json,sys
d=json.load(sys.stdin)
assert d["hookSpecificOutput"]["hookEventName"]=="SessionStart"
assert "Use Postgres for the primary datastore" in d["hookSpecificOutput"]["additionalContext"]
assert "additional_context" not in d' || fail "Claude Code startup output is not the expected hook JSON"
echo "✓ Claude Code startup output is byte-identical and well formed"

echo "→ Claude Code /clear payload"
OUT="$(run_hook "$FIXTURE" "$CLAUDE_CLEAR"; printf x)"; OUT="${OUT%x}"
[[ "$OUT" == "$EXPECTED_CLAUDE_CLEAR" ]] || fail "Claude Code /clear output changed. Got: $OUT"
echo "✓ Claude Code /clear output is byte-identical"

echo "→ Cursor payload from an unrelated working directory"
OUT="$(run_hook "$WORK/elsewhere" "$CURSOR_PAYLOAD")"
printf '%s' "$OUT" | python3 -c 'import json,sys
d=json.load(sys.stdin)
assert list(d.keys())==["additional_context"], d
assert "Use Postgres for the primary datastore" in d["additional_context"]' || fail "Cursor output is not {\"additional_context\": ...} naming the decision. Got: $OUT"
echo "✓ Cursor gets additional_context, found through workspace_roots"

echo "→ Cursor payload with the tolvi CLI on PATH"
mkdir -p "$WORK/stub"
cat > "$WORK/stub/tolvi" <<'STUB'
#!/usr/bin/env bash
printf '%s' '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"STUB RECALL CONTEXT"}}'
STUB
chmod +x "$WORK/stub/tolvi"
OUT="$(run_hook "$WORK/elsewhere" "$CURSOR_PAYLOAD" "$WORK/stub:")"
printf '%s' "$OUT" | python3 -c 'import json,sys
d=json.load(sys.stdin)
assert d=={"additional_context":"STUB RECALL CONTEXT"}, d' || fail "Cursor did not get the CLI recall re-emitted as additional_context. Got: $OUT"
echo "✓ Cursor gets the CLI recall re-emitted as additional_context"

echo "→ No vault"
OUT="$(run_hook "$WORK/novault" "$CLAUDE_STARTUP")"
[[ -z "$OUT" ]] || fail "Claude Code payload with no vault should print nothing. Got: $OUT"
OUT="$(run_hook "$WORK/novault" "{\"cursor_version\":\"2.6.0\",\"workspace_roots\":[\"$WORK/novault\"]}")"
[[ -z "$OUT" ]] || fail "Cursor payload with no vault should print nothing. Got: $OUT"
echo "✓ No vault prints nothing for either agent"

echo "→ Cursor payload with no workspace_roots, run from inside the fixture repo"
OUT="$(run_hook "$FIXTURE" "{\"cursor_version\":\"2.6.0\"}")"
[[ -z "$OUT" ]] || fail "Cursor payload with no workspace_roots should print nothing, even from inside a repo with a vault. Got: $OUT"
echo "✓ Cursor with no workspace_roots prints nothing, even from inside the fixture repo"

echo "→ Cursor payload with workspace_roots not a list, run from inside the fixture repo"
OUT="$(run_hook "$FIXTURE" "{\"cursor_version\":\"2.6.0\",\"workspace_roots\":\"$FIXTURE\"}")"
[[ -z "$OUT" ]] || fail "Cursor payload with workspace_roots as a string should print nothing, even from inside a repo with a vault. Got: $OUT"
echo "✓ Cursor with non-list workspace_roots prints nothing, even from inside the fixture repo"

echo "✓ All recall hook format checks passed."
